package lefthook

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/spf13/afero"
	"gopkg.in/alessio/shellescape.v1"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/git"
	"github.com/evilmartians/lefthook/internal/log"
	"github.com/evilmartians/lefthook/internal/version"
)

var (
	spinner = NewSpinner()

	mutex        sync.Mutex
	okList       []string
	failList     []string
	isPipeBroken bool

	successSprint outputDisablerFn
)

const (
	execMode os.FileMode = 0751
	execMask os.FileMode = 0111
)

func Run(opts *Options, hookName string, gitArgs []string) error {
	if os.Getenv("LEFTHOOK") == "0" {
		return nil
	}

	lefthook, err := initialize(opts)
	if err != nil {
		return err
	}

	return lefthook.Run(hookName, gitArgs)
}

func (l *Lefthook) Run(hookName string, gitArgs []string) error {
	cfg, err := config.Load(l.Fs, l.repo.RootPath)
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		log.Info(log.Red(fmt.Sprintf("Config error! %s", err)))
		return err
	}

	successSprint = outputDisabler(outputSuccess, cfg.SkipOutput)

	hook, ok := cfg.Hooks[hookName]
	if !ok {
		log.Info(log.Red(fmt.Sprintf("Can't find hook by name '%s'.", hookName)))
		return errors.New("hook is not found")
	}
	if err := hook.Validate(); err != nil {
		log.Info(log.Red(fmt.Sprintf("Config error! %s", err)))
		return err
	}

	startTime := time.Now()

	if !isOutputDisabled(outputMeta, cfg.SkipOutput) {
		log.Info(log.Cyan("Lefthook v" + version.Version()))
		log.Info(log.Cyan("RUNNING HOOK:"), log.Bold(hookName))
	}

	spinner.Start()

	var wg sync.WaitGroup

	l.executeScripts(filepath.Join(cfg.SourceDir, hookName), gitArgs, hook, &wg)
	l.executeScripts(filepath.Join(cfg.SourceDirLocal, hookName), gitArgs, hook, &wg)

	for _, name := range hook.CommandsSorted {
		wg.Add(1)

		if hook.Parallel {
			go l.executeCommand(name, gitArgs, hook, &wg)
		} else {
			l.executeCommand(name, gitArgs, hook, &wg)
		}
	}

	wg.Wait()
	spinner.Stop()

	if !isOutputDisabled(outputSummary, cfg.SkipOutput) {
		printSummary(time.Since(startTime))
	}

	if len(failList) == 0 {
		return nil
	}
	return errors.New("Have failed script")
}

func (l *Lefthook) executeCommand(commandName string, gitArgs []string, hook *config.Hook, wg *sync.WaitGroup) {
	defer wg.Done()

	if hook.Piped && isPipeBroken {
		spinner.RestartWithMsg("\n", log.Bold(commandName), log.Red("(SKIP BY BROKEN PIPE)"))
		return
	}

	command := hook.Commands[commandName]

	if err := command.Validate(); err != nil {
		appendFailList(commandName)
		setPipeBroken()
		spinner.RestartWithMsg(commandName, log.Red(fmt.Sprintf("(SKIP BY CONFIG ERROR: %s)", err)))
		return
	}
	if command.DoSkip(l.repo.State()) {
		spinner.RestartWithMsg(successSprint(log.Bold(commandName), log.Red("(SKIP BY SETTINGS)")))
		return
	}
	if isContainsExcludedTags(command.Tags, hook.ExcludeTags) {
		spinner.RestartWithMsg(successSprint(log.Bold(commandName), log.Red("(SKIP BY TAGS)")))
		return
	}

	filesCommand := hook.Files
	if command.Files != "" {
		filesCommand = command.Files
	}

	runner := command.Run
	if strings.Contains(runner, config.SubStagedFiles) {
		files, err := l.repo.StagedFiles()
		files = prepareFiles(command, files)
		if err == nil {
			runner = strings.ReplaceAll(runner, config.SubStagedFiles, strings.Join(files, " "))
		}
	}
	if strings.Contains(runner, config.SubFiles) || filesCommand != "" {
		files, err := git.FilesByCommand(filesCommand)
		files = prepareFiles(command, files)
		if err == nil {
			runner = strings.ReplaceAll(runner, config.SubFiles, strings.Join(files, " "))
		}
	}
	if strings.Contains(runner, config.PushFiles) {
		files, err := l.repo.PushFiles()
		files = prepareFiles(command, files)
		if err == nil {
			runner = strings.ReplaceAll(runner, config.PushFiles, strings.Join(files, " "))
		}
	}
	if strings.Contains(runner, config.SubAllFiles) {
		files, err := l.repo.AllFiles()
		files = prepareFiles(command, files)
		if err == nil {
			runner = strings.ReplaceAll(runner, config.SubAllFiles, strings.Join(files, " "))
		}
	}

	runner = strings.ReplaceAll(runner, "{0}", strings.Join(gitArgs, " "))
	for gitArgIndex, gitArg := range gitArgs {
		runner = strings.ReplaceAll(runner, fmt.Sprintf("{%d}", gitArgIndex+1), gitArg)
	}

	commandOutput, isRun, err := RunCommand(runner, command.Root)

	stageName := fmt.Sprintln(log.Cyan("\n  EXECUTE >"), log.Bold(commandName))
	if !isRun {
		appendFailList(commandName)
		setPipeBroken()
		spinner.RestartWithMsg(stageName, err)
		return
	}

	if err == nil {
		spinner.RestartWithMsg(successSprint(stageName, commandOutput.String()))

		appendOkList(commandName)
	} else {
		spinner.RestartWithMsg(stageName, commandOutput.String())

		appendFailList(commandName)
		setPipeBroken()
	}
}

func prepareFiles(command *config.Command, files []string) []string {
	log.Debug("\nFiles before filters: \n", files)

	files = FilterGlob(files, command.Glob)
	files = FilterExclude(files, command.Exclude)
	files = FilterRelative(files, command.Root)

	log.Debug("\nFiles after filters: \n", files)

	var filesEsc []string
	for _, fileName := range files {
		if len(fileName) > 0 {
			filesEsc = append(filesEsc, shellescape.Quote(fileName))
		}
	}
	files = filesEsc

	log.Debug("Files after escaping: \n", files)

	return files
}

func (l *Lefthook) executeScripts(sourcePath string, gitArgs []string, hook *config.Hook, wg *sync.WaitGroup) {
	executables, err := afero.ReadDir(l.Fs, sourcePath)
	if err != nil || len(executables) == 0 {
		return
	}

	for _, executable := range executables {
		wg.Add(1)

		if hook.Parallel {
			go l.executeScript(executable, gitArgs, sourcePath, hook, wg)
		} else {
			l.executeScript(executable, gitArgs, sourcePath, hook, wg)
		}
	}
}

func (l *Lefthook) executeScript(executable os.FileInfo, gitArgs []string, sourcePath string, hook *config.Hook, wg *sync.WaitGroup) {
	defer wg.Done()

	executableName := executable.Name()

	if hook.Piped && isPipeBroken {
		spinner.RestartWithMsg("\n", log.Bold(executableName), log.Red("(SKIP BY BROKEN PIPE)"))
		return
	}

	script, ok := hook.Scripts[executableName]
	if !ok {
		spinner.RestartWithMsg(successSprint(log.Bold(executableName), log.Red("(SKIP BY NOT EXIST IN CONFIG)")))
		return
	}
	if script.DoSkip(l.repo.State()) {
		spinner.RestartWithMsg(successSprint(log.Bold(executableName), log.Red("(SKIP BY SETTINGS)")))
		return
	}
	if isContainsExcludedTags(script.Tags, hook.ExcludeTags) {
		spinner.RestartWithMsg(successSprint(log.Bold(executableName), log.Red("(SKIP BY TAGS)")))
		return
	}

	executablePath := filepath.Join(sourcePath, executableName)
	if err := isExecutable(executable); err != nil {
		if err := l.Fs.Chmod(executablePath, execMode); err != nil {
			appendFailList(executableName)

			spinner.RestartWithMsg(log.Bold(executableName), err, log.Red("(SKIP BY CHMOD ERROR)"))
			setPipeBroken()
			return
		}
	}

	command := exec.Command(executablePath, gitArgs...)
	if script.Runner != "" {
		runnerArg := append(strings.Split(script.Runner, " "), executablePath)
		runnerArg = append(runnerArg, gitArgs...)

		command = exec.Command(runnerArg[0], runnerArg[1:]...)
	}

	commandOutput, isRun, err := RunPlainCommand(command)

	stageName := fmt.Sprintln(log.Cyan("\n  EXECUTE >"), log.Bold(executableName))
	if !isRun {
		if os.IsPermission(err) {
			spinner.RestartWithMsg(successSprint(stageName, log.Red("(SKIP NOT EXECUTABLE FILE)")))
			return
		}

		appendFailList(executableName)
		spinner.RestartWithMsg(stageName, err, log.Red("TIP: Command start failed. Checkout `runner:` option for this script"))
		setPipeBroken()
		return
	}

	if err == nil {
		spinner.RestartWithMsg(successSprint(stageName, commandOutput.String()))

		appendOkList(executableName)
	} else {
		spinner.RestartWithMsg(stageName, commandOutput.String())

		appendFailList(executableName)
		setPipeBroken()
	}
}

func setPipeBroken() {
	isPipeBroken = true
}

func appendOkList(task string) {
	mutex.Lock()
	defer mutex.Unlock()
	okList = append(okList, task)
}

func appendFailList(task string) {
	mutex.Lock()
	defer mutex.Unlock()
	failList = append(failList, task)
}

func isExecutable(executable os.FileInfo) error {
	mode := executable.Mode()

	if !mode.IsRegular() {
		return os.ErrPermission
	}
	if (mode & execMask) == 0 {
		return os.ErrPermission
	}
	return nil
}

func isContainsExcludedTags(entityTagsList []string, excludedTags []string) bool {
	entityTags := make(map[string]struct{}, len(entityTagsList))
	for _, tag := range entityTagsList {
		entityTags[tag] = struct{}{}
	}
	for _, tag := range excludedTags {
		if _, ok := entityTags[tag]; ok {
			return true
		}
	}
	return false
}

func printSummary(execTime time.Duration) {
	if len(okList) == 0 && len(failList) == 0 {
		log.Info(log.Cyan("\nSUMMARY:"), log.Red("(SKIP EMPTY)"))
	} else {
		log.Info(log.Cyan(fmt.Sprintf("\nSUMMARY: (done in %.2f seconds)", execTime.Seconds())))
	}

	for _, fileName := range okList {
		log.Infof("✔️  %s\n", log.Green(fileName))
	}

	for _, fileName := range failList {
		log.Infof("🥊  %s", log.Red(fileName))
	}
}
