package cmd

import (
	"log"
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

// uninstallCmd represents the uninstall command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Revert install command",
	Run: func(cmd *cobra.Command, args []string) {
		uninstallCmdExecutor(appFs)
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}

func uninstallCmdExecutor(fs afero.Fs) {
	deleteGitHooks(fs)
	revertOldGitHooks(fs)
	deleteConfig(fs)
	deleteSourceDirs(fs)
}

func deleteConfig(fs afero.Fs) {
	err := fs.Remove(getConfigYamlPath())
	if err == nil {
		log.Println(getConfigYamlPath(), "removed")
	}

	err = fs.Remove(getConfigLocalYamlPath())
	if err == nil {
		log.Println(getConfigLocalYamlPath(), "removed")
	}
}

func deleteSourceDirs(fs afero.Fs) {
	configExists, _ := afero.DirExists(fs, filepath.Join(getRootPath(), ".lefthook"))
	err := fs.RemoveAll(filepath.Join(getRootPath(), ".lefthook"))
	if err == nil && configExists {
		log.Println(filepath.Join(getRootPath(), ".lefthook"), "removed")
	}

	localConfigExists, _ := afero.DirExists(fs, filepath.Join(getRootPath(), ".lefthook-local"))
	err = fs.RemoveAll(filepath.Join(getRootPath(), ".lefthook-local"))
	if err == nil && localConfigExists {
		log.Println(filepath.Join(getRootPath(), ".lefthook-local"), "removed")
	}
}

func deleteGitHooks(fs afero.Fs) {
	hookGroups, _ := afero.ReadDir(fs, getSourceDir())

	if len(hookGroups) == 0 {
		return
	}

	hooksPath := filepath.Join(getRootPath(), ".git", "hooks")
	for _, file := range hookGroups {
		hookFilePath := filepath.Join(hooksPath, file.Name())

		err := fs.Remove(hookFilePath)
		if err == nil {
			log.Println(hookFilePath, "removed")
		}
	}

}

func revertOldGitHooks(fs afero.Fs) {
	hookGroups, _ := afero.ReadDir(fs, getSourceDir())

	if len(hookGroups) == 0 {
		return
	}

	hooksPath := filepath.Join(getRootPath(), ".git", "hooks")
	for _, file := range hookGroups {
		hookFilePath := filepath.Join(hooksPath, file.Name()+".old")

		err := fs.Rename(hookFilePath, filepath.Join(hooksPath, file.Name()))
		if err == nil {
			log.Println(hookFilePath, "renamed to", file.Name())
		}
	}
}
