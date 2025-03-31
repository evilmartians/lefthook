package log

import (
	"fmt"
	"strings"
)

type builder interface {
	Add(string, interface{}) builder
	String() string
	Log()
}

type dummyBuilder struct{}

type logBuilder struct {
	level   Level
	builder strings.Builder
}

func Builder(level Level) builder {
	if !std.IsLevelEnabled(level) {
		return dummyBuilder{}
	}

	return &logBuilder{
		level:   level,
		builder: strings.Builder{},
	}
}

func (b *logBuilder) Add(prefix string, data interface{}) builder {
	var lines []string
	switch v := data.(type) {
	case string:
		lines = strings.Split(strings.TrimSpace(v), "\n")
	case []string:
		lines = v
	default:
		lines = strings.Split(fmt.Sprint(data), "\n")
	}
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if i == 0 {
			b.builder.WriteString(prefix + line + "\n")
		} else {
			b.builder.WriteString(strings.Repeat(" ", len(prefix)) + line + "\n")
		}
	}

	return b
}

func (b *logBuilder) Log() {
	switch b.level {
	case DebugLevel:
		Debug(b.builder.String())
	case InfoLevel:
		Info(b.builder.String())
	case ErrorLevel:
		Error(b.builder.String())
	case WarnLevel:
		Warn(b.builder.String())
	}
}

func (b *logBuilder) String() string {
	return b.builder.String()
}

func (d dummyBuilder) Add(_ string, _ interface{}) builder { return d }
func (dummyBuilder) Log()                                  {}
func (dummyBuilder) String() string                        { return "" }
