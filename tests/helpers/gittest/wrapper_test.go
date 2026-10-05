package gittest_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/evilmartians/lefthook/v2/internal/git"
	"github.com/evilmartians/lefthook/v2/internal/git/wrapper"
	"github.com/evilmartians/lefthook/v2/tests/helpers/gittest"
)

var _ git.Wrapper = (*gittest.StubWrapper)(nil)

func TestStubWrapper(t *testing.T) {
	files := []string{"a", "b"}

	for name, tt := range map[string]struct {
		setup func(w *gittest.StubWrapper, calls *[]any)
		call  func(w *gittest.StubWrapper) any
		want  []any
	}{
		"version": {
			setup: func(w *gittest.StubWrapper, calls *[]any) {
				w.VersionFunc = func() (string, error) { return "2.0", nil }
			},
			call: func(w *gittest.StubWrapper) any { v, _ := w.Version(); return v },
			want: []any{"2.0"},
		},
		"paths": {
			setup: func(w *gittest.StubWrapper, calls *[]any) {
				w.PathsFunc = func() (*wrapper.Paths, error) { return &wrapper.Paths{Root: "root"}, nil }
			},
			call: func(w *gittest.StubWrapper) any { p, _ := w.Paths(); return p.Root },
			want: []any{"root"},
		},
		"local-hooks-path": {
			setup: func(w *gittest.StubWrapper, calls *[]any) {
				w.LocalHooksPathFunc = func() string { return "local" }
			},
			call: func(w *gittest.StubWrapper) any { return w.LocalHooksPath() },
			want: []any{"local"},
		},
		"unset-local-hooks-path": {
			setup: func(w *gittest.StubWrapper, calls *[]any) {
				w.UnsetLocalHooksPathFunc = func() error { *calls = append(*calls, "unset-local"); return nil }
			},
			call: func(w *gittest.StubWrapper) any { return w.UnsetLocalHooksPath() },
			want: []any{"unset-local", nil},
		},
		"global-hooks-path": {
			setup: func(w *gittest.StubWrapper, calls *[]any) {
				w.StubGlobalHooksPath = func() string { return "global" }
			},
			call: func(w *gittest.StubWrapper) any { return w.GlobalHooksPath() },
			want: []any{"global"},
		},
		"unset-global-hooks-path": {
			setup: func(w *gittest.StubWrapper, calls *[]any) {
				w.UnsetGlobalHooksPathFunc = func() error { *calls = append(*calls, "unset-global"); return nil }
			},
			call: func(w *gittest.StubWrapper) any { return w.UnsetGlobalHooksPath() },
			want: []any{"unset-global", nil},
		},
		"all-files": {
			setup: func(w *gittest.StubWrapper, calls *[]any) {
				w.AllFilesFunc = func() ([]string, error) { return files, nil }
			},
			call: func(w *gittest.StubWrapper) any { f, _ := w.AllFiles(); return f },
			want: []any{files},
		},
		"staged-files": {
			setup: func(w *gittest.StubWrapper, calls *[]any) {
				w.StagedFilesFunc = func() ([]string, error) { return files, nil }
			},
			call: func(w *gittest.StubWrapper) any { f, _ := w.StagedFiles(); return f },
			want: []any{files},
		},
		"staged-files-with-deleted": {
			setup: func(w *gittest.StubWrapper, calls *[]any) {
				w.StagedFilesWithDeletedFunc = func() ([]string, error) { return files, nil }
			},
			call: func(w *gittest.StubWrapper) any { f, _ := w.StagedFilesWithDeleted(); return f },
			want: []any{files},
		},
		"push-files": {
			setup: func(w *gittest.StubWrapper, calls *[]any) {
				w.PushFilesFunc = func() ([]string, error) { return files, nil }
			},
			call: func(w *gittest.StubWrapper) any { f, _ := w.PushFiles(); return f },
			want: []any{files},
		},
		"files-by-command-relative": {
			setup: func(w *gittest.StubWrapper, calls *[]any) {
				w.FilesByCommandRelativeFunc = func(cmd string, dir string) ([]string, error) {
					*calls = append(*calls, cmd, dir)
					return files, nil
				}
			},
			call: func(w *gittest.StubWrapper) any { f, _ := w.FilesByCommandRelative("ls", "dir"); return f },
			want: []any{"ls", "dir", files},
		},
		"status-short": {
			setup: func(w *gittest.StubWrapper, calls *[]any) {
				w.StatusShortFunc = func() ([]wrapper.FileStatus, error) {
					return []wrapper.FileStatus{{Path: "a"}}, nil
				}
			},
			call: func(w *gittest.StubWrapper) any { s, _ := w.StatusShort(); return s[0].Path },
			want: []any{"a"},
		},
		"diff": {
			setup: func(w *gittest.StubWrapper, calls *[]any) {
				w.DiffFunc = func(f []string, colors bool) (string, error) {
					*calls = append(*calls, f, colors)
					return "diff", nil
				}
			},
			call: func(w *gittest.StubWrapper) any { d, _ := w.Diff(files, true); return d },
			want: []any{files, true, "diff"},
		},
		"save-unstaged-diff": {
			setup: func(w *gittest.StubWrapper, calls *[]any) {
				w.SaveUnstagedDiffFunc = func(f []string) error { *calls = append(*calls, f); return nil }
			},
			call: func(w *gittest.StubWrapper) any { return w.SaveUnstagedDiff(files) },
			want: []any{files, nil},
		},
		"unstaged-diff-applicable": {
			setup: func(w *gittest.StubWrapper, calls *[]any) {
				w.UnstagedDiffApplicableFunc = func() bool { return true }
			},
			call: func(w *gittest.StubWrapper) any { return w.UnstagedDiffApplicable() },
			want: []any{true},
		},
		"apply-unstaged-diff": {
			setup: func(w *gittest.StubWrapper, calls *[]any) {
				w.ApplyUnstagedDiffFunc = func(all bool) error { *calls = append(*calls, all); return nil }
			},
			call: func(w *gittest.StubWrapper) any { return w.ApplyUnstagedDiff(true) },
			want: []any{true, nil},
		},
		"store-stash": {
			setup: func(w *gittest.StubWrapper, calls *[]any) {
				w.StoreStashFunc = func() error { *calls = append(*calls, "store"); return nil }
			},
			call: func(w *gittest.StubWrapper) any { return w.StoreStash() },
			want: []any{"store", nil},
		},
		"drop-stash": {
			setup: func(w *gittest.StubWrapper, calls *[]any) {
				w.DropStashFunc = func() error { *calls = append(*calls, "drop"); return nil }
			},
			call: func(w *gittest.StubWrapper) any { return w.DropStash() },
			want: []any{"drop", nil},
		},
		"discard-unstaged-changes": {
			setup: func(w *gittest.StubWrapper, calls *[]any) {
				w.DiscardUnstagedChangesFunc = func(f []string) error { *calls = append(*calls, f); return nil }
			},
			call: func(w *gittest.StubWrapper) any { return w.DiscardUnstagedChanges(files) },
			want: []any{files, nil},
		},
		"discard-all-unstaged": {
			setup: func(w *gittest.StubWrapper, calls *[]any) {
				w.DiscardAllUnstagedChangesFunc = func() error { *calls = append(*calls, "discard-all"); return nil }
			},
			call: func(w *gittest.StubWrapper) any { return w.DiscardAllUnstagedChanges() },
			want: []any{"discard-all", nil},
		},
		"stage-files": {
			setup: func(w *gittest.StubWrapper, calls *[]any) {
				w.StageFilesFunc = func(f []string) error { *calls = append(*calls, f); return nil }
			},
			call: func(w *gittest.StubWrapper) any { return w.StageFiles(files) },
			want: []any{files, nil},
		},
		"hash-objects": {
			setup: func(w *gittest.StubWrapper, calls *[]any) {
				w.HashObjectsFunc = func(f []string) ([]string, error) {
					*calls = append(*calls, f)
					return []string{"h1", "h2"}, nil
				}
			},
			call: func(w *gittest.StubWrapper) any { h, _ := w.HashObjects(files); return h },
			want: []any{files, []string{"h1", "h2"}},
		},
		"state": {
			setup: func(w *gittest.StubWrapper, calls *[]any) {
				w.StateFunc = func() wrapper.State { return wrapper.State{Branch: "main"} }
			},
			call: func(w *gittest.StubWrapper) any { return w.State().Branch },
			want: []any{"main"},
		},
		"clone": {
			setup: func(w *gittest.StubWrapper, calls *[]any) {
				w.CloneFunc = func(a wrapper.CloneArgs) error { *calls = append(*calls, a.Url); return nil }
			},
			call: func(w *gittest.StubWrapper) any { return w.Clone(wrapper.CloneArgs{Url: "url"}) },
			want: []any{"url", nil},
		},
		"pull": {
			setup: func(w *gittest.StubWrapper, calls *[]any) {
				w.PullFunc = func(path string) error { *calls = append(*calls, path); return nil }
			},
			call: func(w *gittest.StubWrapper) any { return w.Pull("path") },
			want: []any{"path", nil},
		},
		"fetch": {
			setup: func(w *gittest.StubWrapper, calls *[]any) {
				w.FetchFunc = func(path string, ref string) error { *calls = append(*calls, path, ref); return nil }
			},
			call: func(w *gittest.StubWrapper) any { return w.Fetch("path", "ref") },
			want: []any{"path", "ref", nil},
		},
	} {
		t.Run(name, func(t *testing.T) {
			w := gittest.NewStubWrapper()
			calls := make([]any, 0, 1)
			tt.setup(w, &calls)

			calls = append(calls, tt.call(w))

			if !cmp.Equal(calls, tt.want) {
				t.Errorf("StubWrapper calls = %v, want %v", calls, tt.want)
			}
		})
	}
}
