package cmd

import (
	"github.com/urfave/cli/v3"

	ver "github.com/evilmartians/lefthook/internal/version"
)

func Lefthook() *cli.Command {
	return &cli.Command{
		Name:                  "lefthook",
		Usage:                 "Git hooks manager",
		Version:               ver.Version(true),
		Commands:              commands,
		EnableShellCompletion: true,
	}
}
