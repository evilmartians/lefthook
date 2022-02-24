package templates

import (
	"bytes"
	"embed"
	"fmt"
	"runtime"
	"text/template"

	"github.com/evilmartians/lefthook/internal/config"
)

const checksumFormat = "# lefthook_version: %s"

//go:embed *
var templatesFS embed.FS

type hookTmplData struct {
	AutoInstall string
	HookName    string
	Extension   string
}

func Hook(hookName, configChecksum string) []byte {
	buf := &bytes.Buffer{}
	t := template.Must(template.ParseFS(templatesFS, "hook.tmpl"))
	err := t.Execute(buf, hookTmplData{
		AutoInstall: autoInstall(hookName, configChecksum),
		HookName:    hookName,
		Extension:   getExtension(),
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

func autoInstall(hookName, configChecksum string) string {
	if hookName != config.ChecksumHookName {
		return ""
	}

	return fmt.Sprintf(checksumFormat, configChecksum)
}

func getExtension() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}
