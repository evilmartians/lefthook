//go:build no_self_update

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/evilmartians/lefthook/internal/lefthook"
)

type command interface {
	New(*lefthook.Options) *cobra.Command
}

var commands = [...]command{
	version{},
	add{},
	install{},
	uninstall{},
	run{},
	dump{},
}
