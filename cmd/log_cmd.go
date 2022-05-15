package cmd

import (
	"docker/container/container_info"
	"fmt"
	"github.com/urfave/cli"
	"io/ioutil"
	"os"
)

var LogCommand = cli.Command{
	Name:   "logs",
	Usage:  "print logs of a container",
	Action: logCmdAction,
}

// mdocker logs 命令逻辑入口
func logCmdAction(ctx *cli.Context) error {
	if len(ctx.Args()) < 1 {
		return fmt.Errorf("missing container name")
	}

	containerName := ctx.Args().Get(0)

	return logContainer(containerName)
}

// 打印容器日志内容
func logContainer(containerName string) error {
	// 获取日志文件路径
	logFileLocation := container_info.GetContainerLogFilePath(containerName)
	file, err := os.Open(logFileLocation)
	defer file.Close()
	if err != nil {
		return fmt.Errorf("log container open file %s error %v", logFileLocation, err)
	}

	content, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("log container read file %s error %v", logFileLocation, err)
	}

	_, _ = fmt.Fprint(os.Stdout, string(content))

	return nil
}
