package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/star/star-yi-cli/internal/app"
	"github.com/star/star-yi-cli/internal/meta"
	"github.com/star/star-yi-cli/internal/names"
	"github.com/star/star-yi-cli/internal/prompt"
)

var appProject string

func init() {
	appCmd := &cobra.Command{
		Use:   "app",
		Short: "在已有项目里添加、删除、查看应用模块",
		Long: `在已有项目里添加、删除、查看应用模块。项目根目录的 star-yi.cmd / star-yi.sh 调用的就是这个命令。

不带子命令时进入菜单。--project 省略时从当前目录向上查找项目根。`,
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := openProject()
			if err != nil {
				return err
			}
			return appMenu(p)
		},
	}
	appCmd.PersistentFlags().StringVar(&appProject, "project", "", "项目根目录，默认从当前目录向上查找")

	var spec app.Spec
	var depends string
	var skipVerify bool
	addCmd := &cobra.Command{
		Use:     "add",
		Aliases: []string{"add-app"},
		Short:   "添加应用模块",
		Example: `  star-yi.cmd add-app --module-id order-service
  star-yi.cmd add-app --module-id order-service --package order --api-path /orders --depends user-center`,
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := openProject()
			if err != nil {
				return err
			}
			if spec.ID == "" {
				if !prompt.IsTerminal() {
					return fmt.Errorf("缺少 --module-id")
				}
				return interactiveAdd(p)
			}
			spec.Depends = splitList(depends)
			return runAdd(p, spec, skipVerify)
		},
	}
	addCmd.Flags().StringVarP(&spec.ID, "module-id", "m", "", "模块名，例如 order-service")
	addCmd.Flags().StringVar(&spec.Package, "package", "", "包名后缀，拼在 com.star 后面；默认由模块名推出")
	addCmd.Flags().StringVar(&spec.Package, "package-suffix", "", "同 --package")
	_ = addCmd.Flags().MarkHidden("package-suffix")
	addCmd.Flags().StringVar(&spec.APIPath, "api-path", "", "接口前缀，默认 /<模块名>")
	addCmd.Flags().StringVar(&depends, "depends", "", "依赖的其他应用模块，逗号分隔")
	addCmd.Flags().BoolVar(&skipVerify, "skip-verify", false, "跳过编译检查")

	var removeID string
	var force, yes bool
	removeCmd := &cobra.Command{
		Use:     "remove",
		Aliases: []string{"remove-app", "rm"},
		Short:   "删除应用模块：去掉构建文件里的引用并删除目录",
		Example: `  star-yi.cmd remove-app --module-id order-service`,
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := openProject()
			if err != nil {
				return err
			}
			if removeID == "" {
				if !prompt.IsTerminal() {
					return fmt.Errorf("缺少 --module-id")
				}
				return interactiveRemove(p)
			}
			if !yes && prompt.IsTerminal() && !prompt.YesNo(fmt.Sprintf("删除 %s，包括目录里的全部代码？", p.ModuleRel(removeID)), false) {
				fmt.Println("已取消。")
				return nil
			}
			return runRemove(p, removeID, force)
		},
	}
	removeCmd.Flags().StringVarP(&removeID, "module-id", "m", "", "要删除的模块名")
	removeCmd.Flags().BoolVar(&force, "force", false, "有其他模块依赖它时也删除")
	removeCmd.Flags().BoolVarP(&yes, "yes", "y", false, "不再确认")

	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"list-apps", "ls"},
		Short:   "查看应用模块",
		RunE: func(cmd *cobra.Command, args []string) error {
			p, err := openProject()
			if err != nil {
				return err
			}
			return printModules(p)
		},
	}

	appCmd.AddCommand(addCmd, removeCmd, listCmd)
	rootCmd.AddCommand(appCmd)
}

func openProject() (*app.Project, error) {
	root := appProject
	if root == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		root, err = meta.FindRoot(wd)
		if err != nil {
			return nil, err
		}
	}
	return app.Open(root)
}

func splitList(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func printHeader(p *app.Project) {
	color.Cyan("%s  %s:%s:%s", p.Info.DisplayName, p.Info.GroupID, p.Info.Name, p.Info.Version)
	fmt.Printf("项目根 %s\n构建 %s · 语言 %s · 应用目录 %s\n", p.Root, p.Meta.Build, p.Meta.Language, p.Meta.AppsDir)
	if p.Meta.Detected {
		color.Yellow("没有 .star-yi/project.yaml，按目录结构推断。可以运行 star-yi-cli init 补上。")
	}
}

func appMenu(p *app.Project) error {
	printHeader(p)
	for {
		fmt.Println()
		var err error
		switch prompt.Choose("要做什么：", []string{"添加应用", "删除应用", "查看应用", "退出"}) {
		case 1:
			err = interactiveAdd(p)
		case 2:
			err = interactiveRemove(p)
		case 3:
			err = printModules(p)
		default:
			return nil
		}
		if err != nil {
			color.Red("失败：%v", err)
		}
	}
}

func interactiveAdd(p *app.Project) error {
	fmt.Println()
	color.Cyan("── 添加应用 ──")
	id := prompt.AskValid("模块名（例如 order-service）", "", func(s string) error {
		spec := app.Spec{ID: s}
		return p.Normalize(&spec)
	})
	spec := app.Spec{ID: id}
	spec.Package = prompt.AskValid("包名后缀（拼在 "+p.Meta.BasePackage+" 后面）", names.DefaultPackage(id), names.ValidatePackage)
	spec.APIPath = prompt.Ask("接口前缀", "/"+id)
	if mods, _ := p.Adapter.Modules(p.Root, p.Meta); len(mods) > 0 {
		fmt.Printf("已有应用：%s\n", strings.Join(mods, "、"))
		spec.Depends = prompt.List("依赖哪些应用（逗号分隔，可留空）")
	}
	verify := prompt.YesNo("添加后编译检查", true)

	fmt.Println()
	fmt.Printf("  目录  %s\n  包    %s.%s\n  接口  %s\n", p.ModuleRel(id), p.Meta.BasePackage, spec.Package, spec.APIPath)
	if !prompt.YesNo("确认添加", true) {
		fmt.Println("已取消。")
		return nil
	}
	return runAdd(p, spec, !verify)
}

func runAdd(p *app.Project, spec app.Spec, skipVerify bool) error {
	res, err := p.Add(spec, app.AddOptions{SkipVerify: skipVerify, Log: logLine})
	if err != nil {
		return err
	}
	switch {
	case res.VerifySkipped != "":
		color.Yellow("没有编译检查（%s）。手动检查：%s", res.VerifySkipped, res.VerifyCommand)
	case res.Verified:
		color.Green("编译通过")
	default:
		color.Yellow("编译没有通过，模块已经加好，请按输出排查：")
		fmt.Println(strings.TrimSpace(res.VerifyOutput))
	}
	color.Green("已添加 %s，包 %s", p.ModuleRel(spec.ID), res.Package)
	return nil
}

func interactiveRemove(p *app.Project) error {
	fmt.Println()
	color.Cyan("── 删除应用 ──")
	mods, err := p.List()
	if err != nil {
		return err
	}
	if len(mods) == 0 {
		fmt.Printf("%s 下没有应用模块。\n", p.Meta.AppsDir)
		return nil
	}
	opts := make([]string, 0, len(mods)+1)
	for _, m := range mods {
		opts = append(opts, m.ID+moduleNote(m))
	}
	opts = append(opts, "取消")
	n := prompt.Choose("删除哪个：", opts)
	if n > len(mods) {
		return nil
	}
	id := mods[n-1].ID
	if !prompt.YesNo(fmt.Sprintf("删除 %s，包括目录里的全部代码？", p.ModuleRel(id)), false) {
		fmt.Println("已取消。")
		return nil
	}
	err = runRemove(p, id, false)
	if err != nil && strings.Contains(err.Error(), "仍依赖") {
		color.Yellow("%v", err)
		if prompt.YesNo("仍然删除", false) {
			return runRemove(p, id, true)
		}
		return nil
	}
	return err
}

func runRemove(p *app.Project, id string, force bool) error {
	if err := p.Remove(id, app.RemoveOptions{Force: force, Log: logLine}); err != nil {
		return err
	}
	color.Green("已删除 %s", id)
	return nil
}

func moduleNote(m app.Module) string {
	switch {
	case !m.Registered:
		return "（目录存在，构建文件里没有登记）"
	case !m.DirExists:
		return "（已登记，目录不存在）"
	}
	return ""
}

func printModules(p *app.Project) error {
	mods, err := p.List()
	if err != nil {
		return err
	}
	if len(mods) == 0 {
		fmt.Printf("%s 下还没有应用模块。\n", p.Meta.AppsDir)
		return nil
	}
	for _, m := range mods {
		fmt.Printf("  %s%s\n", p.ModuleRel(m.ID), moduleNote(m))
	}
	return nil
}
