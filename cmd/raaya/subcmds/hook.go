// cmd/raaya/subcmds/hook.go
package subcmds

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const preCommitScript = `#!/bin/sh
# Installed by Raaya
raaya check
`

func NewHookCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "hook",
		Short: "Manage Git hooks",
		Subcommands: []*cobra.Command{
			{
				Use:   "install",
				Short: "Install local git pre-commit hook automatically",
				RunE: func(cmd *cobra.Command, args []string) error {
					hookPath := ".git/hooks/pre-commit"
					if err := os.WriteFile(hookPath, []byte(preCommitScript), 0755); err != nil {
						return fmt.Errorf("failed installing git hook: %w", err)
					}
					fmt.Println("✓ Successfully installed local git pre-commit hook at .git/hooks/pre-commit")
					return nil
				},
			},
		},
	}
}
