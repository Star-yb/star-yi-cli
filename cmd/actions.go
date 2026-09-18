package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/star/star-yi-cli/pkg/app"
	"github.com/star/star-yi-cli/pkg/project"
	"github.com/star/star-yi-cli/pkg/projecttool"
)

// RunNewProject 执行 new 逻辑并输出结果。
func RunNewProject(opts project.NewOptions) error {
	if err := opts.Validate(); err != nil {
		return err
	}
	target, err := project.CreateNewProject(opts)
	if err != nil {
		return err
	}

	toolRes, toolErr := projecttool.InitProjectTool(projecttool.InitOptions{
		ProjectRoot: target,
		CLICommand:  "star-yi-cli",
		Force:       true,
	})

	color.Green("✓ 项目已生成：%s", target)
	if toolErr != nil {
		color.Yellow("! 项目专用脚本生成失败：%v", toolErr)
	} else {
		color.Green("✓ 已自动生成项目专用脚本：")
		for _, p := range toolRes.Generated {
			fmt.Printf("  %s\n", p)
		}
		if toolRes.CLIPath != "" {
			fmt.Printf("  %s\n", toolRes.CLIPath)
		}
	}
	fmt.Println()
	color.Cyan("下一步：")
	fmt.Printf("  cd %s\n", target)
	fmt.Println("  mvn clean compile")
	fmt.Println("  mvn -pl yi-admin spring-boot:run")
	return nil
}

// RunAddAppModule 执行 add app 逻辑并输出结果。
func RunAddAppModule(opts app.AddOptions) error {
	if err := opts.Validate(); err != nil {
		return err
	}
	if err := app.AddAppModule(opts); err != nil {
		return err
	}
	color.Green("✓ 业务模块 apps/%s 已创建并注册", opts.ModuleID)
	fmt.Println()
	fmt.Println("启动项目：mvn -pl yi-admin spring-boot:run")
	return nil
}

// RunRemoveAppModule 执行 remove app 逻辑并输出结果。
func RunRemoveAppModule(opts app.RemoveOptions) error {
	if err := app.RemoveAppModule(opts); err != nil {
		return err
	}
	color.Green("✓ 业务模块 apps/%s 已删除并从 pom 中移除", opts.ModuleID)
	return nil
}
