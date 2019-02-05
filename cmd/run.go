package cmd

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/logrusorgru/aurora"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	hooksGroup     string
	executableName string
	sourcePath     string
	okList         []string
	failList       []string
)

const (
	runnerConfigKey     string = "runner"
	runnerArgsConfigKey string = "runner_args"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Execute group of hooks",
	Long: `Example:

hookah run pre-commit

It will run all executables in folder`,
	Run: func(cmd *cobra.Command, args []string) {
		err := RunCmdExecutor(args, appFs)
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}

// RunCmdExecutor run executables in hook groups
func RunCmdExecutor(args []string, fs afero.Fs) error {
	setHooksGroup(args[0])

	sourcePath := filepath.Join(getSourceDir(), getHooksGroup())
	executables, err := afero.ReadDir(fs, sourcePath)
	check(err)

	log.Println(aurora.Bold("Running " + getHooksGroup() + " hooks"))

	for _, executable := range executables {
		execute(sourcePath, executable)
	}

	sourcePath = filepath.Join(getLocalSourceDir(), getHooksGroup())
	executables, err = afero.ReadDir(fs, sourcePath)
	if err == nil {
		log.Println(aurora.Bold("Running local " + getHooksGroup() + " hooks"))

		for _, executable := range executables {
			execute(sourcePath, executable)
		}
	}

	printSummary()

	if len(failList) == 0 {
		return nil
	}
	return errors.New("Have failed script")
}

func execute(source string, executable os.FileInfo) {
	setExecutableName(executable.Name())

	log.Println(aurora.Bold("Execute " + getExecutableName()))

	pathToExecutable := filepath.Join(source, getExecutableName())

	command := exec.Command(pathToExecutable)

	if haveRunner() {
		command = exec.Command(
			getRunner(),
			getRunnerArgs(),
			pathToExecutable,
		)
	}

	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin

	command.Start()

	err := command.Wait()
	if err == nil {
		okList = append(okList, getExecutableName())
	} else {
		failList = append(failList, getExecutableName())
	}
}

func haveRunner() (out bool) {
	if runner := getRunner(); runner != "" {
		out = true
	}
	return
}

func getRunner() string {
	key := strings.Join([]string{getHooksGroup(), getExecutableName(), runnerConfigKey}, ".")
	return viper.GetString(key)
}

func getRunnerArgs() string {
	key := strings.Join([]string{getHooksGroup(), getExecutableName(), runnerArgsConfigKey}, ".")
	return viper.GetString(key)
}

func setHooksGroup(str string) {
	hooksGroup = str
}

func getHooksGroup() string {
	return hooksGroup
}

func setExecutableName(name string) {
	executableName = name
}

func getExecutableName() string {
	return executableName
}

func printSummary() {
	for _, fileName := range failList {
		log.Printf("[ %s ] %s\n", aurora.Red("FAIL"), fileName)
	}

	for _, fileName := range okList {
		log.Printf("[ %s ] %s\n", aurora.Green("OK"), fileName)
	}
}
