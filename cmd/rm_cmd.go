package cmd

import (
	"docker/container/cgroups"
	"docker/container/container_info"
	"docker/container/container_init"
	"fmt"
	"github.com/urfave/cli"
	"os"
)

var RmCommand = cli.Command{
	Name:   "rm",
	Usage:  "delete a container info",
	Action: rmCmdAction,
}

// mdocker rm 命令逻辑入口
func rmCmdAction(ctx *cli.Context) error {
	if len(ctx.Args()) < 1 {
		return fmt.Errorf("missing container name")
	}

	containerName := ctx.Args().Get(0)

	return removeContainer(containerName)
}

// 删除指定name的容器
func removeContainer(containerName string) error {
	containerInfo, err := container_info.GetContainerInfoByContainerName(containerName)
	if err != nil {
		return err
	}

	if containerInfo.Status == container_info.StatusRunning {
		return fmt.Errorf("cannot remove running container")
	}

	containerInfoDir := container_info.GetContainerInfoDirPath(containerName)
	if err = os.RemoveAll(containerInfoDir); err != nil {
		return fmt.Errorf("container remove error, %v", err)
	}

	// 移除cgroup path
	cgroups.RemoveContainerCgroup(containerName)
	// 移除文件系统
	if err = container_init.DeleteWorkSpace(containerInfo.Name, containerInfo.Volume); err != nil {
		return err
	}

	return nil
}
