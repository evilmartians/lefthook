package cmd

import (
	"bytes"
	"embed"
	"text/template"

	"github.com/spf13/afero"
)

//go:embed templates/*
var templatesFS embed.FS

type hookTmplData struct {
	AutoInstall string
	HookName    string
}

func hookTemplate(hookName string, fs afero.Fs) []byte {
	buf := &bytes.Buffer{}
	t := template.Must(template.ParseFS(templatesFS, "templates/hook.tmpl"))
	err := t.Execute(buf, hookTmplData{
		AutoInstall: autoInstall(hookName, fs),
		HookName:    hookName,
	})
	check(err)

	return buf.Bytes()
}

func configTemplate() []byte {
	tmpl, err := templatesFS.ReadFile("templates/config.tmpl")
	check(err)

	return tmpl
}

func autoInstall(hookName string, fs afero.Fs) string {
	if hookName != checkSumHook {
		return ""
	}

	return "# lefthook_version: " + configChecksum(fs) + "\n\ncall_lefthook \"lefthook install\""
}
