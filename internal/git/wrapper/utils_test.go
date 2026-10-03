package wrapper

import (
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/v2/tests/helpers/cmdtest"
	"github.com/evilmartians/lefthook/v2/tests/helpers/loggertest"
)

func Test_trimmed(t *testing.T) {
	for name, tt := range map[string]struct {
		input  []string
		output []string
	}{
		"nil": {
			input:  nil,
			output: nil,
		},
		"empty": {
			input:  []string{},
			output: []string(nil),
		},
		"without-spaces": {
			input:  []string{"a", "b", "c"},
			output: []string{"a", "b", "c"},
		},
		"with-spaces": {
			input:  []string{"  a", "", " ", "b  ", " c "},
			output: []string{"a", "b", "c"},
		},
		"with-newlines": {
			input:  []string{"\n a", "b  \n", " \n", "\nc\n"},
			output: []string{"a", "b", "c"},
		},
		"mixed": {
			input:  []string{"\n a \t", "\tb  \n", "\n\tc\n", "\t\n   "},
			output: []string{"a", "b", "c"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			output := slices.Collect(trimmed(slices.Values(tt.input)))

			if !cmp.Equal(output, tt.output) {
				t.Errorf("trimmed() = %v, want %v", output, tt.output)
			}
		})
	}
}

func Test_unquoted(t *testing.T) {
	for name, tt := range map[string]struct {
		input  []string
		output []string
	}{
		"nil": {
			input:  nil,
			output: nil,
		},
		"empty": {
			input:  []string{},
			output: []string(nil),
		},
		"quoted": {
			input:  []string{`"a"`, "'b'", "`c`"},
			output: []string{"a", "b", "c"},
		},
		"unquoted": {
			input:  []string{"a\"", "b'", "`c", "d"},
			output: []string{"a\"", "b'", "`c", "d"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			output := slices.Collect(unquoted(slices.Values(tt.input)))

			if !cmp.Equal(output, tt.output) {
				t.Errorf("unquoted() = %v, want %v", output, tt.output)
			}
		})
	}
}

func TestWrapper_selectFiles(t *testing.T) {
	for name, tt := range map[string]struct {
		input     []string
		filepaths []string
		output    []string
	}{
		"nil": {
			input:     nil,
			filepaths: nil,
			output:    nil,
		},
		"empty": {
			input:     []string{},
			filepaths: []string{},
			output:    []string(nil),
		},
		"no-files": {
			input:  []string{"a.txt", "b.txt", "c.txt"},
			output: []string(nil),
		},
		"with-files": {
			input:     []string{"a.txt", "b.txt", "c.txt"},
			filepaths: []string{"a.txt", "b.txt", "d.txt"},
			output:    []string{"a.txt", "b.txt"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			w := New(fs, cmdtest.NewStubCmd(), loggertest.New())
			for _, filepath := range tt.filepaths {
				if err := afero.WriteFile(fs, filepath, []byte("data"), 0o644); err != nil {
					t.Fatalf("unexpected error = %v", err)
				}
			}
			output := slices.Collect(w.selectFiles(slices.Values(tt.input)))

			if !cmp.Equal(output, tt.output) {
				t.Errorf("wrapper.selectFiles() = %v, want %v", output, tt.output)
			}
		})
	}
}
