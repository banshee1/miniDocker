package main

import (
	_ "docker/boot"
	"docker/cmd"
	"docker/config"
	"docker/utils"
	"github.com/urfave/cli"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = config.AppName
	app.Usage = config.Usage

	// 绑定commands
	app.Commands = []cli.Command{
		cmd.RunCommand,
		cmd.InitCmd,
		cmd.SaveCommand,
		cmd.LoadCommand,
		cmd.ListCommand,
		cmd.LogCommand,
		cmd.ExecCommand,
		cmd.StopCommand,
		cmd.RmCommand,
		cmd.NetworkCmd,
	}

	if err := app.Run(os.Args); err != nil {
		utils.LoggerUtil.Fatalf(err.Error())
	}
}
