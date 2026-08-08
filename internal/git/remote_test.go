package git

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func TestRemoteDirectoryName(t *testing.T) {
	for i, tt := range [...]struct {
		url    string
		ref    string
		result string
	}{
		{
			url:    "https://github.com/evilmartians/lefthook.git",
			ref:    "",
			result: "lefthook",
		},
		{
			url:    "https://github.com/evilmartians/lefthook.git",
			ref:    "main",
			result: "lefthook-main",
		},
		{
			// A ref containing a slash (e.g. a branch named "feat/x") must
			// not turn into a path separator: it would make RemoteFolder()
			// build a nested directory instead of a single flat one.
			// See https://github.com/evilmartians/lefthook/issues/1478
			url:    "https://github.com/scop/lefthook-test.git",
			ref:    "feat/test-branch",
			result: "lefthook-test-feat-test-branch",
		},
		{
			url:    "https://github.com/scop/lefthook-test.git",
			ref:    "release/2.0/rc1",
			result: "lefthook-test-release-2.0-rc1",
		},
	} {
		t.Run(fmt.Sprintf("%d %s %s", i, tt.url, tt.ref), func(t *testing.T) {
			result := RemoteDirectoryName(tt.url, tt.ref)
			if result != tt.result {
				t.Errorf("expected %q, got %q", tt.result, result)
			}

			// The result must always be usable as a single path component:
			// joining it onto a directory must not create any extra
			// intermediate directories.
			if filepath.Base(result) != result || strings.ContainsAny(result, `/\`) {
				t.Errorf("result %q is not a single path component", result)
			}
		})
	}
}
