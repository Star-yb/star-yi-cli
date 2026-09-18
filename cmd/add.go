package cmd

import "github.com/spf13/cobra"

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "在已有 Star-Yi 项目中添加资源",
}

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "在已有 Star-Yi 项目中删除资源",
}

func init() {
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(removeCmd)
}
