package cmd

import (
	"docker/container/container_init"
	"github.com/urfave/cli"
)

var InitCmd = cli.Command{
	Name: "init",
	Usage: `Init container process, and run user process in that.`,
	Action: initCmdAction,
}

// logic entry for `mDocker init` command
func initCmdAction(ctx *cli.Context) error {
	return container_init.ContainerProcessInit()
}