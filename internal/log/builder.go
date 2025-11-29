package log

import (
	"fmt"
	"strings"
)

type builder interface {
	Add(string, any) builder
	String() string
	Log()
}

type dummyBuilder struct{}

type logBuilder struct {
	level   Level
	prefix  string
	builder strings.Builder
}

func Builder(level Level, prefix string) builder {
	if !std.IsLevelEnabled(level) {
		return dummyBuilder{}
	}

	return &logBuilder{
		prefix:  prefix,
		level:   level,
		builder: strings.Builder{},
	}
}

func (b *logBuilder) Add(prefix string, data any) builder {
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
		if len(line) == 0 {
			continue
		}

		switch {
		case b.builder.Len() == 0:
			b.builder.WriteString(b.prefix + prefix + line + "\n")
		case i == 0:
			b.builder.WriteString(strings.Repeat(" ", len(b.prefix)) + prefix + line + "\n")
		default:
			b.builder.WriteString(strings.Repeat(" ", len(b.prefix)+len(prefix)) + line + "\n")
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

func (d dummyBuilder) Add(_ string, _ any) builder { return d }
func (dummyBuilder) Log()                          {}
func (dummyBuilder) String() string                { return "" }
