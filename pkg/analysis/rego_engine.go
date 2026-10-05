//go:build rego

package analysis

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/nareshgcv/raaya/pkg/analysis/policies"
	"github.com/open-policy-agent/opa/v1/rego"
)

// RegoEnabled reports whether this binary can evaluate Rego policies.
const RegoEnabled = true

// EvaluatePolicies evaluates the bundled default policy plus files against
// input and returns everything in data.raaya.deny as findings.
func EvaluatePolicies(ctx context.Context, files []string, input PolicyInput) ([]Finding, error) {
	// Round-trip through JSON so policies see exactly the documented shape.
	raw, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	var doc any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, err
	}

	opts := []func(*rego.Rego){
		rego.Query("data.raaya.deny"),
		rego.Module("raaya/default.rego", policies.Default),
		rego.Input(doc),
	}
	for _, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		opts = append(opts, rego.Module(f, string(src)))
	}

	rs, err := rego.New(opts...).Eval(ctx)
	if err != nil {
		return nil, fmt.Errorf("rego: %w", err)
	}
	var out []Finding
	for _, result := range rs {
		for _, expr := range result.Expressions {
			items, ok := expr.Value.([]any)
			if !ok {
				continue
			}
			for _, item := range items {
				out = append(out, policyFinding(item))
			}
		}
	}
	return out, nil
}
