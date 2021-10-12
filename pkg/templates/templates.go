package templates

import (
	"bytes"
	"embed"
	"runtime"
	"text/template"

	"github.com/spf13/afero"
)

//go:embed *
var templatesFS embed.FS

type hookTmplData struct {
	AutoInstall string
	HookName    string
	Extension   string
}

func Hook(hookName string, fs afero.Fs) []byte {
	buf := &bytes.Buffer{}
	t := template.Must(template.ParseFS(templatesFS, "hook.tmpl"))
	err := t.Execute(buf, hookTmplData{
		AutoInstall: autoInstall(hookName, fs),
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

func autoInstall(hookName string, fs afero.Fs) string {
	//if hookName != checkSumHook {
	//	return ""
	//}
	//
	//return "# lefthook_version: " + configChecksum(fs) + "\n\ncall_lefthook \"install\""

	return ""
}

func getExtension() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}
