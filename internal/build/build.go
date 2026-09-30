// Package build 定义构建工具适配器：下载后的重命名、业务模块的注册与注销。
package build

import (
	"fmt"

	"github.com/star/star-yi-cli/internal/config"
	"github.com/star/star-yi-cli/internal/meta"
)

// Info 项目当前的坐标。
type Info struct {
	Name        string
	DisplayName string
	GroupID     string
	Version     string
}

// RenameSpec 下载后要改成的坐标。
type RenameSpec struct {
	Name        string
	DisplayName string
	GroupID     string
	Version     string
}

// ModuleSpec 新业务模块的构建信息。
type ModuleSpec struct {
	ID      string
	Depends []string
}

// Adapter 一种构建工具。
type Adapter interface {
	// Name 构建工具名，maven 或 gradle。
	Name() string
	// Validate 确认目录是这种构建工具的 Star-Yi 骨架。
	Validate(root string, p *meta.Project) error
	// Info 读取项目坐标。
	Info(root string) (Info, error)
	// Rename 把骨架的父工程名、groupId、版本改成新项目的。
	Rename(root string, p *meta.Project, spec RenameSpec) error
	// BuildFile 模块目录里的构建文件名。
	BuildFile() string
	// WriteModuleBuild 在模块目录写构建文件。
	WriteModuleBuild(root string, p *meta.Project, spec ModuleSpec, info Info) error
	// Register 在根构建文件登记模块，并让启动模块依赖它。
	Register(root string, p *meta.Project, id string, info Info) error
	// Unregister 去掉 Register 写入的全部引用。
	Unregister(root string, p *meta.Project, id string, info Info) error
	// Modules 已登记在应用目录下的模块。
	Modules(root string, p *meta.Project) ([]string, error)
	// Dependents 依赖 id 的其他业务模块。
	Dependents(root string, p *meta.Project, id string, info Info) ([]string, error)
	// VerifyCommand 编译这个模块用的命令。
	VerifyCommand(p *meta.Project, id string) (string, []string)
}

// For 按构建方式取适配器。
func For(name string) (Adapter, error) {
	switch name {
	case config.BuildMaven:
		return Maven{}, nil
	case config.BuildGradle:
		return Gradle{}, nil
	}
	return nil, fmt.Errorf("不支持的构建方式 %q，可选 %s、%s", name, config.BuildMaven, config.BuildGradle)
}
