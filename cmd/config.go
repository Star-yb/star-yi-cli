package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/star/star-yi-cli/internal/config"
	"github.com/star/star-yi-cli/internal/source"
)

func init() {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "查看或修改版本设置（仓库地址、分支、应用目录等）",
	}

	pathCmd := &cobra.Command{
		Use:   "path",
		Short: "显示配置文件位置",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := config.Path()
			if err != nil {
				return err
			}
			fmt.Println(p)
			return nil
		},
	}

	showCmd := &cobra.Command{
		Use:   "show",
		Short: "显示全部版本设置",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			p, _ := config.Path()
			fmt.Printf("配置文件 %s\n", p)
			for _, e := range cfg.Editions {
				fmt.Println()
				color.Cyan("%s  (%s)", e.Title, e.ID)
				fmt.Printf("  仓库      %s\n", e.Repo)
				fmt.Printf("  分支      %s\n", orDefault(e.Branch, "默认分支"))
				fmt.Printf("  构建      %s · %s\n", e.Build, e.Language)
				fmt.Printf("  应用目录  %s\n", e.AppsDir)
				fmt.Printf("  坐标默认  %s · %s\n", e.GroupID, e.Version)
			}
			return nil
		},
	}

	var edition string
	var set config.Edition
	setCmd := &cobra.Command{
		Use:     "set",
		Short:   "修改一个版本的设置，只改传入的项",
		Example: `  star-yi-cli config set --edition star-yi --repo https://github.com/you/Star-Yi.git --branch dev
  star-yi-cli config set --edition star-yi-arc --apps-dir modules`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			e, err := cfg.Edition(edition)
			if err != nil {
				return err
			}
			f := cmd.Flags()
			apply := func(name string, dst *string, val string) {
				if f.Changed(name) {
					*dst = val
				}
			}
			apply("title", &e.Title, set.Title)
			apply("description", &e.Description, set.Description)
			apply("repo", &e.Repo, set.Repo)
			apply("branch", &e.Branch, set.Branch)
			apply("build", &e.Build, set.Build)
			apply("language", &e.Language, set.Language)
			apply("apps-dir", &e.AppsDir, set.AppsDir)
			apply("group-id", &e.GroupID, set.GroupID)
			apply("version", &e.Version, set.Version)
			if err := config.Save(cfg); err != nil {
				return err
			}
			color.Green("已保存 %s", e.ID)
			return nil
		},
	}
	sf := setCmd.Flags()
	sf.StringVarP(&edition, "edition", "e", "", "版本 ID")
	_ = setCmd.MarkFlagRequired("edition")
	sf.StringVar(&set.Title, "title", "", "名称")
	sf.StringVar(&set.Description, "description", "", "说明")
	sf.StringVar(&set.Repo, "repo", "", "仓库地址或本机路径")
	sf.StringVar(&set.Branch, "branch", "", "分支，留空表示默认分支")
	sf.StringVar(&set.Build, "build", "", "构建方式：maven 或 gradle")
	sf.StringVar(&set.Language, "language", "", "新模块的语言：java 或 kotlin")
	sf.StringVar(&set.AppsDir, "apps-dir", "", "新项目的业务模块目录")
	sf.StringVar(&set.GroupID, "group-id", "", "默认 groupId")
	sf.StringVar(&set.Version, "version", "", "默认版本")

	resetCmd := &cobra.Command{
		Use:   "reset <edition>",
		Short: "把内置版本恢复为出厂设置",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			d, ok := config.DefaultEdition(args[0])
			if !ok {
				return fmt.Errorf("%s 不是内置版本", args[0])
			}
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			for i, e := range cfg.Editions {
				if e.ID == d.ID {
					cfg.Editions[i] = d
				}
			}
			if err := config.Save(cfg); err != nil {
				return err
			}
			color.Green("已恢复 %s", d.ID)
			return nil
		},
	}

	checkCmd := &cobra.Command{
		Use:   "check <edition>",
		Short: "检查版本的仓库地址能否访问",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			e, err := cfg.Edition(args[0])
			if err != nil {
				return err
			}
			msg, err := source.Check(e.Repo, e.Branch)
			if err != nil {
				return err
			}
			color.Green(msg)
			return nil
		},
	}

	configCmd.AddCommand(pathCmd, showCmd, setCmd, resetCmd, checkCmd)
	rootCmd.AddCommand(configCmd)
}

func orDefault(s, d string) string {
	if s == "" {
		return d
	}
	return s
}
