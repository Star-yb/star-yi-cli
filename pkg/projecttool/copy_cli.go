package projecttool

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

// ProjectCLIDir 项目内 CLI 工具目录（与业务模块分离，保持根目录整洁）。
const ProjectCLIDir = "common-core"

// CLIBinaryName 项目内 CLI 可执行文件名。
func CLIBinaryName() string {
	if runtime.GOOS == "windows" {
		return "star-yi-cli.exe"
	}
	return "star-yi-cli"
}

// CLIDir 返回项目内 CLI 目录绝对路径。
func CLIDir(projectRoot string) (string, error) {
	root, err := filepath.Abs(projectRoot)
	if err != nil {
		return "", err
	}
	return filepath.Join(root, ProjectCLIDir), nil
}

// CopyCLIToProject 将当前运行的 star-yi-cli 复制到项目 common-core 目录。
func CopyCLIToProject(projectRoot string) (string, error) {
	root, err := filepath.Abs(projectRoot)
	if err != nil {
		return "", err
	}
	src, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("定位当前 CLI 失败: %w", err)
	}
	src, err = filepath.Abs(src)
	if err != nil {
		return "", err
	}

	toolDir := filepath.Join(root, ProjectCLIDir)
	dest := filepath.Join(toolDir, CLIBinaryName())
	if sameFile(src, dest) {
		return dest, nil
	}

	in, err := os.Open(src)
	if err != nil {
		return "", fmt.Errorf("读取 CLI 失败: %w", err)
	}
	defer in.Close()

	if err := os.MkdirAll(toolDir, 0o755); err != nil {
		return "", err
	}
	out, err := os.Create(dest)
	if err != nil {
		return "", fmt.Errorf("写入 %s 失败: %w", dest, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return "", err
	}
	if err := out.Close(); err != nil {
		return "", err
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(dest, 0o755); err != nil {
			return "", err
		}
	}
	return dest, nil
}

func sameFile(a, b string) bool {
	a, err1 := filepath.Abs(a)
	b, err2 := filepath.Abs(b)
	if err1 != nil || err2 != nil {
		return false
	}
	return a == b
}
