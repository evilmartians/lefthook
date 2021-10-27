package cmd

import (
	"github.com/spf13/afero"
)

// Application global options
type Options struct {
	fs afero.Fs

	Verbose  bool
	NoColors bool

	// Deprecated options
	Force      bool
	Aggressive bool
}
