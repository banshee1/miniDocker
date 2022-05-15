package cmd

import (
	"docker/container/cgroups"
	"docker/container/cgroups/subsystems"
	"docker/container/container_info"
	"docker/container/container_init"
	"docker/container/network"
	"docker/utils"
	"fmt"
	"github.com/urfave/cli"
	"os"
	"strconv"
	"strings"
	"time"
)

// RunCommand run命令定义
var RunCommand = cli.Command{
	Name: "run",
	Usage: `Create a container with namespace and cGroups limit
			docker run -ti [command]`,
	Flags:  runCmdFlags,
	Action: runCmdAction,
}

// run 命令逻辑入口
func runCmdAction(ctx *cli.Context) error {
	if len(ctx.Args()) < 1 {
		return fmt.Errorf("container command missing")
	}

	// run命令参数校验
	if err := validRunCmdFlags(ctx); err != nil {
		utils.LoggerUtil.Errorf("invalid command flag: %v", err)
		return err
	}

	// 主逻辑入口
	err := run(ctx)
	if err != nil {
		utils.LoggerUtil.Errorf("mdocker run failed: %v", err)
	}

	return err
}

// 校验run command flag参数
func validRunCmdFlags(ctx *cli.Context) error {
	if ctx.Bool("ti") && ctx.Bool("d") {
		return fmt.Errorf("ti and d param both set")
	}

	return nil
}

// run命令主逻辑
func run(ctx *cli.Context) error {
	// 生成容器名
	containerName := ctx.String(runCmdFlagName)
	id := randStringBytes(containerNameLength)
	if containerName == "" {
		containerName = id
	}
	// todo: 判断containerName是否存在

	// 解析参数
	imageName, cmdArr, err := parseCmdArg(ctx.Args())
	if err != nil {
		return err
	}

	// 使用init初始化容器, 初始化完成后在容器内执行用户命令
	volume := ctx.String(runCmdFlagVolume)
	initCmd, initPipe, err := container_init.NewContainerProcess(ctx.Bool(runCmdFlagTty), volume, containerName, imageName)
	if err != nil {
		return err
	}
	if err = initCmd.Start(); err != nil {
		return err
	}
	// 构造containerInfo
	cInfo := &container_info.ContainerInfo{
		Id:          id,
		Pid:         strconv.Itoa(initCmd.Process.Pid),
		Command:     strings.Join(ctx.Args(), ""),
		CreatedTime: time.Now().Format("2006-01-02 15:04:05"),
		Name:        containerName,
		Volume:      volume,
	}
	if ctx.IsSet(runCmdFlagPortMap) {
		cInfo.PortMap = ctx.StringSlice(runCmdFlagPortMap)
	}

	// 为container创建cgroup
	cgroupManager, err := handleCgroupSet(initCmd.Process.Pid, containerName, ctx)
	if err != nil {
		return err
	}
	defer containerExitProcess(cgroupManager, ctx, cInfo)

	// 设置容器网络
	if err = initNetNs(ctx, cInfo); err != nil {
		return err
	}

	// 记录container信息
	if err = container_info.RecordContainerInfo(cInfo); err != nil {
		return err
	}

	// 将用户命令指定命令通过pipe传递给init进程
	err = sendInitCommandParams(cmdArr, initPipe)
	if err != nil {
		return err
	}

	// 当设置了ti参数时, 等待init进程退出
	if ctx.Bool(runCmdFlagTty) {
		_ = initCmd.Wait()
	}

	return nil
}

// 解析命令行参数
func parseCmdArg(args []string) (string, []string, error) {
	if len(args) < 2 {
		return "", nil, fmt.Errorf("invalid args")
	}

	return args[0], args[1:], nil
}

// handle init cgroup configuration for the container
func handleCgroupSet(pid int, containerName string, ctx *cli.Context) (*cgroups.CgroupManager, error) {
	cgroupManager := cgroups.NewCgroupManager(containerName)
	// set cgroup resource limit config
	if err := cgroupManager.Set(getResourceConfFromCtx(ctx)); err != nil {
		return nil, err
	}

	// apply init process to the cgroup just created
	if err := cgroupManager.Apply(pid); err != nil {
		return nil, err
	}

	return cgroupManager, nil
}

// parse resource config from cli context
func getResourceConfFromCtx(ctx *cli.Context) *subsystems.ResourceConfig {
	return &subsystems.ResourceConfig{
		MemoryLimit: ctx.String(runCmdCgroupMemory),
		CpuSet:      ctx.String(rumCmdCgroupCpuSet),
		CpuShare:    ctx.String(rumCmdCgroupCpuShare),
	}
}

// write the command param array to the writing pipe
func sendInitCommandParams(commandArr []string, initPipe *os.File) error {
	commandStr := strings.Join(commandArr, " ")
	if _, err := initPipe.WriteString(commandStr); err != nil {
		return err
	}

	if err := initPipe.Close(); err != nil {
		return err
	}

	return nil
}

// mdocker run进程退出时触发动作
func containerExitProcess(cgroupManager *cgroups.CgroupManager, ctx *cli.Context, cInfo *container_info.ContainerInfo) {
	if ctx.Bool(runCmdFlagTty) { // 非后台运行container, 退出后删除容器信息
		// 清除容器fs
		_ = container_init.DeleteWorkSpace(cInfo.Name, ctx.String(runCmdFlagVolume))
		// 删除cgroup
		cgroupManager.Destroy()

		// 清除容器网络环境
		_ = network.NetworkManager.DisConnect(cInfo)

		// 删除containerInfo
		containerInfoPath := container_info.GetContainerInfoDirPath(cInfo.Name)
		if err := os.RemoveAll(containerInfoPath); err != nil {
			utils.LoggerUtil.Errorf("remove container info dir %s error %v", containerInfoPath, err)
		}
	}
}

// 初始化容器网络
func initNetNs(ctx *cli.Context, containerInfo *container_info.ContainerInfo) error {
	nw := ctx.String(runCmdFlagNetwork)
	if nw == "" {
		return nil
	}

	// 将容器加入指定网络
	if err := network.NetworkManager.Connect(nw, containerInfo); err != nil {
		return fmt.Errorf("container nw connect error, %v", err)
	}

	return nil
}
