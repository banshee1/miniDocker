package config

import (
	"syscall"
)

const (
	AppName    = "docker"
	Usage      = `Docker runtime implementation`
	CgroupRoot = "mdocker"

	ProcessMountInfoPath = "/proc/self/mountinfo"

	PathMnt       = "/var/lib/mdocker/overlay2/mnt"
	PathReadWrite = "/var/lib/mdocker/overlay2/rw"
	PathImage     = "/var/lib/mdocker/overlay2/image"

	ProcessCloneFlags = syscall.CLONE_NEWNS | syscall.CLONE_NEWPID | syscall.CLONE_NEWUTS |
		syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC

	MountFlagsDefault = syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	MountFlagsPrivate = syscall.MS_REC | syscall.MS_PRIVATE

	// exec 命令相关环境变量
	EnvExecPid = "mdocker_pid"
	EnvExecCmd = "mdocker_cmd"
)
