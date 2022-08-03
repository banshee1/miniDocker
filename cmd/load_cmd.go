package cmd

import (
	"docker/config"
	"docker/utils"
	"fmt"
	"github.com/urfave/cli"
	"os"
	"os/exec"
	"path"
)

var LoadCommand = cli.Command{
	Name:   "load",
	Usage:  "load a image from tar pack",
	Action: loadCmdAction,
}

// mdocker load命令主逻辑入口
func loadCmdAction(ctx *cli.Context) error {
	if len(ctx.Args()) < 2 {
		return fmt.Errorf("invalid params")
	}

	imagePackPath := ctx.Args().Get(0)
	imageName := ctx.Args().Get(1)

	return loadImage(imagePackPath, imageName)
}

// 加载镜像
func loadImage(imagePackPath, imageName string) error {
	imagePath := path.Join(config.PathImage, imageName)
	pathExists, err := utils.GeneralUtils.IsDirExists(imagePath)
	if err != nil {
		return fmt.Errorf("image path existence judge fail, %v", err)
	}
	if pathExists {
		return fmt.Errorf("image already exists")
	}

	// 创建image path
	err = os.MkdirAll(imagePath, 0755)
	if err != nil {
		return fmt.Errorf("image path mkdir error, %v", err)
	}

	// 解压镜像
	_, err = exec.Command("tar", "-xvf", imagePackPath, "-C", imagePath).CombinedOutput()
	if err != nil {
		return fmt.Errorf("image load error, %v", err)
	}

	return nil
}
