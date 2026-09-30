// Package meta 读写项目根目录的 .star-yi/project.yaml。
//
// 这份文件在创建项目时写入，记下该项目实际使用的版本、构建方式和应用目录。
// 以后修改创建器里的默认值，不会影响已经生成的项目。
package meta

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/star/star-yi-cli/internal/config"
	"github.com/star/star-yi-cli/internal/names"
	"gopkg.in/yaml.v3"
)

const (
	DirName  = ".star-yi"
	FileName = "project.yaml"

	DefaultAdminModule  = "yi-admin"
	DefaultCommonModule = "yi-common"
	DefaultBasePackage  = "com.star"
)

// Source 项目是从哪里下载的。
type Source struct {
	Repo   string `yaml:"repo"`
	Branch string `yaml:"branch,omitempty"`
}

// Project 一个已生成项目的固定信息。
type Project struct {
	Edition      string  `yaml:"edition"`
	Build        string  `yaml:"build"`
	Language     string  `yaml:"language"`
	AppsDir      string  `yaml:"appsDir"`
	BasePackage  string  `yaml:"basePackage"`
	AdminModule  string  `yaml:"adminModule"`
	CommonModule string  `yaml:"commonModule"`
	Source       *Source `yaml:"source,omitempty"`
	CreatedAt    string  `yaml:"createdAt,omitempty"`

	// Detected 为 true 表示没有 project.yaml，是按目录结构推断出来的。
	Detected bool `yaml:"-"`
}

// FromEdition 按版本设置生成项目信息。
func FromEdition(e *config.Edition, appsDir string) *Project {
	if appsDir == "" {
		appsDir = e.AppsDir
	}
	return &Project{
		Edition:      e.ID,
		Build:        e.Build,
		Language:     e.Language,
		AppsDir:      appsDir,
		BasePackage:  DefaultBasePackage,
		AdminModule:  DefaultAdminModule,
		CommonModule: DefaultCommonModule,
		Source:       &Source{Repo: e.Repo, Branch: e.Branch},
		CreatedAt:    time.Now().Format(time.RFC3339),
	}
}

// Path 项目信息文件路径。
func Path(root string) string {
	return filepath.Join(root, DirName, FileName)
}

// Write 写入 .star-yi/project.yaml。
func Write(root string, p *Project) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(root, DirName), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(p)
	if err != nil {
		return err
	}
	header := "# 由 star-yi-cli 生成。项目根目录的 star-yi 命令按这里的内容添加、删除应用模块。\n" +
		"# appsDir 是业务模块所在目录；改动前先把已有模块挪过去，并同步构建文件里的路径。\n"
	return os.WriteFile(Path(root), append([]byte(header), data...), 0o644)
}

// Load 读取项目信息；没有文件时按目录结构推断。
func Load(root string) (*Project, error) {
	data, err := os.ReadFile(Path(root))
	if errors.Is(err, os.ErrNotExist) {
		return Detect(root)
	}
	if err != nil {
		return nil, err
	}
	var p Project
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("解析 %s: %w", Path(root), err)
	}
	p.fillDefaults()
	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", Path(root), err)
	}
	return &p, nil
}

func (p *Project) fillDefaults() {
	if p.AdminModule == "" {
		p.AdminModule = DefaultAdminModule
	}
	if p.CommonModule == "" {
		p.CommonModule = DefaultCommonModule
	}
	if p.BasePackage == "" {
		p.BasePackage = DefaultBasePackage
	}
	if p.AppsDir == "" {
		p.AppsDir = "apps"
	}
	if p.Language == "" {
		if p.Build == config.BuildGradle {
			p.Language = config.LangKotlin
		} else {
			p.Language = config.LangJava
		}
	}
}

// Validate 校验项目信息。
func (p *Project) Validate() error {
	if p.Build != config.BuildMaven && p.Build != config.BuildGradle {
		return fmt.Errorf("build 只能是 maven 或 gradle，当前是 %q", p.Build)
	}
	if p.Language != config.LangJava && p.Language != config.LangKotlin {
		return fmt.Errorf("language 只能是 java 或 kotlin，当前是 %q", p.Language)
	}
	dir, err := names.CleanAppsDir(p.AppsDir)
	if err != nil {
		return err
	}
	p.AppsDir = dir
	return nil
}

// Detect 没有 project.yaml 时，按构建文件判断版本。兼容旧创建器生成的项目。
func Detect(root string) (*Project, error) {
	exists := func(rel string) bool {
		_, err := os.Stat(filepath.Join(root, rel))
		return err == nil
	}
	p := &Project{Detected: true}
	switch {
	case exists("pom.xml") && exists(filepath.Join(DefaultAdminModule, "pom.xml")):
		p.Edition = config.EditionStarYi
		p.Build = config.BuildMaven
	case exists("settings.gradle.kts") && exists(filepath.Join(DefaultAdminModule, "build.gradle.kts")):
		p.Edition = config.EditionStarYiArc
		p.Build = config.BuildGradle
	default:
		return nil, fmt.Errorf("%s 不是 Star-Yi 项目：没有 %s，也找不到 yi-admin 的 pom.xml 或 build.gradle.kts", root, filepath.Join(DirName, FileName))
	}
	p.fillDefaults()
	return p, nil
}

// FindRoot 从 start 向上查找项目根目录。
func FindRoot(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(Path(dir)); err == nil {
			return dir, nil
		}
		if _, err := Detect(dir); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("从 %s 向上没有找到 Star-Yi 项目根目录", start)
		}
		dir = parent
	}
}
