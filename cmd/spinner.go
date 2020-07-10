package cmd

import (
	spin "github.com/briandowns/spinner"
	"log"
	"time"
)

const (
	spinnerRefreshRate time.Duration = 100 * time.Millisecond
	spinnerCharSet     int           = 14
)

type Spinner struct {
	extSpinner *spin.Spinner
}

func NewSpinner() *Spinner {
	return &Spinner{spin.New(spin.CharSets[spinnerCharSet], spinnerRefreshRate)}
}

func (s *Spinner) Start() {
	s.extSpinner.Suffix = " waiting"
	s.extSpinner.Start()
}

func (s *Spinner) Stop() {
	s.extSpinner.Stop()
}

func (s *Spinner) RestartWithMsg(msgs ...interface{}) {
	s.extSpinner.Stop()

	if len(msgs) == 1 && msgs[0] == "" {
		s.extSpinner.Start()
		return
	}

	log.Println(msgs...)
	s.extSpinner.Start()
}
