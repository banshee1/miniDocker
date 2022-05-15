package network

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path"
	"strings"
)

// ip地址管理
type IpAddressManager struct {
	ConfigFilePath string
	Subnets        map[string]string
}

// 从config文件中restore network信息
func (ipam *IpAddressManager) load() error {
	if _, err := os.Stat(ipam.ConfigFilePath); err != nil {
		if os.IsNotExist(err) {
			return nil
		} else {
			return err
		}
	}

	// 读取文件内容
	subnetJson, err := os.ReadFile(ipam.ConfigFilePath)
	if err != nil {
		return err
	}

	// 反序列化
	ipam.Subnets = make(map[string]string)
	err = json.Unmarshal(subnetJson, &ipam.Subnets)
	if err != nil {
		return err
	}

	return nil
}

// 将network信息写入到文件中
func (ipam *IpAddressManager) dump() error {
	ipamConfigFileDir, _ := path.Split(ipam.ConfigFilePath)
	if _, err := os.Stat(ipamConfigFileDir); err != nil {
		if !os.IsNotExist(err) { // stat报错
			return err
		}
		// 创建path
		err = os.MkdirAll(ipamConfigFileDir, 0644)
		if err != nil {
			return err
		}
	}
	subnetConfigFile, err := os.OpenFile(ipam.ConfigFilePath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	defer subnetConfigFile.Close()
	if err != nil {
		return err
	}

	ipamConfigJson, err := json.Marshal(ipam.Subnets)
	if err != nil {
		return err
	}

	_, err = subnetConfigFile.Write(ipamConfigJson)
	if err != nil {
		return err
	}

	return nil
}

// 分配ip
func (ipam *IpAddressManager) Allocate(subnet *net.IPNet) (net.IP, error) {
	// 存放网段中地址分配信息的数组
	ipam.Subnets = map[string]string{}

	// 从文件中加载已经分配的网段信息
	err := ipam.load()
	if err != nil {
		return nil, fmt.Errorf("error dump allocation info, %v", err)
	}

	// 解析获取子网对象
	subnetStr := subnet.String()
	one, size := subnet.Mask.Size() // 解析掩码长度信息

	// 初始化, 构造 2^n长度的0/1串, n为子网掩码长度
	if _, exist := ipam.Subnets[subnetStr]; !exist {
		ipam.Subnets[subnet.String()] = strings.Repeat("0", 1<<uint8(size-one))
	}

	var ip net.IP
	for c := range ipam.Subnets[subnetStr] {
		if ipam.Subnets[subnetStr][c] == '1' {
			continue
		}

		// 将该位置置成 1
		ipAlloc := []byte(ipam.Subnets[subnetStr])
		ipAlloc[c] = '1'
		ipam.Subnets[subnetStr] = string(ipAlloc)
		ip = subnet.IP
		// idx移位后取后8位, 加在baseIp上, 即为分配的ip
		for t := uint(4); t > 0; t -= 1 {
			[]byte(ip)[4-t] += uint8(c >> ((t - 1) * 8))
		}
		ip[3] += 1

		break
	}

	// 刷新ip配置
	err = ipam.dump()
	if err != nil {
		return nil, fmt.Errorf("ipam config dump error, %v", err)
	}

	return ip, nil
}

// ip释放
func (ipam *IpAddressManager) Release(subnet *net.IPNet, ipAddr *net.IP) error {
	// 读取配置文件
	err := ipam.load()
	if err != nil {
		return err
	}

	// 计算ip的索引
	ipIdx := 0
	releaseIP := ipAddr.To4()
	releaseIP[3] -= 1
	baseIp := subnet.IP.To4()
	for t := uint(4); t > 0; t -= 1 {
		c := int(releaseIP[t-1]-baseIp[t-1]) << ((4 - t) * 8)
		ipIdx += c
	}

	ipAlloc := []byte(ipam.Subnets[subnet.String()])
	ipAlloc[ipIdx] = '0'
	ipam.Subnets[subnet.String()] = string(ipAlloc)

	err = ipam.dump()

	return err
}
