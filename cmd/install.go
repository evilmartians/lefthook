package cmd

import (
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

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

// AddConfigYaml write hookah.yml in root project directory
func AddConfigYaml(fs afero.Fs) {
	template := `source_dir: ".hookah"
source_dir_local: ".hookah-local"
`
	err := afero.WriteFile(fs, getConfigYamlPath(), []byte(template), defaultDirPermission)
	check(err)
}

// AddGitHooks write existed directories in source_dir as hooks in .git/hooks
func AddGitHooks(fs afero.Fs) {
	dirs, err := afero.ReadDir(fs, getSourceDir())
	if err != nil {
		return
	}

	for _, f := range dirs {
		if f.IsDir() {
			addHook(f.Name(), fs)
		}
	}
}

func getConfigYamlPath() string {
	return filepath.Join(getRootPath(), configFileName) + configExtension
}

func getConfigLocalYamlPath() string {
	return filepath.Join(getRootPath(), configLocalFileName) + configExtension
}
