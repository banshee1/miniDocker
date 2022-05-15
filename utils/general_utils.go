package utils

import "os"

type generalUtils struct{}

var GeneralUtils = generalUtils{}

// 判断dir是否存在
func (util *generalUtils) IsDirExists(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) { // 目录不存在
			return false, nil
		} else { // stat调用报错
			return false, err
		}
	}
	if !stat.IsDir() { // 路径为文件
		return false, nil
	}

	return true, nil
}
