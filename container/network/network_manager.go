package network

import (
	"docker/container/container_info"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"text/tabwriter"
)

type networkManager struct {
	drivers  map[string]NetworkDriver
	networks map[string]*Network
	ipam     *IpAddressManager
}

var NetworkManager = networkManager{}

// 初始化NetworkMgr
func init() {
	// 初始化sub modules
	NetworkManager.networks = make(map[string]*Network)
	NetworkManager.drivers = make(map[string]NetworkDriver)
	NetworkManager.ipam = &IpAddressManager{
		ConfigFilePath: IpAddrConfigFilePath,
	}

	// 设置drivers map
	NetworkManager.drivers[DriverNameBridge] = &BridgeNetworkDriver{}

	// test and create path
	if _, err := os.Stat(DefaultNetworkPath); err != nil {
		if !os.IsNotExist(err) {
			return
		}

		if err = os.MkdirAll(DefaultNetworkPath, 0644); err != nil {
			return
		}
	}

	// 遍历路径读取配置文件
	fileList, err := ioutil.ReadDir(DefaultNetworkPath)
	if err != nil {
		return
	}
	// restore网络配置
	for _, fileInfo := range fileList {
		_ = NetworkManager.restoreNetworkFromConfFile(path.Join(DefaultNetworkPath, fileInfo.Name()))
	}
}

// 从指定文件中读取网络配置
func (n *networkManager) restoreNetworkFromConfFile(filePath string) error {
	if strings.HasSuffix(filePath, "/") {
		return nil
	}

	_, nwName := path.Split(filePath)
	nw := &Network{
		Name: nwName,
	}

	if err := nw.load(filePath); err != nil {
		logrus.Errorf("error load network: %s", err)
	}
	n.networks[nwName] = nw

	return nil
}

// 新建network
func (n *networkManager) CreateNetwork(driver, subnet, name string) error {
	// 解析cidr子网
	_, cidr, _ := net.ParseCIDR(subnet)
	ip, err := n.ipam.Allocate(cidr)
	if err != nil {
		return err
	}
	cidr.IP = ip

	if _, ok := n.drivers[driver]; !ok {
		return fmt.Errorf("invalid driver type: %s", driver)
	}
	// 调用对应的驱动
	nw, err := n.drivers[driver].Create(cidr.String(), name)
	if err != nil {
		return err
	}

	return nw.dump(DefaultNetworkPath)
}

// 打印manager中保存的网络信息
func (n *networkManager) ListNetwork() {
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	_, _ = fmt.Fprint(w, "NAME\tIpRange\tDriver\n")
	for _, nw := range n.networks {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n",
			nw.Name,
			nw.IpRange.String(),
			nw.Driver,
		)
	}
	if err := w.Flush(); err != nil {
		logrus.Errorf("Flush error %v", err)
		return
	}
}

// 指定Name删除network
func (n *networkManager) DeleteNetwork(networkName string) error {
	nw, ok := n.networks[networkName]
	if !ok {
		return fmt.Errorf("no Such Network: %s", networkName)
	}

	if err := n.ipam.Release(nw.IpRange, &nw.IpRange.IP); err != nil {
		return fmt.Errorf("error Remove Network gateway ip: %s", err)
	}

	if err := n.drivers[nw.Driver].Delete(*nw); err != nil {
		return fmt.Errorf("error Remove Network DriverError: %s", err)
	}

	return nw.remove(DefaultNetworkPath)
}

// 将容器加入到指定的network下
func (n *networkManager) Connect(networkName string, cInfo *container_info.ContainerInfo) error {
	// 1. 获取网络对象
	network, ok := n.networks[networkName]
	if !ok { // 网络不存在
		return fmt.Errorf("no Such Network: %s", networkName)
	}

	// 2. 分配容器IP地址
	_, ipRange, _ := net.ParseCIDR(network.IpNetStr) // 重新生成一个ipNet, 因为需要修改IpNet的ip
	ipRange.IP = ipRange.IP.To4()
	ip, err := n.ipam.Allocate(ipRange)
	if err != nil {
		return err
	}
	ipRange.IP = ip
	cInfo.IpAddr = ipRange.String() // 记录container ip信息

	// 3. 创建container endpoint
	ep := &Endpoint{
		Id:          fmt.Sprintf("%s-%s", cInfo.Id, networkName),
		IPAddress:   ip,
		Network:     network,
		PortMapping: cInfo.PortMap,
	}

	// 4. 调用网络驱动挂载和配置网络端点
	if _, ok = n.drivers[network.Driver]; !ok { // 驱动不存在
		return fmt.Errorf("no such network driver: %s", network.Driver)
	}
	if err = n.drivers[network.Driver].Connect(network, ep); err != nil {
		return fmt.Errorf("bridge driver connect error, %v", err)
	}

	// 5. 到容器的namespace配置容器网络设备IP地址
	if err = n.configEndpointIpAddressAndRoute(ep, cInfo); err != nil {
		return fmt.Errorf("net ns configure error, %v", err)
	}

	return n.configPortMapping(ep)
}

// 将veth peer挂载到container中
func (n *networkManager) configEndpointIpAddressAndRoute(ep *Endpoint, cInfo *container_info.ContainerInfo) error {
	peerLink, err := netlink.LinkByName(ep.Device.PeerName)
	if err != nil {
		return fmt.Errorf("fail config endpoint: %v", err)
	}
	//  进入container的namespace
	exitNsHandler, err := n.enterContainerNetns(&peerLink, cInfo)
	if err != nil {
		return fmt.Errorf("enter netns error, %v", err)
	}
	defer exitNsHandler()

	interfaceIP := *ep.Network.IpRange
	interfaceIP.IP = ep.IPAddress

	// 给container中的veth设置ip
	if err = setInterfaceIP(ep.Device.PeerName, &interfaceIP); err != nil {
		return fmt.Errorf("%v,%s", ep.Network, err)
	}

	// 启动veth
	if err = setInterfaceUP(ep.Device.PeerName); err != nil {
		return err
	}

	// 启动loopback设备
	if err = setInterfaceUP("lo"); err != nil {
		return err
	}

	_, cidr, _ := net.ParseCIDR("0.0.0.0/0")
	defaultRoute := &netlink.Route{
		LinkIndex: peerLink.Attrs().Index,
		Gw:        ep.Network.IpRange.IP, //gateway
		Dst:       cidr,
	}

	if err = netlink.RouteAdd(defaultRoute); err != nil {
		return fmt.Errorf("route add error, %v", err)
	}

	return nil
}

// 进入container的netns进行环境设置
func (n *networkManager) enterContainerNetns(enLink *netlink.Link, cInfo *container_info.ContainerInfo) (func(), error) {
	f, err := os.OpenFile(fmt.Sprintf("/proc/%s/ns/net", cInfo.Pid), os.O_RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("error get container net namespace, %v", err)
	}

	nsFD := f.Fd()
	runtime.LockOSThread()

	// ip link set veth-peer netns *
	if err = netlink.LinkSetNsFd(*enLink, int(nsFD)); err != nil {
		return nil, fmt.Errorf("error set link netns , %v", err)
	}

	// 获取当前的网络namespace
	origNs, err := netns.Get()
	if err != nil {
		return nil, fmt.Errorf("error get current netns, %v", err)
	}

	// 设置当前进程到新的网络namespace，并在函数执行完成之后再恢复到之前的namespace
	if err = netns.Set(netns.NsHandle(nsFD)); err != nil {
		return nil, fmt.Errorf("error set netns, %v", err)
	}

	return func() {
		netns.Set(origNs)
		origNs.Close()
		runtime.UnlockOSThread()
		f.Close()
	}, nil
}

// 通过iptables设置接口转发
func (n *networkManager) configPortMapping(ep *Endpoint) error {
	for _, pm := range ep.PortMapping {
		portMapping := strings.Split(pm, ":")
		if len(portMapping) != 2 {
			return fmt.Errorf("port mapping format error, %v", pm)
		}

		iptablesCmd := fmt.Sprintf("-t nat -A PREROUTING -p tcp -m tcp --dport %s -j DNAT --to-destination %s:%s",
			portMapping[0], ep.IPAddress.String(), portMapping[1])
		cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("iptables Output, %v", output)
		}
	}

	return nil
}

// 将container从network中断开, 并清理container网络环境设置
func (n *networkManager) DisConnect(cInfo *container_info.ContainerInfo) error {
	// 解析cidr地址
	ip, ipNet, err := net.ParseCIDR(cInfo.IpAddr)
	if err != nil {
		return err
	}

	// 释放ip地址
	if err = n.ipam.Release(ipNet, &ip); err != nil {
		return err
	}

	// 清除iptables port转发规则
	for _, pm := range cInfo.PortMap {
		portMapping := strings.Split(pm, ":")
		if len(portMapping) != 2 {
			return fmt.Errorf("port mapping format error, %v", pm)
		}

		iptablesCmd := fmt.Sprintf("-t nat -D PREROUTING -p tcp -m tcp --dport %s -j DNAT --to-destination %s:%s",
			portMapping[0], ip.String(), portMapping[1])
		cmd := exec.Command("iptables", strings.Split(iptablesCmd, " ")...)
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("iptables Output, %v", output)
		}
	}

	return nil
}
