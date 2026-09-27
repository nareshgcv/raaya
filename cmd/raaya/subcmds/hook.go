package subcmds

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var HookCmd = &cobra.Command{
	Use:   "hook",
	Short: "Install Git pre-commit hook for automated scanning",
	Run: func(cmd *cobra.Command, args []string) {
		hookScript := "#!/bin/sh\nraaya check\n"
		path := ".git/hooks/pre-commit"

		err := os.WriteFile(path, []byte(hookScript), 0755)
		if err != nil {
			fmt.Printf("❌ Failed to install Git hook: %v\n", err)
			return
		}
		fmt.Println("✅ Pre-commit hook installed successfully at .git/hooks/pre-commit")
	},
}
