package utils

// FirstNonBlank returns first non-empty string from given args.
func FirstNonBlank(args ...string) string {
	for _, a := range args {
		if len(a) > 0 {
			return a
		}
	}

	return ""
}
