package templates

import (
	"bytes"
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func renderHook(t *testing.T, data hookTmplData) string {
	t.Helper()
	buf := &bytes.Buffer{}
	tmpl := template.Must(template.New("hook.tmpl").Funcs(template.FuncMap{
		"shellescape": shellescape,
	}).ParseFS(templatesFS, "hook.tmpl"))
	require.NoError(t, tmpl.ExecuteTemplate(buf, "hook.tmpl", data))
	return buf.String()
}

func TestShellescape(t *testing.T) {
	for name, tt := range map[string]struct {
		value string
		want  string
	}{
		"plain path":            {value: "/usr/bin/lefthook", want: `'/usr/bin/lefthook'`},
		"path with space":       {value: "/home/my user/bin/lefthook", want: `'/home/my user/bin/lefthook'`},
		"embedded single quote": {value: "/home/o'brien/lefthook", want: `'/home/o'\''brien/lefthook'`},
		"metacharacters":        {value: "/tmp/$(touch pwned)", want: `'/tmp/$(touch pwned)'`},
		"empty":                 {value: "", want: `''`},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, shellescape(tt.value))
		})
	}
}

func TestHookQuotesAutodetectedExecutablePath(t *testing.T) {
	hook := renderHook(t, hookTmplData{
		HookName:            "pre-commit",
		LefthookPathCurrent: "/home/my user/bin/lefthook",
	})

	assert.Contains(t, hook, `elif '/home/my user/bin/lefthook' -h >/dev/null 2>&1`)
	assert.Contains(t, hook, `'/home/my user/bin/lefthook' "$@"`)
	assert.NotContains(t, hook, `/home/my user/bin/lefthook -h`)
}

// The `lefthook` and `rc` config values are documented as commands and
// shell-expanded paths, so they must reach the generated hook unquoted.
func TestHookPreservesConfiguredCommandsAndExpansion(t *testing.T) {
	for name, tt := range map[string]struct {
		data hookTmplData
		want string
	}{
		"lefthook path is a command with arguments": {
			data: hookTmplData{HookName: "pre-commit", LefthookPath: "bundle exec lefthook"},
			want: `bundle exec lefthook "$@"`,
		},
		"lefthook path carries an env prefix": {
			data: hookTmplData{HookName: "pre-commit", LefthookPath: "LEFTHOOK_VERBOSE=1 lefthook"},
			want: `LEFTHOOK_VERBOSE=1 lefthook "$@"`,
		},
		"rc path keeps tilde for the shell to expand": {
			data: hookTmplData{HookName: "pre-commit", Rc: "~/.lefthookrc"},
			want: `[ -f ~/.lefthookrc ] && . ~/.lefthookrc`,
		},
		"rc path keeps environment expansion": {
			data: hookTmplData{HookName: "pre-commit", Rc: `"${XDG_CONFIG_HOME:-$HOME/.config}/lefthookrc"`},
			want: `. "${XDG_CONFIG_HOME:-$HOME/.config}/lefthookrc"`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Contains(t, renderHook(t, tt.data), tt.want)
		})
	}
}

func TestHookPreservesUnaffectedContent(t *testing.T) {
	hook := string(Hook("pre-commit", Args{}))
	assert.True(t, strings.HasPrefix(hook, "#!/bin/sh"))
	assert.Contains(t, hook, `call_lefthook run "pre-commit" "$@"`)
}
