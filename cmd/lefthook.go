package cmd

import (
	"github.com/urfave/cli/v3"

	ver "github.com/evilmartians/lefthook/v2/internal/version"
)

func Lefthook() *cli.Command {
	return &cli.Command{
		Name:     "lefthook",
		Usage:    "Git hooks manager",
		Version:  ver.Version(true),
		Commands: commands,
		Description: `... of supported ENV variables:

LEFTHOOK         - set to '0' or 'false' to disable lefthook execution
LEFTHOOK_CONFIG  - override main config path
LEFTHOOK_OUTPUT  - control printed sections (see config option 'output')
LEFTHOOK_VERBOSE - enable debug logs`,
		EnableShellCompletion: true,
		Suggest:               true,
	}
}
