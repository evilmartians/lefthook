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
		glob           string
	}{
		{
			source: []string{"folder/subfolder/0.rb", "1.txt", "2.RB", "3.rbs"},
			glob:   "",
			result: []string{"folder/subfolder/0.rb", "1.txt", "2.RB", "3.rbs"},
		},
		{
			source: []string{"folder/subfolder/0.rb", "1.txt", "2.RB", "3.rbs"},
			glob:   "*.rb",
			result: []string{"folder/subfolder/0.rb", "2.RB"},
		},
		{
			source: []string{"folder/subfolder/0.rb", "1.rbs"},
			glob:   "**/*.rb",
			result: []string{"folder/subfolder/0.rb"},
		},
		{
			source: []string{"folder/0.rb", "1.rBs", "2.rbv"},
			glob:   "*.rb?",
			result: []string{"1.rBs", "2.rbv"},
		},
		{
			source: []string{"f.a", "f.b", "f.c", "f.cn"},
			glob:   "*.{a,b,cn}",
			result: []string{"f.a", "f.b", "f.cn"},
		},
	} {
		t.Run(fmt.Sprintf("%d:", i), func(t *testing.T) {
			res := byGlob(tt.source, tt.glob)
			if !slicesEqual(res, tt.result) {
				t.Errorf("expected %v to be equal to %v", res, tt.result)
			}
		})
	}
}

func TestByExclude(t *testing.T) {
	for i, tt := range [...]struct {
		source, result []string
		exclude        string
	}{
		{
			source:  []string{"folder/subfolder/0.rb", "1.txt", "2.RB", "3.rb"},
			exclude: "",
			result:  []string{"folder/subfolder/0.rb", "1.txt", "2.RB", "3.rb"},
		},
		{
			source:  []string{"folder/subfolder/0.rb", "1.txt", "2.RB", "3.rb"},
			exclude: "^[^/]*\\.rb",
			result:  []string{"folder/subfolder/0.rb", "1.txt", "2.RB"},
		},
		{
			source:  []string{"folder/subfolder/0.rb", "1.rb"},
			exclude: "^.+/.+.*\\.rb",
			result:  []string{"1.rb"},
		},
		{
			source:  []string{"folder/0.rb", "1.rBs", "2.rbv"},
			exclude: ".*\\.rb.?",
			result:  []string{"1.rBs"},
		},
		{
			source:  []string{"f.a", "f.b", "f.c", "f.cn"},
			exclude: ".*\\.(a|b|cn)",
			result:  []string{"f.c"},
		},
	} {
		t.Run(fmt.Sprintf("%d:", i), func(t *testing.T) {
			res := byExclude(tt.source, tt.exclude)
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
