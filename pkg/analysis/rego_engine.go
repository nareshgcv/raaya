package analysis

import (
	"context"
	"fmt"

	"github.com/open-policy-agent/opa/rego"
	"raaya/pkg/analysis/policies"
	"raaya/pkg/analysis/rules"
	"raaya/pkg/graph"
)

type RegoEngine struct {
	query rego.PreparedEvalQuery
}

func NewRegoEngine() (*RegoEngine, error) {
	policyText, err := policies.GetDefaultPolicy()
	if err != nil {
		return nil, fmt.Errorf("failed to load embedded rego policy: %w", err)
	}

	ctx := context.Background()
	r := rego.New(
		rego.Query("data.raaya.security.deny"),
		rego.Module("default.rego", policyText),
	)

	query, err := r.PrepareForEval(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare rego query: %w", err)
	}

	return &RegoEngine{query: query}, nil
}

func (e *RegoEngine) Evaluate(sg *graph.SecurityGraph) ([]rules.Finding, error) {
	ctx := context.Background()

	// Pass SecurityGraph directly as Rego input document
	results, err := e.query.Eval(ctx, rego.EvalWithInput(sg))
	if err != nil {
		return nil, fmt.Errorf("rego evaluation failed: %w", err)
	}

	var findings []rules.Finding

	if len(results) > 0 && len(results[0].Expressions) > 0 {
		if denyMsgs, ok := results[0].Expressions[0].Value.([]interface{}); ok {
			for _, item := range denyMsgs {
				if msg, isStr := item.(string); isStr {
					findings = append(findings, rules.Finding{
						RuleID:   "RAAYA-REGO-001",
						Severity: rules.SeverityHigh,
						Message:  msg,
					})
				}
			}
		}
	}

	return findings, nil
}
