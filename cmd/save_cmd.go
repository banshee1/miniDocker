package cmd

import (
	"docker/config"
	"docker/utils"
	"fmt"
	"github.com/urfave/cli"
	"os/exec"
	"path"
)

var SaveCommand = cli.Command{
	Name:   "save",
	Usage:  "save a container into image",
	Action: saveCmdAction,
}

// mdocker save命令逻辑入口
func saveCmdAction(ctx *cli.Context) error {
	if len(ctx.Args()) < 2 {
		return fmt.Errorf("invalid params")
	}

	containerName := ctx.Args().Get(0)
	imagePath := ctx.Args().Get(1)

	return saveContainerIntoTar(containerName, imagePath)
}

// 将指定容器打包成tar
func saveContainerIntoTar(containerName, imagePath string) error {
	// 判断容器目录是否存在
	containerMntPath := path.Join(config.PathMnt, containerName)
	pathExists, err := utils.GeneralUtils.IsDirExists(containerMntPath)
	if err != nil {
		return fmt.Errorf("container mnt path existence judge fail, %v", err)
	}
	if !pathExists {
		return fmt.Errorf("container mnt path not exists")
	}

	utils.LoggerUtil.Infof("image save path: %s", imagePath)

	_, err = exec.Command(
		"tar", "-czf", imagePath, "-C", containerMntPath, ".",
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("image save fail: %v", err)
	}

	return err
}
