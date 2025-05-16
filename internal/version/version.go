package version

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
)

const version = "1.11.13"

var (
	// Is set via -X github.com/evilmartians/lefthook/internal/version.commit={commit}.
	commit string

	versionRegexp = regexp.MustCompile(
		`^(?P<major>\d+)(?:\.(?P<minor>\d+)(?:\.(?P<patch>\d+))?)?$`,
	)

	ErrInvalidVersion   = errors.New("invalid version format")
	ErrUncoveredVersion = errors.New("version is lower than required")

	errInvalidMinVersion = errors.New("format of 'min_version' setting is incorrect")
)

func Version(verbose bool) string {
	if verbose {
		return version + " " + commit
	}

	return version
}

func Check(wanted, given string) error {
	if !versionRegexp.MatchString(given) {
		return ErrInvalidVersion
	}

	major, minor, patch, err := parseVersion(given)
	if err != nil {
		return ErrInvalidVersion
	}

	wantMajor, wantMinor, wantPatch, err := parseVersion(wanted)
	if err != nil {
		return ErrInvalidVersion
	}

	switch {
	case major > wantMajor:
		return nil
	case major < wantMajor:
		return ErrUncoveredVersion
	case minor > wantMinor:
		return nil
	case minor < wantMinor:
		return ErrUncoveredVersion
	case patch >= wantPatch:
		return nil
	default:
		return ErrUncoveredVersion
	}
}

// CheckCovered returns true if given version is less or equal than current
// and false otherwise.
func CheckCovered(targetVersion string) error {
	if len(targetVersion) == 0 {
		return nil
	}

	err := Check(targetVersion, version)

	if errors.Is(err, ErrUncoveredVersion) {
		execPath, oserr := os.Executable()
		if oserr != nil {
			execPath = "<unknown>"
		}

		return fmt.Errorf("required lefthook version (%s) is higher than current (%s) at %s", targetVersion, version, execPath)
	} else if errors.Is(err, ErrInvalidVersion) {
		return errInvalidMinVersion
	}

	return err
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
