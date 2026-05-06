package logger

import (
	"fmt"
	"strings"
)

// Builder introduces a way to build a multiline log message and print it at once.
type Builder struct {
	logger  logger
	level   Level
	builder strings.Builder
	prefix  string
}

type logger interface {
	log(Level, ...any)
}

func NewBuilder(logger logger) *Builder {
	return &Builder{
		logger:  logger,
		level:   LevelInfo,
		builder: strings.Builder{},
	}
}

func (b *Builder) WithPrefix(prefix string) *Builder {
	b.prefix = prefix
	b.builder.WriteString(b.prefix)
	return b
}

func (b *Builder) WithLevel(level Level) *Builder {
	b.level = level
	return b
}

func (b *Builder) WriteLines(prefix string, out any) *Builder {
	var lines []string
	switch v := out.(type) {
	case string:
		lines = strings.Split(strings.TrimSpace(v), "\n")
	case []string:
		lines = v
	default:
		lines = strings.Split(fmt.Sprint(out), "\n")
	}

	var i int
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		if i == 0 {
			b.builder.WriteString(strings.Repeat(" ", len(b.prefix)))
			b.builder.WriteString(prefix)
		} else {
			b.builder.WriteString(strings.Repeat(" ", len(b.prefix)+len(prefix)))
		}

		b.builder.WriteString(line)
		b.builder.WriteString("\n")
		i++
	}

	return b
}

func (b *Builder) Log() {
	b.logger.log(b.level, b.builder.String())
}
