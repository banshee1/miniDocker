package subsystems

import "path"

type MemorySubSystem struct {
}

func (s *MemorySubSystem) Set(containerName string, res *ResourceConfig) error {
	subsysCgroupPath, err := GetCgroupPath(s.Name(), containerName, true)
	if err != nil {
		return err
	}

	if res.MemoryLimit == "" {
		return nil
	}

	return writeSubsystemFile(path.Join(subsysCgroupPath, "memory.limit_in_bytes"), []byte(res.MemoryLimit), 0644)
}

func (s *MemorySubSystem) Remove(containerName string) error {
	return removeCgroupAtPath(s.Name(), containerName)
}

func (s *MemorySubSystem) Apply(containerName string, pid int) error {
	return applyPidToCgroup(s.Name(), containerName, pid, 0644)
}

// Name return the name of memory subsystem
func (s *MemorySubSystem) Name() string {
	return "memory"
}
