package network

import (
	"encoding/json"
	"github.com/vishvananda/netlink"
	"net"
	"os"
	"path"
)

// 容器网络
type Network struct {
	Name     string `json:"name"`
	IpRange  *net.IPNet
	Driver   string `json:"driver"`
	IpNetStr string `json:"ip_net_str"`
}

// 将network信息保存成文件存储到指定路径下
func (nw *Network) dump(dumpPath string) error {
	if _, err := os.Stat(dumpPath); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(dumpPath, 0644)
		} else {
			return err
		}
	}
	nw.IpNetStr = nw.IpRange.String()

	nwPath := path.Join(dumpPath, nw.Name)
	nwFile, err := os.OpenFile(nwPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer nwFile.Close()

	nwJson, err := json.Marshal(nw)
	if err != nil {
		return err
	}

	_, err = nwFile.Write(nwJson)
	if err != nil {
		return err
	}
	return nil
}

// 删除指定网络的存储文件
func (nw *Network) remove(dumpPath string) error {
	if _, err := os.Stat(path.Join(dumpPath, nw.Name)); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	} else {
		return os.Remove(path.Join(dumpPath, nw.Name))
	}
}

// 从指定路径的文件中读取网络配置
func (nw *Network) load(nwFilePath string) error {
	nwConfigFile, err := os.Open(nwFilePath)
	defer nwConfigFile.Close()
	if err != nil {
		return err
	}
	nwJson := make([]byte, 2000)
	n, err := nwConfigFile.Read(nwJson)
	if err != nil {
		return err
	}

	err = json.Unmarshal(nwJson[:n], nw)
	if err != nil {
		return err
	}
	nw.IpRange, _ = netlink.ParseIPNet(nw.IpNetStr)
	nw.IpRange.IP = nw.IpRange.IP.To4()

	return nil
}
