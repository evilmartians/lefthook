package pathtest

import (
	"path/filepath"
	"strings"
)

func AsOSPaths(paths []string) []string {
	result := make([]string, 0, len(paths))
	for _, path := range paths {
		result = append(result, AsOS(path))
	}

	return result
}

func AsOS(path string) string {
	return strings.Join(
		strings.Split(path, "/"),
		string(filepath.Separator),
	)
}
