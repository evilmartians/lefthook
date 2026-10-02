package wrapper

import (
	"runtime"
	"strings"
)

func (w *Wrapper) FilesByCommandRelative(cmd string, dir string) ([]string, error) {
	var args []string
	if runtime.GOOS == "windows" {
		args = strings.Split(cmd, " ")
	} else {
		args = []string{"sh", "-c", cmd}
	}

	return w.FilesRelative(args, dir)
}
