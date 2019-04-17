package cmd

import (
	"log"
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var availableHooks = [...]string{
	"applypatch-msg",
	"pre-applypatch",
	"post-applypatch",
	"pre-commit",
	"prepare-commit-msg",
	"commit-msg",
	"post-commit",
	"pre-rebase",
	"post-checkout",
	"post-merge",
	"pre-push",
	"pre-receive",
	"update",
	"post-receive",
	"post-update",
	"pre-auto-gc",
	"post-rewrite",
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Write basic configuration file in your project repository. Or initialize existed config",
	Run: func(cmd *cobra.Command, args []string) {
		InstallCmdExecutor(args, appFs)
	},
}

var appFs = afero.NewOsFs()

func init() {
	rootCmd.AddCommand(installCmd)
}

// InstallCmdExecutor execute basic configuration
func InstallCmdExecutor(args []string, fs afero.Fs) {
	if yes, _ := afero.Exists(fs, getConfigYamlPath()); yes {
		AddGitHooks(fs)
	} else {
		AddConfigYaml(fs)
	}
}

// AddConfigYaml write lefthook.yml in root project directory
func AddConfigYaml(fs afero.Fs) {
	template := ""
	err := afero.WriteFile(fs, getConfigYamlPath(), []byte(template), defaultDirPermission)
	check(err)
	log.Println("Added config: ", getConfigYamlPath())
}

// AddGitHooks write existed directories in source_dir as hooks in .git/hooks
func AddGitHooks(fs afero.Fs) {
	// add directory hooks
	dirs, err := afero.ReadDir(fs, getSourceDir())
	if err == nil {
		for _, f := range dirs {
			if f.IsDir() {
				addHook(f.Name(), fs)
			}
		}
	}

	// add config hooks
	var dirsHooks []string
	for _, dir := range dirs {
		dirsHooks = append(dirsHooks, dir.Name())
	}

	var configHooks []string
	for _, key := range availableHooks {
		if viper.Get(key) != nil {
			configHooks = append(configHooks, key)
		}
	}

	for _, key := range configHooks {
		if !contains(dirsHooks, key) {
			addHook(key, fs)
		}
	}
}

func getConfigYamlPath() string {
	return filepath.Join(getRootPath(), configFileName) + configExtension
}

func getConfigLocalYamlPath() string {
	return filepath.Join(getRootPath(), configLocalFileName) + configExtension
}

func contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}
