package helpers

import (
	"bytes"
	"strings"

	"github.com/goccy/go-yaml"

	"github.com/evilmartians/lefthook/internal/config"
)

// ParseHook simplifies config.Hook definition with YAML string.
func ParseHook(str string) *config.Hook {
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
	hook := config.Hook{}
	err := yaml.Unmarshal(cleanBuffer.Bytes(), &hook)
	if err != nil {
		panic("Failed to parse test data: " + err.Error())
	}
	return &hook
}
