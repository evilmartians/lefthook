package action

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getNChars(t *testing.T) {
	for i, tt := range [...]struct {
		source, cut, rest []string
		n                 int
	}{
		{
			source: []string{"str1", "str2", "str3"},
			n:      0,
			cut:    []string{"str1"},
			rest:   []string{"str2", "str3"},
		},
		{
			source: []string{"str1", "str2", "str3"},
			n:      4,
			cut:    []string{"str1"},
			rest:   []string{"str2", "str3"},
		},
		{
			source: []string{"str1", "str2", "str3"},
			n:      6,
			cut:    []string{"str1"},
			rest:   []string{"str2", "str3"},
		},
		{
			source: []string{"str1", "str2", "str3"},
			n:      8,
			cut:    []string{"str1"}, // because we need to add a space
			rest:   []string{"str2", "str3"},
		},
		{
			source: []string{"str1", "str2", "str3"},
			n:      9,
			cut:    []string{"str1", "str2"},
			rest:   []string{"str3"},
		},
		{
			source: []string{"str1", "str2", "str3"},
			n:      500,
			cut:    []string{"str1", "str2", "str3"},
			rest:   nil,
		},
		{
			source: nil,
			n:      2,
			cut:    nil,
			rest:   nil,
		},
	} {
		t.Run(fmt.Sprintf("test %d", i), func(t *testing.T) {
			assert := assert.New(t)
			cut, rest := getNChars(tt.source, tt.n)

			assert.EqualValues(cut, tt.cut)
			assert.EqualValues(rest, tt.rest)
		})
	}
}

func Test_replaceInChunks(t *testing.T) {
	for i, tt := range [...]struct {
		str       string
		templates map[string]*template
		maxlen    int
		action    *Action
	}{
		{
			str: "echo {staged_files}",
			templates: map[string]*template{
				"{staged_files}": {
					files: []string{"file1", "file2", "file3"},
					cnt:   1,
				},
			},
			maxlen: 300,
			action: &Action{
				Execs: []string{"echo file1 file2 file3"},
				Files: []string{"file1", "file2", "file3"},
			},
		},
		{
			str: "echo {staged_files}",
			templates: map[string]*template{
				"{staged_files}": {
					files: []string{"file1", "file2", "file3"},
					cnt:   1,
				},
			},
			maxlen: 10,
			action: &Action{
				Execs: []string{
					"echo file1",
					"echo file2",
					"echo file3",
				},
				Files: []string{"file1", "file2", "file3"},
			},
		},
		{
			str: "echo {files} && git add {files}",
			templates: map[string]*template{
				"{files}": {
					files: []string{"file1", "file2", "file3"},
					cnt:   2,
				},
			},
			maxlen: 49, // (49 - 17(len of command without templates)) / 2 = 16, but we need 17 (3 words + 2 spaces)
			action: &Action{
				Execs: []string{
					"echo file1 file2 && git add file1 file2",
					"echo file3 && git add file3",
				},
				Files: []string{"file1", "file2", "file3"},
			},
		},
		{
			str: "echo {files} && git add {files}",
			templates: map[string]*template{
				"{files}": {
					files: []string{"file1", "file2", "file3"},
					cnt:   2,
				},
			},
			maxlen: 51,
			action: &Action{
				Execs: []string{
					"echo file1 file2 file3 && git add file1 file2 file3",
				},
				Files: []string{"file1", "file2", "file3"},
			},
		},
		{
			str: "echo {push_files} && git add {files}",
			templates: map[string]*template{
				"{push_files}": {
					files: []string{"push-file"},
					cnt:   1,
				},
				"{files}": {
					files: []string{"file1", "file2"},
					cnt:   1,
				},
			},
			maxlen: 10,
			action: &Action{
				Execs: []string{
					"echo push-file && git add file1",
					"echo push-file && git add file2",
				},
				Files: []string{"push-file", "file1", "file2"},
			},
		},
		{
			str: "echo {push_files} && git add {files}",
			templates: map[string]*template{
				"{push_files}": {
					files: []string{"push1", "push2", "push3"},
					cnt:   1,
				},
				"{files}": {
					files: []string{"file1", "file2"},
					cnt:   1,
				},
			},
			maxlen: 27,
			action: &Action{
				Execs: []string{
					"echo push1 && git add file1",
					"echo push2 && git add file2",
					"echo push3 && git add file2",
				},
				Files: []string{"push1", "push2", "push3", "file1", "file2"},
			},
		},
	} {
		t.Run(fmt.Sprintf("test %d", i), func(t *testing.T) {
			assert := assert.New(t)
			action := replaceInChunks(tt.str, tt.templates, tt.maxlen)

			assert.ElementsMatch(action.Files, tt.action.Files)
			assert.Equal(action.Execs, tt.action.Execs)
		})
	}
}

func Test_replaceQuoted(t *testing.T) {
	for i, tt := range [...]struct {
		name, source, substitution string
		files                      []string
		result                     string
	}{
		{
			name:         "without substitutions",
			source:       "echo",
			substitution: "{staged_files}",
			files:        []string{"a", "b"},
			result:       "echo",
		},
		{
			name:         "with simple substitution",
			source:       "echo {staged_files}",
			substitution: "{staged_files}",
			files:        []string{"test.rb", "README"},
			result:       "echo test.rb README",
		},
		{
			name:         "with single quoted substitution",
			source:       "echo '{staged_files}'",
			substitution: "{staged_files}",
			files:        []string{"test.rb", "README"},
			result:       "echo 'test.rb' 'README'",
		},
		{
			name:         "with double quoted substitution",
			source:       `echo "{staged_files}"`,
			substitution: "{staged_files}",
			files:        []string{"test.rb", "README"},
			result:       `echo "test.rb" "README"`,
		},
		{
			name:         "with escaped files double quoted",
			source:       `echo "{staged_files}"`,
			substitution: "{staged_files}",
			files:        []string{"'test me.rb'", "README"},
			result:       `echo "test me.rb" "README"`,
		},
		{
			name:         "with escaped files single quoted",
			source:       "echo '{staged_files}'",
			substitution: "{staged_files}",
			files:        []string{"'test me.rb'", "README"},
			result:       `echo 'test me.rb' 'README'`,
		},
		{
			name:         "with escaped files",
			source:       "echo {staged_files}",
			substitution: "{staged_files}",
			files:        []string{"'test me.rb'", "README"},
			result:       `echo 'test me.rb' README`,
		},
		{
			name:         "with many substitutions",
			source:       `echo "{staged_files}" {staged_files}`,
			substitution: "{staged_files}",
			files:        []string{"'test me.rb'", "README"},
			result:       `echo "test me.rb" "README" 'test me.rb' README`,
		},
	} {
		t.Run(fmt.Sprintf("%d: %s", i, tt.name), func(t *testing.T) {
			result := replaceQuoted(tt.source, tt.substitution, tt.files)
			assert.Equal(t, result, tt.result)
		})
	}
}
