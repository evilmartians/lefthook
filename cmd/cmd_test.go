package cmd

import (
	"bytes"
	"log"
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
		"prepare-commit-msg",
	}

	files, err := afero.ReadDir(fs, getGitHooksPath())
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

	files, _ := afero.ReadDir(fs, getGitHooksPath())
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

	files, _ = afero.ReadDir(fs, getGitHooksPath())
	actualFiles = []string{}
	for _, f := range files {
		actualFiles = append(actualFiles, f.Name())
	}

	assert.Equal(t, expectedFiles, actualFiles, "Haven`t renamed file with .old extension")
}

func TestRunCmdExecutor(t *testing.T) {
	fs := afero.NewMemMapFs()

	var yamlExample = []byte(`
pre-commit:
  commands:
    test:
      run: echo 'test passed'
`)
	viper.SetConfigType("yaml")
	_ = viper.ReadConfig(bytes.NewBuffer(yamlExample))

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	err := RunCmdExecutor([]string{"pre-commit"}, fs)
	assert.NoError(t, err)

	assert.Contains(t, buf.String(), "test passed")
}

func TestExtendsProperty(t *testing.T) {
	var yamlExample = "extends:"
	var yamlExampleArray = []byte(yamlExample + "\n- c1.yml\n- c2.yml")
	var yamlExampleString = []byte(yamlExample + " 'c3.yml'")

	var expectedPathsArray = []string{"c1.yml", "c2.yml"}
	var expectedPathsString = []string{"c3.yml"}
	viper.SetConfigType("yaml")

	_ = viper.ReadConfig(bytes.NewBuffer([]byte("")))
	assert.False(t, isConfigExtends(), "Should not detect extends property")

	_ = viper.ReadConfig(bytes.NewBuffer(yamlExampleString))
	paths := getExtendsPath()

	assert.True(t, isConfigExtends(), "Should detect extends property")
	assert.Equal(t, paths, expectedPathsString, "Extends path does not match for string value")

	_ = viper.ReadConfig(bytes.NewBuffer(yamlExampleArray))
	paths = getExtendsPath()
	assert.Equal(t, paths, expectedPathsArray, "Extends path does not match for array value")
}

func presetConfig(fs afero.Fs) {
	viper.SetDefault(configSourceDirKey, ".lefthook")

	AddConfigYaml(fs)

	_ = fs.Mkdir(filepath.Join(getRootPath(), ".lefthook/commit-msg"), defaultFilePermission)
	_ = fs.Mkdir(filepath.Join(getRootPath(), ".lefthook/pre-commit"), defaultFilePermission)

	setGitHooksPath(".git/hooks")
	_ = fs.MkdirAll(getGitHooksPath(), defaultFilePermission)
}
