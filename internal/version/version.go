package version

import (
	"errors"
	"regexp"
	"strconv"
)

const version = "1.5.0"

var (
	// Is set via -X github.com/evilmartians/lefthook/internal/version.commit={commit}.
	commit string

	versionRegexp = regexp.MustCompile(
		`^(?P<major>\d+)(?:\.(?P<minor>\d+)(?:\.(?P<patch>\d+))?)?$`,
	)

	errIncorrectVersion = errors.New("format of 'min_version' setting is incorrect")
	errUncovered        = errors.New("required Lefthook version is higher than current")
)

func Version(verbose bool) string {
	if verbose {
		return version + " " + commit
	}

	return version
}

// CheckCovered returns true if given version is less or equal than current
// and false otherwise.
func CheckCovered(targetVersion string) error {
	if len(targetVersion) == 0 {
		return nil
	}

	if !versionRegexp.MatchString(targetVersion) {
		return errIncorrectVersion
	}

	major, minor, patch, err := parseVersion(version)
	if err != nil {
		return err
	}

	tMajor, tMinor, tPatch, err := parseVersion(targetVersion)
	if err != nil {
		return err
	}

	switch {
	case major > tMajor:
		return nil
	case major < tMajor:
		return errUncovered
	case minor > tMinor:
		return nil
	case minor < tMinor:
		return errUncovered
	case patch >= tPatch:
		return nil
	default:
		return errUncovered
	}
}

// parseVersion parses the version string of "1.2.3", "1.2", or just "1" and
// returns the major, minor and patch versions accordingly.
func parseVersion(version string) (major, minor, patch int, err error) {
	matches := versionRegexp.FindStringSubmatch(version)

	majorID := versionRegexp.SubexpIndex("major")
	minorID := versionRegexp.SubexpIndex("minor")
	patchID := versionRegexp.SubexpIndex("patch")

	major, err = strconv.Atoi(matches[majorID])

	if len(matches[minorID]) > 0 {
		minor, err = strconv.Atoi(matches[minorID])
	}

	if len(matches[patchID]) > 0 {
		patch, err = strconv.Atoi(matches[patchID])
	}

	return
}
