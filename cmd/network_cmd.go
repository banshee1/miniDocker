package cmd

import (
	"docker/container/network"
	"fmt"
	"github.com/urfave/cli"
)

// mdocker network命令
var NetworkCmd = cli.Command{
	Name:  "network",
	Usage: "container network commands",
	Subcommands: []cli.Command{
		{
			Name:  "create",
			Usage: "create a container network",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "driver",
					Usage: "network driver",
				},
				cli.StringFlag{
					Name:  "subnet",
					Usage: "subnet cidr",
				},
			},
			Action: NetworkCreateAction,
		},
		{
			Name:   "list",
			Usage:  "list container network",
			Action: NetworkListAction,
		},
		{
			Name:   "remove",
			Usage:  "remove container network",
			Action: NetworkRemoveAction,
		},
	},
}

// mdocker network create命令
func NetworkCreateAction(ctx *cli.Context) error {
	if len(ctx.Args()) < 1 {
		return fmt.Errorf("missing network name")
	}

	err := network.NetworkManager.CreateNetwork(
		ctx.String("driver"), ctx.String("subnet"), ctx.Args()[0])
	if err != nil {
		return fmt.Errorf("create network error: %+v", err)
	}

	return nil
}

// mdocker network create命令
func NetworkListAction(ctx *cli.Context) error {
	network.NetworkManager.ListNetwork()

	return nil
}

// mdocker network remove 命令
func NetworkRemoveAction(ctx *cli.Context) error {
	if len(ctx.Args()) < 1 {
		return fmt.Errorf("missing network name")
	}

	err := network.NetworkManager.DeleteNetwork(ctx.Args()[0])
	if err != nil {
		return fmt.Errorf("remove network error: %+v", err)
	}
	return nil
}
