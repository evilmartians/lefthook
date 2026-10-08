package command

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// commandScopedHooksPath returns core.hooksPath if it is set at command scope,
// i.e. via `git -c core.hooksPath=...` or GIT_CONFIG_COUNT/GIT_CONFIG_KEY_n/
// GIT_CONFIG_VALUE_n. Either triggers the same bug, so we check both.
//
// When git runs a hook, it passes command-scoped config to the hook's environment
// via GIT_CONFIG_PARAMETERS (the internal serialization of `-c` pairs). Users may
// also supply GIT_CONFIG_COUNT/GIT_CONFIG_KEY_n/GIT_CONFIG_VALUE_n directly.
func commandScopedHooksPath() string {
	// GIT_CONFIG_COUNT / GIT_CONFIG_KEY_n / GIT_CONFIG_VALUE_n
	if count, err := strconv.Atoi(os.Getenv("GIT_CONFIG_COUNT")); err == nil && count > 0 {
		for i := range count {
			key := os.Getenv(fmt.Sprintf("GIT_CONFIG_KEY_%d", i))
			if strings.EqualFold(key, "core.hooksPath") {
				return os.Getenv(fmt.Sprintf("GIT_CONFIG_VALUE_%d", i))
			}
		}
	}

	// GIT_CONFIG_PARAMETERS is the internal format git uses to propagate -c pairs
	// to child processes (including hook scripts). Format: 'key=value' 'key2=v2'
	// where literal single quotes inside a value are escaped as '\''
	if params := os.Getenv("GIT_CONFIG_PARAMETERS"); params != "" {
		for _, kv := range parseGitConfigParameters(params) {
			before, after, ok := strings.Cut(kv, "=")
			if !ok {
				continue
			}
			if strings.EqualFold(before, "core.hooksPath") {
				return after
			}
		}
	}

	return ""
}

// parseGitConfigParameters parses the GIT_CONFIG_PARAMETERS environment variable.
// The value is a series of space-separated shell-quoted 'key=value' pairs, where
// a literal single quote inside a quoted segment is represented as: '\”.
// (closing quote, backslash-escaped quote, reopening quote).
func parseGitConfigParameters(params string) []string {
	var pairs []string
	var buf strings.Builder
	i := 0
	for i < len(params) {
		switch {
		case params[i] == ' ':
			if buf.Len() > 0 {
				pairs = append(pairs, buf.String())
				buf.Reset()
			}
			i++

		case params[i] == '\'':
			// Single-quoted segment: read until the next unescaped single quote.
			i++ // skip opening quote
			for i < len(params) && params[i] != '\'' {
				buf.WriteByte(params[i])
				i++
			}
			if i < len(params) {
				i++ // skip closing quote
			}

		case params[i] == '\\' && i+1 < len(params):
			// Backslash escape outside of quotes (e.g. the \' in '\'')
			buf.WriteByte(params[i+1])
			i += 2

		default:
			buf.WriteByte(params[i])
			i++
		}
	}
	if buf.Len() > 0 {
		pairs = append(pairs, buf.String())
	}
	return pairs
}
