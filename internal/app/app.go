// Package app 在已有项目里添加、删除、列出业务模块。
package app

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/star/star-yi-cli/internal/build"
	"github.com/star/star-yi-cli/internal/meta"
	"github.com/star/star-yi-cli/internal/names"
	"github.com/star/star-yi-cli/internal/source"
)

// Spec 添加模块的参数。
type Spec struct {
	ID      string
	Package string
	APIPath string
	Depends []string

	ClassPrefix string
}

// Project 打开的项目。
type Project struct {
	Root    string
	Meta    *meta.Project
	Adapter build.Adapter
	Info    build.Info
}

// Open 读取项目信息并选好构建适配器。
func Open(root string) (*Project, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	m, err := meta.Load(abs)
	if err != nil {
		return nil, err
	}
	a, err := build.For(m.Build)
	if err != nil {
		return nil, err
	}
	if err := a.Validate(abs, m); err != nil {
		return nil, err
	}
	info, err := a.Info(abs)
	if err != nil {
		return nil, err
	}
	return &Project{Root: abs, Meta: m, Adapter: a, Info: info}, nil
}

// ModuleDir 模块的绝对路径。
func (p *Project) ModuleDir(id string) string {
	return filepath.Join(p.Root, filepath.FromSlash(path.Join(p.Meta.AppsDir, id)))
}

// ModuleRel 模块相对项目根的路径。
func (p *Project) ModuleRel(id string) string {
	return path.Join(p.Meta.AppsDir, id)
}

// Module 一个业务模块的状态。
type Module struct {
	ID         string
	Registered bool
	DirExists  bool
}

// List 列出已登记的模块，以及应用目录里没登记的文件夹。
func (p *Project) List() ([]Module, error) {
	reg, err := p.Adapter.Modules(p.Root, p.Meta)
	if err != nil {
		return nil, err
	}
	byID := map[string]*Module{}
	for _, id := range reg {
		byID[id] = &Module{ID: id, Registered: true}
	}
	entries, err := os.ReadDir(filepath.Join(p.Root, filepath.FromSlash(p.Meta.AppsDir)))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		if m, ok := byID[e.Name()]; ok {
			m.DirExists = true
		} else {
			byID[e.Name()] = &Module{ID: e.Name(), DirExists: true}
		}
	}
	for _, m := range byID {
		if m.Registered && !m.DirExists {
			if _, err := os.Stat(p.ModuleDir(m.ID)); err == nil {
				m.DirExists = true
			}
		}
	}
	out := make([]Module, 0, len(byID))
	for _, m := range byID {
		out = append(out, *m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

// Normalize 补全默认值并校验参数。
func (p *Project) Normalize(spec *Spec) error {
	spec.ID = strings.TrimSpace(spec.ID)
	if err := names.ValidateModuleID(spec.ID); err != nil {
		return err
	}
	if spec.ID == p.Meta.AdminModule || spec.ID == p.Meta.CommonModule || spec.ID == p.Info.Name {
		return fmt.Errorf("模块名 %q 和项目已有模块重名", spec.ID)
	}
	spec.Package = strings.TrimSpace(spec.Package)
	if spec.Package == "" {
		spec.Package = names.DefaultPackage(spec.ID)
	}
	if err := names.ValidatePackage(spec.Package); err != nil {
		return err
	}
	spec.APIPath = strings.TrimSpace(spec.APIPath)
	if spec.APIPath == "" {
		spec.APIPath = "/" + spec.ID
	}
	if !strings.HasPrefix(spec.APIPath, "/") {
		spec.APIPath = "/" + spec.APIPath
	}
	spec.APIPath = strings.TrimRight(spec.APIPath, "/")
	if spec.APIPath == "" || strings.ContainsAny(spec.APIPath, " \"\\{}") {
		return fmt.Errorf("接口前缀 %q 不合法，应类似 /orders", spec.APIPath)
	}
	if spec.ClassPrefix == "" {
		spec.ClassPrefix = names.ClassName(spec.ID)
	}

	var deps []string
	seen := map[string]bool{}
	for _, d := range spec.Depends {
		d = strings.TrimSpace(d)
		if d == "" || seen[d] {
			continue
		}
		if d == spec.ID {
			return fmt.Errorf("模块不能依赖自己")
		}
		seen[d] = true
		deps = append(deps, d)
	}
	spec.Depends = deps
	return nil
}

// AddOptions 添加时的附加选项。
type AddOptions struct {
	SkipVerify bool
	// Output 编译命令的输出写到这里；为空时丢弃，失败时仍会返回输出。
	Output io.Writer
	Log    func(string)
}

// AddResult 添加结果。
type AddResult struct {
	Dir           string
	Package       string
	VerifyCommand string
	Verified      bool
	VerifySkipped string
	VerifyOutput  string
}

// Add 生成模块并登记到构建文件。
func (p *Project) Add(spec Spec, opts AddOptions) (*AddResult, error) {
	log := func(format string, a ...any) {
		if opts.Log != nil {
			opts.Log(fmt.Sprintf(format, a...))
		}
	}
	if err := p.Normalize(&spec); err != nil {
		return nil, err
	}
	registered, err := p.Adapter.Modules(p.Root, p.Meta)
	if err != nil {
		return nil, err
	}
	regSet := map[string]bool{}
	for _, r := range registered {
		regSet[r] = true
	}
	if regSet[spec.ID] {
		return nil, fmt.Errorf("模块 %s 已经登记在构建文件里", spec.ID)
	}
	for _, d := range spec.Depends {
		if !regSet[d] {
			return nil, fmt.Errorf("依赖的模块 %s 不在 %s 下，或还没有登记", d, p.Meta.AppsDir)
		}
	}
	dir := p.ModuleDir(spec.ID)
	if _, err := os.Stat(dir); err == nil {
		return nil, fmt.Errorf("%s 已存在，换一个模块名，或先删除这个目录", p.ModuleRel(spec.ID))
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	cleanup := func() { _ = source.RemoveAll(dir) }
	if err := p.Adapter.WriteModuleBuild(p.Root, p.Meta, build.ModuleSpec{ID: spec.ID, Depends: spec.Depends}, p.Info); err != nil {
		cleanup()
		return nil, fmt.Errorf("写 %s: %w", p.Adapter.BuildFile(), err)
	}
	if err := writeSources(dir, p.Meta, spec); err != nil {
		cleanup()
		return nil, fmt.Errorf("生成源码目录: %w", err)
	}
	log("已生成 %s", p.ModuleRel(spec.ID))

	if err := p.Adapter.Register(p.Root, p.Meta, spec.ID, p.Info); err != nil {
		_ = p.Adapter.Unregister(p.Root, p.Meta, spec.ID, p.Info)
		cleanup()
		return nil, fmt.Errorf("登记模块失败，已撤回: %w", err)
	}
	log("已登记到构建文件，%s 依赖 %s", p.Meta.AdminModule, spec.ID)

	res := &AddResult{Dir: dir, Package: p.Meta.BasePackage + "." + spec.Package}
	name, args := p.Adapter.VerifyCommand(p.Meta, spec.ID)
	res.VerifyCommand = name + " " + strings.Join(args, " ")
	if opts.SkipVerify {
		res.VerifySkipped = "按参数跳过"
		return res, nil
	}
	exe, err := exec.LookPath(name)
	if err != nil {
		res.VerifySkipped = "本机没有找到 " + name
		return res, nil
	}
	log("编译检查：%s", res.VerifyCommand)
	cmd := exec.Command(exe, args...)
	cmd.Dir = p.Root
	var buf strings.Builder
	if opts.Output != nil {
		cmd.Stdout = io.MultiWriter(&buf, opts.Output)
		cmd.Stderr = io.MultiWriter(&buf, opts.Output)
	} else {
		cmd.Stdout = &buf
		cmd.Stderr = &buf
	}
	err = cmd.Run()
	res.VerifyOutput = buf.String()
	res.Verified = err == nil
	return res, nil
}

// RemoveOptions 删除时的选项。
type RemoveOptions struct {
	// Force 有其他模块依赖它时仍然删除。
	Force bool
	Log   func(string)
}

// Remove 注销模块并删除目录。
func (p *Project) Remove(id string, opts RemoveOptions) error {
	log := func(format string, a ...any) {
		if opts.Log != nil {
			opts.Log(fmt.Sprintf(format, a...))
		}
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("模块名不能为空")
	}
	if names.ReservedModules[id] || id == p.Meta.AdminModule || id == p.Meta.CommonModule {
		return fmt.Errorf("%s 是骨架模块，不能用这个命令删除", id)
	}
	if strings.ContainsAny(id, `/\`) || id == "." || id == ".." {
		return fmt.Errorf("模块名 %q 不合法", id)
	}
	registered, err := p.Adapter.Modules(p.Root, p.Meta)
	if err != nil {
		return err
	}
	isReg := false
	for _, r := range registered {
		if r == id {
			isReg = true
		}
	}
	dir := p.ModuleDir(id)
	_, statErr := os.Stat(dir)
	dirExists := statErr == nil
	if !isReg && !dirExists {
		return fmt.Errorf("%s 下没有模块 %s", p.Meta.AppsDir, id)
	}

	deps, err := p.Adapter.Dependents(p.Root, p.Meta, id, p.Info)
	if err != nil {
		return err
	}
	if len(deps) > 0 && !opts.Force {
		return fmt.Errorf("%s 仍依赖 %s。先去掉这些依赖，或加 --force 强制删除", strings.Join(deps, "、"), id)
	}

	if err := p.Adapter.Unregister(p.Root, p.Meta, id, p.Info); err != nil {
		return fmt.Errorf("清理构建文件: %w", err)
	}
	log("已从构建文件去掉 %s", id)
	if dirExists {
		if err := source.RemoveAll(dir); err != nil {
			return fmt.Errorf("删除目录 %s: %w", dir, err)
		}
		log("已删除 %s", p.ModuleRel(id))
	}
	return nil
}
