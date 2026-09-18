package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/star/star-yi-cli/pkg/app"
	"github.com/star/star-yi-cli/pkg/fsutil"
	"github.com/star/star-yi-cli/pkg/pom"
	"github.com/star/star-yi-cli/pkg/project"
	"github.com/star/star-yi-cli/pkg/prompt"
)

// RunInteractive 问答式主菜单。
func RunInteractive() error {
	color.Cyan("╔══════════════════════════════════════╗")
	color.Cyan("║     Star-Yi 项目创建器 (交互模式)     ║")
	color.Cyan("╚══════════════════════════════════════╝")
	fmt.Println("直接运行 star-yi-cli 进入本模式；也可使用命令行参数，见: star-yi-cli --help")
	fmt.Println()

	for {
		choice := prompt.Choose("请选择要执行的操作：", []string{
			"创建新项目 (new)",
			"添加 apps 业务模块 (add app)",
			"删除 apps 业务模块 (remove app)",
			"退出",
		})

		var err error
		switch choice {
		case 1:
			err = interactiveNewProject()
		case 2:
			err = interactiveAddApp()
		case 3:
			err = interactiveRemoveApp()
		case 4:
			color.Cyan("再见。")
			return nil
		}

		if err != nil {
			color.Red("✗ 操作失败: %v", err)
		}
		fmt.Println()
		if !prompt.AskYesNo("是否返回主菜单继续", true) {
			return nil
		}
	}
}

func interactiveNewProject() error {
	fmt.Println()
	color.Yellow("── 创建新项目 ──")
	fmt.Println("只需填写创建位置与项目名称，其余使用默认（模板、groupId、版本等）。")

	tpl := prompt.DefaultTemplatePath()
	if tpl == "" {
		return fmt.Errorf("未找到默认模板，请将 templates/Star-Yi.zip 放在 exe 同目录下")
	}

	parentDir := prompt.Ask("创建位置（父目录，留空=当前目录）", ".")
	projectName := prompt.AskRequired("项目名称（文件夹名 / Maven artifactId）")

	target, err := project.ProjectDir(parentDir, projectName)
	if err != nil {
		return err
	}

	force := false
	nonEmpty, err := fsutil.DirNotEmpty(target)
	if err != nil {
		return err
	}
	if nonEmpty {
		color.Yellow("目录已存在且非空: %s", target)
		if !prompt.AskYesNo("是否覆盖", false) {
			fmt.Println("已取消。")
			return nil
		}
		force = true
	}

	fmt.Println()
	color.Cyan("将创建: %s", target)
	fmt.Printf("  Maven: com.star:%s:0.0.1-SNAPSHOT\n", projectName)

	opts, err := project.DefaultNewOptions(tpl, parentDir, projectName, force)
	if err != nil {
		return err
	}
	return RunNewProject(opts)
}

func interactiveAddApp() error {
	fmt.Println()
	color.Yellow("── 添加 apps 业务模块 ──")
	fmt.Println("只需填写 3 项：项目根目录、模块名、包后缀。")

	projectRoot := prompt.Ask("Star-Yi 项目根目录（留空=自动搜索当前目录第一个项目）", "")
	projectRoot = strings.TrimSpace(projectRoot)
	if projectRoot == "" {
		autoRoot, err := findFirstStarYiProject(".")
		if err != nil {
			return fmt.Errorf("未找到有效项目根目录: %w", err)
		}
		projectRoot = autoRoot
		color.Cyan("已自动选择项目根目录: %s", projectRoot)
	} else if err := pom.ValidateStarYiProject(projectRoot); err != nil {
		return fmt.Errorf("未找到有效项目根目录: %w", err)
	}
	moduleID := prompt.AskRequired("模块 artifactId（apps 目录名）")
	packageSuffix := prompt.Ask("Java 包后缀（留空默认 xxx）", "xxx")

	fmt.Println()
	color.Cyan("即将添加模块，请确认：")
	fmt.Printf("  项目根: %s\n", projectRoot)
	fmt.Printf("  模块: apps/%s\n", moduleID)
	fmt.Printf("  包名: com.star.%s\n", packageSuffix)
	fmt.Println("  编译验证: 强制执行 mvn compile")
	if !prompt.AskYesNo("确认执行", true) {
		fmt.Println("已取消。")
		return nil
	}

	return RunAddAppModule(app.AddOptions{
		ProjectRoot:   projectRoot,
		ModuleID:      moduleID,
		PackageSuffix: packageSuffix,
		SkipCompile:   false,
	})
}

func interactiveRemoveApp() error {
	fmt.Println()
	color.Yellow("── 删除 apps 业务模块 ──")
	fmt.Println("填写项目根目录和模块名，工具将自动删除目录并清理 POM。")

	projectRoot := prompt.Ask("Star-Yi 项目根目录（留空=自动搜索当前目录第一个项目）", "")
	projectRoot = strings.TrimSpace(projectRoot)
	if projectRoot == "" {
		autoRoot, err := findFirstStarYiProject(".")
		if err != nil {
			return fmt.Errorf("未找到有效项目根目录: %w", err)
		}
		projectRoot = autoRoot
		color.Cyan("已自动选择项目根目录: %s", projectRoot)
	} else if err := pom.ValidateStarYiProject(projectRoot); err != nil {
		return fmt.Errorf("未找到有效项目根目录: %w", err)
	}
	moduleID := prompt.AskRequired("要删除的模块 artifactId（apps 目录名）")

	fmt.Println()
	color.Cyan("即将删除模块，请确认：")
	fmt.Printf("  项目根: %s\n", projectRoot)
	fmt.Printf("  删除: apps/%s\n", moduleID)
	fmt.Println("  将同步移除父 pom.xml / yi-admin/pom.xml 的依赖引用")
	if !prompt.AskYesNo("确认执行", true) {
		fmt.Println("已取消。")
		return nil
	}

	return RunRemoveAppModule(app.RemoveOptions{
		ProjectRoot: projectRoot,
		ModuleID:    moduleID,
	})
}

// ShouldRunInteractive 无子命令或显式 interactive 时进入问答模式。
func ShouldRunInteractive() bool {
	if len(os.Args) <= 1 {
		return true
	}
	switch os.Args[1] {
	case "interactive", "-i", "--interactive":
		return true
	}
	return false
}

func findFirstStarYiProject(root string) (string, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}

	// 先检查当前目录本身
	if err := pom.ValidateStarYiProject(absRoot); err == nil {
		return absRoot, nil
	}

	entries, err := os.ReadDir(absRoot)
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		candidate := filepath.Join(absRoot, entry.Name())
		if err := pom.ValidateStarYiProject(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("在 %s 及其一级子目录中未找到 Star-Yi 项目", absRoot)
}
