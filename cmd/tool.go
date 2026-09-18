package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/star/star-yi-cli/pkg/projecttool"
)

var toolCmd = &cobra.Command{
	Use:   "tool",
	Short: "生成项目专用工具脚本",
}

var toolInitOpts projecttool.InitOptions

var toolInitCmd = &cobra.Command{
	Use:   "init",
	Short: "在项目根目录生成跨平台工具脚本",
	Long: `生成当前项目专用脚本（Windows/macOS/Linux），
后续通过脚本固定操作该项目，无需重复传 --project 参数。`,
	Example: `  star-yi-cli tool init --project .
  star-yi-cli tool init --project D:\JAVA_File\111\sss --cli star-yi-cli.exe --force
  star-yi-cli tool init --project . --all-platform`,
	RunE: func(cmd *cobra.Command, args []string) error {
		res, err := projecttool.InitProjectTool(toolInitOpts)
		if err != nil {
			return err
		}
		color.Green("✓ 已生成项目专用工具脚本：%s", res.ProjectRoot)
		for _, p := range res.Generated {
			fmt.Println("  - " + p)
		}
		if res.CLIPath != "" {
			fmt.Println("  - " + res.CLIPath)
		}
		fmt.Println()
		color.Cyan("示例：")
		fmt.Println(`  star-yi-project-tool.cmd add-app --module-id ed-pp --package-suffix aaa`)
		fmt.Println(`  ./star-yi-project-tool.sh remove-app --module-id ed-pp`)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(toolCmd)
	toolCmd.AddCommand(toolInitCmd)

	toolInitCmd.Flags().StringVar(&toolInitOpts.ProjectRoot, "project", ".", "Star-Yi 项目根目录")
	toolInitCmd.Flags().StringVar(&toolInitOpts.CLICommand, "cli", "star-yi-cli", "脚本中使用的 CLI 命令名")
	toolInitCmd.Flags().BoolVar(&toolInitOpts.Force, "force", false, "覆盖已存在的脚本文件")
	toolInitCmd.Flags().BoolVar(&toolInitOpts.AllPlatforms, "all-platform", false, "同时生成 Windows/macOS/Linux 三类脚本")
}

