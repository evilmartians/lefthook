package loggertest_test

import (
	"testing"

	"github.com/evilmartians/lefthook/v2/tests/helpers/loggertest"
)

func TestNew(t *testing.T) {
	l := loggertest.New()

	if !l.NoColors() {
		t.Errorf("loggertest.New().NoColors() = false, want true")
	}
}

func TestNewWithColors(t *testing.T) {
	l := loggertest.NewWithColors()

	if l.NoColors() {
		t.Errorf("loggertest.NewWithColors().NoColors() = true, want false")
	}

	if !l.ColorsForced() {
		t.Errorf("loggertest.NewWithColors().ColorsForced() = false, want true")
	}
}

func TestNewExecution(t *testing.T) {
	if l := loggertest.NewExecution(); l == nil {
		t.Errorf("loggertest.NewExecution() = nil, want not nil")
	}
}
