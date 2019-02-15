package cmd

import (
	"log"
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

const (
	gitHooksDir string = ".git/hooks"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "This command add a hook directory to a repository",
	Long: `Example:
hookah add pre-commit

This command will try to build the foolowing structure in repository:
├───.git
│   └───hooks
│       └───pre-commit // this executable will be added. Existed file with
│                      // same name will be renamed to pre-commit.old
...
│
├───.hookah            // directory for project level hooks
│   └───pre-commit     // directory with hooks executables
├───.hookah-local      // directory for personal hooks add it in .gitignore
│   └───pre-commit
`,
	Run: func(cmd *cobra.Command, args []string) {
		addCmdExecutor(args, appFs)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}

func addCmdExecutor(args []string, fs afero.Fs) {
	addHook(args[0], fs)
	addProjectHookDir(args[0], fs)
	addLocalHookDir(args[0], fs)
}

func addHook(hookName string, fs afero.Fs) {
	// TODO: text/template
	template := `#!/bin/bash
# If can't find hookah in global scope
# we suppose it in local npm package
if ! type hookah 2>/dev/null
then
  exec npx hookah run ` + hookName + "\nelse\n  exec hookah run " + hookName + "\nfi"

	pathToFile := filepath.Join(getGitHooksDir(), hookName)

	if yes, _ := afero.Exists(fs, pathToFile); yes {
		if yes, _ := afero.Exists(fs, pathToFile+".old"); yes {
			panic("Can`t rename " + hookName + " to " + hookName + ".old File already exists")
		}
		e := fs.Rename(pathToFile, pathToFile+".old")
		log.Println("Existed " + hookName + " hook renamed to " + hookName + ".old")
		check(e)
	}

	err := afero.WriteFile(fs, pathToFile, []byte(template), defaultFilePermission)
	check(err)
	log.Println("Added hook: ", pathToFile)
}

func addProjectHookDir(hookName string, fs afero.Fs) {
	err := fs.MkdirAll(filepath.Join(getSourceDir(), hookName), defaultFilePermission)
	check(err)
}

func addLocalHookDir(hookName string, fs afero.Fs) {
	err := fs.MkdirAll(filepath.Join(getLocalSourceDir(), hookName), defaultFilePermission)
	check(err)
}

func getGitHooksDir() string {
	return filepath.Join(getRootPath(), gitHooksDir)
}
