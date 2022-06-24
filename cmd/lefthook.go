package cmd

import (
	"os"

	"github.com/evilmartians/lefthook/internal/log"
)

func Lefthook() {
	rootCmd := newRootCmd()

	if err := rootCmd.Execute(); err != nil {
		if err.Error() != "" {
			log.Errorf("Error: %s", err)
		}
		os.Exit(1)
	}
}
