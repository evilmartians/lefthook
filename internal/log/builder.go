package log

import (
	"fmt"
	"strings"
)

type Builder interface {
	Add(string, interface{}) Builder
	Log()
}

type dummyBuilder struct{}

type debugBuilder struct {
	builder strings.Builder
}

func DebugBuilder() Builder {
	if !std.IsLevelEnabled(DebugLevel) {
		return dummyBuilder{}
	}

	return &debugBuilder{
		builder: strings.Builder{},
	}
}

func (b *debugBuilder) Add(prefix string, data interface{}) Builder {
	var lines []string
	switch v := data.(type) {
	case string:
		lines = strings.Split(v, "\n")
	case []string:
		lines = v
	default:
		lines = strings.Split(fmt.Sprint(data), "\n")
	}
	for i, line := range lines {
		if i == 0 {
			b.builder.WriteString(prefix + line + "\n")
		} else {
			b.builder.WriteString(strings.Repeat(" ", len(prefix)) + line + "\n")
		}
	}

	return b
}

func (b *debugBuilder) Log() {
	Debug(b.builder.String())
}

func (d dummyBuilder) Add(_ string, _ interface{}) Builder { return d }
func (dummyBuilder) Log()                                  {}
