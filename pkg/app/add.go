package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/star/star-yi-cli/pkg/pom"
)

// AddOptions add app 子命令参数。
type AddOptions struct {
	ProjectRoot   string
	ModuleID      string
	PackageSuffix string
	APIPath       string
	Depends       []string
	SkipCompile   bool
}

// RemoveOptions remove app 子命令参数。
type RemoveOptions struct {
	ProjectRoot string
	ModuleID    string
}

// Validate 校验参数。
func (o *AddOptions) Validate() error {
	if strings.TrimSpace(o.ModuleID) == "" {
		return fmt.Errorf("--module-id 不能为空")
	}
	if strings.TrimSpace(o.PackageSuffix) == "" {
		return fmt.Errorf("--package-suffix 不能为空")
	}
	if strings.ContainsAny(o.ModuleID, `/\ `) {
		return fmt.Errorf("--module-id 不能包含路径分隔符或空格")
	}
	if strings.ContainsAny(o.PackageSuffix, `/\ .`) {
		return fmt.Errorf("--package-suffix 应为合法 Java 包段，如 order、edschool")
	}
	return nil
}

// Validate 校验 remove 参数。
func (o *RemoveOptions) Validate() error {
	if strings.TrimSpace(o.ModuleID) == "" {
		return fmt.Errorf("--module-id 不能为空")
	}
	if strings.ContainsAny(o.ModuleID, `/\ `) {
		return fmt.Errorf("--module-id 不能包含路径分隔符或空格")
	}
	return nil
}

// AddAppModule 创建 apps 模块并更新 POM。
func AddAppModule(opts AddOptions) error {
	root, err := filepath.Abs(opts.ProjectRoot)
	if err != nil {
		return err
	}
	if err := pom.ValidateStarYiProject(root); err != nil {
		return err
	}

	moduleDir := filepath.Join(root, "apps", opts.ModuleID)
	if _, err := os.Stat(moduleDir); err == nil {
		return fmt.Errorf("apps/%s 已存在，请更换 --module-id 或手动删除后重试", opts.ModuleID)
	} else if !os.IsNotExist(err) {
		return err
	}

	coords, err := pom.LoadCoordinates(root)
	if err != nil {
		return err
	}

	apiPath := opts.APIPath
	if apiPath == "" {
		apiPath = "/" + opts.ModuleID
	}

	if err := os.MkdirAll(moduleDir, 0o755); err != nil {
		return err
	}

	pomContent := renderModulePOM(coords, opts.ModuleID, opts.Depends)
	if err := os.WriteFile(filepath.Join(moduleDir, "pom.xml"), []byte(pomContent), 0o644); err != nil {
		return err
	}

	if err := createPackageDirs(moduleDir, opts.PackageSuffix); err != nil {
		return err
	}
	if err := createTestController(moduleDir, opts.PackageSuffix, opts.ModuleID, apiPath); err != nil {
		return err
	}

	readme := fmt.Sprintf(`# %s

- Java 包根: com.star.%s
- 建议 API 前缀: %s
- 下一步: 使用 star-yi-cli add crud（规划中）或参考 yi-demo 的 DemoArticle 五层结构
`, opts.ModuleID, opts.PackageSuffix, apiPath)
	if err := os.WriteFile(filepath.Join(moduleDir, "README.md"), []byte(readme), 0o644); err != nil {
		return err
	}

	if err := pom.RegisterAppModule(root, opts.ModuleID, coords); err != nil {
		_ = os.RemoveAll(moduleDir)
		return fmt.Errorf("更新 POM 失败（已尝试创建目录，请检查 pom.xml）: %w", err)
	}

	if !opts.SkipCompile {
		runMavenCompile(root, opts.ModuleID)
	}

	return nil
}

// RemoveAppModule 删除 apps 模块目录并移除 POM 注册。
func RemoveAppModule(opts RemoveOptions) error {
	if err := opts.Validate(); err != nil {
		return err
	}
	root, err := filepath.Abs(opts.ProjectRoot)
	if err != nil {
		return err
	}
	if err := pom.ValidateStarYiProject(root); err != nil {
		return err
	}

	moduleDir := filepath.Join(root, "apps", opts.ModuleID)
	if _, err := os.Stat(moduleDir); os.IsNotExist(err) {
		return fmt.Errorf("apps/%s 不存在", opts.ModuleID)
	} else if err != nil {
		return err
	}

	if err := pom.UnregisterAppModule(root, opts.ModuleID); err != nil {
		return err
	}
	if err := os.RemoveAll(moduleDir); err != nil {
		return fmt.Errorf("删除目录失败: %w", err)
	}
	return nil
}

func createPackageDirs(moduleDir, packageSuffix string) error {
	base := filepath.Join(moduleDir, "src", "main", "java", "com", "star", packageSuffix)
	dirs := []string{
		filepath.Join(base, "model", "entity"),
		filepath.Join(base, "dao"),
		filepath.Join(base, "repository"),
		filepath.Join(base, "service"),
		filepath.Join(base, "service", "impl"),
		filepath.Join(base, "controller"),
		filepath.Join(moduleDir, "src", "main", "dto"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
		keep := filepath.Join(d, ".gitkeep")
		if err := os.WriteFile(keep, []byte{}, 0o644); err != nil {
			return err
		}
	}
	return nil
}

func createTestController(moduleDir, packageSuffix, moduleID, apiPath string) error {
	controllerDir := filepath.Join(moduleDir, "src", "main", "java", "com", "star", packageSuffix, "controller")
	className := toClassName(moduleID) + "TestController"
	content := fmt.Sprintf(`package com.star.%s.controller;

import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

@RestController
@RequestMapping("%s")
public class %s {

    @RequestMapping("/test")
    public String test() {
        return "test";
    }
}
`, packageSuffix, apiPath, className)
	return os.WriteFile(filepath.Join(controllerDir, className+".java"), []byte(content), 0o644)
}

func toClassName(moduleID string) string {
	parts := strings.FieldsFunc(moduleID, func(r rune) bool {
		return r == '-' || r == '_' || r == '.'
	})
	var b strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		b.WriteString(strings.ToUpper(p[:1]))
		if len(p) > 1 {
			b.WriteString(p[1:])
		}
	}
	if b.Len() == 0 {
		return "App"
	}
	return b.String()
}

func runMavenCompile(projectRoot, moduleID string) {
	mvn, err := exec.LookPath("mvn")
	if err != nil {
		color.Yellow("未检测到 mvn，已跳过编译验证。安装 Maven 后可执行：")
		fmt.Printf("  cd %s && mvn compile -pl apps/%s -am\n", projectRoot, moduleID)
		return
	}
	cmd := exec.Command(mvn, "compile", "-pl", "apps/"+moduleID, "-am", "-q")
	cmd.Dir = projectRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		color.Yellow("mvn compile 未通过（模块已创建，请本地排查）：\n%s", string(out))
		return
	}
	color.Green("mvn compile -pl apps/%s -am 验证通过", moduleID)
}
