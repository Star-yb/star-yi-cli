package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/star/star-yi-cli/internal/build"
	"github.com/star/star-yi-cli/internal/meta"
	"github.com/star/star-yi-cli/internal/names"
	"github.com/star/star-yi-cli/internal/scripts"
)

func init() {
	var root, appsDir string
	var force bool
	cmd := &cobra.Command{
		Use:   "init",
		Short: "给已有项目补上 .star-yi/project.yaml 和 star-yi 命令",
		Long: `给旧创建器生成的项目，或直接克隆下来的骨架，补上 .star-yi/project.yaml 和 star-yi.cmd / star-yi.sh。

版本按构建文件判断：有 pom.xml 是 Star-Yi，有 settings.gradle.kts 是 Star-Yi Arc。
已有 project.yaml 时只重写两个命令脚本，除非加 --force。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			abs, err := filepath.Abs(root)
			if err != nil {
				return err
			}
			p, err := meta.Load(abs)
			if err != nil {
				return err
			}
			a, err := build.For(p.Build)
			if err != nil {
				return err
			}
			if err := a.Validate(abs, p); err != nil {
				return err
			}
			if cmd.Flags().Changed("apps-dir") {
				d, err := names.CleanAppsDir(appsDir)
				if err != nil {
					return err
				}
				p.AppsDir = d
			}
			if p.Detected || force || cmd.Flags().Changed("apps-dir") {
				if err := meta.Write(abs, p); err != nil {
					return err
				}
				logLine("已写入 " + meta.Path(abs))
			}
			if err := os.MkdirAll(filepath.Join(abs, filepath.FromSlash(p.AppsDir)), 0o755); err != nil {
				return err
			}
			files, err := scripts.Write(abs)
			if err != nil {
				return err
			}
			for _, f := range files {
				logLine("已写入 " + f)
			}
			color.Green("完成。版本 %s，构建 %s，应用目录 %s", p.Edition, p.Build, p.AppsDir)
			fmt.Printf("旧的 star-yi-project-tool.* 和 common-core/ 可以删掉，改用 %s。\n", scripts.CmdName)
			return nil
		},
	}
	cmd.Flags().StringVar(&root, "project", ".", "项目根目录")
	cmd.Flags().StringVar(&appsDir, "apps-dir", "apps", "业务模块目录")
	cmd.Flags().BoolVar(&force, "force", false, "已有 project.yaml 时也重写")
	rootCmd.AddCommand(cmd)
}
