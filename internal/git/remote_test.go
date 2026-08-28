package git

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
			// A ref containing a slash (e.g. a branch named "feat/x") must
			// not turn into a path separator: it would make RemoteFolder()
			// build a nested directory instead of a single flat one.
			// See https://github.com/evilmartians/lefthook/issues/1478
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
			result := RemoteDirectoryName(tt.url, tt.ref)
			assert.Equal(t, tt.result, result)

			// The result must always be usable as a single path component:
			// joining it onto a directory must not create any extra
			// intermediate directories.
			assert.Equal(t, result, filepath.Base(result), "result %q is not a single path component", result)
			assert.False(t, strings.ContainsAny(result, `/\`), "result %q is not a single path component", result)
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
		result := RemoteDirectoryName(url, ref)
		if other, ok := seen[result]; ok {
			t.Errorf("refs %q and %q both sanitize to %q", other, ref, result)
		}
		seen[result] = ref
	}
}
