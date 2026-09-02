package templates

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
)

const checksumFormat = "%s %d %s\n"

//go:embed *
var templatesFS embed.FS

// hookExecutable is os.Executable in production; tests may override it.
var hookExecutable = os.Executable

type Args struct {
	Rc                      string
	LefthookPath            string
	AssertLefthookInstalled bool
	Roots                   []string
}

type hookTmplData struct {
	HookName                string
	Extension               string
	LefthookPathTest        string
	LefthookPathInvoke      string
	LefthookPathCurrentTest string
	LefthookPathCurrent     string
	Rc                      string
	Roots                   []string
	AssertLefthookInstalled bool
}

func Hook(hookName string, args Args) []byte {
	lefthookPathCurrent, err := hookExecutable()
	if err != nil {
		lefthookPathCurrent = ""
	}
	lefthookPathCurrent = filepath.ToSlash(lefthookPathCurrent)

	lefthookPath := filepath.ToSlash(strings.ReplaceAll(strings.TrimSpace(args.LefthookPath), "\n", ";"))

	buf := &bytes.Buffer{}
	t := template.Must(template.ParseFS(templatesFS, "hook.tmpl"))
	if err = t.Execute(buf, hookTmplData{
		HookName:                hookName,
		Extension:               getExtension(),
		Rc:                      shellWordForRc(filepath.ToSlash(args.Rc)),
		AssertLefthookInstalled: args.AssertLefthookInstalled,
		Roots:                   args.Roots,
		LefthookPathTest:        shellWordForTest(lefthookPath),
		LefthookPathInvoke:      shellInvokeForLefthookPath(lefthookPath),
		LefthookPathCurrentTest: shellWordForPath(lefthookPathCurrent),
		LefthookPathCurrent:     shellWordForPath(lefthookPathCurrent),
	}); err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func Config() []byte {
	tmpl, err := templatesFS.ReadFile("config.tmpl")
	if err != nil {
		panic(err)
	}

	return tmpl
}

func Checksum(checksum string, timestamp int64, hooks []string) []byte {
	return fmt.Appendf(nil, checksumFormat, checksum, timestamp, strings.Join(hooks, ","))
}

func getExtension() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

// shellWordForRc returns a /bin/sh word for an rc file path or documented shell expression.
func shellWordForRc(value string) string {
	if value == "" {
		return ""
	}
	trimmed := strings.TrimSpace(value)
	if strings.HasPrefix(trimmed, `"`) && strings.HasSuffix(trimmed, `"`) {
		return trimmed
	}
	if strings.Contains(trimmed, "$") {
		if strings.Contains(trimmed, " ") {
			return `"` + shellEscapeDoubleQuote(trimmed) + `"`
		}
		return trimmed
	}

	return shellWordForPath(trimmed)
}

// shellWordForPath returns a /bin/sh word for a filesystem path, quoting only when needed.
func shellWordForPath(path string) string {
	if path == "" {
		return ""
	}
	if !strings.Contains(path, " ") {
		return path
	}

	return `"` + shellEscapeDoubleQuote(path) + `"`
}

// shellWordForTest returns a double-quoted /bin/sh word for test -n / test -f checks.
func shellWordForTest(value string) string {
	if value == "" {
		return `""`
	}

	return `"` + shellEscapeDoubleQuote(value) + `"`
}

// shellInvokeForLefthookPath preserves arbitrary shell commands while quoting path-like values.
func shellInvokeForLefthookPath(value string) string {
	if value == "" {
		return ""
	}
	if looksLikeFilePath(value) {
		return shellWordForPath(value)
	}

	return value
}

func looksLikeFilePath(value string) bool {
	if value == "" || containsShellCommandSyntax(value) {
		return false
	}
	if strings.Contains(value, " ") {
		return strings.Contains(value, "/")
	}
	if filepath.IsAbs(value) {
		return true
	}
	if strings.HasPrefix(value, "./") || strings.HasPrefix(value, "../") {
		return true
	}

	return strings.Contains(value, "/")
}

func containsShellCommandSyntax(value string) bool {
	if strings.Contains(value, ";") {
		return true
	}
	lower := " " + strings.ToLower(value) + " "
	for _, token := range []string{" exec ", " run ", " tool ", " env "} {
		if strings.Contains(lower, token) {
			return true
		}
	}
	if strings.Contains(value, " --") || strings.Contains(value, " -") {
		return true
	}
	if strings.Contains(value, " ") && strings.Contains(value, "=") {
		return true
	}
	if idx := strings.Index(value, " "); idx > 0 {
		first := value[:idx]
		if !strings.Contains(first, "/") && !filepath.IsAbs(first) {
			return true
		}
	}

	return false
}

// shellEscapeDoubleQuote escapes characters that break double-quoted /bin/sh words.
func shellEscapeDoubleQuote(value string) string {
	replacer := strings.NewReplacer(
		`\`, `\\`,
		`"`, `\"`,
	)
	return replacer.Replace(value)
}
