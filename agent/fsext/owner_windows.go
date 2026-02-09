//go:build windows

package fsext

import "os"

// Owner 检索指定路径下文件或目录的所有者的用户ID。
func Owner(path string) (int, error) {
	_, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return -1, nil
}
