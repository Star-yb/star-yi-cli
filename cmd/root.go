package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "star-yi-cli",
	Short: "Star-Yi 项目脚手架：从模板创建工程或在已有项目中添加 apps 业务模块",
	Long: `star-yi-cli 用于基于 Star-Yi Maven 多模块骨架快速创建业务项目。

使用方式：
  直接运行 star-yi-cli          进入交互式问答菜单（推荐新手）
  star-yi-cli interactive       同上
  star-yi-cli new ...           命令行参数创建新项目
  star-yi-cli add app ...       命令行参数添加 apps 模块
  star-yi-cli tool init ...     生成项目专用工具脚本`,
}

// Execute 运行根命令。
func Execute() error {
	if ShouldRunInteractive() {
		return RunInteractive()
	}
	return rootCmd.Execute()
}

func exitErr(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
