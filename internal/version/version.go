package version

import (
	"errors"

	"golang.org/x/mod/semver"
)

const version = "2.0.8"

var (
	// Is set via -X github.com/evilmartians/lefthook/internal/version.commit={commit}.
	commit string

	ErrInvalidVersion   = errors.New("invalid version format")
	ErrUncoveredVersion = errors.New("version is lower than required")
)

func Version(verbose bool) string {
	if verbose {
		return version + " " + commit
	}

	return version
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
