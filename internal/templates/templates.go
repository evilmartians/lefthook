package templates

import (
	"bytes"
	"embed"
	"fmt"
	"runtime"
	"strings"
	"text/template"
)

const checksumFormat = "%s %d %s\n"

//go:embed *
var templatesFS embed.FS

type Args struct {
	Rc                      string
	LefthookExe             string
	AssertLefthookInstalled bool
	Roots                   []string
}

type hookTmplData struct {
	HookName                string
	Extension               string
	LefthookExe             string
	Rc                      string
	Roots                   []string
	AssertLefthookInstalled bool
}

func Hook(hookName string, args Args) []byte {
	buf := &bytes.Buffer{}
	t := template.Must(template.ParseFS(templatesFS, "hook.tmpl"))
	err := t.Execute(buf, hookTmplData{
		HookName:                hookName,
		Extension:               getExtension(),
		Rc:                      args.Rc,
		AssertLefthookInstalled: args.AssertLefthookInstalled,
		Roots:                   args.Roots,
		LefthookExe:             strings.ReplaceAll(strings.TrimSpace(args.LefthookExe), "\n", ";"),
	})
	if err != nil {
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
