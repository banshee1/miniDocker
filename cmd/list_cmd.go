package cmd

import (
	"docker/container/container_info"
	"fmt"
	"github.com/urfave/cli"
	"os"
	"text/tabwriter"
)

// ListCommand `mdocker ps`命令定义
var ListCommand = cli.Command{
	Name:   "ps",
	Usage:  `list and print the info of all containers`,
	Action: listCmdAction,
}

// `mdocker ps`命令主逻辑入口
func listCmdAction(ctx *cli.Context) error {
	// 获取所有容器信息
	containers, err := container_info.GetContainerInfoAll()
	if err != nil {
		return err
	}

	// 打印容器信息
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	_, _ = fmt.Fprint(w, "ID\tNAME\tPID\tSTATUS\tCOMMAND\tCREATED\n")
	for _, item := range containers {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Id,
			item.Name,
			item.Pid,
			item.Status,
			item.Command,
			item.CreatedTime)
	}
	if err = w.Flush(); err != nil {
		return fmt.Errorf("tabwriter flush error %v", err)
	}

	return nil
}
