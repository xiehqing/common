package util

import (
	"fmt"
	"io"
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

// CopyDirectory 复制整个目录到目标位置
func CopyDirectory(src, dst string) error {
	// 获取源目录信息
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("获取源目录信息失败: %w", err)
	}

	// 确保源是目录
	if !srcInfo.IsDir() {
		return fmt.Errorf("源路径不是目录: %s", src)
	}

	// 创建目标目录（如果不存在）
	err = os.MkdirAll(dst, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("创建目标目录失败: %w", err)
	}

	// 遍历源目录
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 计算相对路径
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		// 跳过根目录
		if relPath == "." {
			return nil
		}

		// 构建目标路径
		dstPath := filepath.Join(dst, relPath)

		// 如果是目录，创建目录
		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		// 如果是文件，复制文件
		return copyFile(path, dstPath, info.Mode())
	})
}

// copyFile 复制单个文件
func copyFile(src, dst string, mode os.FileMode) error {
	// 打开源文件
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("打开源文件失败: %w", err)
	}
	defer srcFile.Close()

	// 创建目标文件
	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer dstFile.Close()

	// 复制内容
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("复制文件内容失败: %w", err)
	}

	return nil
}

// CopyDirectoryPreserveTime 复制目录并保持文件时间
func CopyDirectoryPreserveTime(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			err := os.MkdirAll(dstPath, info.Mode())
			if err != nil {
				return err
			}
		} else {
			err := copyFile(path, dstPath, info.Mode())
			if err != nil {
				return err
			}
		}

		// 保持文件修改时间
		return os.Chtimes(dstPath, info.ModTime(), info.ModTime())
	})
}

// PathExists 判断文件或目录是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil // 路径存在
	}
	if os.IsNotExist(err) {
		return false, nil // 路径不存在
	}
	return false, err // 其他错误（如权限不足）
}

// FileExists 判断文件是否存在（排除目录）
func FileExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return !info.IsDir(), nil // 是文件且存在
}

// DirExists 判断目录是否存在
func DirExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return info.IsDir(), nil // 是目录且存在
}
