package main

import (
	"os"

	"github.com/star/star-yi-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
