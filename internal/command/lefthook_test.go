package command

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/evilmartians/lefthook/v2/internal/config"
	"github.com/evilmartians/lefthook/v2/internal/templates"
	"github.com/evilmartians/lefthook/v2/tests/helpers/cmdtest"
	"github.com/evilmartians/lefthook/v2/tests/helpers/gittest"
	"github.com/evilmartians/lefthook/v2/tests/helpers/loggertest"
)

const (
	symlinkFileMode fs.FileMode = 0o777
	maxSymlinkDepth             = 16
)

type symlinkFileInfo struct {
	name string
}

func (s symlinkFileInfo) Name() string       { return s.name }
func (s symlinkFileInfo) Size() int64        { return 0 }
func (s symlinkFileInfo) Mode() os.FileMode  { return os.ModeSymlink | symlinkFileMode }
func (s symlinkFileInfo) ModTime() time.Time { return time.Time{} }
func (s symlinkFileInfo) IsDir() bool        { return false }
func (s symlinkFileInfo) Sys() any           { return nil }

// symlinkMemFs wraps afero.MemMapFs with POSIX-like symlink behavior so
// symlink tests run consistently without host OS symlink privileges.
type symlinkMemFs struct {
	afero.Fs
	symlinks map[string]string
}

func newSymlinkMemFs() *symlinkMemFs {
	return &symlinkMemFs{
		Fs:       afero.NewMemMapFs(),
		symlinks: make(map[string]string),
	}
}

func (s *symlinkMemFs) addSymlink(linkPath, targetPath string) {
	s.symlinks[filepath.Clean(linkPath)] = targetPath
}

func (s *symlinkMemFs) resolve(name string) string {
	clean := filepath.Clean(name)
	for range maxSymlinkDepth {
		target, ok := s.symlinks[clean]
		if !ok {
			return clean
		}
		if !filepath.IsAbs(target) {
			target = filepath.Join(filepath.Dir(clean), target)
		}
		clean = filepath.Clean(target)
	}
	return clean
}

func (s *symlinkMemFs) isSymlink(name string) bool {
	_, ok := s.symlinks[filepath.Clean(name)]
	return ok
}

func (s *symlinkMemFs) LstatIfPossible(name string) (os.FileInfo, bool, error) {
	clean := filepath.Clean(name)
	if _, ok := s.symlinks[clean]; ok {
		return symlinkFileInfo{name: filepath.Base(clean)}, true, nil
	}
	info, err := s.Fs.Stat(clean)
	return info, true, err
}

func (s *symlinkMemFs) Stat(name string) (os.FileInfo, error) {
	return s.Fs.Stat(s.resolve(name))
}

type symlinkDirFile struct {
	afero.File
	extra []os.FileInfo
}

func (d *symlinkDirFile) Readdir(count int) ([]os.FileInfo, error) {
	entries, err := d.File.Readdir(count)
	if err != nil {
		return entries, err
	}
	return append(entries, d.extra...), nil
}

func (s *symlinkMemFs) Open(name string) (afero.File, error) {
	resolved := s.resolve(name)
	file, err := s.Fs.Open(resolved)
	if err != nil {
		return nil, err
	}
	info, statErr := file.Stat()
	if statErr != nil || !info.IsDir() {
		return file, nil
	}

	var extra []os.FileInfo
	dirClean := filepath.Clean(resolved)
	for linkPath := range s.symlinks {
		if filepath.Dir(linkPath) == dirClean {
			extra = append(extra, symlinkFileInfo{name: filepath.Base(linkPath)})
		}
	}
	return &symlinkDirFile{File: file, extra: extra}, nil
}

func (s *symlinkMemFs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	return s.Fs.OpenFile(s.resolve(name), flag, perm)
}

func (s *symlinkMemFs) Remove(name string) error {
	clean := filepath.Clean(name)
	if _, ok := s.symlinks[clean]; ok {
		delete(s.symlinks, clean)
		return nil
	}
	return s.Fs.Remove(clean)
}

func (s *symlinkMemFs) Rename(oldname, newname string) error {
	oldClean := filepath.Clean(oldname)
	newClean := filepath.Clean(newname)
	if target, ok := s.symlinks[oldClean]; ok {
		delete(s.symlinks, oldClean)
		if rmErr := s.Fs.Remove(newClean); rmErr != nil && !errors.Is(rmErr, os.ErrNotExist) {
			return rmErr
		}
		s.symlinks[newClean] = target
		return nil
	}
	delete(s.symlinks, newClean)
	return s.Fs.Rename(oldClean, newClean)
}

func newSymlinkTestLefthook(t *testing.T, root string, mem *symlinkMemFs) *Lefthook {
	t.Helper()

	gitCmds := []cmdtest.Out{
		{Command: "git config --local core.hooksPath"},
		{Command: "git config --global core.hooksPath"},
	}
	repo := gittest.NewRepositoryBuilder().
		Root(root).
		Fs(mem).
		Cmd(cmdtest.NewFakeCmd(t, gitCmds)).
		Build()

	return &Lefthook{
		logger: loggertest.New(),
		fs:     mem,
		repo:   repo,
	}
}

func TestInstallDoesNotWriteThroughSymlink(t *testing.T) {
	root, err := filepath.Abs("src")
	require.NoError(t, err)

	sharedTarget := filepath.Join(filepath.Dir(root), "shared", "hook")
	sharedContent := "#!/bin/sh\necho shared hook\n"
	danglingTarget := filepath.Join(filepath.Dir(root), "elsewhere", "commit-msg")

	hookPath := func(hook string) string {
		return filepath.Join(gittest.GitPath(root), "hooks", hook)
	}

	t.Run("ghost hook live symlink is backed up and target stays untouched", func(t *testing.T) {
		mem := newSymlinkMemFs()
		lh := newSymlinkTestLefthook(t, root, mem)

		require.NoError(t, mem.MkdirAll(filepath.Dir(sharedTarget), hooksDirMode))
		require.NoError(t, afero.WriteFile(mem.Fs, sharedTarget, []byte(sharedContent), hookFileMode))
		require.NoError(t, mem.MkdirAll(lh.repo.Paths.Hooks, hooksDirMode))
		require.NoError(t, afero.WriteFile(mem, filepath.Join(root, "lefthook.yml"), []byte("min_version: 2.0.0\n"), checksumFileMode))

		ghostPath := hookPath(config.GhostHookName)
		mem.addSymlink(ghostPath, sharedTarget)

		require.NoError(t, lh.Install(t.Context(), InstallArgs{}, nil))

		assert.False(t, mem.isSymlink(ghostPath), "ghost hook should be a regular file, not a symlink")
		assert.True(t, mem.isSymlink(ghostPath+oldHookPostfix), "existing symlink should be backed up as .old")

		gotTarget, readErr := afero.ReadFile(mem.Fs, sharedTarget)
		require.NoError(t, readErr)
		assert.Equal(t, sharedContent, string(gotTarget))

		gotGhost, readErr := afero.ReadFile(mem, ghostPath)
		require.NoError(t, readErr)
		assert.Contains(t, string(gotGhost), hookContentFingerprint)
	})

	t.Run("dangling symlink for configured hook is backed up and target is not created", func(t *testing.T) {
		mem := newSymlinkMemFs()
		lh := newSymlinkTestLefthook(t, root, mem)

		require.NoError(t, mem.MkdirAll(filepath.Dir(danglingTarget), hooksDirMode))
		require.NoError(t, mem.MkdirAll(lh.repo.Paths.Hooks, hooksDirMode))
		require.NoError(t, afero.WriteFile(mem, filepath.Join(root, "lefthook.yml"), []byte("commit-msg:\n  jobs:\n    - run: echo hi\n"), checksumFileMode))

		commitMsgPath := hookPath("commit-msg")
		mem.addSymlink(commitMsgPath, danglingTarget)

		require.NoError(t, lh.Install(t.Context(), InstallArgs{}, nil))

		assert.False(t, mem.isSymlink(commitMsgPath), "installed hook should replace symlink with regular file")
		assert.True(t, mem.isSymlink(commitMsgPath+oldHookPostfix), "dangling symlink should be backed up as .old")

		targetCreated, existsErr := afero.Exists(mem.Fs, danglingTarget)
		require.NoError(t, existsErr)
		assert.False(t, targetCreated, "dangling symlink target must not be created")

		gotHook, readErr := afero.ReadFile(mem, commitMsgPath)
		require.NoError(t, readErr)
		assert.Contains(t, string(gotHook), hookContentFingerprint)

		require.NoError(t, lh.Uninstall(t.Context(), UninstallArgs{}))
		assert.True(t, mem.isSymlink(commitMsgPath), "uninstall should restore backed-up dangling symlink")
		assert.False(t, mem.isSymlink(commitMsgPath+oldHookPostfix))
	})

	t.Run("dangling symlink backup fails without force when .old exists", func(t *testing.T) {
		mem := newSymlinkMemFs()
		lh := newSymlinkTestLefthook(t, root, mem)

		require.NoError(t, mem.MkdirAll(filepath.Dir(danglingTarget), hooksDirMode))
		require.NoError(t, mem.MkdirAll(lh.repo.Paths.Hooks, hooksDirMode))
		require.NoError(t, afero.WriteFile(mem, filepath.Join(root, "lefthook.yml"), []byte("commit-msg:\n  jobs:\n    - run: echo hi\n"), checksumFileMode))

		commitMsgPath := hookPath("commit-msg")
		mem.addSymlink(commitMsgPath, danglingTarget)
		mem.addSymlink(commitMsgPath+oldHookPostfix, filepath.Join(filepath.Dir(root), "elsewhere", "old-dangling"))

		installErr := lh.Install(t.Context(), InstallArgs{}, nil)
		require.Error(t, installErr)
		assert.True(t, mem.isSymlink(commitMsgPath))
		assert.True(t, mem.isSymlink(commitMsgPath+oldHookPostfix))

		targetCreated, existsErr := afero.Exists(mem.Fs, danglingTarget)
		require.NoError(t, existsErr)
		assert.False(t, targetCreated)
	})

	t.Run("dangling symlink backup overwrites .old with force", func(t *testing.T) {
		mem := newSymlinkMemFs()
		lh := newSymlinkTestLefthook(t, root, mem)

		require.NoError(t, mem.MkdirAll(filepath.Dir(danglingTarget), hooksDirMode))
		require.NoError(t, mem.MkdirAll(lh.repo.Paths.Hooks, hooksDirMode))
		require.NoError(t, afero.WriteFile(mem, filepath.Join(root, "lefthook.yml"), []byte("commit-msg:\n  jobs:\n    - run: echo hi\n"), checksumFileMode))

		commitMsgPath := hookPath("commit-msg")
		mem.addSymlink(commitMsgPath, danglingTarget)
		mem.addSymlink(commitMsgPath+oldHookPostfix, filepath.Join(filepath.Dir(root), "elsewhere", "old-dangling"))

		require.NoError(t, lh.Install(t.Context(), InstallArgs{Force: true}, nil))
		assert.False(t, mem.isSymlink(commitMsgPath))
		assert.True(t, mem.isSymlink(commitMsgPath+oldHookPostfix))

		targetCreated, existsErr := afero.Exists(mem.Fs, danglingTarget)
		require.NoError(t, existsErr)
		assert.False(t, targetCreated)
	})

	t.Run("addHook alone replaces symlink without writing through target", func(t *testing.T) {
		mem := newSymlinkMemFs()
		lh := newSymlinkTestLefthook(t, root, mem)

		require.NoError(t, mem.MkdirAll(filepath.Dir(sharedTarget), hooksDirMode))
		require.NoError(t, afero.WriteFile(mem.Fs, sharedTarget, []byte(sharedContent), hookFileMode))
		require.NoError(t, mem.MkdirAll(lh.repo.Paths.Hooks, hooksDirMode))

		preCommitPath := hookPath("pre-commit")
		mem.addSymlink(preCommitPath, sharedTarget)

		require.NoError(t, lh.addHook("pre-commit", templates.Args{}))

		assert.False(t, mem.isSymlink(preCommitPath), "addHook must replace symlink instead of writing through it")
		gotTarget, readErr := afero.ReadFile(mem.Fs, sharedTarget)
		require.NoError(t, readErr)
		assert.Equal(t, sharedContent, string(gotTarget))
	})
}

func TestAddHookLeavesNoTempFilesOnOsFs(t *testing.T) {
	root := t.TempDir()
	hooksDir := filepath.Join(root, ".git", "hooks")
	fs := afero.NewOsFs()
	require.NoError(t, fs.MkdirAll(hooksDir, hooksDirMode))

	repo := gittest.NewRepositoryBuilder().Root(root).Fs(fs).Build()
	lh := &Lefthook{
		logger: loggertest.New(),
		fs:     fs,
		repo:   repo,
	}

	require.NoError(t, lh.addHook("pre-commit", templates.Args{}))

	entries, err := afero.ReadDir(fs, hooksDir)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "pre-commit", entries[0].Name())

	content, err := afero.ReadFile(fs, filepath.Join(hooksDir, "pre-commit"))
	require.NoError(t, err)
	assert.Contains(t, string(content), hookContentFingerprint)
}
