package subsystems

import (
	"bufio"
	"docker/config"
	"docker/utils"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
)

// FindCgroupMountPoint find the cgroup path where the specified subsystem mounted
func FindCgroupMountPoint(subsystem string) string {
	f, err := os.Open(config.ProcessMountInfoPath)
	if err != nil { // mountinfo open failed
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		txt := scanner.Text()
		fields := strings.Split(txt, " ")
		// find the row of memory cgroup
		for _, opt := range strings.Split(fields[len(fields)-1], ",") {
			if opt == subsystem {
				return path.Join(fields[4], config.CgroupRoot)
			}
		}
	}
	// check scanner errors
	if err = scanner.Err(); err != nil {
		utils.LoggerUtil.Errorf("proc mountinfo read err: %v", err)
	}

	return ""
}

// GetCgroupPath get the absolute path of the specified cgroup
// auto create when path not exists and autoCreate set to true
func GetCgroupPath(subsystem string, containerName string, autoCreate bool) (string, error) {
	cgroupRoot := FindCgroupMountPoint(subsystem)

	_, err := os.Stat(path.Join(cgroupRoot, containerName))
	if err != nil {
		if !autoCreate || !os.IsNotExist(err) {
			return "", fmt.Errorf("cgroup path error %v", err)
		}

		if err = os.MkdirAll(path.Join(cgroupRoot, containerName), 0755); err != nil {
			return "", fmt.Errorf("cgroup create error %v", err)
		}
	}

	return path.Join(cgroupRoot, containerName), nil
}

// do subsystem file write opr
func writeSubsystemFile(filePath string, content []byte, fileMode fs.FileMode) error {
	// write memory limit for the cgroup
	err := ioutil.WriteFile(filePath, content, fileMode)
	if err != nil {
		return fmt.Errorf("cgroup %s write fail: %v", filePath, err)
	}

	return nil
}

// delete a cgroup node in hierarchy
func removeCgroupAtPath(subsysName, containerName string) error {
	subsysCgroupPath, err := GetCgroupPath(subsysName, containerName, false)
	if err != nil {
		return err
	}

	return os.RemoveAll(subsysCgroupPath)
}

// write the pid to the cgroupPath/cgroup.procs
func applyPidToCgroup(subsysName, containerName string, pid int, perm fs.FileMode) error {
	subsysCgroupPath, err := GetCgroupPath(subsysName, containerName, false)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(
		path.Join(subsysCgroupPath, "tasks"), []byte(strconv.Itoa(pid)), perm,
	)
}
