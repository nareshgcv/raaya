package analysis

import (
	"context"
	
	"raaya/pkg/analysis/policies"
	"raaya/pkg/graph"
)

type RegoEngine struct{}

func NewRegoEngine() *RegoEngine {
	return &RegoEngine{}
}

func (r *RegoEngine) Evaluate(ctx context.Context, g *graph.Graph) ([]Finding, error) {
	_ = policies.DefaultPolicy // Access embedded default.rego bytes
	
	var findings []Finding
	// Check for unauthenticated tool exposure across edges
	for _, edge := range g.Edges {
		if edge.Permission == graph.PermAdmin && edge.Transitive {
			findings = append(findings, Finding{
				ID:         "RAAYA-001",
				RuleID:     "transitive-admin-access",
				Severity:   "HIGH",
				Message:    "Transitive path exposes ADMIN privilege without explicit check.",
				TargetNode: edge.TargetID,
			})
		}
	}

	return findings, nil
}
