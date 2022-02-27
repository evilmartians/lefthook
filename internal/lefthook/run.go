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

	envExcludedTags []string

	successSprint outputDisablerFn
)

const (
	subFiles       string = "{files}"
	subAllFiles    string = "{all_files}"
	subStagedFiles string = "{staged_files}"
	pushFiles      string = "{push_files}"

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
		log.Info(log.Brown(fmt.Sprintf("Config error! %s", err)))
		return err
	}

	if tags := os.Getenv("LEFTHOOK_EXCLUDE"); tags != "" {
		envExcludedTags = strings.Split(tags, ",")
	}

	var envDisabledOutputs []string
	if tags := os.Getenv("LEFTHOOK_QUIET"); tags != "" {
		envDisabledOutputs = strings.Split(tags, ",")
	}
	successSprint = outputDisabler(outputSuccess, cfg.SkipOutput, envDisabledOutputs)

	hook, ok := cfg.Hooks[hookName]
	if !ok {
		log.Info(log.Brown(fmt.Sprintf("Can't find hook by name '%s'.", hookName)))
		return errors.New("hook is not found")
	}
	if err := hook.Validate(); err != nil {
		log.Info(log.Brown(fmt.Sprintf("Config error! %s", err)))
		return err
	}

	startTime := time.Now()

	if !isOutputDisabled(outputMeta, cfg.SkipOutput, envDisabledOutputs) {
		log.Info(log.Cyan("Lefthook v" + version.Version()))
		log.Info(log.Cyan("RUNNING HOOK:"), log.Bold(hookName))
	}

	spinner.Start()

	var wg sync.WaitGroup
	l.executeScripts(hookName, gitArgs, cfg.SourceDir, hook, &wg)
	l.executeScripts(hookName, gitArgs, cfg.SourceDirLocal, hook, &wg)

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

	if !isOutputDisabled(outputSummary, cfg.SkipOutput, envDisabledOutputs) {
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
		spinner.RestartWithMsg("\n", log.Bold(commandName), log.Brown("(SKIP BY BROKEN PIPE)"))
		return
	}

	command := hook.Commands[commandName]

	if command.DoSkip(l.repo.State()) {
		spinner.RestartWithMsg(successSprint(log.Bold(commandName), log.Brown("(SKIP BY SETTINGS)")))
		return
	}
	if isContainsExcludedTags(command.Tags, hook.ExcludeTags) {
		spinner.RestartWithMsg(successSprint(log.Bold(commandName), log.Brown("(SKIP BY TAGS)")))
		return
	}

	runner := command.GetRunner()

	var files []string
	filesCommand := hook.Files
	if command.Files != "" {
		filesCommand = command.Files
	}

	if strings.Contains(runner, subStagedFiles) {
		files, _ = l.repo.StagedFiles()
	} else if strings.Contains(runner, subFiles) || filesCommand != "" {
		files, _ = git.FilesByCommand(filesCommand)
	} else if strings.Contains(runner, pushFiles) {
		files, _ = l.repo.PushFiles()
	} else {
		files, _ = l.repo.AllFiles()
	}

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

	runner = strings.Replace(runner, pushFiles, strings.Join(files, " "), -1)
	runner = strings.Replace(runner, subStagedFiles, strings.Join(files, " "), -1)
	runner = strings.Replace(runner, subAllFiles, strings.Join(files, " "), -1)
	runner = strings.Replace(runner, subFiles, strings.Join(files, " "), -1)
	runner = strings.Replace(runner, "{0}", strings.Join(gitArgs, " "), -1)
	for gitArgIndex, gitArg := range gitArgs {
		runner = strings.Replace(runner, fmt.Sprintf("{%d}", gitArgIndex+1), gitArg, -1)
	}

	commandOutput, wait, err := RunCommand(runner, command.Root)

	mutex.Lock()
	defer mutex.Unlock()

	stageName := fmt.Sprintln(log.Cyan("\n  EXECUTE >"), log.Bold(commandName))
	if err != nil {
		failList = append(failList, commandName)
		setPipeBroken()
		spinner.RestartWithMsg(stageName, err)
		return
	}

	if wait() == nil {
		spinner.RestartWithMsg(successSprint(stageName, commandOutput.String()))

		okList = append(okList, commandName)
	} else {
		spinner.RestartWithMsg(stageName, commandOutput.String())

		failList = append(failList, commandName)
		setPipeBroken()
	}
}

func (l *Lefthook) executeScripts(hookName string, gitArgs []string, sourceDir string, hook *config.Hook, wg *sync.WaitGroup) {
	sourcePath := filepath.Join(sourceDir, hookName)
	executables, err := afero.ReadDir(l.Fs, sourcePath)
	if err != nil || len(executables) == 0 {
		return
	}

	for _, executable := range executables {
		wg.Add(1)

		f := func() {
			defer wg.Done()

			executableName := executable.Name()

			if hook.Piped && isPipeBroken {
				spinner.RestartWithMsg("\n", log.Bold(executableName), log.Brown("(SKIP BY BROKEN PIPE)"))
				return
			}

			script, ok := hook.Scripts[executableName]
			if !ok {
				spinner.RestartWithMsg(successSprint(log.Bold(executableName), log.Brown("(SKIP BY NOT EXIST IN CONFIG)")))
				return
			}
			if script.DoSkip(l.repo.State()) {
				spinner.RestartWithMsg(successSprint(log.Bold(executableName), log.Brown("(SKIP BY SETTINGS)")))
				return
			}
			if isContainsExcludedTags(script.Tags, hook.ExcludeTags) {
				spinner.RestartWithMsg(successSprint(log.Bold(executableName), log.Brown("(SKIP BY TAGS)")))
				return
			}

			executablePath := filepath.Join(sourcePath, executableName)
			if err := isExecutable(executable); err != nil {
				if err := l.Fs.Chmod(executablePath, execMode); err != nil {
					log.Error(err)
					panic(err)
				}
			}

			command := exec.Command(executablePath, gitArgs...)
			if runner := script.GetRunner(); runner != "" {
				runnerArg := append(strings.Split(runner, " "), executablePath)
				runnerArg = append(runnerArg, gitArgs...)

				command = exec.Command(runnerArg[0], runnerArg[1:]...)
			}

			commandOutput, wait, err := RunPlainCommand(command)

			mutex.Lock()
			defer mutex.Unlock()

			stageName := fmt.Sprintln(log.Cyan("\n  EXECUTE >"), log.Bold(executableName))
			if os.IsPermission(err) {
				spinner.RestartWithMsg(successSprint(stageName, log.Brown("(SKIP NOT EXECUTABLE FILE)")))
				return
			}
			if err != nil {
				failList = append(failList, executableName)
				spinner.RestartWithMsg(stageName, err, log.Brown("TIP: Command start failed. Checkout `runner:` option for this script"))
				setPipeBroken()
				return
			}

			if wait() == nil {
				spinner.RestartWithMsg(successSprint(stageName, commandOutput.String()))

				okList = append(okList, executableName)
			} else {
				spinner.RestartWithMsg(stageName, commandOutput.String())

				failList = append(failList, executableName)
				setPipeBroken()
			}
		}

		if hook.Parallel {
			go f()
		} else {
			f()
		}
	}
}

func setPipeBroken() {
	isPipeBroken = true
}

func isExecutable(executable os.FileInfo) error {
	mode := executable.Mode()

	if !mode.IsRegular() {
		return errors.New("ErrPermission")
	}
	if (mode & execMask) == 0 {
		return errors.New("ErrPermission")
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
	for _, tag := range envExcludedTags {
		if _, ok := entityTags[tag]; ok {
			return true
		}
	}
	return false
}

func printSummary(execTime time.Duration) {
	if len(okList) == 0 && len(failList) == 0 {
		log.Info(log.Cyan("\nSUMMARY:"), log.Brown("(SKIP EMPTY)"))
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
