package command

import (
	"testing"
)

func Test_parseGitConfigParameters(t *testing.T) {
	for name, tt := range map[string]struct {
		input string
		want  []string
	}{
		"empty": {
			input: "",
			want:  nil,
		},
		"single pair": {
			input: "'core.hooksPath=/tmp/hooks'",
			want:  []string{"core.hooksPath=/tmp/hooks"},
		},
		"multiple pairs": {
			input: "'core.hooksPath=/tmp/hooks' 'safe.directory=/repo'",
			want:  []string{"core.hooksPath=/tmp/hooks", "safe.directory=/repo"},
		},
		"value with escaped single quote": {
			// git encodes a literal ' as '\'' (close-quote, backslash-quote, reopen-quote)
			input: `'core.hooksPath=/path/it'\'` + `'s/hooks'`,
			want:  []string{"core.hooksPath=/path/it's/hooks"},
		},
		"value with spaces via multiple segments": {
			input: "'some.key=hello world'",
			want:  []string{"some.key=hello world"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			got := parseGitConfigParameters(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("parseGitConfigParameters(%q) = %v (len %d), want %v (len %d)",
					tt.input, got, len(got), tt.want, len(tt.want))
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("parseGitConfigParameters(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func Test_commandScopedHooksPath(t *testing.T) {
	for name, tt := range map[string]struct {
		envGitConfigParameters string
		envGitConfigCount      string
		envGitConfigKeys       map[string]string
		want                   string
	}{
		"no env vars": {
			want: "",
		},
		"GIT_CONFIG_PARAMETERS with hooksPath": {
			envGitConfigParameters: "'core.hooksPath=/tmp/git-hooks'",
			want:                   "/tmp/git-hooks",
		},
		"GIT_CONFIG_PARAMETERS with other key": {
			envGitConfigParameters: "'safe.directory=/repo'",
			want:                   "",
		},
		"GIT_CONFIG_PARAMETERS case insensitive key": {
			envGitConfigParameters: "'core.hookspath=/tmp/hooks'",
			want:                   "/tmp/hooks",
		},
		"GIT_CONFIG_PARAMETERS multiple pairs, hooksPath second": {
			envGitConfigParameters: "'safe.directory=/repo' 'core.hooksPath=/tmp/hooks'",
			want:                   "/tmp/hooks",
		},
		"GIT_CONFIG_COUNT with hooksPath": {
			envGitConfigCount: "1",
			envGitConfigKeys: map[string]string{
				"GIT_CONFIG_KEY_0":   "core.hooksPath",
				"GIT_CONFIG_VALUE_0": "/tmp/count-hooks",
			},
			want: "/tmp/count-hooks",
		},
		"GIT_CONFIG_COUNT with other key": {
			envGitConfigCount: "1",
			envGitConfigKeys: map[string]string{
				"GIT_CONFIG_KEY_0":   "safe.directory",
				"GIT_CONFIG_VALUE_0": "/repo",
			},
			want: "",
		},
		"GIT_CONFIG_COUNT case insensitive": {
			envGitConfigCount: "1",
			envGitConfigKeys: map[string]string{
				"GIT_CONFIG_KEY_0":   "CORE.HOOKSPATH",
				"GIT_CONFIG_VALUE_0": "/tmp/hooks",
			},
			want: "/tmp/hooks",
		},
		"GIT_CONFIG_COUNT multiple entries": {
			envGitConfigCount: "2",
			envGitConfigKeys: map[string]string{
				"GIT_CONFIG_KEY_0":   "safe.directory",
				"GIT_CONFIG_VALUE_0": "/repo",
				"GIT_CONFIG_KEY_1":   "core.hooksPath",
				"GIT_CONFIG_VALUE_1": "/tmp/hooks",
			},
			want: "/tmp/hooks",
		},
	} {
		t.Run(name, func(t *testing.T) {
			if tt.envGitConfigParameters != "" {
				t.Setenv("GIT_CONFIG_PARAMETERS", tt.envGitConfigParameters)
			}
			if tt.envGitConfigCount != "" {
				t.Setenv("GIT_CONFIG_COUNT", tt.envGitConfigCount)
				for k, v := range tt.envGitConfigKeys {
					t.Setenv(k, v)
				}
			}

			got := commandScopedHooksPath()
			if got != tt.want {
				t.Errorf("commandScopedHooksPath() = %q, want %q", got, tt.want)
			}
		})
	}
}
