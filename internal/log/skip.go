package log

const (
	skipMeta      = 0b00001
	skipSuccess   = 0b00010
	skipFailure   = 0b00100
	skipSummary   = 0b01000
	skipExecution = 0b10000
)

type SkipLogSettings struct {
	skips int8
}

func (s *SkipLogSettings) ApplySetting(setting string) {
	switch setting {
	case "meta":
		s.skips |= skipMeta
	case "success":
		s.skips |= skipSuccess
	case "failure":
		s.skips |= skipFailure
	case "summary":
		s.skips |= skipSummary
	case "execution":
		s.skips |= skipExecution
	}
}

func (s SkipLogSettings) SkipSuccess() bool {
	return s.doSkip(skipSuccess)
}

func (s SkipLogSettings) SkipFailure() bool {
	return s.doSkip(skipFailure)
}

func (s SkipLogSettings) SkipSummary() bool {
	return s.doSkip(skipSummary)
}

func (s SkipLogSettings) SkipMeta() bool {
	return s.doSkip(skipMeta)
}

func (s SkipLogSettings) SkipExecution() bool {
	return s.doSkip(skipExecution)
}

func (s SkipLogSettings) doSkip(option int8) bool {
	return s.skips&option != 0
}
