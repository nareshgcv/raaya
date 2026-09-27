package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"raaya/cmd/raaya/subcmds"
)

var rootCmd = &cobra.Command{
	Use:   "raaya",
	Short: "Raaya Policy Engine — Local-first security & dependency graph analysis for AI agents",
}

func main() {
	rootCmd.AddCommand(
		subcmds.NewCheckCmd(),
		subcmds.NewGraphCmd(),
		subcmds.NewDoctorCmd(),
		subcmds.NewHookCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
