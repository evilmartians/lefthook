package run

import (
	"fmt"
	"testing"
)

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	r := make(map[string]struct{})

	for _, item := range a {
		r[item] = struct{}{}
	}

	for _, item := range b {
		if _, ok := r[item]; !ok {
			return false
		}
	}

	return true
}

func TestGetNChars(t *testing.T) {
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
			cut, rest := getNChars(tt.source, tt.n)

			if !slicesEqual(cut, tt.cut) {
				t.Errorf("expected cut %v to be equal to %v", cut, tt.cut)
			}
			if !slicesEqual(rest, tt.rest) {
				t.Errorf("expected rest %v to be equal to %v", rest, tt.rest)
			}
		})
	}
}

func TestReplaceInChunks(t *testing.T) {
	for i, tt := range [...]struct {
		str       string
		templates map[string]*template
		maxlen    int
		res       *run
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
			res: &run{
				commands: []string{"echo file1 file2 file3"},
				files:    []string{"file1", "file2", "file3"},
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
			res: &run{
				commands: []string{
					"echo file1",
					"echo file2",
					"echo file3",
				},
				files: []string{"file1", "file2", "file3"},
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
			res: &run{
				commands: []string{
					"echo file1 file2 && git add file1 file2",
					"echo file3 && git add file3",
				},
				files: []string{"file1", "file2", "file3"},
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
			res: &run{
				commands: []string{
					"echo file1 file2 file3 && git add file1 file2 file3",
				},
				files: []string{"file1", "file2", "file3"},
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
			res: &run{
				commands: []string{
					"echo push-file && git add file1",
					"echo push-file && git add file2",
				},
				files: []string{"push-file", "file1", "file2"},
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
			res: &run{
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
			res := replaceInChunks(tt.str, tt.templates, tt.maxlen)
			if !slicesEqual(res.files, tt.res.files) {
				t.Errorf("expected files %v to be equal to %v", res.files, tt.res.files)
			}

			if len(res.commands) != len(tt.res.commands) {
				t.Errorf("expected commands to be %d instead of %d", len(tt.res.commands), len(res.commands))
			} else {
				for i, command := range res.commands {
					if command != tt.res.commands[i] {
						t.Errorf("expected command %v to be equal to %v", command, tt.res.commands[i])
					}
				}
			}
		})
	}
}
