package logger

import (
	"time"

	"github.com/briandowns/spinner"
)

const (
	spinnerCharSet     = 14
	spinnerRefreshRate = 100 * time.Millisecond
	spinnerText        = " waiting"
)

type Spinner struct {
	spinner *spinner.Spinner
	names   []string
}

func NewSpinner() *Spinner {
	return &Spinner{
		names: make([]string, 0, 10), // reduce extra allocations
		spinner: spinner.New(
			spinner.CharSets[spinnerCharSet],
			spinnerRefreshRate,
			spinner.WithSuffix(spinnerText),
		),
	}
}

func (s *Spinner) Start() {
	s.spinner.Start()
}

func (s *Spinner) Stop() {
	s.spinner.Stop()
}

func (s *Spinner) active() bool {
	return s.spinner.Active()
}
