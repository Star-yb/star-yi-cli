package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/star/star-yi-cli/internal/config"
	"github.com/star/star-yi-cli/internal/create"
	"github.com/star/star-yi-cli/internal/names"
	"github.com/star/star-yi-cli/internal/prompt"
	"github.com/star/star-yi-cli/internal/scripts"
)

func init() {
	var o create.Options
	var edition, branch string
	cmd := &cobra.Command{
		Use:   "new",
		Short: "从仓库创建新项目",
		Example: `  star-yi-cli new --edition star-yi --parent D:\work --name my-platform
  star-yi-cli new --edition star-yi-arc --name my-arc --display-name "我的平台"
  star-yi-cli new --edition star-yi --name demo --repo D:\JAVA_File\springboot_web_file\Star-Yi`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if o.Name == "" {
				if !prompt.IsTerminal() {
					return fmt.Errorf("缺少 --name")
				}
				askNew(cfg, &edition, &o)
			}
			e, err := cfg.Edition(edition)
			if err != nil {
				return err
			}
			o.Edition = e
			if cmd.Flags().Changed("branch") {
				o.Branch = &branch
			}
			o.Log = logLine
			res, err := create.Run(o)
			if err != nil {
				return err
			}
			printCreated(res)
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVarP(&edition, "edition", "e", config.EditionStarYi, "版本 ID，见 star-yi-cli config show")
	f.StringVarP(&o.ParentDir, "parent", "p", ".", "创建位置（父目录）")
	f.StringVarP(&o.Name, "name", "n", "", "项目名：文件夹名，同时是 Maven artifactId / Gradle rootProject.name")
	f.StringVar(&o.DisplayName, "display-name", "", "展示名，写进 Maven <name> / Gradle description，默认同项目名")
	f.StringVar(&o.GroupID, "group-id", "", "groupId，默认取版本设置")
	f.StringVar(&o.Version, "version", "", "项目版本，默认取版本设置")
	f.StringVar(&o.Repo, "repo", "", "本次改用的仓库地址或本机路径，不写入设置")
	f.StringVar(&branch, "branch", "", "本次改用的分支，不写入设置")
	f.StringVar(&o.AppsDir, "apps-dir", "", "业务模块目录，默认取版本设置")
	f.BoolVar(&o.Force, "force", false, "目标目录非空时覆盖同名文件")
	f.BoolVar(&o.GitInit, "git-init", false, "创建后执行 git init")
	rootCmd.AddCommand(cmd)
}

func askNew(cfg *config.Config, edition *string, o *create.Options) {
	color.Cyan("── 创建新项目 ──")
	opts := make([]string, len(cfg.Editions))
	for i, e := range cfg.Editions {
		opts[i] = fmt.Sprintf("%s  %s", e.Title, e.Description)
	}
	e := cfg.Editions[prompt.Choose("选择版本：", opts)-1]
	*edition = e.ID
	o.ParentDir = prompt.Ask("创建位置（父目录）", o.ParentDir)
	o.Name = prompt.AskValid("项目名", "", names.ValidateProjectName)
	o.DisplayName = prompt.Ask("展示名", o.Name)
}

func printCreated(res *create.Result) {
	color.Green("创建完成：%s", res.Target)
	fmt.Printf("  坐标      %s:%s:%s\n", res.GroupID, res.Name, res.Version)
	fmt.Printf("  应用目录  %s\n", res.AppsDir)
	fmt.Println()
	fmt.Println("在项目里添加应用：")
	fmt.Printf("  cd \"%s\"\n", res.Target)
	fmt.Printf("  .\\%s add-app --module-id order-service\n", scripts.CmdName)
	fmt.Printf("  ./%s add-app --module-id order-service\n", filepath.ToSlash(scripts.ShName))
}
