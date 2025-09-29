package cmd

import (
	"github.com/urfave/cli/v3"
)

func Lefthook() *cli.Command {
	return &cli.Command{
		Name:                  "lefthook",
		Commands:              commands,
		EnableShellCompletion: true,
	}
}
