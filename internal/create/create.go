// Package create 从仓库下载骨架并改成新项目。
package create

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/star/star-yi-cli/internal/build"
	"github.com/star/star-yi-cli/internal/config"
	"github.com/star/star-yi-cli/internal/meta"
	"github.com/star/star-yi-cli/internal/names"
	"github.com/star/star-yi-cli/internal/scripts"
	"github.com/star/star-yi-cli/internal/source"
)

// Options 一次创建。空字段取版本设置里的默认值。
type Options struct {
	Edition     *config.Edition `json:"-"`
	ParentDir   string          `json:"parentDir"`
	Name        string          `json:"name"`
	DisplayName string          `json:"displayName"`
	GroupID     string          `json:"groupId"`
	Version     string          `json:"version"`
	Repo        string          `json:"repo"`
	Branch      *string         `json:"branch"`
	AppsDir     string          `json:"appsDir"`
	Force       bool            `json:"force"`
	GitInit     bool            `json:"gitInit"`

	Log func(string) `json:"-"`
}

// Result 创建结果。
type Result struct {
	Target      string `json:"target"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	GroupID     string `json:"groupId"`
	Version     string `json:"version"`
	AppsDir     string `json:"appsDir"`
	Build       string `json:"build"`
	GitInit     string `json:"gitInit,omitempty"`
}

// Plan 补全默认值并校验，返回目标目录。
func (o *Options) Plan() (string, error) {
	e := o.Edition
	if e == nil {
		return "", errors.New("没有选择版本")
	}
	if err := e.Validate(); err != nil {
		return "", err
	}
	o.Name = strings.TrimSpace(o.Name)
	if err := names.ValidateProjectName(o.Name); err != nil {
		return "", err
	}
	o.DisplayName = strings.TrimSpace(o.DisplayName)
	if o.DisplayName == "" {
		o.DisplayName = o.Name
	}
	if err := names.ValidateDisplayName(o.DisplayName); err != nil {
		return "", err
	}
	o.GroupID = strings.TrimSpace(o.GroupID)
	if o.GroupID == "" {
		o.GroupID = e.GroupID
	}
	if err := names.ValidateGroupID(o.GroupID); err != nil {
		return "", err
	}
	o.Version = strings.TrimSpace(o.Version)
	if o.Version == "" {
		o.Version = e.Version
	}
	if err := names.ValidateVersion(o.Version); err != nil {
		return "", err
	}
	o.Repo = strings.TrimSpace(o.Repo)
	if o.Repo == "" {
		o.Repo = e.Repo
	}
	if o.Branch == nil {
		b := e.Branch
		o.Branch = &b
	} else {
		b := strings.TrimSpace(*o.Branch)
		o.Branch = &b
	}
	if strings.TrimSpace(o.AppsDir) == "" {
		o.AppsDir = e.AppsDir
	}
	dir, err := names.CleanAppsDir(o.AppsDir)
	if err != nil {
		return "", err
	}
	o.AppsDir = dir

	parent := strings.TrimSpace(o.ParentDir)
	if parent == "" {
		parent = "."
	}
	parent, err = filepath.Abs(parent)
	if err != nil {
		return "", err
	}
	o.ParentDir = parent
	return filepath.Join(parent, o.Name), nil
}

func (o *Options) log(format string, a ...any) {
	if o.Log != nil {
		o.Log(fmt.Sprintf(format, a...))
	}
}

// Run 执行创建。先在父目录的临时目录里做完全部改动，成功后再移到目标位置。
func Run(o Options) (*Result, error) {
	target, err := o.Plan()
	if err != nil {
		return nil, err
	}
	if info, err := os.Stat(target); err == nil {
		if !info.IsDir() {
			return nil, fmt.Errorf("%s 已存在并且是文件", target)
		}
		entries, _ := os.ReadDir(target)
		if len(entries) > 0 && !o.Force {
			return nil, fmt.Errorf("%s 已存在且不为空。换一个项目名，或勾选覆盖", target)
		}
	}
	if err := os.MkdirAll(o.ParentDir, 0o755); err != nil {
		return nil, fmt.Errorf("创建父目录: %w", err)
	}

	work, err := os.MkdirTemp(o.ParentDir, ".star-yi-new-*")
	if err != nil {
		return nil, fmt.Errorf("在 %s 建临时目录: %w", o.ParentDir, err)
	}
	defer source.RemoveAll(work)
	stage := filepath.Join(work, o.Name)

	e := o.Edition
	o.log("版本：%s（%s）", e.Title, e.Build)
	if err := source.Fetch(source.Request{Repo: o.Repo, Branch: *o.Branch, Dest: stage, Log: o.Log}); err != nil {
		return nil, err
	}

	p := meta.FromEdition(e, o.AppsDir)
	p.Source = &meta.Source{Repo: o.Repo, Branch: *o.Branch}
	adapter, err := build.For(p.Build)
	if err != nil {
		return nil, err
	}
	if err := adapter.Validate(stage, p); err != nil {
		return nil, fmt.Errorf("下载的内容和版本设置不符（构建方式 %s）：%w", p.Build, err)
	}
	before, err := adapter.Info(stage)
	if err != nil {
		return nil, err
	}

	spec := build.RenameSpec{Name: o.Name, DisplayName: o.DisplayName, GroupID: o.GroupID, Version: o.Version}
	if err := adapter.Rename(stage, p, spec); err != nil {
		return nil, fmt.Errorf("重命名: %w", err)
	}
	o.log("重命名：%s → %s，坐标 %s:%s:%s", before.Name, o.Name, o.GroupID, o.Name, o.Version)

	appsAbs := filepath.Join(stage, filepath.FromSlash(o.AppsDir))
	if err := os.MkdirAll(appsAbs, 0o755); err != nil {
		return nil, err
	}
	if entries, _ := os.ReadDir(appsAbs); len(entries) == 0 {
		if err := os.WriteFile(filepath.Join(appsAbs, ".gitkeep"), nil, 0o644); err != nil {
			return nil, err
		}
	}
	if err := meta.Write(stage, p); err != nil {
		return nil, err
	}
	if _, err := scripts.Write(stage); err != nil {
		return nil, err
	}
	o.log("已写入 %s、%s、%s", filepath.ToSlash(filepath.Join(meta.DirName, meta.FileName)), scripts.CmdName, scripts.ShName)

	if err := place(stage, target); err != nil {
		return nil, err
	}
	o.log("项目位置：%s", target)

	res := &Result{
		Target: target, Name: o.Name, DisplayName: o.DisplayName,
		GroupID: o.GroupID, Version: o.Version, AppsDir: o.AppsDir, Build: p.Build,
	}
	if o.GitInit {
		res.GitInit = gitInit(target)
		o.log("%s", res.GitInit)
	}
	return res, nil
}

// place 把准备好的目录放到目标位置。目标已存在时逐个文件覆盖。
func place(stage, target string) error {
	if _, err := os.Stat(target); errors.Is(err, os.ErrNotExist) {
		if err := os.Rename(stage, target); err == nil {
			return nil
		}
	}
	if err := source.CopyTree(stage, target, false); err != nil {
		return fmt.Errorf("复制到 %s: %w", target, err)
	}
	return nil
}

func gitInit(dir string) string {
	if !source.GitAvailable() {
		return "没有找到 git，跳过 git init"
	}
	if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
		return "目录里已有 .git，跳过 git init"
	}
	cmd := exec.Command("git", "init", "--quiet")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		return "git init 失败：" + strings.TrimSpace(string(out))
	}
	return "已执行 git init"
}
