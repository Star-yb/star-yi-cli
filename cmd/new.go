package cmd

import (
	"github.com/spf13/cobra"
	"github.com/star/star-yi-cli/pkg/project"
)

var newOpts project.NewOptions

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "从模板创建全新的 Star-Yi 项目",
	Long: `从本地文件夹或 .zip 模板复制项目结构，并替换占位符后输出到目标目录。

占位符：{projectArtifactId}、{projectName}、{groupId}、{version}、
{appsModulePrefix}、{defaultApiPrefix}、{sqlSubDir}`,
	Example: `  star-yi-cli new --template ./star-yi-template --target ./my-project --artifact-id My-Platform
  star-yi-cli new --template ~/template.zip --target ./edu --artifact-id Education-Yi --name "教育平台"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return RunNewProject(newOpts)
	},
}

func init() {
	rootCmd.AddCommand(newCmd)

	newCmd.Flags().StringVar(&newOpts.Template, "template", "", "模板路径（文件夹或 .zip）【必填】")
	newCmd.Flags().StringVar(&newOpts.Target, "target", "", "目标项目根目录【必填】")
	newCmd.Flags().StringVar(&newOpts.ArtifactID, "artifact-id", "", "Maven artifactId【必填】")
	newCmd.Flags().StringVar(&newOpts.GroupID, "group-id", "com.star", "Maven groupId")
	newCmd.Flags().StringVar(&newOpts.Version, "version", "0.0.1-SNAPSHOT", "项目版本")
	newCmd.Flags().StringVar(&newOpts.Name, "name", "", "项目展示名称（默认与 artifact-id 相同）")
	newCmd.Flags().StringVar(&newOpts.AppsPrefix, "apps-prefix", "", "业务模块命名习惯（仅写入占位符，可选）")
	newCmd.Flags().StringVar(&newOpts.APIPrefix, "api-prefix", "", "默认 API 前缀（可选）")
	newCmd.Flags().StringVar(&newOpts.SQLSubdir, "sql-subdir", "", "SQL 子目录名（可选）")
	newCmd.Flags().BoolVar(&newOpts.Force, "force", false, "目标目录非空时强制覆盖")

	_ = newCmd.MarkFlagRequired("template")
	_ = newCmd.MarkFlagRequired("target")
	_ = newCmd.MarkFlagRequired("artifact-id")
}
