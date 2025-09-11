package config

import (
	"cmp"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/evilmartians/lefthook/internal/git"
	"github.com/evilmartians/lefthook/internal/system"
)

type Script struct {
	Runner string `json:"runner,omitempty" mapstructure:"runner" toml:"runner,omitempty" yaml:"runner,omitempty"`

	Skip     interface{}       `json:"skip,omitempty"     jsonschema:"oneof_type=boolean;array" mapstructure:"skip"       toml:"skip,omitempty,inline" yaml:",omitempty"`
	Only     interface{}       `json:"only,omitempty"     jsonschema:"oneof_type=boolean;array" mapstructure:"only"       toml:"only,omitempty,inline" yaml:",omitempty"`
	Tags     []string          `json:"tags,omitempty"     jsonschema:"oneof_type=string;array"  mapstructure:"tags"       toml:"tags,omitempty"        yaml:",omitempty"`
	Env      map[string]string `json:"env,omitempty"      mapstructure:"env"                    toml:"env,omitempty"      yaml:",omitempty"`
	Priority int               `json:"priority,omitempty" mapstructure:"priority"               toml:"priority,omitempty" yaml:",omitempty"`

	FailText    string `json:"fail_text,omitempty"   koanf:"fail_text"          mapstructure:"fail_text"     toml:"fail_text,omitempty"   yaml:"fail_text,omitempty"`
	Interactive bool   `json:"interactive,omitempty" mapstructure:"interactive" toml:"interactive,omitempty" yaml:",omitempty"`
	UseStdin    bool   `json:"use_stdin,omitempty"   koanf:"use_stdin"          mapstructure:"use_stdin"     toml:"use_stdin,omitempty"   yaml:"use_stdin,omitempty"`
	StageFixed  bool   `json:"stage_fixed,omitempty" koanf:"stage_fixed"        mapstructure:"stage_fixed"   toml:"stage_fixed,omitempty" yaml:"stage_fixed,omitempty"`
}

func (s Script) DoSkip(state func() git.State) bool {
	skipChecker := NewSkipChecker(system.Cmd)
	return skipChecker.Check(state, s.Skip, s.Only)
}

func ScriptsToJobs(scripts map[string]*Script) []*Job {
	jobs := make([]*Job, 0, len(scripts))
	for name, script := range scripts {
		jobs = append(jobs, &Job{
			Name:        name,
			Script:      name,
			Runner:      script.Runner,
			FailText:    script.FailText,
			Tags:        script.Tags,
			Env:         script.Env,
			Interactive: script.Interactive,
			UseStdin:    script.UseStdin,
			StageFixed:  script.StageFixed,
			Skip:        script.Skip,
			Only:        script.Only,
		})
	}

	// ASC
	slices.SortFunc(jobs, func(i, j *Job) int {
		a := scripts[i.Name]
		b := scripts[j.Name]

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

func parseNum(str string) int {
	numEnds := -1
	for idx, ch := range str {
		if unicode.IsDigit(ch) {
			numEnds = idx
		} else {
			break
		}
	}

	if numEnds == -1 {
		return -1
	}
	num, err := strconv.Atoi(str[:numEnds+1])
	if err != nil {
		return -1
	}

	return num
}
