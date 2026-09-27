package reporter

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"raaya/pkg/analysis/static"
)

var (
	errStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9"))
	warnStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11"))
	codeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	pathStyle = lipgloss.NewStyle().Underline(true)
	hintStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
)

func RenderTerminal(w io.Writer, findings []static.Finding) bool {
	if len(findings) == 0 {
		fmt.Fprintln(w, lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("✓ No security or integrity issues found."))
		return false
	}

	hasErrors := false
	for _, f := range findings {
		badge := errStyle.Render("[ERROR]")
		if f.Severity == static.SeverityWarning {
			badge = warnStyle.Render("[WARN]")
		} else {
			hasErrors = true
		}

		fmt.Fprintf(w, "%s:%d:%d %s %s (%s)\n",
			pathStyle.Render(f.FilePath),
			f.Line,
			f.Col,
			badge,
			f.Message,
			codeStyle.Render(f.RuleID),
		)

		if f.Snippet != "" {
			fmt.Fprintf(w, "  %s | %s\n", codeStyle.Render(fmt.Sprintf("%4d", f.Line)), f.Snippet)
			fmt.Fprintf(w, "       | %s\n", errStyle.Render(strings.Repeat("^", len(f.Snippet))))
		}

		if f.FixHint != "" {
			fmt.Fprintf(w, "  └─ %s: %s\n", hintStyle.Render("Suggestion"), f.FixHint)
		}
		fmt.Fprintln(w)
	}

	return hasErrors
}
