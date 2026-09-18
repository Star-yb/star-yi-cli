package cmd

import (
	"github.com/spf13/cobra"
)

var interactiveCmd = &cobra.Command{
	Use:   "interactive",
	Short: "交互式问答模式（也可直接运行 star-yi-cli 无参数进入）",
	Aliases: []string{"i"},
	RunE: func(cmd *cobra.Command, args []string) error {
		return RunInteractive()
	},
}

func init() {
	rootCmd.AddCommand(interactiveCmd)
}
