package replacer

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/evilmartians/lefthook/v2/internal/config"
	"github.com/evilmartians/lefthook/v2/internal/run/controller/filter"
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

func Test_ReplaceAndSplit(t *testing.T) {
	type result struct {
		commands []string
		files    []string
	}
	for i, tt := range [...]struct {
		command string
		maxlen  int
		cache   map[string]*entry
		result  result
	}{
		{
			command: "echo {staged_files}",
			cache: map[string]*entry{
				"{staged_files}": {
					items: []string{"file1", "file2", "file3"},
					cnt:   1,
				},
			},
			maxlen: 300,
			result: result{
				commands: []string{"echo file1 file2 file3"},
				files:    []string{"file1", "file2", "file3"},
			},
		},
		{
			command: "echo {staged_files}",
			cache: map[string]*entry{
				"{staged_files}": {
					items: []string{"file1", "file2", "file3"},
					cnt:   1,
				},
			},
			maxlen: 10,
			result: result{
				commands: []string{
					"echo file1",
					"echo file2",
					"echo file3",
				},
				files: []string{"file1", "file2", "file3"},
			},
		},
		{
			command: "echo {files} && git add {files}",
			cache: map[string]*entry{
				"{files}": {
					items: []string{"file1", "file2", "file3"},
					cnt:   2,
				},
			},
			maxlen: 49, // (49 - 17(len of command without templates)) / 2 = 16, but we need 17 (3 words + 2 spaces)
			result: result{
				commands: []string{
					"echo file1 file2 && git add file1 file2",
					"echo file3 && git add file3",
				},
				files: []string{"file1", "file2", "file3"},
			},
		},
		{
			command: "echo {files} && git add {files}",
			cache: map[string]*entry{
				"{files}": {
					items: []string{"file1", "file2", "file3"},
					cnt:   2,
				},
			},
			maxlen: 51,
			result: result{
				commands: []string{
					"echo file1 file2 file3 && git add file1 file2 file3",
				},
				files: []string{"file1", "file2", "file3"},
			},
		},
		{
			command: "echo {push_files} && git add {files}",
			cache: map[string]*entry{
				"{push_files}": {
					items: []string{"push-file"},
					cnt:   1,
				},
				"{files}": {
					items: []string{"file1", "file2"},
					cnt:   1,
				},
			},
			maxlen: 10,
			result: result{
				commands: []string{
					"echo push-file && git add file1",
					"echo push-file && git add file2",
				},
				files: []string{"push-file", "file1", "file2"},
			},
		},
		{
			command: "echo {push_files} && git add {files}",
			cache: map[string]*entry{
				"{push_files}": {
					items: []string{"push1", "push2", "push3"},
					cnt:   1,
				},
				"{files}": {
					items: []string{"file1", "file2"},
					cnt:   1,
				},
			},
			maxlen: 27,
			result: result{
				commands: []string{
					"echo push1 && git add file1",
					"echo push2 && git add file2",
					"echo push3 && git add file2",
				},
				files: []string{"push1", "push2", "push3", "file1", "file2"},
			},
		},
	} {
		t.Run(fmt.Sprintf("test %d", i), func(t *testing.T) {
			assert := assert.New(t)
			r := Replacer{
				cache: tt.cache,
				files: map[string]func() ([]string, error){
					config.SubStagedFiles: func() ([]string, error) { return nil, nil },
					config.SubPushFiles:   func() ([]string, error) { return nil, nil },
					config.SubAllFiles:    func() ([]string, error) { return nil, nil },
					config.SubFiles:       func() ([]string, error) { return nil, nil },
				},
			}
			commands, files := r.ReplaceAndSplit(tt.command, tt.maxlen)

			assert.ElementsMatch(files, tt.result.files)
			assert.Equal(commands, tt.result.commands)
		})
	}
}

func Test_ReplaceAndSplit_CustomTemplates(t *testing.T) {
	t.Run("custom templates should not be escaped", func(t *testing.T) {
		assert := assert.New(t)

		// Create a replacer with custom templates (note: keys include braces)
		r := NewMocked([]string{"file1.js"}).AddTemplates(
			map[string]string{
				"use-mise": `eval "$(mise env)"`,
			},
		)

		// Discover templates in the command (use empty filter)
		emptyFilter := &filter.Filter{}
		err := r.Discover("{use-mise} prettier {staged_files}", emptyFilter)
		assert.NoError(err)

		// Replace templates
		commands, files := r.ReplaceAndSplit("{use-mise} prettier {staged_files}", 300)

		// Custom template should NOT be escaped (no quotes around it)
		assert.Equal([]string{`eval "$(mise env)" prettier file1.js`}, commands)
		assert.Equal([]string{"file1.js"}, files)
	})

	t.Run("file templates should still be escaped", func(t *testing.T) {
		assert := assert.New(t)

		// Create a replacer with a file that needs escaping
		r := NewMocked([]string{"file with spaces.js"})

		// Discover templates in the command (use empty filter)
		emptyFilter := &filter.Filter{}
		err := r.Discover("prettier {staged_files}", emptyFilter)
		assert.NoError(err)

		// Replace templates
		commands, _ := r.ReplaceAndSplit("prettier {staged_files}", 300)

		// File template SHOULD be escaped (with quotes)
		assert.Equal([]string{`prettier 'file with spaces.js'`}, commands)
	})
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
