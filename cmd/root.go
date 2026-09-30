// Package cmd 命令行入口。
package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/star/star-yi-cli/internal/config"
)

// Version 由构建参数 -ldflags "-X github.com/star/star-yi-cli/cmd.Version=..." 覆盖。
var Version = "2.0.0"

var rootCmd = &cobra.Command{
	Use:   "star-yi-cli",
	Short: "Star-Yi 创建器：从仓库创建 Star-Yi / Star-Yi Arc 项目，并在项目里增删应用模块",
	Long: `star-yi-cli 从配置的仓库下载骨架，改成新项目的名字，再在项目根目录留下 star-yi 命令。

  star-yi-cli                  打开创建器页面（创建项目、修改版本设置）
  star-yi-cli new ...          在终端创建项目
  star-yi-cli app ...          在已有项目里添加、删除、查看应用模块
  star-yi-cli config ...       查看或修改版本设置
  star-yi-cli init ...         给旧项目补上 .star-yi/project.yaml 和 star-yi 命令`,
	Version:       Version,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runUI(uiOptions{OpenBrowser: true})
	},
}

// Execute 运行命令。
func Execute() int {
	_ = config.RecordCLIPath()
	if err := rootCmd.Execute(); err != nil {
		color.New(color.FgRed).Fprintln(os.Stderr, "错误：", err)
		return 1
	}
	return 0
}

func logLine(s string) {
	color.New(color.FgCyan).Print("· ")
	fmt.Println(s)
}
