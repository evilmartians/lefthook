package templates

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLooksLikeFilePath(t *testing.T) {
	for name, tt := range map[string]struct {
		value string
		want  bool
	}{
		"absolute path":                     {value: "/usr/local/bin/lefthook", want: true},
		"relative path":                       {value: "./bin/lefthook", want: true},
		"spaced filesystem path":            {value: "/tmp/my project/bin/lefthook", want: true},
		"bundle exec command":               {value: "bundle exec lefthook", want: false},
		"env wrapper command":               {value: "/usr/bin/env lefthook", want: false},
		"relative wrapper with flag":        {value: "./wrapper --flag", want: false},
		"go tool command":                   {value: "go tool lefthook", want: false},
		"node relative path command":        {value: "node ./bin/lefthook", want: false},
		"relative wrapper short flag":       {value: "./wrapper -f", want: false},
		"plain command without slash":       {value: "lefthook", want: false},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, looksLikeFilePath(tt.value))
		})
	}
}

func TestShellWordForRc(t *testing.T) {
	for name, tt := range map[string]struct {
		in   string
		want string
	}{
		"empty":                         {in: "", want: ""},
		"plain path":                    {in: "/home/user/.lefthookrc", want: "/home/user/.lefthookrc"},
		"spaced path":                   {in: "/tmp/my project/.lefthookrc", want: `"/tmp/my project/.lefthookrc"`},
		"already quoted spaced path":    {in: `"/tmp/my project/.lefthookrc"`, want: `"/tmp/my project/.lefthookrc"`},
		"expansion without spaces":      {in: "${XDG_CONFIG_HOME:-$HOME/.config}/lefthookrc", want: "${XDG_CONFIG_HOME:-$HOME/.config}/lefthookrc"},
		"expansion with spaces":         {in: `${XDG_CONFIG_HOME:-$HOME/.config}/my lefthookrc`, want: `"${XDG_CONFIG_HOME:-$HOME/.config}/my lefthookrc"`},
		"already quoted expansion path": {in: `"${XDG_CONFIG_HOME:-$HOME/.config}/my lefthookrc"`, want: `"${XDG_CONFIG_HOME:-$HOME/.config}/my lefthookrc"`},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, shellWordForRc(tt.in))
		})
	}
}

func TestShellWordForPath(t *testing.T) {
	for name, tt := range map[string]struct {
		in   string
		want string
	}{
		"empty":                {in: "", want: ""},
		"no spaces":            {in: "/usr/bin/lefthook", want: "/usr/bin/lefthook"},
		"spaces need quoting":  {in: "/tmp/my project/lefthook", want: `"/tmp/my project/lefthook"`},
		"preserves expansion":  {in: "${HOME}/my lefthookrc", want: `"${HOME}/my lefthookrc"`},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, shellWordForPath(tt.in))
		})
	}
}

func TestShellInvokeForLefthookPath(t *testing.T) {
	for name, tt := range map[string]struct {
		in   string
		want string
	}{
		"spaced path":     {in: "/tmp/my project/bin/lefthook", want: `"/tmp/my project/bin/lefthook"`},
		"shell command":   {in: "bundle exec lefthook", want: "bundle exec lefthook"},
		"env wrapper":     {in: "/usr/bin/env lefthook", want: "/usr/bin/env lefthook"},
		"wrapper flag":    {in: "./wrapper --flag", want: "./wrapper --flag"},
		"node wrapper":    {in: "node ./bin/lefthook", want: "node ./bin/lefthook"},
		"wrapper -f":      {in: "./wrapper -f", want: "./wrapper -f"},
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tt.want, shellInvokeForLefthookPath(tt.in))
		})
	}
}

func TestShellEscapeDoubleQuoteEscapesEmbeddedQuotes(t *testing.T) {
	assert.Equal(t, `a\"b\\c$`, shellEscapeDoubleQuote(`a"b\c$`))
}

func TestHook(t *testing.T) {
	binPath := filepath.Join("/tmp", "my project", "bin", "lefthook")
	slashPath := filepath.ToSlash(binPath)
	rcWithSpaces := filepath.ToSlash(filepath.Join("/tmp", "my project", ".lefthookrc"))

	for name, tt := range map[string]struct {
		args        Args
		mustContain []string
		mustExclude []string
	}{
		"quotes configured path with spaces": {
			args: Args{LefthookPath: binPath},
			mustContain: []string{
				`"` + slashPath + `" "$@"`,
			},
			mustExclude: []string{
				"elif " + slashPath + " -h",
			},
		},
		"quotes rc path with spaces": {
			args: Args{Rc: filepath.Join("/tmp", "my project", ".lefthookrc")},
			mustContain: []string{
				`[ -f "` + rcWithSpaces + `" ] && . "` + rcWithSpaces + `"`,
			},
		},
		"preserves rc shell expansion": {
			args: Args{Rc: `${XDG_CONFIG_HOME:-$HOME/.config}/lefthookrc`},
			mustContain: []string{
				`[ -f ${XDG_CONFIG_HOME:-$HOME/.config}/lefthookrc ]`,
			},
			mustExclude: []string{
				`\$HOME`,
				`\$XDG_CONFIG_HOME`,
			},
		},
		"quotes spaced rc path with shell expansion": {
			args: Args{Rc: `${XDG_CONFIG_HOME:-$HOME/.config}/my lefthookrc`},
			mustContain: []string{
				`[ -f "${XDG_CONFIG_HOME:-$HOME/.config}/my lefthookrc" ]`,
			},
			mustExclude: []string{
				`\$HOME`,
			},
		},
		"preserves configured shell command": {
			args: Args{LefthookPath: "bundle exec lefthook"},
			mustContain: []string{
				`bundle exec lefthook "$@"`,
				`elif test -n "bundle exec lefthook"`,
			},
			mustExclude: []string{
				`"bundle exec lefthook" "$@"`,
			},
		},
		"preserves env wrapper command": {
			args: Args{LefthookPath: "/usr/bin/env lefthook"},
			mustContain: []string{
				`/usr/bin/env lefthook "$@"`,
			},
			mustExclude: []string{
				`"/usr/bin/env lefthook" "$@"`,
			},
		},
		"preserves relative wrapper command": {
			args: Args{LefthookPath: "./wrapper --flag"},
			mustContain: []string{
				`./wrapper --flag "$@"`,
			},
			mustExclude: []string{
				`"./wrapper --flag" "$@"`,
			},
		},
		"preserves node wrapper command": {
			args: Args{LefthookPath: "node ./bin/lefthook"},
			mustContain: []string{
				`node ./bin/lefthook "$@"`,
			},
			mustExclude: []string{
				`"node ./bin/lefthook" "$@"`,
			},
		},
		"preserves already quoted rc path": {
			args: Args{Rc: `"/tmp/my project/.lefthookrc"`},
			mustContain: []string{
				`[ -f "/tmp/my project/.lefthookrc" ] && . "/tmp/my project/.lefthookrc"`,
			},
			mustExclude: []string{
				`\"`,
			},
		},
		"preserves already quoted rc expansion path": {
			args: Args{Rc: `"${XDG_CONFIG_HOME:-$HOME/.config}/my lefthookrc"`},
			mustContain: []string{
				`[ -f "${XDG_CONFIG_HOME:-$HOME/.config}/my lefthookrc" ]`,
			},
			mustExclude: []string{
				`\"`,
				`\$HOME`,
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			out := string(Hook("pre-commit", tt.args))
			for _, want := range tt.mustContain {
				assert.Contains(t, out, want)
			}
			for _, omit := range tt.mustExclude {
				assert.NotContains(t, out, omit)
			}
		})
	}
}

func TestHook_QuotesExecutablePathWithSpaces(t *testing.T) {
	binDir := filepath.Join(t.TempDir(), "space dir")
	require.NoError(t, os.MkdirAll(binDir, 0o755))
	bin := filepath.Join(binDir, "lefthook")
	require.NoError(t, os.WriteFile(bin, []byte("#!/bin/sh\n"), 0o755))

	old := hookExecutable
	hookExecutable = func() (string, error) { return bin, nil }
	t.Cleanup(func() { hookExecutable = old })

	out := string(Hook("pre-commit", Args{}))
	assert.Contains(t, out, `"`+filepath.ToSlash(bin)+`" "$@"`)
}
