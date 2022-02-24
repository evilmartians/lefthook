package main

import (
	"os"

	"github.com/evilmartians/lefthook/internal/log"
)

func main() {
	rootCmd := newRootCmd()

	if err := rootCmd.Execute(); err != nil {
		log.Errorf("Error: %s", err)
		os.Exit(1)
	}
}
