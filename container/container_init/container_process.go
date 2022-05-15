package container_init

import (
	"docker/config"
	"docker/container/container_info"
	"docker/utils"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// NewContainerProcess 重新创建容器父进程:
// 			1. 创建Namespace
//			2. 创建一个fifo管道, 将管道的读取端fd设置给新创建的init进程, 返回读取端fd供写入参数
func NewContainerProcess(tty bool, volume, containerName, imageName string) (*exec.Cmd, *os.File, error) {
	// 尝试创建获取一个pipe
	readPipe, writePipe, err := newPipe()
	if err != nil {
		return nil, nil, err
	}

	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: config.ProcessCloneFlags,
	}

	// 处理init进程的io
	if err = procInitProcessIO(cmd, tty, containerName); err != nil {
		return nil, nil, err
	}

	// 将创建的pipe读取端fd赋给init process
	cmd.ExtraFiles = []*os.File{readPipe}

	// 在指定挂载点上创建容器的文件视图
	mntPath, err := newWorkSpace(containerName, imageName, volume)
	if err != nil {
		return nil, nil, err
	}
	cmd.Dir = mntPath // 修改init cmd的执行目录为挂载点

	return cmd, writePipe, nil
}

// create and return a pipeline
func newPipe() (*os.File, *os.File, error) {
	return os.Pipe()
}

// 设置container init进程的io
func procInitProcessIO(cmd *exec.Cmd, tty bool, containerName string) error {
	if tty {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		return nil
	}

	containerInfoPath := container_info.GetContainerInfoDirPath(containerName)
	if err := os.MkdirAll(containerInfoPath, 0622); err != nil {
		return fmt.Errorf("NewParentProcess mkdir %s error %v", containerInfoPath, err)
	}
	stdLogFilePath := container_info.GetContainerLogFilePath(containerName)
	stdLogFile, err := os.Create(stdLogFilePath)
	if err != nil {
		return fmt.Errorf("NewParentProcess create file %s error %v", stdLogFilePath, err)
	}
	cmd.Stdout = stdLogFile

	return nil
}

// ContainerProcessInit 容器init进程初始化
func ContainerProcessInit() error {
	utils.LoggerUtil.Infof("container init start")

	// process will stuck here waiting for the reading the pipe
	cmdArray := readUserCommand()
	if cmdArray == nil || len(cmdArray) == 0 {
		return fmt.Errorf("container init command read nil")
	}

	// 初始化容器mount
	if err := mountInit(); err != nil {
		return fmt.Errorf("mount init failed: %v", err)
	}

	// look up for the absolute path of cmd in the current PATH env var
	cmdPath, err := exec.LookPath(cmdArray[0])
	if err != nil {
		return err
	}

	// call exec to replace current process, cmdArray: exec file path(only used for display), params...
	if err = syscall.Exec(cmdPath, cmdArray[0:], os.Environ()); err != nil {
		utils.LoggerUtil.Fatalf("mount fail: %s", err.Error())
	}

	return nil
}

func readUserCommand() []string {
	// open the 4th fd of the process
	readPipe := os.NewFile(uintptr(3), "pipe")
	msg, err := ioutil.ReadAll(readPipe)
	if err != nil {
		utils.LoggerUtil.Errorf("init read pipe error: %v", err)

		return nil
	}

	return strings.Split(string(msg), " ")
}
