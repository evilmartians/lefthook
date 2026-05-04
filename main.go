package main

import (
	"context"
	"fmt"
	"os"

	"github.com/evilmartians/lefthook/v2/cmd"
)

func main() {
	if err := cmd.Lefthook().Run(context.Background(), os.Args); err != nil {
		if err.Error() != "" {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		}
		os.Exit(1)
	}
}
