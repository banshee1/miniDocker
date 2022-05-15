package container_info

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
)

// 保存容器信息
func RecordContainerInfo(containerInfo *ContainerInfo) error {
	// 构造容器信息
	containerInfo.Status = StatusRunning
	jsonBytes, err := json.Marshal(containerInfo) // json序列化
	if err != nil {
		return fmt.Errorf("container info marshal error: %v", err)
	}
	jsonStr := string(jsonBytes)

	// 写入container info json file
	infoDirPath := GetContainerInfoDirPath(containerInfo.Name)
	if err = os.MkdirAll(infoDirPath, 0622); err != nil {
		return fmt.Errorf("container info mkdir error, %v", err)
	}
	infoFileName := path.Join(infoDirPath, ContainerConfigName)
	file, err := os.Create(infoFileName) // 创建文件
	if err != nil {
		return fmt.Errorf("container info file %s create error: %v", infoFileName, err)
	}
	defer file.Close()

	if _, err = file.WriteString(jsonStr); err != nil { // 写入container info
		return fmt.Errorf("container info file write %s error: %v", infoFileName, err)
	}

	return nil
}

// 更新容器信息
func UpdateContainerInfo(containerName string, containerInfo *ContainerInfo) error {
	jsonBytes, err := json.Marshal(containerInfo) // json序列化
	if err != nil {
		return fmt.Errorf("container info marshal error: %v", err)
	}
	jsonStr := string(jsonBytes)

	// 写入container info json file
	infoFileName := GetContainerInfoFilePath(containerName)
	file, err := os.Create(infoFileName) // 创建文件
	if err != nil {
		return fmt.Errorf("container info file %s create error: %v", infoFileName, err)
	}
	defer file.Close()

	if _, err = file.WriteString(jsonStr); err != nil { // 写入container info
		return fmt.Errorf("container info file write %s error: %v", infoFileName, err)
	}

	return nil
}

// 根据containerName获取容器pid
func GetContainerPidByName(containerName string) (string, error) {
	containerInfoDir := path.Join(ContainerInfoLocation, containerName)
	configFilePath := path.Join(containerInfoDir, ContainerConfigName)
	contentBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return "", err
	}
	var containerInfo ContainerInfo
	if err = json.Unmarshal(contentBytes, &containerInfo); err != nil {
		return "", err
	}
	return containerInfo.Pid, nil
}

// 获取当前运行的所有容器信息
func GetContainerInfoAll() ([]*ContainerInfo, error) {
	// 读取container信息路径下 文件/路径 列表
	containerInfoDir := ContainerInfoLocation
	files, err := ioutil.ReadDir(containerInfoDir)
	if err != nil {
		return nil, fmt.Errorf("read dir %s error %v", containerInfoDir, err)
	}

	// info路径下所有container info文件
	var containers []*ContainerInfo
	for _, file := range files {
		curContainerInfo, err := getContainerInfo(file)
		if err != nil {
			log.Errorf("Get container info error %v", err)
			continue
		}
		containers = append(containers, curContainerInfo)
	}

	return containers, nil
}

// 读取指定文件内容, 解析出容器相关信息
func getContainerInfo(file os.FileInfo) (*ContainerInfo, error) {
	// 读取container info文件内容
	containerName := file.Name()
	configFileDir := path.Join(path.Join(ContainerInfoLocation, containerName), ContainerConfigName)
	content, err := ioutil.ReadFile(configFileDir)
	if err != nil {
		return nil, fmt.Errorf("read file %s error %v", configFileDir, err)
	}

	// 解析文件json内容
	containerInfo := &ContainerInfo{}
	if err = json.Unmarshal(content, containerInfo); err != nil {
		return nil, fmt.Errorf("json unmarshal error %v", err)
	}

	return containerInfo, nil
}

// 根据containerName获取容器运行时相关信息
func GetContainerInfoByContainerName(containerName string) (*ContainerInfo, error) {
	// 读取文件内容
	configFilePath := GetContainerInfoFilePath(containerName)
	content, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("container config file read error: %v", err)
	}

	// 反序列化json文件
	containerInfo := &ContainerInfo{}
	err = json.Unmarshal(content, containerInfo)
	if err != nil {
		return nil, fmt.Errorf("container info file unmarshal error: %v", err)
	}

	return containerInfo, err
}

// 获取container info文件路径
func GetContainerInfoFilePath(containerName string) string {
	containerInfoPath := GetContainerInfoDirPath(containerName)

	return path.Join(containerInfoPath, ContainerConfigName)
}

// 获取container log文件路径
func GetContainerLogFilePath(containerName string) string {
	containerInfoPath := GetContainerInfoDirPath(containerName)

	return path.Join(containerInfoPath, ContainerLogFileName)
}

// 获取容器info信息文件存放路径
func GetContainerInfoDirPath(containerName string) string {
	return path.Join(ContainerInfoLocation, containerName)
}
