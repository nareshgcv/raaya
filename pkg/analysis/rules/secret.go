package rules

import (
	"fmt"

	"raaya/pkg/graph"
)

type Severity string

const (
	SeverityCritical Severity = "CRITICAL"
	SeverityHigh     Severity = "HIGH"
	SeverityMedium   Severity = "MEDIUM"
)

type Finding struct {
	RuleID           string   `json:"rule_id"`
	Severity         Severity `json:"severity"`
	Message          string   `json:"message"`
	AssetID          string   `json:"asset_id"`
	FilePath         string   `json:"file_path"`
	AutofixAvailable bool     `json:"autofix_available"`
}

type HardcodedSecretRule struct{}

func (r *HardcodedSecretRule) ID() string { return "RAAYA-001-PLAINTEXT-SECRET" }

func (r *HardcodedSecretRule) Evaluate(sg *graph.SecurityGraph) []Finding {
	var findings []Finding
	for _, node := range sg.Nodes {
		if node.Kind == graph.KindSecret {
			findings = append(findings, Finding{
				RuleID:           r.ID(),
				Severity:         SeverityCritical,
				Message:          fmt.Sprintf("Plaintext credential detected in key '%s'", node.Name),
				AssetID:          node.ID,
				FilePath:         node.SourceFile,
				AutofixAvailable: true,
			})
		}
	}
	return findings
}
