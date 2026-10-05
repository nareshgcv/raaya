//go:build !rego

package analysis

import "context"

// RegoEnabled reports whether this binary can evaluate Rego policies.
const RegoEnabled = false

// EvaluatePolicies is a no-op in builds without Rego. Asking for policy
// files is an error, so a CI job can't silently skip them.
func EvaluatePolicies(_ context.Context, files []string, _ PolicyInput) ([]Finding, error) {
	if len(files) > 0 {
		return nil, ErrRegoUnavailable
	}
	return nil, nil
}
