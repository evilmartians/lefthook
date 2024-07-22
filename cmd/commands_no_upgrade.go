//go:build no_upgrade

package cmd

import (
	"github.com/evilmartians/lefthook/internal/lefthook"
	"github.com/spf13/cobra"
)

var commands = [...]func(*lefthook.Options) *cobra.Command{
	newVersionCmd,
	newAddCmd,
	newInstallCmd,
	newUninstallCmd,
	newRunCmd,
	newDumpCmd,
}
