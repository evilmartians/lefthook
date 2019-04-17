package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestInstallCmdExecutor(t *testing.T) {
	// if branch
	fs := afero.NewMemMapFs()

	InstallCmdExecutor([]string{}, fs)

	expectedFile := "lefthook.yml"

	_, err := fs.Stat(expectedFile)
	assert.Equal(t, os.IsNotExist(err), false, "lefthook.yml not exists after install command")

	// else branch
	fs = afero.NewMemMapFs()
	presetConfig(fs)

	InstallCmdExecutor([]string{}, fs)

	expectedFiles := []string{
		"commit-msg",
		"pre-commit",
	}

	files, err := afero.ReadDir(fs, filepath.Join(getRootPath(), ".git/hooks"))
	assert.NoError(t, err)

	actualFiles := []string{}
	for _, f := range files {
		actualFiles = append(actualFiles, f.Name())
	}
	assert.Equal(t, expectedFiles, actualFiles, "Expected files not exists")
}

func TestAddCmdExecutor(t *testing.T) {
	fs := afero.NewMemMapFs()
	presetConfig(fs)

	addCmdExecutor([]string{"pre-push"}, fs)

	expectedFiles := []string{
		"pre-push",
	}

	expectedDirs := []string{
		"commit-msg",
		"pre-commit",
	}

	files, _ := afero.ReadDir(fs, filepath.Join(getRootPath(), ".git/hooks"))
	actualFiles := []string{}
	for _, f := range files {
		actualFiles = append(actualFiles, f.Name())
	}

	dirs, _ := afero.ReadDir(fs, filepath.Join(getRootPath(), ".lefthook"))
	actualDirs := []string{}
	for _, f := range dirs {
		actualDirs = append(actualDirs, f.Name())
	}

	assert.Equal(t, expectedFiles, actualFiles, "Expected files not exists")
	assert.Equal(t, expectedDirs, actualDirs, "Expected dirs not exists")

	addCmdExecutor(expectedFiles, fs)

	expectedFiles = []string{
		"pre-push",
		"pre-push.old",
	}

	files, _ = afero.ReadDir(fs, filepath.Join(getRootPath(), ".git/hooks"))
	actualFiles = []string{}
	for _, f := range files {
		actualFiles = append(actualFiles, f.Name())
	}

	assert.Equal(t, expectedDirs, actualDirs, "Haven`t renamed file with .old extension")
}

func presetConfig(fs afero.Fs) {
	viper.SetDefault(configSourceDirKey, ".lefthook")

	AddConfigYaml(fs)

	fs.Mkdir(filepath.Join(getRootPath(), ".lefthook/commit-msg"), defaultFilePermission)
	fs.Mkdir(filepath.Join(getRootPath(), ".lefthook/pre-commit"), defaultFilePermission)

	fs.Mkdir(filepath.Join(getRootPath(), ".git/hooks"), defaultFilePermission)
}

func presetExecutable(hookName string, hookGroup string, exitCode string, fs afero.Fs) {
	template := "#!/bin/sh\nexit " + exitCode + "\n"
	pathToFile := filepath.Join(".lefthook", hookGroup, hookName)
	err := afero.WriteFile(fs, pathToFile, []byte(template), defaultFilePermission)
	check(err)
}
