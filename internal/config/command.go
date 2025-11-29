package config

import (
	"cmp"
	"errors"
	"slices"
	"strings"
)

var ErrFilesIncompatible = errors.New("one of your runners contains incompatible file types")

type Command struct {
	Run      string `json:"run"                 mapstructure:"run"   toml:"run"               yaml:"run"`
	Files    string `json:"files,omitempty"     mapstructure:"files" toml:"files,omitempty"   yaml:",omitempty"`
	Root     string `json:"root,omitempty"      mapstructure:"root"  toml:"root,omitempty"    yaml:",omitempty"`
	FailText string `json:"fail_text,omitempty" koanf:"fail_text"    mapstructure:"fail_text" toml:"fail_text,omitempty" yaml:"fail_text,omitempty"`

	Skip any `json:"skip,omitempty" jsonschema:"oneof_type=boolean;array" mapstructure:"skip" toml:"skip,omitempty,inline" yaml:",omitempty"`
	Only any `json:"only,omitempty" jsonschema:"oneof_type=boolean;array" mapstructure:"only" toml:"only,omitempty,inline" yaml:",omitempty"`

	Tags      []string `json:"tags,omitempty"       jsonschema:"oneof_type=string;array" mapstructure:"tags"       toml:"tags,omitempty"       yaml:",omitempty"`
	FileTypes []string `json:"file_types,omitempty" koanf:"file_types"                   mapstructure:"file_types" toml:"file_types,omitempty" yaml:"file_types,omitempty"`
	Glob      []string `json:"glob,omitempty"       jsonschema:"oneof_type=string;array" mapstructure:"glob"       toml:"glob,omitempty"       yaml:",omitempty"`
	Exclude   []string `json:"exclude,omitempty"    jsonschema:"oneof_type=string;array" mapstructure:"exclude"    toml:"exclude,omitempty"    yaml:",omitempty"`

	Env map[string]string `json:"env,omitempty" mapstructure:"env" toml:"env,omitempty" yaml:",omitempty"`

	Priority    int  `json:"priority,omitempty"    mapstructure:"priority"    toml:"priority,omitempty"    yaml:",omitempty"`
	Interactive bool `json:"interactive,omitempty" mapstructure:"interactive" toml:"interactive,omitempty" yaml:",omitempty"`
	UseStdin    bool `json:"use_stdin,omitempty"   koanf:"use_stdin"          mapstructure:"use_stdin"     toml:"use_stdin,omitempty"   yaml:"use_stdin,omitempty"`
	StageFixed  bool `json:"stage_fixed,omitempty" koanf:"stage_fixed"        mapstructure:"stage_fixed"   toml:"stage_fixed,omitempty" yaml:"stage_fixed,omitempty"`
}

func CommandsToJobs(commands map[string]*Command) []*Job {
	jobs := make([]*Job, 0, len(commands))
	for name, command := range commands {
		jobs = append(jobs, &Job{
			Name:        name,
			Run:         command.Run,
			Glob:        command.Glob,
			Root:        command.Root,
			Files:       command.Files,
			FailText:    command.FailText,
			Tags:        command.Tags,
			FileTypes:   command.FileTypes,
			Env:         command.Env,
			Interactive: command.Interactive,
			UseStdin:    command.UseStdin,
			StageFixed:  command.StageFixed,
			Exclude:     command.Exclude,
			Skip:        command.Skip,
			Only:        command.Only,
		})
	}

	// ASC
	slices.SortFunc(jobs, func(i, j *Job) int {
		a := commands[i.Name]
		b := commands[j.Name]

		if a.Priority != 0 || b.Priority != 0 {
			// Script without a priority must be the last
			if a.Priority == 0 {
				return 1
			}
			if b.Priority == 0 {
				return -1
			}

			return cmp.Compare(a.Priority, b.Priority)
		}

		iNum := parseNum(i.Name)
		jNum := parseNum(j.Name)

		if iNum == -1 && jNum == -1 {
			return strings.Compare(i.Name, j.Name)
		}

		if iNum == -1 {
			return 1
		}

		if jNum == -1 {
			return -1
		}

		return cmp.Compare(iNum, jNum)
	})

	return jobs
}
