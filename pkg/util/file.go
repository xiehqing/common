package util

import (
	"os"
	"path/filepath"
)

// AppendFile 追加文件
func AppendFile(name string, data []byte) error {
	f, err := os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err1 := f.Close(); err1 != nil && err == nil {
		err = err1
	}
	return err
}

// WriteFile 写文件
func WriteFile(filePath, content string) error {
	if _, err := os.Stat(filepath.Dir(filePath)); err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
			if err != nil {
				return err
			}
		}
	}
	err := os.WriteFile(filePath, []byte(content), os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

// CheckAndCreateDir 检查并创建目录
func CheckAndCreateDir(dir string) {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			os.MkdirAll(dir, os.ModePerm)
		}
	}
}
