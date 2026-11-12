package wrapper

import (
	"regexp"
	"strings"
)

var (
	cmdListStash   = []string{"git", "stash", "list", "--format=%gd %H %gs"}
	reStashMessage = regexp.MustCompile(`^(?P<stash>[^ ]+) (?P<hash>[^ ]+) (?P<message>.*)$`)
)

func (w *Wrapper) DropStash() error {
	lines, err := w.cmd.cmdLines(cmdListStash)
	if err != nil {
		return err
	}

	for i := range lines {
		line := lines[len(lines)-i-1]
		matches := reStashMessage.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		stash := matches[reStashMessage.SubexpIndex("stash")]
		hash := matches[reStashMessage.SubexpIndex("hash")]
		message := matches[reStashMessage.SubexpIndex("message")]

		if len(w.cache.backupStashHash) > 0 {
			if w.cache.backupStashHash == hash {
				_, err := w.cmd.cmd([]string{
					"git",
					"stash",
					"drop",
					"--quiet",
					"--",
					stash,
				})
				return err
			}

			continue
		}

		if strings.Contains(message, stashMessage) {
			_, err := w.cmd.cmd([]string{
				"git",
				"stash",
				"drop",
				"--quiet",
				"--",
				stash,
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}
