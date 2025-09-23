package command

// SkipError implements error interface but indicates that the execution needs to be skipped.
type SkipError struct {
	reason string
}

func (r SkipError) Error() string {
	return r.reason
}
