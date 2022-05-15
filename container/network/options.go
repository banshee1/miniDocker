package network

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"net"
	"time"
)

// container端点
type Endpoint struct {
	Id          string           `json:"id"`
	Device      netlink.Veth     `json:"dev"`
	IPAddress   net.IP           `json:"ip"`
	MacAddress  net.HardwareAddr `json:"mac"`
	PortMapping []string         `json:"port_mapping"`
	Network     *Network
}

// 网络驱动抽象接口
type NetworkDriver interface {
	Name() string
	Create(subnet string, name string) (*Network, error)
	Delete(network Network) error
	Connect(network *Network, endpoint *Endpoint) error
	Disconnect(network Network, endpoint *Endpoint) error
}

const (
	// 子网ip地址配置文件path
	IpAddrConfigFilePath = "/var/run/mdocker/network/ipam/subnet.json"
	DefaultNetworkPath   = "/var/run/mdocker/network/instances/"

	// 驱动名字
	DriverNameBridge = "bridge"
)

// 启动虚拟设备
func setInterfaceUP(interfaceName string) error {
	iface, err := netlink.LinkByName(interfaceName)
	if err != nil {
		return fmt.Errorf("error retrieving a link named [ %s ]: %v", interfaceName, err)
	}

	if err = netlink.LinkSetUp(iface); err != nil {
		return fmt.Errorf("error enabling interface for %s: %v", interfaceName, err)
	}

	return nil
}

// 给虚拟设备 设置ip
func setInterfaceIP(name string, gwIpNet *net.IPNet) error {
	// 获取bridge设备
	var iface netlink.Link
	var err error
	for i := 0; i < 2; i++ {
		iface, err = netlink.LinkByName(name)
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return fmt.Errorf("abandoning retrieving the new bridge link from netlink, "+
			"Run [ ip link ] to troubleshoot the error: %v", err)
	}

	// 构造ip和子网range
	addr := &netlink.Addr{
		IPNet: gwIpNet,
		Label: "",
		Flags: 0,
		Scope: 0,
		Peer:  nil,
	}

	return netlink.AddrAdd(iface, addr)
}
