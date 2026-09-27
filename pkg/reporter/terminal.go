package reporter

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"raaya/pkg/analysis/rules"
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).ForegroundColor(lipgloss.Color("86")).MarginBottom(1)
	criticalStyle = lipgloss.NewStyle().Bold(true).ForegroundColor(lipgloss.Color("196"))
	fileStyle     = lipgloss.NewStyle().Underline(true).ForegroundColor(lipgloss.Color("244"))
)

func PrintTerminalReport(findings []rules.Finding) {
	fmt.Println(titleStyle.Render("🔍 Raaya Security Diagnostics"))

	if len(findings) == 0 {
		fmt.Println("✨ No security issues found!")
		return
	}

	for _, f := range findings {
		sev := criticalStyle.Render(string(f.Severity))
		fmt.Printf("[%s] %s\n", sev, f.Message)
		fmt.Printf("  ├─ Rule: %s\n", f.RuleID)
		fmt.Printf("  └─ File: %s\n\n", fileStyle.Render(f.FilePath))
	}
}
