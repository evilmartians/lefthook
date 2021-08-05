package cmd

import (
	"log"
	"time"

	spin "github.com/briandowns/spinner"
)

const (
	spinnerRefreshRate time.Duration = 100 * time.Millisecond
	spinnerCharSet     int           = 14
)

type Spinner struct {
	extSpinner *spin.Spinner
	animate bool
}

func NewSpinner(animate bool) *Spinner {
	return &Spinner{
		spin.New(spin.CharSets[spinnerCharSet], spinnerRefreshRate),
		animate,
	}
}

func (s *Spinner) Start() {
	if s.animate {
		s.extSpinner.Suffix = " waiting"
		s.extSpinner.Start()
	} else {
		log.Println("waiting")
	}
}

func (s *Spinner) Stop() {
	if s.animate {
		s.extSpinner.Stop()
	}
}

func (s *Spinner) RestartWithMsg(msgs ...interface{}) {
	hasMsg := !(len(msgs) == 1 && msgs[0] == "")

	if !s.animate {
		if hasMsg {
			log.Println(msgs...)
		}
		log.Println("waiting")
	} else {
		s.extSpinner.Stop()

		if hasMsg {
			log.Println(msgs...)
		}

		s.extSpinner.Start()
	}
}
