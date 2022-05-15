package cmd

import (
	"docker/container/container_info"
	"docker/container/network"
	"fmt"
	"github.com/urfave/cli"
	"strconv"
	"syscall"
)

var StopCommand = cli.Command{
	Name:   "stop",
	Usage:  "stop a container",
	Action: stopCmdAction,
}

// mdocker命令逻辑入口
func stopCmdAction(ctx *cli.Context) error {
	// 缺少containerName
	if len(ctx.Args()) < 1 {
		return fmt.Errorf("lack of containerName")
	}

	containerName := ctx.Args().Get(0)

	return stopContainer(containerName)
}

// 停止容器
func stopContainer(containerName string) error {
	// 通过containerName, 读取containerInfo获取pid
	cInfo, err := container_info.GetContainerInfoByContainerName(containerName)
	if err != nil {
		return err
	}
	pid, err := strconv.Atoi(cInfo.Pid)
	if err != nil {
		return fmt.Errorf("pid parse error, %v", err)
	}

	// 给container的init进程(pid=1的进程)发送终止信号
	if err = syscall.Kill(pid, syscall.SIGTERM); err != nil {
		return fmt.Errorf("stop container error %v", err)
	}

	// 清除容器的网络环境
	if err = network.NetworkManager.DisConnect(cInfo); err != nil {
		return fmt.Errorf("network disconnect error, %v", err)
	}

	cInfo.Pid = ""
	cInfo.Status = container_info.StatusStop
	if err = container_info.UpdateContainerInfo(containerName, cInfo); err != nil {
		return fmt.Errorf("container info update error, %v", err)
	}

	return nil
}
