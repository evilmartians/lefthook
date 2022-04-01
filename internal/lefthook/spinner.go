package lefthook

import (
	"sync"
	"time"

	spin "github.com/briandowns/spinner"

	"github.com/evilmartians/lefthook/internal/log"
)

const (
	spinnerRefreshRate     = 100 * time.Millisecond
	spinnerCharSet     int = 14
)

type Spinner struct {
	sync.Mutex
	extSpinner *spin.Spinner
}

func NewSpinner() *Spinner {
	return &Spinner{sync.Mutex{}, spin.New(spin.CharSets[spinnerCharSet], spinnerRefreshRate)}
}

func (s *Spinner) Start() {
	s.extSpinner.Suffix = " waiting"
	s.extSpinner.Start()
}

func (s *Spinner) Stop() {
	s.extSpinner.Stop()
}

func (s *Spinner) RestartWithMsg(msgs ...interface{}) {
	s.Lock()
	defer s.Unlock()

	s.extSpinner.Stop()

	if len(msgs) == 1 && msgs[0] == "" {
		s.extSpinner.Start()
		return
	}

	log.Println(msgs...)
	s.extSpinner.Start()
}
