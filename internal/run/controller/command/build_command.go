package command

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alessio/shellescape"

	"github.com/evilmartians/lefthook/v2/internal/config"
	"github.com/evilmartians/lefthook/v2/internal/run/controller/command/replacer"
	"github.com/evilmartians/lefthook/v2/internal/run/controller/filter"
	"github.com/evilmartians/lefthook/v2/internal/system"
)

func (b *Builder) buildCommand(params *JobParams) ([]string, []string, error) {
	if err := params.validateCommand(); err != nil {
		return nil, nil, err
	}

	// TODO(mrexox): Port this logic to buildScript and add tests
	replacer := b.buildReplacer(params)
	filter := filter.New(b.git.Fs, filter.Params{
		Glob:         params.Glob,
		ExcludeFiles: params.ExcludeFiles,
		Root:         params.Root,
		FileTypes:    params.FileTypes,
		GlobMatcher:  b.opts.GlobMatcher,
	})

	command := strings.Join([]string{params.Run, params.Args}, " ")
	err := replacer.Discover(command, filter)
	if err != nil {
		return nil, nil, err
	}

	// Checking substitutions and skipping execution if it is empty.
	//
	// Special case when `files` option specified but not referenced in `run`: return if the result is empty.
	if !b.opts.Force && len(params.FilesCmd) > 0 && replacer.Empty(config.SubFiles) {
		files, err := replacer.Files(config.SubFiles, filter)
		if err != nil {
			return nil, nil, err
		}

		if len(files) == 0 {
			return nil, nil, SkipError{"no files for inspection"}
		}
	}

	commands, replacedFiles := replacer.ReplaceAndSplit(command, system.MaxCmdLen())

	if b.opts.Force || len(replacedFiles) != 0 {
		return commands, replacedFiles, nil
	}

	// Skip if no files were staged (including deleted)
	if config.HookUsesStagedFiles(b.opts.HookName) {
		files, err := replacer.Files(config.SubStagedFiles, filter)
		if err != nil {
			return nil, nil, err
		}

		if len(files) == 0 {
			files, err = b.git.StagedFilesWithDeleted()
			if err != nil {
				return nil, nil, err
			}

			if len(filter.Apply(files)) == 0 {
				return nil, nil, SkipError{"no matching staged files"}
			}
		}
	}

	// Skip if no files were to be pushed
	if config.HookUsesPushFiles(b.opts.HookName) {
		files, err := replacer.Files(config.SubPushFiles, filter)
		if err != nil {
			return nil, nil, err
		}

		if len(files) == 0 {
			return nil, nil, SkipError{"no matching push files"}
		}
	}

	return commands, replacedFiles, nil
}

func replacePositionalArguments(str string, args []string) string {
	str = strings.ReplaceAll(str, "{0}", strings.Join(args, " "))
	for i, arg := range args {
		str = strings.ReplaceAll(str, fmt.Sprintf("{%d}", i+1), arg)
	}
	return str
}

func (b *Builder) buildReplacer(params *JobParams) replacer.Replacer {
	predefined := make(map[string]string)
	predefined["{0}"] = strings.Join(b.opts.GitArgs, " ")
	for i, arg := range b.opts.GitArgs {
		predefined["{"+strconv.Itoa(i+1)+"}"] = arg
	}
	for key, replacement := range b.opts.Templates {
		predefined["{"+key+"}"] = replacement
	}
	predefined["{lefthook_job_name}"] = shellescape.Quote(params.Name)

	if len(b.opts.ForceFiles) > 0 {
		return replacer.NewMocked(b.opts.ForceFiles, predefined)
	}

	return replacer.New(b.git, params.Root, params.FilesCmd, predefined)
}
