package cmd

import (
	"github.com/urfave/cli"
	"math/rand"
	"time"
)

const (
	// mdocker run 相关参数
	runCmdFlagTty     = "ti"
	runCmdFlagVolume  = "v"
	runCmdFlagDetach  = "d"
	runCmdFlagName    = "name"
	runCmdFlagNetwork = "net"
	runCmdFlagPortMap = "p"

	// cgroup subsystem限制参数
	runCmdCgroupMemory   = "m"
	rumCmdCgroupCpuShare = "cpushare"
	rumCmdCgroupCpuSet   = "cpuset"

	containerNameLength = 10
)

var (
	runCmdFlags = []cli.Flag{
		cli.BoolFlag{
			Name:  runCmdFlagTty,
			Usage: "enable tty",
		},
		cli.BoolFlag{
			Name:  runCmdFlagDetach,
			Usage: "run container detachedly",
		},
		cli.StringFlag{
			Name:  runCmdFlagVolume,
			Usage: "volume mount",
		},
		cli.StringFlag{
			Name:  runCmdFlagName,
			Usage: "container name",
		},
		cli.StringFlag{
			Name:  runCmdFlagNetwork,
			Usage: "network name",
		},
		cli.StringFlag{
			Name:  runCmdFlagPortMap,
			Usage: "port map",
		},
		// cgroup subsystem flag
		cli.StringFlag{
			Name:  runCmdCgroupMemory,
			Usage: "memory bytes limit",
		},
		cli.StringFlag{
			Name:  rumCmdCgroupCpuShare,
			Usage: "cpushare limit",
		},
		cli.StringFlag{
			Name:  rumCmdCgroupCpuSet,
			Usage: "cpuset limit",
		},
	}
)

// 生成指定长度随机字符串
func randStringBytes(n int) string {
	letterBytes := "ABCDEFGHJKMNPQRSTWXYZabcdefhijkmnprstwxyz2345678"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	return string(b)
}
