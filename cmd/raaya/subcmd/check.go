package subcmds

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"raaya/pkg/analysis/fixer"
	"raaya/pkg/analysis/static"
	"raaya/pkg/discovery"
	"raaya/pkg/reporter"
)

func NewCheckCmd() *cobra.Command {
	var fix bool
	var format string

	cmd := &cobra.Command{
		Use:   "check [path]",
		Short: "Run fast static security and integrity checks against local agent & MCP assets",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) > 0 {
				dir = args[0]
			}

			// 1. Discover assets
			mcpConfigs, prompts, err := discovery.ScanWorkspace(dir)
			if err != nil {
				return fmt.Errorf("failed scanning workspace: %w", err)
			}

			// 2. Run Native Rules (RAA001 - RAA005)
			engine := static.NewEngine()
			findings := engine.Evaluate(mcpConfigs, prompts)

			// 3. Handle Auto-Fix if requested
			if fix {
				fixedCount, err := fixer.ApplyFixes(dir, findings)
				if err != nil {
					return fmt.Errorf("auto-fix failed: %w", err)
				}
				fmt.Printf("✓ Successfully auto-fixed %d issue(s).\n\n", fixedCount)
				// Re-run evaluation post-fix
				mcpConfigs, prompts, _ = discovery.ScanWorkspace(dir)
				findings = engine.Evaluate(mcpConfigs, prompts)
			}

			// 4. Report findings
			if format == "sarif" {
				return reporter.RenderSARIF(os.Stdout, findings)
			}

			hasErrors := reporter.RenderTerminal(os.Stdout, findings)
			if hasErrors {
				os.Exit(1)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&fix, "fix", false, "Automatically fix deterministic issues (e.g., extract secrets to .env)")
	cmd.Flags().StringVarP(&format, "format", "f", "terminal", "Output format: terminal | sarif | json")
	return cmd
}
