package cmd

import "github.com/evilmartians/lefthook/internal/log"

func Lefthook() int {
	rootCmd := newRootCmd()

	if err := rootCmd.Execute(); err != nil {
		if err.Error() != "" {
			log.Errorf("Error: %s", err)
		}
		return 1
	}

	return 0
}
