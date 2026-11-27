package controller

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_withTemplates(t *testing.T) {
	templates := map[string]string{
		"slow": `[ "$CI" != "1" ]`,
		"name": "world",
	}

	for name, tt := range map[string]struct {
		condition any
		templates map[string]string
		want      any
	}{
		"bool": {
			condition: true,
			templates: templates,
			want:      true,
		},
		"git-state": {
			condition: "merge",
			templates: templates,
			want:      "merge",
		},
		"no-templates": {
			condition: []any{map[string]any{"run": "{slow}"}},
			want:      []any{map[string]any{"run": "{slow}"}},
		},
		"run-with-template": {
			condition: []any{map[string]any{"run": "{slow}"}},
			templates: templates,
			want:      []any{map[string]any{"run": `[ "$CI" != "1" ]`}},
		},
		"run-with-several-templates": {
			condition: []any{map[string]any{"run": "echo {name} && {slow}"}},
			templates: templates,
			want:      []any{map[string]any{"run": `echo world && [ "$CI" != "1" ]`}},
		},
		"unknown-template": {
			condition: []any{map[string]any{"run": "test {unknown}"}},
			templates: templates,
			want:      []any{map[string]any{"run": "test {unknown}"}},
		},
		"mixed-conditions": {
			condition: []any{
				"rebase",
				map[string]any{"ref": "main"},
				map[string]any{"ref": "dev", "run": "{slow}"},
			},
			templates: templates,
			want: []any{
				"rebase",
				map[string]any{"ref": "main"},
				map[string]any{"ref": "dev", "run": `[ "$CI" != "1" ]`},
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			got := withTemplates(tt.condition, tt.templates)

			if !cmp.Equal(got, tt.want) {
				t.Errorf("withTemplates() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_withTemplates_keepsConfigIntact(t *testing.T) {
	cond := map[string]any{"run": "{slow}"}

	withTemplates([]any{cond}, map[string]string{"slow": "true"})

	if cond["run"] != "{slow}" {
		t.Errorf("withTemplates() changed the config condition to %v", cond["run"])
	}
}
