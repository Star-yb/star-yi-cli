package cmd

import (
	"github.com/spf13/cobra"
	"github.com/star/star-yi-cli/pkg/app"
)

var addAppOpts app.AddOptions

var addAppCmd = &cobra.Command{
	Use:   "app",
	Short: "添加 apps 业务模块",
	Long: `在已有 Star-Yi 项目中创建 apps/<module-id> 模块，并更新父 pom 与 yi-admin 依赖。

要求当前目录（或 --project）为 Star-Yi 根目录，且存在 pom.xml、yi-admin 子模块。`,
	Example: `  star-yi-cli add app --module-id order-service --package-suffix order --api-path /orders
  star-yi-cli add app --project . --module-id ed-school --package-suffix edschool --depends inventory`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return RunAddAppModule(addAppOpts)
	},
}

func init() {
	addCmd.AddCommand(addAppCmd)

	addAppCmd.Flags().StringVar(&addAppOpts.ProjectRoot, "project", ".", "Star-Yi 项目根目录")
	addAppCmd.Flags().StringVar(&addAppOpts.ModuleID, "module-id", "", "Maven artifactId，同时作为 apps 目录名【必填】")
	addAppCmd.Flags().StringVar(&addAppOpts.PackageSuffix, "package-suffix", "", "Java 包后缀，如 order → com.star.order【必填】")
	addAppCmd.Flags().StringVar(&addAppOpts.APIPath, "api-path", "", "Controller @RequestMapping 前缀（默认 /{module-id}）")
	addAppCmd.Flags().StringSliceVar(&addAppOpts.Depends, "depends", nil, "依赖的同仓库 apps 模块列表")
	addAppCmd.Flags().BoolVar(&addAppOpts.SkipCompile, "skip-compile", false, "跳过 mvn compile 验证")

	_ = addAppCmd.MarkFlagRequired("module-id")
	_ = addAppCmd.MarkFlagRequired("package-suffix")
}
