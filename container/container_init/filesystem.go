package container_init

import (
	"docker/config"
	"docker/utils"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
)

// region 容器初始化, 创建文件系统

// 创建容器的工作目录
func newWorkSpace(containerName, imageName, volume string) (string, error) {
	// 判断镜像是否存在
	imagePath := getImagePath(imageName)
	imageExists, err := utils.GeneralUtils.IsDirExists(imagePath)
	if err != nil {
		return "", fmt.Errorf("image path existence judge fail, %v", err)
	}
	if imageExists == false {
		return "", fmt.Errorf("image not exists")
	}

	// 创建aufs读写层branch
	if err = createOverlay2Layers(containerName); err != nil {
		return "", err
	}

	// aufs联合挂载
	mntPath, err := createMountPoint(containerName, imagePath)
	if err != nil {
		return "", err
	}

	// 挂载用户指定volume
	if err = handleUserVolume(getMntPointPath(containerName), volume); err != nil {
		return "", err
	}

	return mntPath, nil
}

// 创建容器读写层
func createOverlay2Layers(containerName string) error {
	// rwdir
	writeURL := getRwLayerPath(containerName)
	if err := os.Mkdir(writeURL, 0777); err != nil {
		return fmt.Errorf("mkdir dir %s error. %v", writeURL, err)
	}

	// workdir
	workPath := getWorkDirPath(containerName)
	if err := os.Mkdir(workPath, 0777); err != nil {
		return fmt.Errorf("mkdir dir %s error: $v", workPath, err)
	}

	return nil
}

// 使用aufs挂载容器文件视图
func createMountPoint(containerName, imagePath string) (string, error) {
	// 创建挂载点
	mntPath := getMntPointPath(containerName)
	if err := os.Mkdir(mntPath, 0777); err != nil {
		return "", fmt.Errorf("mkdir dir %s error. %v", mntPath, err)
	}

	// 挂载unionFs
	dirs := getContainerMountParam(containerName, imagePath)
	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", dirs, mntPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("mount aufs failed, %v", err)
	}

	return mntPath, nil
}

// 处理用户volume挂载
func handleUserVolume(mntPath, volume string) error {
	if volume == "" {
		return nil
	}

	volumeUrls := strings.Split(volume, ":")
	if len(volumeUrls) != 2 || volumeUrls[0] == "" || volumeUrls[1] == "" {
		return fmt.Errorf("invalid volume params: %v", volumeUrls)
	}

	return doMountVolumeOpr(mntPath, volumeUrls)
}

// 挂载用户指定volume目录到container挂载点下
func doMountVolumeOpr(mntPath string, volumeUrls []string) error {
	// 尝试创建宿主机目录
	if err := os.Mkdir(volumeUrls[0], 0777); err != nil && !os.IsExist(err) { // 非重复目录报错异常
		return fmt.Errorf("host path create fail: %v", err)
	}

	// 创建容器目录
	containerPath := path.Join(mntPath, volumeUrls[1])
	if err := os.Mkdir(containerPath, 0777); err != nil {
		return fmt.Errorf("mkdir container volume dir error: %v", err)
	}

	// 利用aufs进行挂载
	dirParam := fmt.Sprintf("upperdir=%s", volumeUrls[0])
	cmd := exec.Command("mount", "-t", "overlay", "overlay", "-o", dirParam, containerPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("user volume mount error: %v", err)
	}

	return nil
}

// endregion

// region 容器退出, 清理容器文件系统

// DeleteWorkSpace 容器退出时候删除aufs挂载目录
func DeleteWorkSpace(containerName, volume string) error {
	// 清楚用户挂载volume
	mntPath := getMntPointPath(containerName)
	// 需要先取消用户挂载目录, 再取消根目录挂载
	if err := deleteUserVolume(mntPath, volume); err != nil {
		return err
	}

	// 取消容器文件系统挂载
	if err := deleteMountPoint(mntPath); err != nil {
		return err
	}

	// 删除相关挂载目录
	delDirs := []string{getRwLayerPath(containerName), getWorkDirPath(containerName), getMntPointPath(containerName)}
	for _, dir := range delDirs {
		// 删除容器文件系统挂载点
		if err := rmDirAll(dir); err != nil {
			return err
		}
	}

	return nil
}

func deleteUserVolume(mntUrl, volume string) error {
	if volume == "" {
		return nil
	}

	volumeUrls := strings.Split(volume, ":")
	if len(volumeUrls) != 2 || volumeUrls[0] == "" || volumeUrls[1] == "" { // 跳过, 无挂载
		return nil
	}

	containerUrl := path.Join(mntUrl, volumeUrls[1])
	cmd := exec.Command("umount", containerUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("volume umount error %v", err)
	}

	return nil
}

// 取消容器挂载视图
func deleteMountPoint(mntUrl string) error {
	cmd := exec.Command("umount", mntUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%v", err)
	}

	return nil
}

func rmDirAll(rwLayerPath string) error {
	if err := os.RemoveAll(rwLayerPath); err != nil {
		return fmt.Errorf("remove dir %s error %v", rwLayerPath, err)
	}

	return nil
}

// endregion

// region 路径获取方法

// 获取指定镜像路径
func getImagePath(imageName string) string {
	return path.Join(config.PathImage, imageName)
}

// 获取container rw layer路径
func getRwLayerPath(containerIdent string) string {
	return path.Join(config.PathReadWrite, containerIdent)
}

// 获取overlay2 workdir路径
func getWorkDirPath(containerIdent string) string {
	return path.Join(config.PathWorkDir, containerIdent)
}

// 获取container挂载点路径
func getMntPointPath(containerIdent string) string {
	return path.Join(config.PathMnt, containerIdent)
}

// 获取容器挂载参数
func getContainerMountParam(containerName, imagePath string) string {
	rwPath := getRwLayerPath(containerName)
	workDirPath := getWorkDirPath(containerName)

	return fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", imagePath, rwPath, workDirPath)
}

// endregion
