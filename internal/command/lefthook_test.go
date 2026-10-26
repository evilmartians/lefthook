package command

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"

	"github.com/evilmartians/lefthook/v2/internal/templates"
	"github.com/evilmartians/lefthook/v2/tests/helpers/cmdtest"
	"github.com/evilmartians/lefthook/v2/tests/helpers/gittest"
	"github.com/evilmartians/lefthook/v2/tests/helpers/loggertest"
)

// spyFs records the operations that would leave the watched path missing or
// half-written for a moment: removing it, renaming it away or opening it for
// writing. Replacing it with a rename is fine.
type spyFs struct {
	afero.Fs
	path   string
	unsafe []string
}

func (s *spyFs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	if name == s.path && flag&(os.O_WRONLY|os.O_RDWR|os.O_CREATE|os.O_TRUNC) != 0 {
		s.unsafe = append(s.unsafe, "OpenFile")
	}
	return s.Fs.OpenFile(name, flag, perm)
}

func (s *spyFs) Create(name string) (afero.File, error) {
	if name == s.path {
		s.unsafe = append(s.unsafe, "Create")
	}
	return s.Fs.Create(name)
}

func (s *spyFs) Remove(name string) error {
	if name == s.path {
		s.unsafe = append(s.unsafe, "Remove")
	}
	return s.Fs.Remove(name)
}

func (s *spyFs) RemoveAll(path string) error {
	if path == s.path {
		s.unsafe = append(s.unsafe, "RemoveAll")
	}
	return s.Fs.RemoveAll(path)
}

func (s *spyFs) Rename(oldname, newname string) error {
	if oldname == s.path {
		s.unsafe = append(s.unsafe, "Rename")
	}
	return s.Fs.Rename(oldname, newname)
}

func newTestLefthook(t *testing.T, fs afero.Fs) *Lefthook {
	t.Helper()

	root, err := filepath.Abs("src")
	if err != nil {
		t.Fatal(err)
	}

	repo := gittest.NewRepositoryBuilder().
		Root(root).
		Fs(fs).
		Cmd(cmdtest.NewFakeCmd(t, nil)).
		Build()

	if err = fs.MkdirAll(repo.Paths.Hooks, hooksDirMode); err != nil {
		t.Fatal(err)
	}

	return &Lefthook{
		logger: loggertest.New(),
		fs:     fs,
		repo:   repo,
	}
}

func TestLefthook_addHook(t *testing.T) {
	const hook = "pre-commit"

	for name, tt := range map[string]struct {
		existing string
	}{
		"new-hook":              {},
		"replace-lefthook-hook": {existing: "#!/bin/sh\n# LEFTHOOK\nlefthook run pre-commit\n"},
		"replace-user-hook":     {existing: "#!/bin/sh\necho user hook\n"},
	} {
		t.Run(name, func(t *testing.T) {
			l := newTestLefthook(t, afero.NewMemMapFs())
			hookPath := filepath.Join(l.repo.Paths.Hooks, hook)
			spy := &spyFs{Fs: l.fs, path: hookPath}
			l.fs = spy

			if tt.existing != "" {
				if err := afero.WriteFile(spy.Fs, hookPath, []byte(tt.existing), hookFileMode); err != nil {
					t.Fatal(err)
				}
			}

			err := l.addHook(hook, templates.Args{})
			if err != nil {
				t.Fatalf("addHook() error = %v", err)
			}

			if len(spy.unsafe) != 0 {
				t.Errorf("addHook() touched %s in place: %v, want only a rename over it", hook, spy.unsafe)
			}

			content, err := afero.ReadFile(l.fs, hookPath)
			if err != nil {
				t.Fatalf("ReadFile(hook) error = %v", err)
			}
			if want := templates.Hook(hook, templates.Args{}); !cmp.Equal(content, want) {
				t.Errorf("hook content = %q, want %q", content, want)
			}

			info, err := l.fs.Stat(hookPath)
			if err != nil {
				t.Fatalf("Stat(hook) error = %v", err)
			}
			if info.Mode().Perm() != hookFileMode {
				t.Errorf("hook mode = %v, want %v", info.Mode().Perm(), os.FileMode(hookFileMode))
			}

			entries, err := afero.ReadDir(l.fs, l.repo.Paths.Hooks)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != 1 {
				t.Errorf("hooks dir has %d files, want only %s", len(entries), hook)
			}
		})
	}
}

func TestLefthook_cleanHook(t *testing.T) {
	const (
		hook        = "pre-commit"
		lefthookOld = "#!/bin/sh\n# LEFTHOOK\nlefthook run pre-commit\n"
		userHook    = "#!/bin/sh\necho user hook\n"
		otherOld    = "#!/bin/sh\necho other\n"
	)

	for name, tt := range map[string]struct {
		existing, existingOld string
		force                 bool
		wantHook, wantOld     string
		wantError             bool
	}{
		"no-hook":       {},
		"lefthook-hook": {existing: lefthookOld, wantHook: lefthookOld},
		"user-hook":     {existing: userHook, wantHook: userHook, wantOld: userHook},
		"old-exists": {
			existing: userHook, existingOld: otherOld,
			wantHook: userHook, wantOld: otherOld, wantError: true,
		},
		"old-exists-force": {
			existing: userHook, existingOld: otherOld, force: true,
			wantHook: userHook, wantOld: userHook,
		},
	} {
		t.Run(name, func(t *testing.T) {
			l := newTestLefthook(t, afero.NewMemMapFs())
			hookPath := filepath.Join(l.repo.Paths.Hooks, hook)
			spy := &spyFs{Fs: l.fs, path: hookPath}
			l.fs = spy

			for path, content := range map[string]string{
				hookPath:                  tt.existing,
				hookPath + oldHookPostfix: tt.existingOld,
			} {
				if content == "" {
					continue
				}
				if err := afero.WriteFile(spy.Fs, path, []byte(content), hookFileMode); err != nil {
					t.Fatal(err)
				}
			}

			err := l.cleanHook(hook, tt.force)
			if (err != nil) != tt.wantError {
				t.Fatalf("cleanHook() error = %v, wantError %v", err, tt.wantError)
			}

			if len(spy.unsafe) != 0 {
				t.Errorf("cleanHook() touched %s in place: %v, want it left for addHook", hook, spy.unsafe)
			}

			for path, want := range map[string]string{
				hookPath:                  tt.wantHook,
				hookPath + oldHookPostfix: tt.wantOld,
			} {
				content, err := afero.ReadFile(l.fs, path)
				switch {
				case want == "" && !errors.Is(err, os.ErrNotExist):
					t.Errorf("ReadFile(%s) error = %v, want %v", filepath.Base(path), err, os.ErrNotExist)
				case want != "" && err != nil:
					t.Errorf("ReadFile(%s) error = %v", filepath.Base(path), err)
				case want != "" && string(content) != want:
					t.Errorf("%s content = %q, want %q", filepath.Base(path), content, want)
				}
			}
		})
	}
}

func TestLefthook_writeFileAtomic(t *testing.T) {
	for name, tt := range map[string]struct {
		existing string
		mode     os.FileMode
	}{
		"new-file":     {mode: 0o644},
		"replace-file": {existing: "old content\n", mode: 0o755},
	} {
		t.Run(name, func(t *testing.T) {
			l := newTestLefthook(t, afero.NewMemMapFs())
			path := filepath.Join(l.repo.Paths.Hooks, "file")
			spy := &spyFs{Fs: l.fs, path: path}
			l.fs = spy

			if tt.existing != "" {
				if err := afero.WriteFile(spy.Fs, path, []byte(tt.existing), 0o600); err != nil {
					t.Fatal(err)
				}
			}

			want := []byte("new content\n")
			if err := l.writeFileAtomic(path, want, tt.mode); err != nil {
				t.Fatalf("writeFileAtomic() error = %v", err)
			}

			if len(spy.unsafe) != 0 {
				t.Errorf("writeFileAtomic() touched the file in place: %v, want only a rename over it", spy.unsafe)
			}

			content, err := afero.ReadFile(l.fs, path)
			if err != nil {
				t.Fatalf("ReadFile() error = %v", err)
			}
			if !cmp.Equal(content, want) {
				t.Errorf("content = %q, want %q", content, want)
			}

			info, err := l.fs.Stat(path)
			if err != nil {
				t.Fatalf("Stat() error = %v", err)
			}
			if info.Mode().Perm() != tt.mode {
				t.Errorf("mode = %v, want %v", info.Mode().Perm(), tt.mode)
			}

			entries, err := afero.ReadDir(l.fs, l.repo.Paths.Hooks)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != 1 {
				t.Errorf("dir has %d files, want 1 (no temporary file left)", len(entries))
			}
		})
	}
}
