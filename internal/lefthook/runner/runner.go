package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/briandowns/spinner"
	"github.com/spf13/afero"
	"gopkg.in/alessio/shellescape.v1"

	"github.com/evilmartians/lefthook/internal/config"
	"github.com/evilmartians/lefthook/internal/git"
	"github.com/evilmartians/lefthook/internal/log"
)

type status int8

const (
	executableFileMode os.FileMode = 0o751
	executableMask     os.FileMode = 0o111

	spinnerCharSet     int = 14
	spinnerRefreshRate     = 100 * time.Millisecond
)

type Runner struct {
	Failed bool

	fs         afero.Fs
	repo       *git.Repository
	spinner    *spinner.Spinner
	hook       *config.Hook
	args       []string
	resultChan chan Result
}

func NewRunner(
	fs afero.Fs,
	repo *git.Repository,
	hook *config.Hook,
	args []string,
	resultChan chan Result,
) *Runner {
	return &Runner{
		fs:   fs,
		repo: repo,
		spinner: spinner.New(
			spinner.CharSets[spinnerCharSet],
			spinnerRefreshRate,
			spinner.WithSuffix(" waiting"),
		),
		hook:       hook,
		args:       args,
		resultChan: resultChan,
	}
}

func (r *Runner) RunAll(scriptDirs []string) {
	for _, dir := range scriptDirs {
		r.runScripts(dir)
	}
	r.runCommands()
	r.spinner.Stop() // make sure spinner is stopped
}

func (r *Runner) fail(name string) {
	r.resultChan <- resultFail(name)
	r.Failed = true
}

func (r *Runner) success(name string) {
	r.resultChan <- resultSuccess(name)
}

func (r *Runner) runScripts(dir string) {
	files, err := afero.ReadDir(r.fs, dir) // ReadDir already sorts files by .Name()
	if err != nil || len(files) == 0 {
		return
	}

	var wg sync.WaitGroup
	for _, file := range files {
		script, ok := r.hook.Scripts[file.Name()]
		if !ok {
			log.Debug(log.Bold(file.Name()), log.Red("(SKIP BY NOT EXIST IN CONFIG)"))
			continue
		}

		if r.Failed && r.hook.Piped {
			log.Info(" ", log.Bold(file.Name()), log.Red("(SKIP BY BROKEN PIPE)"))
			continue
		}

		if r.hook.Parallel {
			wg.Add(1)
			go func(script *config.Script, path string, file os.FileInfo) {
				defer wg.Done()
				r.runScript(script, path, file)
			}(script, filepath.Join(dir, file.Name()), file)
		} else {
			r.runScript(script, filepath.Join(dir, file.Name()), file)
		}
	}

	wg.Wait()
}

func (r *Runner) runScript(script *config.Script, path string, file os.FileInfo) {
	if script.DoSkip(r.repo.State()) {
		log.Info(log.Bold(file.Name()), log.Cyan("(SKIP BY SETTINGS)"))
		return
	}

	if intersect(r.hook.ExcludeTags, script.Tags) {
		log.Info(log.Bold(file.Name()), log.Cyan("(SKIP BY TAGS)"))
	}

	// Skip non-regular files (dirs, symlinks, sockets, etc.)
	if !file.Mode().IsRegular() {
		log.Debugf("File %s is not a regular file, skipping", file.Name())
		return
	}

	// Make sure file is executable
	if (file.Mode() & executableMask) == 0 {
		if err := r.fs.Chmod(path, executableFileMode); err != nil {
			log.Errorf("Couldn't change file mode to make file executable: %s", err)
			r.fail(file.Name())
			return
		}
	}

	var args []string
	if len(script.Runner) > 0 {
		args = strings.Split(script.Runner, " ")
	}

	args = append(args, path)
	args = append(args, r.args[:]...)

	r.run(file.Name(), r.repo.RootPath, args)
}

func (r *Runner) runCommands() {
	var wg sync.WaitGroup

	commands := make([]string, 0, len(r.hook.Commands))
	for name := range r.hook.Commands {
		commands = append(commands, name)
	}

	sort.Strings(commands)

	for _, name := range commands {
		if r.Failed && r.hook.Piped {
			log.Info(" ", log.Bold(name), log.Red("(SKIP BY BROKEN PIPE)"))
			continue
		}

		if r.hook.Parallel {
			wg.Add(1)
			go func(name string, command *config.Command) {
				defer wg.Done()
				r.runCommand(name, command)
			}(name, r.hook.Commands[name])
		} else {
			r.runCommand(name, r.hook.Commands[name])
		}
	}

	wg.Wait()
}

func (r *Runner) runCommand(name string, command *config.Command) {
	if command.DoSkip(r.repo.State()) {
		log.Info(log.Cyan(fmt.Sprintf("%s: (SKIP BY SETTINGS)", name)))
		return
	}

	if intersect(r.hook.ExcludeTags, command.Tags) {
		log.Info(log.Cyan(fmt.Sprintf("%s: (SKIP BY TAGS)", name)))
		return
	}

	if err := command.Validate(); err != nil {
		r.fail(name)
		return
	}

	root := filepath.Join(r.repo.RootPath, command.Root)
	r.run(name, root, r.buildCommandArgs(command))
}

func (r *Runner) buildCommandArgs(command *config.Command) []string {
	filesCommand := r.hook.Files
	if command.Files != "" {
		filesCommand = command.Files
	}

	filesTypeToFn := map[string]func() ([]string, error){
		config.SubStagedFiles: r.repo.StagedFiles,
		config.PushFiles:      r.repo.PushFiles,
		config.SubAllFiles:    r.repo.AllFiles,
		config.SubFiles: func() ([]string, error) {
			return git.FilesByCommand(filesCommand)
		},
	}

	runString := command.Run
	for filesType, filesFn := range filesTypeToFn {
		if strings.Contains(runString, filesType) {
			files, err := filesFn()
			if err != nil {
				continue
			}
			runString = strings.ReplaceAll(
				runString, filesType, prepareFiles(command, files),
			)
		}
	}

	runString = strings.ReplaceAll(runString, "{0}", strings.Join(r.args, " "))
	for i, gitArg := range r.args {
		runString = strings.ReplaceAll(runString, fmt.Sprintf("{%d}", i+1), gitArg)
	}

	return strings.Split(runString, " ")
}

func prepareFiles(command *config.Command, files []string) string {
	if files == nil {
		return ""
	}

	log.Debug("Files before filters:\n", files)

	files = filterGlob(files, command.Glob)
	files = filterExclude(files, command.Exclude)
	files = filterRelative(files, command.Root)

	log.Debug("Files after filters:\n", files)

	// Escape file names to prevent unexpected bugs
	var filesEsc []string
	for _, fileName := range files {
		if len(fileName) > 0 {
			filesEsc = append(filesEsc, shellescape.Quote(fileName))
		}
	}
	files = filesEsc

	log.Debug("Files after escaping:\n", files)

	return strings.Join(files, " ")
}

func (r *Runner) run(name string, root string, args []string) {
	r.spinner.Start()
	out, err := Execute(root, args)
	r.spinner.Stop()

	var execName string
	if err != nil {
		r.fail(name)
		execName = fmt.Sprint(log.Red("\n  EXECUTE >"), log.Bold(name))
	} else {
		r.success(name)
		execName = fmt.Sprint(log.Cyan("\n  EXECUTE >"), log.Bold(name))
	}

	if out != nil {
		log.Infof("%s\n%s\n", execName, out)
	} else if err != nil {
		log.Infof("%s\n%s\n", execName, err)
	}

	r.spinner.Start()
}

// Returns whether two arrays have at least one similar element.
func intersect(a, b []string) bool {
	intersections := make(map[string]struct{}, len(a))

	for _, v := range a {
		intersections[v] = struct{}{}
	}

	for _, v := range b {
		if _, ok := intersections[v]; ok {
			return true
		}
	}

	return false
}
