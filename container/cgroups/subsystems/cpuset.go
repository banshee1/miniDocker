package subsystems

import "path"

type CpusetSubSystem struct {
}

func (s *CpusetSubSystem) Set(containerName string, res *ResourceConfig) error {
	subsysCgroupPath, err := GetCgroupPath(s.Name(), containerName, true)
	if err != nil {
		return err
	}

	// init cgroup mem config
	err = writeSubsystemFile(path.Join(subsysCgroupPath, "cpuset.mems"), []byte("0"), 0644)
	if err != nil {
		return err
	}

	err = writeSubsystemFile(path.Join(subsysCgroupPath, "cpuset.cpus"), []byte("0"), 0644)
	if err != nil {
		return err
	}

	if res.CpuSet == "" {
		return nil
	}

	return writeSubsystemFile(path.Join(subsysCgroupPath, "cpuset.cpus"), []byte(res.CpuSet), 0644)
}

func (s *CpusetSubSystem) Remove(containerName string) error {
	return removeCgroupAtPath(s.Name(), containerName)
}

func (s *CpusetSubSystem) Apply(containerName string, pid int) error {
	return applyPidToCgroup(s.Name(), containerName, pid, 0644)
}

func (s *CpusetSubSystem) Name() string {
	return "cpuset"
}
