package subsystems

// ResourceConfig define the resource limit config
type ResourceConfig struct {
	MemoryLimit string
	CpuShare    string
	CpuSet      string
}

// Subsystem the interface proto of subsystem
type Subsystem interface {
	// Name get the name of the subsystem, eg, cpu, memory
	Name() string
	// Set the resource limit config of the cgroup node
	Set(ContainerName string, res *ResourceConfig) error
	// Apply a process into a cgroup node
	Apply(path string, pid int) error
	// Remove a cgroup
	Remove(path string) error
}

var (
	SubsystemsIns = []Subsystem{
		&CpusetSubSystem{},
		&MemorySubSystem{},
		&CpuSubSystem{},
	}
)
