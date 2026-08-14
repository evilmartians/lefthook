package utils

// Intersect returns true if values of two slices have at least one similar value.
func Intersect[K comparable](a, b []K) bool {
	intersections := make(map[K]struct{}, len(a))

	for _, v := range a {
		intersections[v] = struct{}{}
	}

	for _, v := range b {
		if _, ok := intersections[v]; ok {
			return true
		}
	}

	return false
}
