package version

const version = "1.0.0"

func Version() string {
	return version
}

// IsCovered returns true if given version is less or equal than current
// and false otherwise.
func IsCovered(ver string) bool {
	// Implement comparing the versions
	return true
}
