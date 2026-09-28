package command

type reason uint8

const (
	NoFiles reason = iota
	NoStagedFiles
	NoPushFiles
	InvalidScript
)

// SkipError implements error interface but indicates that the execution needs to be skipped.
type SkipError struct {
	reason reason
}

func (r SkipError) Error() string {
	switch r.reason {
	case NoFiles:
		return "no files for inspection"
	case NoStagedFiles:
		return "no matching staged files"
	case NoPushFiles:
		return "no matching push files"
	case InvalidScript:
		return "script files is not a regular file"
	}

	panic("unknown skip reason")
}
