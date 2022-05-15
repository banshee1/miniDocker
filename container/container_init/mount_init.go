package container_init

import (
	"docker/config"
	"docker/utils"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

func mountInit() error {
	// 改变当前Namespace的Mount传播模式
	err := syscall.Mount("", "/", "", uintptr(config.MountFlagsPrivate), "")
	if err != nil {
		return err
	}

	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get pwd failed: %v", err)
	}
	utils.LoggerUtil.Infof("pwd is %s", pwd)

	// 调用pivot切换rootfs
	err = pivotRoot(pwd)
	if err != nil {
		return err
	}

	// mount /proc directory
	_ = os.Mkdir("/proc", 1555)
	err = syscall.Mount("proc", "/proc", "proc", uintptr(config.MountFlagsDefault), "")
	if err != nil {
		return fmt.Errorf("mount /proc failed: %v", err)
	}

	_ = os.Mkdir("/dev", 1755)
	err = syscall.Mount(
		"tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755",
	)
	if err != nil {
		return fmt.Errorf("mount /dev failed: %v", err)
	}

	return nil
}

// 调用pivotRoot将当前Namespace的 "/" 路径切换为空, 摆脱对宿主机 root目录依赖
func pivotRoot(root string) error {
	// pivotRoot要求put_old和root_new为不同类型文件系统, 所以利用bind重新mount一次
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("mount bind root error: %v", err)
	}

	// 创建 {rootfs}/.pivot_root 存储 old_root
	pivotDir := filepath.Join(root, ".pivot_root")
	if err := os.Mkdir(pivotDir, 0777); err != nil {
		return err
	}
	// 调用pivot_root 将{rootfs}和rootfs/.pivot_root交换
	// 此时挂载点现在依然可以在mount命令中看到
	utils.LoggerUtil.Infof("root: %s, pivot: %s", root, pivotDir)
	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return fmt.Errorf("pivot_root %v", err)
	}

	// 修改当前的工作目录到根目录
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir / %v", err)
	}

	pivotDir = filepath.Join("/", ".pivot_root")
	// umount rootfs/.pivot_root
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount pivot_root dir %v", err)
	}

	// 删除临时文件夹
	return os.Remove(pivotDir)
}
