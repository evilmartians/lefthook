package configtest

import (
	"bytes"
	"strings"

	"github.com/goccy/go-yaml"

	"github.com/evilmartians/lefthook/v2/internal/config"
)

// ParseHook simplifies config.Hook definition with YAML string.
func ParseHook(str string) *config.Hook {
	hook := config.Hook{}
	err := yaml.Unmarshal(stripPadding(str), &hook)
	if err != nil {
		panic("Failed to parse hook: " + err.Error())
	}
	return &hook
}

// ParseJob simplifies config.Job definition with YAML string.
func ParseJob(str string) *config.Job {
	job := config.Job{}
	err := yaml.Unmarshal(stripPadding(str), &job)
	if err != nil {
		panic("Failed to parse job: " + err.Error())
	}
	return &job
}

func stripPadding(str string) []byte {
	str = strings.TrimRight(strings.Trim(str, "\n"), " \t")
	cleanBuffer := new(bytes.Buffer)
	var padding int
	var paddingSet bool
	for line := range strings.Lines(str) {
		var cleanLine string
		if !paddingSet {
			cleanLine = strings.TrimLeft(line, " \t")
			padding = len(line) - len(cleanLine)
			paddingSet = true
		} else {
			cleanLine = line[padding:]
		}
		cleanBuffer.WriteString(cleanLine)
	}

	return cleanBuffer.Bytes()
}
