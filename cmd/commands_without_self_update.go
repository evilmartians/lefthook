//go:build no_self_update && !jsonschema

package cmd

import "github.com/urfave/cli/v3"

var cliCommands = []*cli.Command{
	run(),
	install(),
	uninstall(),
	checkInstall(),
	dump(),
	add(),
	validate(),
	version(),
	// selfUpdate(),
}
