package runner

import (
	"fmt"
	"testing"
)

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
