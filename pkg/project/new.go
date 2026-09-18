package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/schollz/progressbar/v3"
	"github.com/star/star-yi-cli/pkg/fsutil"
	"github.com/star/star-yi-cli/pkg/template"
)

// NewOptions new 子命令参数。
type NewOptions struct {
	Template   string
	Target     string
	ArtifactID string
	GroupID    string
	Version    string
	Name       string
	AppsPrefix string
	APIPrefix  string
	SQLSubdir  string
	Force      bool
}

// Validate 校验必填项。
func (o *NewOptions) Validate() error {
	if strings.TrimSpace(o.Template) == "" {
		return fmt.Errorf("--template 不能为空")
	}
	if strings.TrimSpace(o.Target) == "" {
		return fmt.Errorf("--target 不能为空")
	}
	if strings.TrimSpace(o.ArtifactID) == "" {
		return fmt.Errorf("--artifact-id 不能为空")
	}
	if _, err := os.Stat(o.Template); err != nil {
		return fmt.Errorf("模板路径不可用: %w", err)
	}
	return nil
}

// CreateNewProject 从模板生成项目，返回绝对目标路径。
func CreateNewProject(opts NewOptions) (string, error) {
	target, err := filepath.Abs(opts.Target)
	if err != nil {
		return "", err
	}
	if err := fsutil.ConfirmOverwrite(target, opts.Force); err != nil {
		return "", err
	}

	srcRoot, cleanup, err := resolveTemplateRoot(opts.Template)
	if err != nil {
		return "", err
	}
	if cleanup != nil {
		defer cleanup()
	}

	name := opts.Name
	if name == "" {
		name = opts.ArtifactID
	}
	vars := template.Vars{
		ProjectArtifactID: opts.ArtifactID,
		ProjectName:       name,
		GroupID:           opts.GroupID,
		Version:           opts.Version,
		AppsModulePrefix:  opts.AppsPrefix,
		DefaultAPIPrefix:  opts.APIPrefix,
		SQLSubDir:         opts.SQLSubdir,
	}

	files, err := fsutil.WalkRelFiles(srcRoot)
	if err != nil {
		return "", err
	}

	bar := progressbar.Default(int64(len(files)), "生成项目")
	for _, rel := range files {
		src := filepath.Join(srcRoot, filepath.FromSlash(rel))
		dst := filepath.Join(target, filepath.FromSlash(rel))
		if err := template.ProcessFile(src, dst, vars); err != nil {
			return "", fmt.Errorf("处理 %s: %w", rel, err)
		}
		_ = bar.Add(1)
	}

	return target, nil
}

func resolveTemplateRoot(templatePath string) (root string, cleanup func(), err error) {
	abs, err := filepath.Abs(templatePath)
	if err != nil {
		return "", nil, err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", nil, err
	}

	if info.IsDir() {
		return abs, nil, nil
	}

	if !fsutil.IsZipFile(abs) {
		return "", nil, fmt.Errorf("模板必须是文件夹或 .zip 文件: %s", abs)
	}

	tmp, err := os.MkdirTemp("", "star-yi-template-*")
	if err != nil {
		return "", nil, err
	}
	if err := fsutil.ExtractZip(abs, tmp); err != nil {
		os.RemoveAll(tmp)
		return "", nil, err
	}

	// zip 常含单层根目录，若仅有一个子目录则进入该目录
	entries, err := os.ReadDir(tmp)
	if err != nil {
		os.RemoveAll(tmp)
		return "", nil, err
	}
	if len(entries) == 1 && entries[0].IsDir() {
		return filepath.Join(tmp, entries[0].Name()), func() { os.RemoveAll(tmp) }, nil
	}
	return tmp, func() { os.RemoveAll(tmp) }, nil
}
