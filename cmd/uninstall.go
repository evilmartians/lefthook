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
	if err != nil {
		log.Println(err)
	} else {
		log.Println(getConfigYamlPath(), "removed")
	}

	err = fs.Remove(getConfigLocalYamlPath())
	if err != nil {
		log.Println(err)
	} else {
		log.Println(getConfigLocalYamlPath(), "removed")
	}
}

func deleteSourceDirs(fs afero.Fs) {
	err := fs.RemoveAll(filepath.Join(getRootPath(), ".hookah"))
	if err != nil {
		log.Println(err)
	} else {
		log.Println(filepath.Join(getRootPath(), ".hookah"), "removed")
	}

	err = fs.RemoveAll(filepath.Join(getRootPath(), ".hookah-local"))
	if err != nil {
		log.Println(err)
	} else {
		log.Println(filepath.Join(getRootPath(), ".hookah-local"), "removed")
	}
}

func deleteGitHooks(fs afero.Fs) {
	hookGroups, _ := afero.ReadDir(fs, getSourceDir())

	if len(hookGroups) == 0 {
		log.Println("Hooks not found. Delete skipped")
		return
	}

	hooksPath := filepath.Join(getRootPath(), ".git", "hooks")
	for _, file := range hookGroups {
		hookFilePath := filepath.Join(hooksPath, file.Name())

		err := fs.Remove(hookFilePath)
		if err != nil {
			log.Println(err)
		} else {
			log.Println(hookFilePath, "removed")
		}
	}

}

func revertOldGitHooks(fs afero.Fs) {
	hookGroups, _ := afero.ReadDir(fs, getSourceDir())

	if len(hookGroups) == 0 {
		log.Println("Hooks not found. Renaming skipped")
		return
	}

	hooksPath := filepath.Join(getRootPath(), ".git", "hooks")
	for _, file := range hookGroups {
		hookFilePath := filepath.Join(hooksPath, file.Name()+".old")

		err := fs.Rename(hookFilePath, filepath.Join(hooksPath, file.Name()))
		if err != nil {
			log.Println(err)
		} else {
			log.Println(hookFilePath, "renamed to", file.Name())
		}
	}
}
