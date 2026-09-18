package fsutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DirNotEmpty 判断目录是否存在且包含至少一个条目。
func DirNotEmpty(path string) (bool, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return len(entries) > 0, nil
}

// WalkRelFiles 递归收集 root 下所有文件（相对 root 的路径）。
func WalkRelFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			base := filepath.Base(path)
			if base == ".git" || base == "target" || base == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	return files, err
}

// IsZipFile 根据扩展名判断是否为 zip 模板。
func IsZipFile(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".zip")
}

// ConfirmOverwrite 非空目录时根据 force 决定是否允许继续。
func ConfirmOverwrite(target string, force bool) error {
	nonEmpty, err := DirNotEmpty(target)
	if err != nil {
		return err
	}
	if !nonEmpty {
		return nil
	}
	if force {
		return nil
	}
	return fmt.Errorf("目标目录 %q 已存在且非空；使用 --force 覆盖或换用其它 --target", target)
}
