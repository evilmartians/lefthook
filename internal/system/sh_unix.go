//go:build !windows

package system

// Sh returns `sh` executable name.
func Sh() (string, error) {
	return "sh", nil
}
