package cmd

import (
	"errors"
	"hookah/context"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gobwas/glob"
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
	runnerConfigKey     string      = "runner"
	runnerArgsConfigKey string      = "runner_args"
	scriptsConfigKey    string      = "scripts"
	commandsConfigKey   string      = "commands"
	includeConfigKey    string      = "include"
	excludeConfigKey    string      = "exclude"
	globConfigKey       string      = "glob"
	skipConfigKey       string      = "skip"
	skipEmptyConfigKey  string      = "skip_empty"
	filesConfigKey      string      = "files"
	subFiles            string      = "{files}"
	execMode            os.FileMode = 0751
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
	if err == nil && len(executables) > 0 {
		log.Println(aurora.Cyan("RUNNING HOOKS GROUP:"), aurora.Bold(getHooksGroup()))

		for _, executable := range executables {
			executeScript(sourcePath, executable)
		}
	}

	sourcePath = filepath.Join(getLocalSourceDir(), getHooksGroup())
	executables, err = afero.ReadDir(fs, sourcePath)
	if err == nil && len(executables) > 0 {
		log.Println(aurora.Cyan("RUNNING LOCAL HOOKS GROUP:"), aurora.Bold(getHooksGroup()))

		for _, executable := range executables {
			executeScript(sourcePath, executable)
		}
	}

	commands := getCommands()
	if len(commands) != 0 {
		log.Println(aurora.Cyan("RUNNING COMMANDS HOOKS GROUP:"), aurora.Bold(getHooksGroup()))

		for commandName := range commands {
			executeCommand(commandName)
		}
	}

	printSummary()

	if len(failList) == 0 {
		return nil
	}
	return errors.New("Have failed script")
}

func executeCommand(commandName string) {
	setExecutableName(commandName)

	var files []string
	switch getCommandFiles() {
	case "git_staged":
		files, _ = context.StagedFiles()
	case "all":
		files, _ = context.AllFiles()
	case "none":
		files = []string{}
	default:
		files = []string{}
	}
	files = FilterGlob(files, getCommandGlobRegexp())
	files = FilterInclude(files, getCommandIncludeRegexp()) // NOTE: confusing option, suppose delete it
	files = FilterExclude(files, getCommandExcludeRegexp())

	runner := strings.Replace(getRunner(commandsConfigKey), subFiles, strings.Join(files, " "), -1)
	runnerArg := strings.Split(runner, " ")

	command := exec.Command(runnerArg[0], runnerArg[1:]...)

	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin

	log.Println(aurora.Cyan("  EXECUTE >"), aurora.Bold(getExecutableName()))

	if isSkipCommmand() {
		log.Println(aurora.Brown("(SKIP BY SETTINGS)"))
		return
	}
	if len(files) < 1 && isSkipEmptyCommmand() {
		log.Println(aurora.Brown("(SKIP. NO FILES FOR INSPECTING)"))
		return
	}

	err := command.Start()
	if err != nil {
		log.Println(err)
	}

	err = command.Wait()
	if err == nil {
		okList = append(okList, getExecutableName())
	} else {
		failList = append(failList, getExecutableName())
	}
}

func executeScript(source string, executable os.FileInfo) {
	setExecutableName(executable.Name())

	log.Println(aurora.Cyan("  EXECUTE >"), aurora.Bold(getExecutableName()))

	if isSkipScript() {
		log.Println(aurora.Brown("(SKIP BY SETTINGS)"))
		return
	}

	pathToExecutable := filepath.Join(source, getExecutableName())

	if err := isExecutable(executable); err != nil {
		makeExecutable(pathToExecutable)
	}

	command := exec.Command(pathToExecutable)

	if haveRunner(scriptsConfigKey) {
		runnerArg := strings.Split(getRunner(scriptsConfigKey), " ")
		runnerArg = append(runnerArg, pathToExecutable)

		command = exec.Command(runnerArg[0], runnerArg[1:]...)
	}

	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin

	err := command.Start()
	if os.IsPermission(err) {
		log.Println(aurora.Brown("(SKIP NOT EXECUTABLE FILE)"))
		return
	}
	if err != nil {
		log.Println(err)
	}

	err = command.Wait()
	if err == nil {
		okList = append(okList, getExecutableName())
	} else {
		failList = append(failList, getExecutableName())
	}
}

func haveRunner(source string) (out bool) {
	if runner := getRunner(source); runner != "" {
		out = true
	}
	return
}

func getRunner(source string) string {
	key := strings.Join([]string{getHooksGroup(), source, getExecutableName(), runnerConfigKey}, ".")
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
	if len(okList) == 0 && len(failList) == 0 {
		log.Println(aurora.Cyan("\nSUMMARY:"), aurora.Brown("(SKIP EMPTY)"))
	} else {
		log.Println(aurora.Cyan("\nSUMMARY:"))
	}

	for _, fileName := range okList {
		log.Printf("[  %s  ] %s\n", aurora.Green("OK"), fileName)
	}

	for _, fileName := range failList {
		log.Printf("[ %s ] %s\n", aurora.Red("FAIL"), fileName)
	}
}

func isSkipScript() bool {
	key := strings.Join([]string{getHooksGroup(), scriptsConfigKey, getExecutableName(), skipConfigKey}, ".")
	return viper.GetBool(key)
}

func isSkipCommmand() bool {
	key := strings.Join([]string{getHooksGroup(), commandsConfigKey, getExecutableName(), skipConfigKey}, ".")
	return viper.GetBool(key)
}

// NOTE: confusing option, suppose it unnesecary and should be deleted.
func isSkipEmptyCommmand() bool {
	key := strings.Join([]string{getHooksGroup(), commandsConfigKey, getExecutableName(), skipEmptyConfigKey}, ".")
	if viper.IsSet(key) {
		return viper.GetBool(key)
	}

	key = strings.Join([]string{getHooksGroup(), skipEmptyConfigKey}, ".")
	if viper.IsSet(key) {
		return viper.GetBool(key)
	}

	return true
}

func getCommands() map[string]interface{} {
	key := strings.Join([]string{getHooksGroup(), commandsConfigKey}, ".")
	return viper.GetStringMap(key)
}

func getCommandIncludeRegexp() string {
	key := strings.Join([]string{getHooksGroup(), commandsConfigKey, getExecutableName(), includeConfigKey}, ".")
	return viper.GetString(key)
}

func getCommandExcludeRegexp() string {
	key := strings.Join([]string{getHooksGroup(), commandsConfigKey, getExecutableName(), excludeConfigKey}, ".")
	return viper.GetString(key)
}

func getCommandGlobRegexp() string {
	key := strings.Join([]string{getHooksGroup(), commandsConfigKey, getExecutableName(), globConfigKey}, ".")
	if viper.GetString(key) != "" {
		return viper.GetString(key)
	}

	key = strings.Join([]string{getHooksGroup(), globConfigKey}, ".")
	return viper.GetString(key)
}

func getCommandFiles() string {
	key := strings.Join([]string{getHooksGroup(), commandsConfigKey, getExecutableName(), filesConfigKey}, ".")
	if viper.GetString(key) != "" {
		return viper.GetString(key)
	}

	key = strings.Join([]string{getHooksGroup(), filesConfigKey}, ".")
	if viper.GetString(key) != "" {
		return viper.GetString(key)
	}

	return "git_staged"
}

func FilterGlob(vs []string, matcher string) []string {
	if matcher == "" {
		return vs
	}

	g := glob.MustCompile(matcher)

	vsf := make([]string, 0)
	for _, v := range vs {
		if res := g.Match(v); res {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func FilterInclude(vs []string, matcher string) []string {
	if matcher == "" {
		return vs
	}

	vsf := make([]string, 0)
	for _, v := range vs {
		if res, _ := regexp.MatchString(matcher, v); res {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func FilterExclude(vs []string, matcher string) []string {
	if matcher == "" {
		return vs
	}

	vsf := make([]string, 0)
	for _, v := range vs {
		if res, _ := regexp.MatchString(matcher, v); !res {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func isExecutable(executable os.FileInfo) error {
	mode := executable.Mode()

	if !mode.IsRegular() {
		return errors.New("ErrPermission")
	}
	if (mode & 0111) == 0 {
		return errors.New("ErrPermission")
	}
	return nil
}

func makeExecutable(path string) {
	if err := os.Chmod(path, execMode); err != nil {
		log.Fatal(err)
	}
}
