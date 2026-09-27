// cmd/raaya/subcmds/doctor.go
package subcmds

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func NewDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Run instant 5-second diagnostics on local environment & AI workspace setup",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("🏥 Running Raaya Doctor Diagnostics...\n")

			// Check .mcp config presence
			if _, err := os.Stat(".mcp/config.json"); err == nil {
				fmt.Println("  ✓ Found .mcp/config.json")
			} else {
				fmt.Println("  ⚠️ Missing .mcp/config.json")
			}

			// Check system prompts
			matches, _ := filepath.Glob("prompts/*.txt")
			if len(matches) > 0 {
				fmt.Printf("  ✓ Found %d system prompt definition(s)\n", len(matches))
			} else {
				fmt.Println("  ⚠️ No prompt definitions found in ./prompts")
			}

			fmt.Println("\nSystem ready for `raaya check`.")
		},
	}
}
