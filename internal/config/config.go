// Package config 保存创建器自己的设置：每个版本从哪个仓库下载、新项目的默认值。
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/star/star-yi-cli/internal/names"
	"gopkg.in/yaml.v3"
)

const (
	BuildMaven  = "maven"
	BuildGradle = "gradle"
	LangJava    = "java"
	LangKotlin  = "kotlin"

	EditionStarYi    = "star-yi"
	EditionStarYiArc = "star-yi-arc"
)

// Edition 一个可创建的版本。
type Edition struct {
	ID          string `yaml:"id" json:"id"`
	Title       string `yaml:"title" json:"title"`
	Description string `yaml:"description,omitempty" json:"description"`
	Repo        string `yaml:"repo" json:"repo"`
	Branch      string `yaml:"branch" json:"branch"`
	Build       string `yaml:"build" json:"build"`
	Language    string `yaml:"language" json:"language"`
	AppsDir     string `yaml:"appsDir" json:"appsDir"`
	GroupID     string `yaml:"groupId" json:"groupId"`
	Version     string `yaml:"version" json:"version"`
	BuiltIn     bool   `yaml:"-" json:"builtIn"`
}

// Config 创建器配置文件内容。
type Config struct {
	Editions []*Edition `yaml:"editions" json:"editions"`
}

var editionIDRe = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

// Defaults 内置的两个版本。
func Defaults() []*Edition {
	return []*Edition{
		{
			ID:          EditionStarYi,
			Title:       "Star-Yi",
			Description: "Java 21 · Maven · Spring Boot 4 · MySQL",
			Repo:        "https://github.com/Star-yb/Star-Yi.git",
			Branch:      "main",
			Build:       BuildMaven,
			Language:    LangJava,
			AppsDir:     "apps",
			GroupID:     "com.star",
			Version:     "0.0.1-SNAPSHOT",
			BuiltIn:     true,
		},
		{
			ID:          EditionStarYiArc,
			Title:       "Star-Yi Arc",
			Description: "Kotlin · Gradle · Spring Boot 4 · PostgreSQL",
			Repo:        "https://github.com/Star-yb/Star-Yi-Arc.git",
			Branch:      "main",
			Build:       BuildGradle,
			Language:    LangKotlin,
			AppsDir:     "apps",
			GroupID:     "com.star",
			Version:     "0.0.1-SNAPSHOT",
			BuiltIn:     true,
		},
	}
}

// DefaultEdition 返回内置版本的出厂设置。
func DefaultEdition(id string) (*Edition, bool) {
	for _, e := range Defaults() {
		if e.ID == id {
			return e, true
		}
	}
	return nil, false
}

// Dir 创建器的用户配置目录，例如 %AppData%\star-yi-cli。
func Dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("找不到用户配置目录: %w", err)
	}
	return filepath.Join(base, "star-yi-cli"), nil
}

// Path 配置文件路径。
func Path() (string, error) {
	d, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "config.yaml"), nil
}

// Load 读取配置；文件不存在时返回内置默认值。
func Load() (*Config, error) {
	p, err := Path()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if errors.Is(err, os.ErrNotExist) {
		return &Config{Editions: Defaults()}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("读取配置 %s: %w", p, err)
	}
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("解析配置 %s: %w", p, err)
	}
	c.normalize()
	return &c, nil
}

// normalize 补回被删掉的内置版本，并标记内置项。
func (c *Config) normalize() {
	var kept []*Edition
	for _, e := range c.Editions {
		if e != nil {
			kept = append(kept, e)
		}
	}
	c.Editions = kept
	for i, d := range Defaults() {
		if e := c.find(d.ID); e != nil {
			e.BuiltIn = true
			continue
		}
		pos := i
		if pos > len(c.Editions) {
			pos = len(c.Editions)
		}
		c.Editions = append(c.Editions[:pos], append([]*Edition{d}, c.Editions[pos:]...)...)
	}
	for _, e := range c.Editions {
		e.trim()
	}
}

func (e *Edition) trim() {
	e.ID = strings.TrimSpace(e.ID)
	e.Title = strings.TrimSpace(e.Title)
	e.Description = strings.TrimSpace(e.Description)
	e.Repo = strings.TrimSpace(e.Repo)
	e.Branch = strings.TrimSpace(e.Branch)
	e.Build = strings.ToLower(strings.TrimSpace(e.Build))
	e.Language = strings.ToLower(strings.TrimSpace(e.Language))
	e.AppsDir = strings.TrimSpace(e.AppsDir)
	e.GroupID = strings.TrimSpace(e.GroupID)
	e.Version = strings.TrimSpace(e.Version)
}

func (c *Config) find(id string) *Edition {
	for _, e := range c.Editions {
		if e.ID == id {
			return e
		}
	}
	return nil
}

// Edition 按 ID 取版本。
func (c *Config) Edition(id string) (*Edition, error) {
	if e := c.find(id); e != nil {
		return e, nil
	}
	ids := make([]string, 0, len(c.Editions))
	for _, e := range c.Editions {
		ids = append(ids, e.ID)
	}
	return nil, fmt.Errorf("没有版本 %q，可选：%s", id, strings.Join(ids, "、"))
}

// Validate 校验单个版本的设置。
func (e *Edition) Validate() error {
	e.trim()
	if !editionIDRe.MatchString(e.ID) {
		return fmt.Errorf("版本 ID %q 只能用小写字母开头，后面跟小写字母、数字或连字符", e.ID)
	}
	if e.Title == "" {
		return fmt.Errorf("版本 %s：名称不能为空", e.ID)
	}
	if e.Repo == "" {
		return fmt.Errorf("版本 %s：仓库地址不能为空", e.ID)
	}
	if strings.ContainsAny(e.Branch, " \t") || strings.HasPrefix(e.Branch, "-") {
		return fmt.Errorf("版本 %s：分支名 %q 不合法", e.ID, e.Branch)
	}
	if e.Build != BuildMaven && e.Build != BuildGradle {
		return fmt.Errorf("版本 %s：构建方式只能是 maven 或 gradle", e.ID)
	}
	if e.Language != LangJava && e.Language != LangKotlin {
		return fmt.Errorf("版本 %s：语言只能是 java 或 kotlin", e.ID)
	}
	dir, err := names.CleanAppsDir(e.AppsDir)
	if err != nil {
		return fmt.Errorf("版本 %s：%w", e.ID, err)
	}
	e.AppsDir = dir
	if err := names.ValidateGroupID(e.GroupID); err != nil {
		return fmt.Errorf("版本 %s：%w", e.ID, err)
	}
	if err := names.ValidateVersion(e.Version); err != nil {
		return fmt.Errorf("版本 %s：%w", e.ID, err)
	}
	return nil
}

// Validate 校验全部版本，ID 不能重复。
func (c *Config) Validate() error {
	seen := map[string]bool{}
	for _, e := range c.Editions {
		if err := e.Validate(); err != nil {
			return err
		}
		if seen[e.ID] {
			return fmt.Errorf("版本 ID %q 重复", e.ID)
		}
		seen[e.ID] = true
	}
	return nil
}

// Save 校验后写入配置文件。
func Save(c *Config) error {
	c.normalize()
	if err := c.Validate(); err != nil {
		return err
	}
	p, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	header := "# star-yi-cli 配置。可以在创建器页面的「版本设置」里修改。\n" +
		"# 改动只影响之后新建的项目；已有项目按自己的 .star-yi/project.yaml 增删应用。\n"
	return os.WriteFile(p, append([]byte(header), data...), 0o644)
}

// RecordCLIPath 记下当前可执行文件位置，项目根目录的 star-yi 命令会读这个文件。
func RecordCLIPath() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	if isTempBuild(exe) {
		return nil
	}
	d, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(d, 0o755); err != nil {
		return err
	}
	p := filepath.Join(d, "cli-path.txt")
	if old, err := os.ReadFile(p); err == nil && strings.TrimSpace(string(old)) == exe {
		return nil
	}
	return os.WriteFile(p, []byte(exe), 0o644)
}

// isTempBuild go run / go test 生成的临时可执行文件不记录。
func isTempBuild(exe string) bool {
	tmp := os.TempDir()
	if resolved, err := filepath.EvalSymlinks(tmp); err == nil {
		tmp = resolved
	}
	rel, err := filepath.Rel(tmp, exe)
	return err == nil && !strings.HasPrefix(rel, "..")
}
