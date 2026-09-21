package analysis

import (
	"context"
	"fmt"

	"github.com/ape-sec/ape/pkg/analysis/policies"
	"github.com/ape-sec/ape/pkg/graph"
	"github.com/open-policy-agent/opa/rego"
)

type EvaluationResult struct {
	Allowed    bool     `json:"allowed"`
	Violations []string `json:"violations"`
}

func EvaluateGraph(ctx context.Context, sg *graph.SecurityGraph) (*EvaluationResult, error) {
	regoCode, err := policies.DefaultPolicies.ReadFile("default.rego")
	if err != nil {
		return nil, fmt.Errorf("failed to load embedded rego policy: %w", err)
	}

	query, err := rego.New(
		rego.Query("data.ape.security"),
		rego.Module("default.rego", string(regoCode)),
	).PrepareForEval(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to prepare rego evaluation: %w", err)
	}

	results, err := query.Eval(ctx, rego.EvalWithInput(sg))
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate policy: %w", err)
	}

	res := &EvaluationResult{
		Allowed:    true,
		Violations: make([]string, 0),
	}

	if len(results) > 0 && len(results[0].Expressions) > 0 {
		value, ok := results[0].Expressions[0].Value.(map[string]interface{})
		if ok {
			if violations, exists := value["violations"].([]interface{}); exists {
				for _, v := range violations {
					res.Violations = append(res.Violations, fmt.Sprintf("%v", v))
				}
				if len(res.Violations) > 0 {
					res.Allowed = false
				}
			}
		}
	}

	return res, nil
}
