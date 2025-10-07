package main

import (
	"context"
	"os"

	"github.com/evilmartians/lefthook/cmd"
	"github.com/evilmartians/lefthook/internal/log"
)

func main() {
	if err := cmd.Lefthook().Run(context.Background(), os.Args); err != nil {
		if err.Error() != "" {
			log.Errorf("Error: %s", err)
		}
		os.Exit(1)
	}
}
