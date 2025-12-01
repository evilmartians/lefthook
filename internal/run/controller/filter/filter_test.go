package filter

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

func TestByGlob(t *testing.T) {
	for i, tt := range [...]struct {
		source, result []string
		glob           []string
		globMatcher    string
	}{
		{
			source:      []string{"folder/subfolder/0.rb", "1.txt", "2.RB", "3.rbs"},
			glob:        []string{},
			globMatcher: "",
			result:      []string{"folder/subfolder/0.rb", "1.txt", "2.RB", "3.rbs"},
		},
		{
			source:      []string{"folder/subfolder/0.rb", "1.txt", "2.RB", "3.rbs"},
			glob:        []string{"*.rb"},
			globMatcher: "",
			result:      []string{"folder/subfolder/0.rb", "2.RB"},
		},
		{
			source:      []string{"folder/subfolder/0.rb", "1.rbs"},
			glob:        []string{"**/*.rb"},
			globMatcher: "",
			result:      []string{"folder/subfolder/0.rb"},
		},
		{
			source:      []string{"folder/0.rb", "1.rBs", "2.rbv"},
			glob:        []string{"*.rb?"},
			globMatcher: "",
			result:      []string{"1.rBs", "2.rbv"},
		},
		{
			source:      []string{"f.a", "f.b", "f.c", "f.cn"},
			glob:        []string{"*.{a,b,cn}"},
			globMatcher: "",
			result:      []string{"f.a", "f.b", "f.cn"},
		},
	} {
		t.Run(fmt.Sprintf("%d:", i), func(t *testing.T) {
			res := byGlob(tt.source, tt.glob, tt.globMatcher)
			if !slicesEqual(res, tt.result) {
				t.Errorf("expected %v to be equal to %v", res, tt.result)
			}
		})
	}
}

func TestByGlobDoublestar(t *testing.T) {
	for i, tt := range [...]struct {
		source, result []string
		glob           []string
		globMatcher    string
	}{
		{
			source:      []string{"0.rb", "folder/1.rb", "folder/subfolder/2.rb"},
			glob:        []string{"**/*.rb"},
			globMatcher: "doublestar",
			result:      []string{"0.rb", "folder/1.rb", "folder/subfolder/2.rb"},
		},
		{
			source:      []string{"0.rb", "folder/1.rb", "folder/subfolder/2.rb"},
			glob:        []string{"**/*.rb"},
			globMatcher: "",
			result:      []string{"folder/1.rb", "folder/subfolder/2.rb"},
		},
		{
			source:      []string{"a/b.go", "a/c/d.go", "e.go"},
			glob:        []string{"**/*.go"},
			globMatcher: "doublestar",
			result:      []string{"a/b.go", "a/c/d.go", "e.go"},
		},
		{
			source:      []string{"a/b.go", "a/c/d.go", "e.go"},
			glob:        []string{"**/*.go"},
			globMatcher: "",
			result:      []string{"a/b.go", "a/c/d.go"},
		},
		{
			source:      []string{"test.js", "src/app.js", "src/lib/util.js"},
			glob:        []string{"**/*.js"},
			globMatcher: "doublestar",
			result:      []string{"test.js", "src/app.js", "src/lib/util.js"},
		},
	} {
		t.Run(fmt.Sprintf("doublestar-%d:", i), func(t *testing.T) {
			res := byGlob(tt.source, tt.glob, tt.globMatcher)
			if !slicesEqual(res, tt.result) {
				t.Errorf("expected %v to be equal to %v", res, tt.result)
			}
		})
	}
}

func TestByExclude(t *testing.T) {
	for i, tt := range [...]struct {
		source, result []string
		exclude        []string
		globMatcher    string
	}{
		{
			source:      []string{"folder/subfolder/0.rb", "1.txt", "2.RB", "3.rb"},
			exclude:     []string{},
			globMatcher: "",
			result:      []string{"folder/subfolder/0.rb", "1.txt", "2.RB", "3.rb"},
		},
		{
			source:      []string{"f.a", "f.b", "f.c", "f.cn"},
			exclude:     []string{"*.a", "*.b", "*.cn"},
			globMatcher: "",
			result:      []string{"f.c"},
		},
	} {
		t.Run(fmt.Sprintf("%d:", i), func(t *testing.T) {
			res := byExclude(tt.source, tt.exclude, tt.globMatcher)
			if !slicesEqual(res, tt.result) {
				t.Errorf("expected %v to be equal to %v", res, tt.result)
			}
		})
	}
}

func TestByExcludeDoublestar(t *testing.T) {
	for i, tt := range [...]struct {
		source, result []string
		exclude        []string
		globMatcher    string
	}{
		{
			source:      []string{"0.rb", "folder/1.rb", "folder/subfolder/2.rb", "test.js"},
			exclude:     []string{"**/*.rb"},
			globMatcher: "doublestar",
			result:      []string{"test.js"},
		},
		{
			source:      []string{"0.rb", "folder/1.rb", "folder/subfolder/2.rb", "test.js"},
			exclude:     []string{"**/*.rb"},
			globMatcher: "",
			result:      []string{"0.rb", "test.js"},
		},
		{
			source:      []string{"src/app.js", "src/lib/util.js", "test.py", "src/test.py"},
			exclude:     []string{"**/*.py"},
			globMatcher: "doublestar",
			result:      []string{"src/app.js", "src/lib/util.js"},
		},
		{
			source:      []string{"a.go", "src/b.go", "src/lib/c.go"},
			exclude:     []string{"**/*.go"},
			globMatcher: "doublestar",
			result:      []string{},
		},
	} {
		t.Run(fmt.Sprintf("doublestar-%d:", i), func(t *testing.T) {
			res := byExclude(tt.source, tt.exclude, tt.globMatcher)
			if !slicesEqual(res, tt.result) {
				t.Errorf("expected %v to be equal to %v", res, tt.result)
			}
		})
	}
}

func TestByRoot(t *testing.T) {
	for i, tt := range [...]struct {
		source, result []string
		path           string
	}{
		{
			source: []string{"folder/subfolder/0.rb", "1.txt", "2.RB", "3.rb"},
			path:   "",
			result: []string{"folder/subfolder/0.rb", "1.txt", "2.RB", "3.rb"},
		},
		{
			source: []string{"folder/subfolder/0.rb", "subfolder/1.txt", "folder/2.RB", "3.rbs"},
			path:   "folder",
			result: []string{".//subfolder/0.rb", ".//2.RB"},
		},
		{
			source: []string{"folder/subfolder/0.rb", "folder/1.rbs"},
			path:   "folder/subfolder",
			result: []string{".//0.rb"},
		},
		{
			source: []string{"folder/subfolder/0.rb", "folder/1.rbs"},
			path:   "folder/subfolder/",
			result: []string{"./0.rb"},
		},
	} {
		t.Run(fmt.Sprintf("%d:", i), func(t *testing.T) {
			res := byRoot(tt.source, tt.path)
			if !slicesEqual(res, tt.result) {
				t.Errorf("expected %v to be equal to %v", res, tt.result)
			}
		})
	}
}
