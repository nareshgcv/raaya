package reporter

import (
	"encoding/json"
	"io"

	"raaya/pkg/analysis/static"
)

type JSONOutput struct {
	TotalFindings int              `json:"total_findings"`
	Errors        int              `json:"errors"`
	Warnings      int              `json:"warnings"`
	Findings      []static.Finding `json:"findings"`
}

func RenderJSON(w io.Writer, findings []static.Finding) error {
	errCount := 0
	warnCount := 0

	for _, f := range findings {
		if f.Severity == static.SeverityError {
			errCount++
		} else {
			warnCount++
		}
	}

	output := JSONOutput{
		TotalFindings: len(findings),
		Errors:        errCount,
		Warnings:      warnCount,
		Findings:      findings,
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}
