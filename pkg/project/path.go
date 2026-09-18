package project

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ProjectDir 在父目录下生成项目路径：parent/name，例如 D:\work + sss → D:\work\sss。
func ProjectDir(parentDir, projectName string) (string, error) {
	parentDir = strings.TrimSpace(parentDir)
	projectName = strings.TrimSpace(projectName)
	if parentDir == "" {
		return "", fmt.Errorf("创建位置不能为空")
	}
	if projectName == "" {
		return "", fmt.Errorf("项目名称不能为空")
	}
	if strings.ContainsAny(projectName, `/\`) {
		return "", fmt.Errorf("项目名称不能包含 \\ 或 /")
	}
	return filepath.Join(parentDir, projectName), nil
}

// DefaultNewOptions 根据父目录与项目名填充 new 的默认参数。
func DefaultNewOptions(template, parentDir, projectName string, force bool) (NewOptions, error) {
	target, err := ProjectDir(parentDir, projectName)
	if err != nil {
		return NewOptions{}, err
	}
	return NewOptions{
		Template:   template,
		Target:     target,
		ArtifactID: projectName,
		Name:       projectName,
		GroupID:    "com.star",
		Version:    "0.0.1-SNAPSHOT",
		Force:      force,
	}, nil
}
