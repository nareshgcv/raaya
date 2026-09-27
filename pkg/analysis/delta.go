package analysis

import (
	"raaya/pkg/analysis/rules"
)

// ComputeDelta filters current findings against a baseline set of existing findings
func ComputeDelta(currentFindings []rules.Finding, baselineFindings []rules.Finding) []rules.Finding {
	baselineMap := make(map[string]bool)
	for _, f := range baselineFindings {
		// Unique fingerprinted key for each issue
		key := f.RuleID + ":" + f.FilePath + ":" + f.AssetID
		baselineMap[key] = true
	}

	var newFindings []rules.Finding
	for _, f := range currentFindings {
		key := f.RuleID + ":" + f.FilePath + ":" + f.AssetID
		if !baselineMap[key] {
			newFindings = append(newFindings, f)
		}
	}

	return newFindings
}
