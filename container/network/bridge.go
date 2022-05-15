package network

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"net"
	"os/exec"
	"strings"
)

type BridgeNetworkDriver struct{}

// 创建Bridge网络
func (d *BridgeNetworkDriver) Create(subnet string, name string) (*Network, error) {
	ip, ipRange, _ := net.ParseCIDR(subnet)
	ipRange.IP = ip
	n := &Network{
		Name:    name,
		IpRange: ipRange,
		Driver:  d.Name(),
	}
	err := d.initBridge(n)
	if err != nil {
		return nil, fmt.Errorf("error init bridge: %v", err)
	}

	return n, err
}

// 返回驱动类型
func (d *BridgeNetworkDriver) Name() string {
	return DriverNameBridge
}

func (d *BridgeNetworkDriver) Delete(network Network) error {
	bridgeName := network.Name
	br, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}

	if err = d.restoreIPTables(network.Name, network.IpRange); err != nil {
		return err
	}

	return netlink.LinkDel(br)
}

// 将endpoint 连接到bridge网络上
func (d *BridgeNetworkDriver) Connect(network *Network, endpoint *Endpoint) error {
	bridgeName := network.Name
	br, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return err
	}

	// 构造veth pair attr
	la := netlink.NewLinkAttrs()
	la.Name = endpoint.Id[:5]
	la.MasterIndex = br.Attrs().Index
	endpoint.Device = netlink.Veth{
		LinkAttrs: la,
		PeerName:  "cif-" + endpoint.Id[:5],
	}

	// 添加veth设备
	if err = netlink.LinkAdd(&endpoint.Device); err != nil {
		return fmt.Errorf("error Add Endpoint Device: %v", err)
	}
	// 启动veth设备
	if err = netlink.LinkSetUp(&endpoint.Device); err != nil {
		return fmt.Errorf("error Add Endpoint Device: %v", err)
	}

	return nil
}

func (d *BridgeNetworkDriver) Disconnect(network Network, endpoint *Endpoint) error {
	return nil
}

// 初始化bridge虚拟设备
func (d *BridgeNetworkDriver) initBridge(n *Network) error {
	// 创建虚拟设备
	bridgeName := n.Name
	if err := d.createBridgeInterface(bridgeName); err != nil {
		return fmt.Errorf("error add bridge： %s, Error: %v", bridgeName, err)
	}

	// 设置bridge的子网ip
	gwIPNet := n.IpRange
	gwIPNet.IP = n.IpRange.IP
	if err := setInterfaceIP(bridgeName, gwIPNet); err != nil {
		return fmt.Errorf(
			"error assigning address: %s on bridge: %s with an error of: %v",
			gwIPNet, bridgeName, err,
		)
	}

	// 启动bridge
	if err := setInterfaceUP(bridgeName); err != nil {
		return fmt.Errorf("error set bridge up: %s, Error: %v", bridgeName, err)
	}

	// 设置iptables转发
	if err := d.setupIPTables(bridgeName, n.IpRange); err != nil {
		return fmt.Errorf("error setting iptables for %s: %v", bridgeName, err)
	}

	return nil
}

// 指定名字创建bridge虚拟设备
func (d *BridgeNetworkDriver) createBridgeInterface(bridgeName string) error {
	// 尝试获取设备
	_, err := net.InterfaceByName(bridgeName)
	if err == nil {
		return nil
	}
	if !strings.Contains(err.Error(), "no such network interface") { // 获取设备异常
		return err
	}

	// create *netlink.Bridge object
	la := netlink.NewLinkAttrs()
	la.Name = bridgeName

	br := &netlink.Bridge{LinkAttrs: la}
	if err = netlink.LinkAdd(br); err != nil {
		return fmt.Errorf("bridge creation failed for bridge %s: %v", bridgeName, err)
	}

	return nil
}

// 设置iptables转发
func (d *BridgeNetworkDriver) setupIPTables(bridgeName string, subnet *net.IPNet) error {
	iptablesCmd := fmt.Sprintf("-t nat -A POSTROUTING -s %s ! -o %s -j MASQUERADE", subnet.String(), bridgeName)
	cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("iptables Output error %v", output)
	}

	return err
}

// 还原iptables设置
func (d *BridgeNetworkDriver) restoreIPTables(bridgeName string, subnet *net.IPNet) error {
	iptablesCmd := fmt.Sprintf("-t nat -D POSTROUTING -s %s ! -o %s -j MASQUERADE", subnet.String(), bridgeName)
	cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("iptables Output error %v", output)
	}

	return err
}

// 删除指定的bridge设备
func (d *BridgeNetworkDriver) deleteBridge(n *Network) error {
	bridgeName := n.Name

	// get the link
	l, err := netlink.LinkByName(bridgeName)
	if err != nil {
		return fmt.Errorf("getting link with name %s failed: %v", bridgeName, err)
	}

	// delete the link
	if err = netlink.LinkDel(l); err != nil {
		return fmt.Errorf("failed to remove bridge interface %s delete: %v", bridgeName, err)
	}

	return nil
}
