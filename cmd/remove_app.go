package cmd

import (
	"github.com/spf13/cobra"
	"github.com/star/star-yi-cli/pkg/app"
)

var removeAppOpts app.RemoveOptions

var removeAppCmd = &cobra.Command{
	Use:   "app",
	Short: "删除 apps 业务模块",
	Long: `删除已有 Star-Yi 项目中的 apps/<module-id> 模块目录，
并自动移除父 pom.xml 与 yi-admin/pom.xml 中对应依赖。`,
	Example: `  star-yi-cli remove app --project . --module-id ed-pp`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return RunRemoveAppModule(removeAppOpts)
	},
}

func init() {
	removeCmd.AddCommand(removeAppCmd)

	removeAppCmd.Flags().StringVar(&removeAppOpts.ProjectRoot, "project", ".", "Star-Yi 项目根目录")
	removeAppCmd.Flags().StringVar(&removeAppOpts.ModuleID, "module-id", "", "要删除的模块名（apps 目录名）【必填】")
	_ = removeAppCmd.MarkFlagRequired("module-id")
}

