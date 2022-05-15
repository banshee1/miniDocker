package cgroups

import (
	"docker/container/cgroups/subsystems"
	"docker/utils"
	"os"
)

type CgroupManager struct {
	// ContainerName the relative path of cgroup node in the hierachy
	ContainerName string
	// Resource config
	Resource *subsystems.ResourceConfig
}

// NewCgroupManager Create a New cgroupMgr
func NewCgroupManager(path string) *CgroupManager {
	return &CgroupManager{
		ContainerName: path,
	}
}

// Apply adding a PID into every cgroup in the cgroup list
func (c *CgroupManager) Apply(pid int) error {
	var err error
	for _, subSysIns := range subsystems.SubsystemsIns {
		err = subSysIns.Apply(c.ContainerName, pid)
		if err != nil {
			return err
		}
	}

	return nil
}

// Set set the resource limit config of the cgroup manager
func (c *CgroupManager) Set(res *subsystems.ResourceConfig) error {
	var err error
	for _, subSysIns := range subsystems.SubsystemsIns {
		err = subSysIns.Set(c.ContainerName, res)
		if err != nil {
			return err
		}
	}

	return nil
}

// Destroy Remove all the cgroup created by the manager
func (c *CgroupManager) Destroy() {
	for _, subSysIns := range subsystems.SubsystemsIns {
		if err := subSysIns.Remove(c.ContainerName); err != nil {
			utils.LoggerUtil.Errorf("remove cgroup fail %v", err)
		}
	}
}

// 删除容器的cgroup目录
func RemoveContainerCgroup(containerName string) {
	for _, subSysIns := range subsystems.SubsystemsIns {
		if err := subSysIns.Remove(containerName); err != nil && !os.IsNotExist(err) {
			utils.LoggerUtil.Errorf("remove cgroup fail %v", err)
		}
	}
}
