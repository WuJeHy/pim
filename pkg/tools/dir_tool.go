package tools

import "os"

func IsDirExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

// 创建目录
func CreateDir(path string) error {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

// 使用golang实现校验目录是否存在，不存在就创建
func CheckAndCreateDir(path string) error {
	if IsDirExists(path) {
		return nil
	}
	err := CreateDir(path)
	if err != nil {
		return err
	}
	return nil
}
