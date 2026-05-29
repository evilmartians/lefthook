package version

import (
	"errors"
	"strings"

	"golang.org/x/mod/semver"
)

const version = "2.1.9"

var (
	// Is set via -X github.com/evilmartians/lefthook/v2/internal/version.commit={commit}.
	commit string
	// Is set via -X github.com/evilmartians/lefthook/v2/internal/version.dev=true.
	dev string

	ErrInvalidVersion   = errors.New("invalid version format")
	ErrUncoveredVersion = errors.New("version is lower than required")
)

func Version(verbose bool) string {
	result := strings.Builder{}
	result.WriteString(version)

	if dev == "true" {
		result.WriteString("-dev")
	}

	if verbose {
		result.WriteString(" ")
		result.WriteString(commit)
	}

	return result.String()
}

func Check(wanted, given string) error {
	if wanted[0] != 'v' {
		wanted = "v" + wanted
	}
	if given[0] != 'v' {
		given = "v" + given
	}

	if !semver.IsValid(wanted) {
		return ErrInvalidVersion
	}
	if !semver.IsValid(given) {
		return ErrInvalidVersion
	}
	if cmp := semver.Compare(wanted, given); cmp > 0 {
		return ErrUncoveredVersion
	}

	return nil
}
