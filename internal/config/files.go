package config

import "strings"

const (
	SubFiles       string = "{files}"
	SubAllFiles    string = "{all_files}"
	SubStagedFiles string = "{staged_files}"
	SubPushFiles   string = "{push_files}"
)

func IsRunFilesCompatible(run string) bool {
	return !strings.Contains(run, SubStagedFiles) || !strings.Contains(run, SubPushFiles)
}
