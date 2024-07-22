//go:build !no_upgrade

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/evilmartians/lefthook/internal/lefthook"
)

var commands = [...]func(*lefthook.Options) *cobra.Command{
	newVersionCmd,
	newAddCmd,
	newInstallCmd,
	newUninstallCmd,
	newRunCmd,
	newDumpCmd,
	newUpgradeCmd,
}
