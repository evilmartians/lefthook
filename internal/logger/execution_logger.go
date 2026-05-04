package logger

import "strings"

type ExecutionLogger struct {
	logger   *Logger
	settings *ExecutionSettings
}

func (l *logger) NewExecutionLogger(configs ...any) *ExecutionLogger {
	settings := NewExecutionSettings()

	for _, config := range configs {
		switch c := config.(type) {
		case bool:
			if c {
				settings.enable(executionFull)
			}
		case []any:
			for _, option := range c {
				name, ok := option.(string)
				if !ok {
					logger.Warnf("Unknown output setting: %v", option)
					continue
				}

				setting, err := nameToSetting(name)
				if err != nil {
					logger.Warn(err)
					continue
				}
				settings.enable(setting)
			}
		case string:
			names := strings.Split(name, ",")
			for _, name := range names {
				setting, err := nameToSetting(name)
				if err != nil {
					logger.Warn(err)
					fallthrough
				}
				settings.enable(setting)
			}
		default:
			settings.enable(executionFull)
		}
	}

	return &ExecutionLogger{
		logger:   logger,
		settings: settings,
	}
}

func (el *ExecutionLogger) Enabled(setting ExecutionSetting) bool {
	return el.settings.enabled(setting)
}
