//go:build no_self_update && !jsonschema

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/evilmartians/lefthook/internal/command"
)

type cmd interface {
	New(*command.Options) *cobra.Command
}

var commands = [...]cmd{
	version{},
	add{},
	install{},
	checkInstall{},
	uninstall{},
	run{},
	dump{},
	validate{},
}
