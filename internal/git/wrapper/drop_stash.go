package wrapper

import "regexp"

var cmdListStash = []string{"git", "stash", "list"}
var reStashMessage = regexp.MustCompile(`^(?P<stash>[^ ]+):\s*` + stashMessage)

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

		stashID := reStashMessage.SubexpIndex("stash")

		if len(matches[stashID]) > 0 {
			_, err := w.cmd.cmd([]string{
				"git",
				"stash",
				"drop",
				"--quiet",
				"--",
				matches[stashID],
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}
