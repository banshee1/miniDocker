package subsystems

import "path"

type CpuSubSystem struct {
}

func (s *CpuSubSystem) Set(containerName string, res *ResourceConfig) error {
	subsysCgroupPath, err := GetCgroupPath(s.Name(), containerName, true)
	if err != nil {
		return err
	}

	if res.CpuShare == "" {
		return nil
	}

	return writeSubsystemFile(path.Join(subsysCgroupPath, "cpu.shares"), []byte(res.CpuShare), 0644)
}

func (s *CpuSubSystem) Remove(containerName string) error {
	return removeCgroupAtPath(s.Name(), containerName)

}

func (s *CpuSubSystem) Apply(containerName string, pid int) error {
	return applyPidToCgroup(s.Name(), containerName, pid, 0644)

}

func (s *CpuSubSystem) Name() string {
	return "cpu"
}
