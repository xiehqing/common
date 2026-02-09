package fsext

import (
	"errors"
	"fmt"
	"github.com/xiehqing/common/agent/home"
	"os"
	"path/filepath"
)

// Lookup 从目录 dir 开始搜索目标文件或目录，
// 并向上遍历目录树，直到达到文件系统根目录。
// 它还会检查文件的所有权，以确保搜索不会跨越所有权边界。
// 对于所有权不匹配的情况，它会跳过而不报错。
// 返回找到的目标的完整路径。
// 搜索包括起始目录本身。
func Lookup(dir string, targets ...string) ([]string, error) {
	if len(targets) == 0 {
		return nil, nil
	}
	var found []string
	err := traverseUp(dir, func(cwd string, owner int) error {
		for _, target := range targets {
			fpath := filepath.Join(cwd, target)
			err := probeEnt(fpath, owner)

			// skip to the next file on permission denied
			if errors.Is(err, os.ErrNotExist) ||
				errors.Is(err, os.ErrPermission) {
				continue
			}

			if err != nil {
				return fmt.Errorf("error probing file %s: %w", fpath, err)
			}

			found = append(found, fpath)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return found, nil
}

// LookupClosest 从目录 dir 开始搜索目标文件或目录，
// 并向上遍历目录树，直到找到目标、到达根目录或家目录。
// 它还会检查文件的所有权，以确保搜索不会跨越所有权边界。
// 如果找到目标，则返回目标的完整路径；否则返回空字符串和 false。
// 搜索包括起始目录本身。
func LookupClosest(dir, target string) (string, bool) {
	var found string
	err := traverseUp(dir, func(cwd string, owner int) error {
		fpath := filepath.Join(cwd, target)

		err := probeEnt(fpath, owner)
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		if err != nil {
			return fmt.Errorf("error probing file %s: %w", fpath, err)
		}

		if cwd == home.Dir() {
			return filepath.SkipAll
		}

		found = fpath
		return filepath.SkipAll
	})

	return found, err == nil && found != ""
}

// traverseUp 向上遍历目录 直到文件系统根目录
// 它将当前目录的绝对路径和起始目录的所有者ID传递给回调函数
// 由用户来检查所有权。
func traverseUp(dir string, walkFn func(dir string, owner int) error) error {
	cwd, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("cannot get absolute path of %s, %w", dir, err)
	}
	owner, err := Owner(dir)
	if err != nil {
		return fmt.Errorf("cannot get owner of %s, %w", dir, err)
	}
	for {
		err := walkFn(cwd, owner)
		if err == nil || errors.Is(err, filepath.SkipDir) {
			parent := filepath.Dir(cwd)
			if parent == cwd {
				return nil
			}
			cwd = parent
			continue
		}
		if errors.Is(err, filepath.SkipAll) {
			return nil
		}
		return err
	}

}

// 检查给定路径下的实体是否存在且属于指定的所有者。
func probeEnt(fspath string, owner int) error {
	_, err := os.Stat(fspath)
	if err != nil {
		return fmt.Errorf("cannot stat %s: %w", fspath, err)
	}

	// special case for ownership check bypass
	if owner == -1 {
		return nil
	}

	fowner, err := Owner(fspath)
	if err != nil {
		return fmt.Errorf("cannot get ownership for %s: %w", fspath, err)
	}

	if fowner != owner {
		return os.ErrPermission
	}

	return nil
}
