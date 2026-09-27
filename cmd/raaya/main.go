package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"raaya/cmd/raaya/subcmds"
)

var rootCmd = &cobra.Command{
	Use:   "raaya",
	Short: "Raaya — Local-first security engine for AI agents & MCP",
}

func main() {
	rootCmd.AddCommand(subcmds.CheckCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
