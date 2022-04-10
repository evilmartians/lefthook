package version

import (
	"regexp"
	"strconv"
)

const Version = "1.0.0"

var versionRegexp = regexp.MustCompile(
	`^(?P<major>\d+)(?:\.(?P<minor>\d+)(?:\.(?P<patch>\d+))?)?$`,
)

// IsCovered returns true if given version is less or equal than current
// and false otherwise.
func IsCovered(targetVersion string) bool {
	if len(targetVersion) == 0 {
		return true
	}

	if !versionRegexp.MatchString(targetVersion) {
		return false
	}

	major, minor, patch, err := parseVersion(Version)
	if err != nil {
		return false
	}

	tMajor, tMinor, tPatch, err := parseVersion(targetVersion)
	if err != nil {
		return false
	}

	switch {
	case major > tMajor:
		return true
	case major < tMajor:
		return false
	case minor > tMinor:
		return true
	case minor < tMinor:
		return false
	case patch >= tPatch:
		return true
	default:
		return false
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
