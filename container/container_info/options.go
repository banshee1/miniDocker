package container_info

type ContainerInfo struct {
	Pid         string   `json:"pid"`
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	Command     string   `json:"command"`
	CreatedTime string   `json:"createTime"`
	Status      string   `json:"status"`
	Volume      string   `json:"volume"`
	PortMap     []string `json:"port_map"`
	IpAddr      string   `json:"ip_addr"`
}

const (
	ContainerInfoLocation = "/var/run/mdocker/containers/"
	ContainerConfigName   = "config.json"
	ContainerLogFileName  = "container.log"

	// container 状态
	StatusRunning = "running"
	StatusStop    = "stopped"
	StatusExit    = "exited"
)
