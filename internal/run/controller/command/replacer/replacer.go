package replacer

import (
	"fmt"
	"regexp"
	"runtime"
	"strings"

	"github.com/alessio/shellescape"

	"github.com/evilmartians/lefthook/v2/internal/config"
	"github.com/evilmartians/lefthook/v2/internal/git"
	"github.com/evilmartians/lefthook/v2/internal/log"
	"github.com/evilmartians/lefthook/v2/internal/run/controller/filter"
)

var surroundingQuotesRegexp = regexp.MustCompile(`^'(.*)'$`)

type entry struct {
	items []string
	cnt   int
}

type Replacer struct {
	cache     map[string]*entry
	files     map[string]func() ([]string, error)
	templates map[string]string
}

func New(
	git *git.Repository,
	root string,
	filesCmd string,
	templates map[string]string,
) Replacer {
	var (
		staged = git.StagedFiles
		push   = git.PushFiles
		all    = git.AllFiles
		cmd    = func() ([]string, error) {
			var cmd []string
			if runtime.GOOS == "windows" {
				cmd = strings.Split(filesCmd, " ")
			} else {
				cmd = []string{"sh", "-c", filesCmd}
			}
			return git.FindExistingFiles(cmd, root)
		}
	)

	return Replacer{
		cache: make(map[string]*entry),
		files: map[string]func() ([]string, error){
			config.SubStagedFiles: staged,
			config.SubPushFiles:   push,
			config.SubAllFiles:    all,
			config.SubFiles:       cmd,
		},
		templates: templates,
	}
}

func NewMocked(files []string, templates map[string]string) Replacer {
	forceFilesFn := func() ([]string, error) { return files, nil }

	return Replacer{
		cache: make(map[string]*entry),
		files: map[string]func() ([]string, error){
			config.SubStagedFiles: forceFilesFn,
			config.SubPushFiles:   forceFilesFn,
			config.SubAllFiles:    forceFilesFn,
			config.SubFiles:       forceFilesFn,
		},
		templates: templates,
	}
}

// Discover finds patterns in `source` and caches the results.
func (r Replacer) Discover(source string, filter *filter.Filter) error {
	for template, fn := range r.files {
		cnt := strings.Count(source, template)
		if cnt == 0 {
			continue
		}

		files, err := fn()
		if err != nil {
			return fmt.Errorf("error replacing %s: %w", template, err)
		}

		files = filter.Apply(files)

		r.cache[template] = &entry{items: files, cnt: cnt}
	}

	for template, replacement := range r.templates {
		cnt := strings.Count(source, template)
		if cnt == 0 {
			continue
		}

		r.cache[template] = &entry{items: []string{replacement}, cnt: cnt}
	}

	return nil
}

func (r Replacer) Empty(key string) bool {
	_, ok := r.cache[key]
	return !ok
}

func (r Replacer) Files(template string, filter *filter.Filter) ([]string, error) {
	entry, ok := r.cache[template]
	if ok {
		return entry.items, nil
	}

	fn, ok := r.files[template]
	if !ok {
		panic("filtering: no such files template: " + template)
	}

	files, err := fn()
	if err != nil {
		return nil, err
	}

	return filter.Apply(files), nil
}

func (r Replacer) ReplaceAndSplit(command string, maxlen int) ([]string, []string) {
	if len(r.cache) == 0 {
		return []string{command}, nil
	}

	var cnt int

	allFiles := make([]string, 0)
	for template, entry := range r.cache {
		if entry.cnt == 0 {
			continue
		}

		cnt += entry.cnt
		maxlen += entry.cnt * len(template)
		if _, ok := r.files[template]; ok {
			allFiles = append(allFiles, entry.items...)
		}

		entry.items = escapeFiles(entry.items)
	}

	maxlen -= len(command)

	if cnt > 0 {
		maxlen /= cnt
	}

	var exhausted int
	commands := make([]string, 0)
	for {
		result := command
		for template, entry := range r.cache {
			added, rest := getNChars(entry.items, maxlen)
			if len(rest) == 0 {
				exhausted += 1
			} else {
				entry.items = rest
			}
			result = replaceQuoted(result, template, added)
		}

		log.Debug("[lefthook] job: ", result)
		commands = append(commands, result)
		if exhausted >= len(r.cache) {
			break
		}
	}

	return commands, allFiles
}

// Escape file names to prevent unexpected bugs.
func escapeFiles(files []string) []string {
	var filesEsc []string
	for _, fileName := range files {
		if len(fileName) > 0 {
			filesEsc = append(filesEsc, shellescape.Quote(fileName))
		}
	}

	log.Builder(log.DebugLevel, "[lefthook] ").
		Add("files after escaping: ", filesEsc).
		Log()

	return filesEsc
}

func getNChars(s []string, n int) ([]string, []string) {
	if len(s) == 0 {
		return nil, nil
	}

	var cnt int
	for i, str := range s {
		cnt += len(str)
		if i > 0 {
			cnt += 1 // a space
		}
		if cnt > n {
			if i == 0 {
				i = 1
			}
			return s[:i], s[i:]
		}
	}

	return s, nil
}

func replaceQuoted(source, substitution string, files []string) string {
	for _, elem := range [][]string{
		{"\"", "\"" + substitution + "\""},
		{"'", "'" + substitution + "'"},
		{"", substitution},
	} {
		quote := elem[0]
		sub := elem[1]
		if !strings.Contains(source, sub) {
			continue
		}

		quotedFiles := files
		if len(quote) != 0 {
			quotedFiles = make([]string, 0, len(files))
			for _, fileName := range files {
				quotedFiles = append(quotedFiles,
					quote+surroundingQuotesRegexp.ReplaceAllString(fileName, "$1")+quote)
			}
		}

		source = strings.ReplaceAll(
			source, sub, strings.Join(quotedFiles, " "),
		)
	}

	return source
}
