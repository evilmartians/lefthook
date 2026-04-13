package git

import "github.com/spf13/afero"

type Git struct {
	Fs        afero.Fs
	Commander *Commander
	HooksPath string
	RootPath  string
	GitPath   string
	InfoPath  string
}

func New(fs afero.Fs) *Git {
	return &Git{Fs: fs}
}

func (g *Git) Setup() error {
	return nil
}
