package templates

import (
	"bytes"
	"embed"
	"fmt"
	"runtime"
	"text/template"
)

const checksumFormat = "%s %d\n"

//go:embed *
var templatesFS embed.FS

type hookTmplData struct {
	AutoInstall string
	HookName    string
	Extension   string
}

func Hook(hookName string) []byte {
	buf := &bytes.Buffer{}
	t := template.Must(template.ParseFS(templatesFS, "hook.tmpl"))
	err := t.Execute(buf, hookTmplData{
		HookName:  hookName,
		Extension: getExtension(),
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

func Checksum(checksum string, timestamp int64) []byte {
	return []byte(fmt.Sprintf(checksumFormat, checksum, timestamp))
}

func getExtension() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}
