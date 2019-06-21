package cmd

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

const (
	gitHooksDir string = ".git/hooks"
)

var createDirsFlag bool

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "This command add a hook directory to a repository",
	Long: `This command will try to build the following structure in repository:

├───.git
│   └───hooks
│       └───pre-commit // this executable will be added. Existed file with 
│                      // same name will be renamed to pre-commit.old
(lefthook add this dirs if you run command with -d option)
│
├───.lefthook            // directory for project level hooks
│   └───pre-commit     // directory with hooks executables
├───.lefthook-local      // directory for personal hooks add it in .gitignore
│   └───pre-commit
`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		addCmdExecutor(args, appFs)
	},
}

func init() {
	addCmd.SetUsageTemplate(`Usage:
    lefthook add [hooksGroup]
Example:
    lefthook add pre-commit
`)
	addCmd.PersistentFlags().BoolVarP(&createDirsFlag, "dirs", "d", false, "create directory for scripts")
	rootCmd.AddCommand(addCmd)
}

func addCmdExecutor(args []string, fs afero.Fs) {
	addHook(args[0], fs)
	if createDirsFlag {
		addProjectHookDir(args[0], fs)
		addLocalHookDir(args[0], fs)
	}
}

func addHook(hookName string, fs afero.Fs) {
	if !contains(availableHooks[:], hookName) {
		VerbosePrint("Skip adding, because that name unavailable: ", hookName)
		return
	}
	// TODO: text/template
	template := "#!/bin/bash\n" + autoInstall(hookName, fs) + "\n" +
		`# If can't find lefthook in global scope
# we suppose it in local npm package
if ! type lefthook >/dev/null 2>&1
then
  exec npx lefthook run ` + hookName + " $@\nelse\n  exec lefthook run " + hookName + " $@\nfi\n"

	pathToFile := filepath.Join(getGitHooksDir(), hookName)

	if yes, _ := afero.Exists(fs, pathToFile); yes {
		if isLefthookFile(pathToFile) {
			e := fs.Remove(pathToFile)
			check(e)
		} else {
			if yes, _ := afero.Exists(fs, pathToFile+".old"); yes {
				panic("Can`t rename " + hookName + " to " + hookName + ".old File already exists")
			}
			e := fs.Rename(pathToFile, pathToFile+".old")
			log.Println("Existed " + hookName + " hook renamed to " + hookName + ".old")
			check(e)
		}
	}

	err := afero.WriteFile(fs, pathToFile, []byte(template), defaultFilePermission)
	check(err)
	VerbosePrint("Added hook: ", pathToFile)
}

func autoInstall(hookName string, fs afero.Fs) string {
	if hookName == checkSumHook {
		return "# lefthook_version: " + configChecksum(fs) + "\n" +
			`if ! type lefthook >/dev/null
then
	exec npx lefthook install
else
	exec lefthook install
fi`
	}

	return ""
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

func isLefthookFile(pathFile string) bool {
	file, err := ioutil.ReadFile(pathFile)
	if err != nil {
		return false
	}
	return strings.Contains(string(file), "lefthook")
}
