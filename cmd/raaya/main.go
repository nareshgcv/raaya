package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"raaya/cmd/raaya/subcmds"
)

var rootCmd = &cobra.Command{
	Use:   "raaya",
	Short: "Raaya - Agentic & MCP Security Surface Analyzer",
}

func main() {
	rootCmd.AddCommand(subcmds.CheckCmd)
	rootCmd.AddCommand(subcmds.GraphCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
