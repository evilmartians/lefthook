package git_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/evilmartians/lefthook/v2/internal/git"
	"github.com/google/go-cmp/cmp"
)

func TestRemoteDirectoryName(t *testing.T) {
	for name, tt := range map[string]struct {
		url    string
		ref    string
		result string
	}{
		"no ref": {
			url:    "https://github.com/evilmartians/lefthook.git",
			ref:    "",
			result: "lefthook",
		},
		"plain branch name": {
			url:    "https://github.com/evilmartians/lefthook.git",
			ref:    "main",
			result: "lefthook-main",
		},
		"ref containing a slash": {
			url:    "https://github.com/scop/lefthook-test.git",
			ref:    "feat/test-branch",
			result: "lefthook-test-feat-test--branch",
		},
		"ref containing multiple slashes": {
			url:    "https://github.com/scop/lefthook-test.git",
			ref:    "release/2.0/rc1",
			result: "lefthook-test-release-2.0-rc1",
		},
		"ref containing a literal hyphen": {
			url:    "https://github.com/scop/lefthook-test.git",
			ref:    "feat/a-b",
			result: "lefthook-test-feat-a--b",
		},
	} {
		t.Run(name, func(t *testing.T) {
			result := git.RemoteDirectoryName(tt.url, tt.ref)
			if !cmp.Equal(result, tt.result) {
				t.Errorf("RemoteDirectoryName() = %v, want %v", result, tt.result)
			}

			if result != filepath.Base(result) {
				t.Errorf("%v must be a single path component", result)
			}

			if strings.ContainsAny(result, `/\`) {
				t.Errorf("%v must not contain slashes", result)
			}
		})
	}
}

func TestRemoteDirectoryName_distinctRefsDoNotCollide(t *testing.T) {
	// A slash and a literal hyphen must not be ambiguous with each other:
	// "feat/a-b" and "feat/a/b" are distinct refs and must sanitize to
	// distinct directory names, or synchronizing one ref would silently
	// overwrite the checkout (and cached state) used by the other.
	// See https://github.com/evilmartians/lefthook/pull/1486#discussion (Greptile P1).
	const url = "https://github.com/scop/lefthook-test.git"
	refs := []string{
		"feat/a-b",
		"feat/a/b",
		"feat-a-b",
		"feat-a/b",
		"feat/a--b",
	}

	seen := make(map[string]string, len(refs))
	for _, ref := range refs {
		result := git.RemoteDirectoryName(url, ref)
		if other, ok := seen[result]; ok {
			t.Errorf("refs %q and %q both sanitize to %q", other, ref, result)
		}
		seen[result] = ref
	}
}
