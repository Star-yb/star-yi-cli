package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"github.com/star/star-yi-cli/internal/ui"
)

type uiOptions struct {
	Port        int
	OpenBrowser bool
}

func init() {
	var o uiOptions
	var noBrowser bool
	cmd := &cobra.Command{
		Use:   "ui",
		Short: "打开创建器页面（直接运行 star-yi-cli 也会打开）",
		RunE: func(cmd *cobra.Command, args []string) error {
			o.OpenBrowser = !noBrowser
			return runUI(o)
		},
	}
	cmd.Flags().IntVar(&o.Port, "port", 0, "监听端口，0 表示随机")
	cmd.Flags().BoolVar(&noBrowser, "no-browser", false, "不自动打开浏览器")
	rootCmd.AddCommand(cmd)
}

func runUI(o uiOptions) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	return ui.Run(ctx, ui.Options{
		Port:        o.Port,
		OpenBrowser: o.OpenBrowser,
		Version:     Version,
		Log:         func(s string) { fmt.Println(s) },
	})
}
