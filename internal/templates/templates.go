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

type Args struct {
	Rc                      string
	LefthookPath            string
	AssertLefthookInstalled bool
	Roots                   []string
}

type hookTmplData struct {
	HookName                string
	Extension               string
	LefthookPath            string
	LefthookPathCurrent     string
	Rc                      string
	Roots                   []string
	AssertLefthookInstalled bool
}

func Hook(hookName string, args Args) []byte {
	lefthookPathCurrent, err := os.Executable()
	if err != nil {
		lefthookPathCurrent = ""
	}

	buf := &bytes.Buffer{}
	t := template.Must(template.New("hook.tmpl").Funcs(template.FuncMap{
		"shellescape": shellescape,
	}).ParseFS(templatesFS, "hook.tmpl"))
	if err = t.ExecuteTemplate(buf, "hook.tmpl", hookTmplData{
		HookName:                hookName,
		Extension:               getExtension(),
		Rc:                      args.Rc,
		AssertLefthookInstalled: args.AssertLefthookInstalled,
		Roots:                   args.Roots,
		LefthookPath:            filepath.ToSlash(strings.ReplaceAll(strings.TrimSpace(args.LefthookPath), "\n", ";")),
		LefthookPathCurrent:     filepath.ToSlash(lefthookPathCurrent),
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

// shellescape wraps a value in POSIX single quotes so it is passed to the
// generated hook verbatim. It is used for the auto-detected executable path
// (os.Executable), which the user cannot quote themselves: a path with a
// space was word-split and silently disabled every hook. The user-supplied
// `lefthook` and `rc` values are intentionally not escaped, since they are
// documented as commands and shell-expanded paths.
func shellescape(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}
