package cmd

import (
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var keepConfiguration bool

// uninstallCmd represents the uninstall command
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Revert install command",
	Run: func(cmd *cobra.Command, args []string) {
		uninstallCmdExecutor(appFs)
	},
}

func init() {
	uninstallCmd.PersistentFlags().BoolVarP(&keepConfiguration, "keep-config", "k", false, "keep configuration files and source directories present")
	rootCmd.AddCommand(uninstallCmd)
}

func uninstallCmdExecutor(fs afero.Fs) {
	DeleteGitHooks(fs)
	revertOldGitHooks(fs)
	if !keepConfiguration {
		deleteSourceDirs(fs)
		deleteConfig(fs)
	}
}

func deleteConfig(fs afero.Fs) {
	err := fs.Remove(getConfigYamlPath())
	if err == nil {
		log.Println(getConfigYamlPath(), "removed")
	}

	results, err := afero.Glob(fs, getConfigLocalYamlPattern())
	if err != nil {
		log.Println("Error occurred while remove config file!:", err.Error())
	}
	for _, fileName := range results {
		err = fs.Remove(getConfigLocalYamlPattern())
		if err == nil {
			log.Println(fileName, "removed")
		} else {
			log.Println("Error occurred while remove config file!:", err.Error())
		}
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

// DeleteGitHooks read the config and remove all git hooks except
func DeleteGitHooks(fs afero.Fs) {
	hooksPath := getGitHooksPath()

	hooks, err := afero.ReadDir(fs, hooksPath)
	if err != nil {
		log.Println("⚠️ ", au.Bold(hooksPath), "directory does not exist, creating")
		if err := os.Mkdir(hooksPath, os.ModePerm); err != nil {
			log.Println(au.Brown("🚨 Failed to create"), au.Bold(hooksPath), au.Brown("directory"))
			log.Fatal(err)
		}
	}

	for _, file := range hooks {
		hookFile := filepath.Join(hooksPath, file.Name())
		if isLefthookFile(hookFile) || aggressive {
			err := fs.Remove(hookFile)
			if err == nil {
				VerbosePrint(hookFile, "removed")
			}
		}
	}
}

func revertOldGitHooks(fs afero.Fs) {
	hookGroups, _ := afero.ReadDir(fs, getSourceDir())

	if len(hookGroups) == 0 {
		return
	}

	hooksPath := getGitHooksPath()
	for _, file := range hookGroups {
		hookFilePath := filepath.Join(hooksPath, file.Name()+".old")

		err := fs.Rename(hookFilePath, filepath.Join(hooksPath, file.Name()))
		if err == nil {
			log.Println(hookFilePath, "renamed to", file.Name())
		}
	}
}
