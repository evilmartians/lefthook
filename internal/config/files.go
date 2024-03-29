package config

import "strings"

const (
	SubFiles       string = "{files}"
	SubAllFiles    string = "{all_files}"
	SubStagedFiles string = "{staged_files}"
	SubPushFiles   string = "{push_files}"
)

func isRunnerFilesCompatible(runner string) bool {
	if strings.Contains(runner, SubStagedFiles) && strings.Contains(runner, SubPushFiles) {
		return false
	}
	return true
}
