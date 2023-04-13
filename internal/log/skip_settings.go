package log

const (
	skipMeta = 1 << iota
	skipSuccess
	skipFailure
	skipSummary
	skipExecution
	skipSkips
)

type SkipSettings int8

func (s *SkipSettings) ApplySetting(setting string) {
	switch setting {
	case "meta":
		*s |= skipMeta
	case "success":
		*s |= skipSuccess
	case "failure":
		*s |= skipFailure
	case "summary":
		*s |= skipSummary
	case "skips":
		*s |= skipSkips
	case "execution":
		*s |= skipExecution
	}
}

func (s SkipSettings) SkipSuccess() bool {
	return s.doSkip(skipSuccess)
}

func (s SkipSettings) SkipFailure() bool {
	return s.doSkip(skipFailure)
}

func (s SkipSettings) SkipSummary() bool {
	return s.doSkip(skipSummary)
}

func (s SkipSettings) SkipMeta() bool {
	return s.doSkip(skipMeta)
}

func (s SkipSettings) SkipExecution() bool {
	return s.doSkip(skipExecution)
}

func (s SkipSettings) SkipSkips() bool {
	return s.doSkip(skipSkips)
}

func (s SkipSettings) doSkip(option int8) bool {
	return int8(s)&option != 0
}
