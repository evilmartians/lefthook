package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/evilmartians/lefthook/v2/cmd"
	"github.com/evilmartians/lefthook/v2/internal/command"
)

func main() {
	if err := cmd.Lefthook().Run(context.Background(), os.Args); err != nil {
		var hookExit *command.HookExitError
		if errors.As(err, &hookExit) {
			os.Exit(hookExit.Code)
		}
		if err.Error() != "" {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		}
		os.Exit(1)
	}
}
