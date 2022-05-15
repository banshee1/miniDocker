package cmd

import (
	"docker/config"
	"docker/container/container_info"
	_ "docker/container/nsenter"
	"docker/utils"
	"fmt"
	"github.com/urfave/cli"
	"os"
	"os/exec"
	"strings"
)

var ExecCommand = cli.Command{
	Name:   "exec",
	Usage:  "exec a command into a container",
	Action: execCmdAction,
}

// mdocker exec 命令逻辑入口
func execCmdAction(ctx *cli.Context) error {
	//This is for callback
	if os.Getenv(config.EnvExecPid) != "" {
		utils.LoggerUtil.Infof("pid callback pid %v", os.Getgid())

		return nil
	}

	if len(ctx.Args()) < 2 {
		return fmt.Errorf("missing container name or command")
	}
	containerName := ctx.Args()[0]
	commandArray := ctx.Args()[1:]

	return execContainer(containerName, commandArray)
}

// 在指定name的容器中执行comArr命令
func execContainer(containerName string, comArray []string) error {
	// 获取容器init进程的pid
	pid, err := container_info.GetContainerPidByName(containerName)
	if err != nil {
		return fmt.Errorf("exec container getContainerPidByName %s error %v", containerName, err)
	}
	// 拼接command
	cmdStr := strings.Join(comArray, " ")
	utils.LoggerUtil.Infof("container pid: %s, command: %s", pid, cmdStr)

	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err = os.Setenv(config.EnvExecPid, pid); err != nil {
		return fmt.Errorf("exec pid env set error: %v", err)
	}
	if err = os.Setenv(config.EnvExecCmd, cmdStr); err != nil {
		return fmt.Errorf("exec cmd env set error: %v", err)
	}

	if err = cmd.Run(); err != nil {
		return fmt.Errorf("exec container %s error %v", containerName, err)
	}

	return nil
}
