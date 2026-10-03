package git_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/v2/internal/git"
	"github.com/evilmartians/lefthook/v2/internal/git/wrapper"
	"github.com/evilmartians/lefthook/v2/tests/helpers/gittest"
	"github.com/evilmartians/lefthook/v2/tests/helpers/loggertest"
)

func TestRepo_PartiallyStagedFiles(t *testing.T) {
	logger := loggertest.New()
	w := gittest.NewStubWrapper()
	statuses := []wrapper.FileStatus{
		{Index: ' ', Worktree: ' ', Path: "  "},
		{Index: ' ', Worktree: 'M', Path: " M"},
		{Index: 'M', Worktree: ' ', Path: "M "},
		{Index: 'M', Worktree: 'M', Path: "MM"},
		{Index: 'A', Worktree: 'M', Path: "AM"},
		{Index: '?', Worktree: '?', Path: "??"},
		{Index: 'M', Worktree: '?', Path: "M?"},
		{Index: '?', Worktree: 'M', Path: "?M"},
	}
	w.StatusShortFunc = func() ([]wrapper.FileStatus, error) { return statuses, nil }

	repo := git.NewRepo(
		afero.NewMemMapFs(),
		logger,
		w,
		&git.Paths{},
	)

	result, err := repo.PartiallyStagedFiles()
	want := []string{
		"MM",
		"AM",
	}
	if err != nil {
		t.Errorf("err = %v, want nil", err)
	}

	if !cmp.Equal(result, want) {
		t.Errorf("repo.PartiallyStagedFiles() = %v, want %v", result, want)
	}
}

func TestRepo_Changeset(t *testing.T) {
	for name, tt := range map[string]struct {
		StatusShort []wrapper.FileStatus
		HashObjects []string
		result      map[string]string
	}{
		"no-changes": {
			StatusShort: []wrapper.FileStatus{},
			result:      map[string]string{},
		},
		"modified": {
			StatusShort: []wrapper.FileStatus{
				{Path: "modified.txt", Index: ' ', Worktree: 'M'},
			},
			HashObjects: []string{"123456"},
			result: map[string]string{
				"modified.txt": "123456",
			},
		},
		"deleted": {
			StatusShort: []wrapper.FileStatus{
				{Path: "deleted.txt", Index: 'D', Worktree: ' '},
			},
			result: map[string]string{
				"deleted.txt": "deleted",
			},
		},
		"new": {
			StatusShort: []wrapper.FileStatus{
				{Path: "new.txt", Index: '?', Worktree: '?'},
			},
			HashObjects: []string{"654321"},
			result: map[string]string{
				"new.txt": "654321",
			},
		},
		"dir-new": {
			StatusShort: []wrapper.FileStatus{
				{Path: "new/", Index: '?', Worktree: '?'},
			},
			HashObjects: []string{},
			result: map[string]string{
				"new/": "directory",
			},
		},
		"mixed": {
			StatusShort: []wrapper.FileStatus{
				{Path: "modified.txt", Index: 'M', Worktree: ' '},
				{Path: "copied to", Index: 'C', Worktree: 'T'},
				{Path: "deleted.txt", Index: ' ', Worktree: 'D'},
				{Path: "new.txt", Index: '?', Worktree: '?'},
				{Path: "new-dir/", Index: '?', Worktree: '?'},
				{Path: "new-file", Index: 'R', Worktree: 'M'},
				{Path: "foo -> bar", Index: 'A', Worktree: ' '},
				{Path: "back\\slashes", Index: 'M', Worktree: 'M'},
			},
			HashObjects: []string{
				"123456",
				"c0c0c0",
				"654321",
				"758213",
				"fbfbfb",
				"bbbbbb",
			},
			result: map[string]string{
				"modified.txt": "123456",
				"copied to":    "c0c0c0",
				"deleted.txt":  "deleted",
				"new.txt":      "654321",
				"new-dir/":     "directory",
				"new-file":     "758213",
				"foo -> bar":   "fbfbfb",
				`back\slashes`: "bbbbbb",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			logger := loggertest.New()
			w := gittest.NewStubWrapper()
			w.StatusShortFunc = func() ([]wrapper.FileStatus, error) { return tt.StatusShort, nil }
			w.HashObjectsFunc = func([]string) ([]string, error) { return tt.HashObjects, nil }

			repository := git.NewRepo(
				afero.NewMemMapFs(),
				logger,
				w,
				&git.Paths{},
			)

			result, err := repository.Changeset()
			if err != nil {
				t.Errorf("err = %v, want nil", err)
			}

			if !cmp.Equal(result, tt.result) {
				t.Errorf("repo.Changeset() = %v, want %v", result, tt.result)
			}
		})
	}
}

func TestRepo_RestoreUnstagedChanges(t *testing.T) {
	errApply := errors.New("apply failed")
	errDrop := errors.New("drop failed")

	for name, tt := range map[string]struct {
		applyErr error
		dropErr  error
		dropped  bool
		err      error
	}{
		"restores-diff": {
			dropped: true,
		},
		"no-diff": {
			applyErr: wrapper.ErrNoUnstagedDiff,
		},
		"apply-fails": {
			applyErr: errApply,
			err:      errApply,
		},
		"drop-fails": {
			dropErr: errDrop,
			dropped: true,
			err:     errDrop,
		},
	} {
		t.Run(name, func(t *testing.T) {
			w := gittest.NewStubWrapper()
			w.ApplyUnstagedDiffFunc = func(all bool) error {
				if all {
					t.Errorf("wrapper.ApplyUnstagedDiff(%v), want false", all)
				}
				return tt.applyErr
			}
			dropped := false
			w.DropStashFunc = func() error {
				dropped = true
				return tt.dropErr
			}

			repo := git.NewRepo(
				afero.NewMemMapFs(),
				loggertest.New(),
				w,
				&git.Paths{},
			)

			err := repo.RestoreUnstagedChanges()

			if !errors.Is(err, tt.err) {
				t.Errorf("repo.RestoreUnstagedChanges() error = %v, want %v", err, tt.err)
			}

			if dropped != tt.dropped {
				t.Errorf("stash dropped = %v, want %v", dropped, tt.dropped)
			}
		})
	}
}

func TestRepo_RestoreAllUnstagedChanges(t *testing.T) {
	errApply := errors.New("apply failed")
	errDrop := errors.New("drop failed")

	for name, tt := range map[string]struct {
		applyErr error
		dropErr  error
		dropped  bool
		err      error
	}{
		"restores-diff": {
			dropped: true,
		},
		"no-diff": {
			applyErr: wrapper.ErrNoUnstagedDiff,
		},
		"apply-fails": {
			applyErr: errApply,
			err:      errApply,
		},
		"drop-fails": {
			dropErr: errDrop,
			dropped: true,
			err:     errDrop,
		},
	} {
		t.Run(name, func(t *testing.T) {
			w := gittest.NewStubWrapper()
			w.ApplyUnstagedDiffFunc = func(all bool) error {
				if !all {
					t.Errorf("wrapper.ApplyUnstagedDiff(%v), want true", all)
				}
				return tt.applyErr
			}
			dropped := false
			w.DropStashFunc = func() error {
				dropped = true
				return tt.dropErr
			}

			repo := git.NewRepo(
				afero.NewMemMapFs(),
				loggertest.New(),
				w,
				&git.Paths{},
			)

			err := repo.RestoreAllUnstagedChanges()

			if !errors.Is(err, tt.err) {
				t.Errorf("repo.RestoreAllUnstagedChanges() error = %v, want %v", err, tt.err)
			}

			if dropped != tt.dropped {
				t.Errorf("stash dropped = %v, want %v", dropped, tt.dropped)
			}
		})
	}
}
