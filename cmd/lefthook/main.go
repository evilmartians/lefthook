package main

import (
	"os"

	"github.com/evilmartians/lefthook/pkg/cmd"
)

func main() {
	rootCmd := cmd.NewRootCmd()

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
