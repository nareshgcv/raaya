package reporter

import (
	"encoding/json"
	"os"

	"raaya/pkg/analysis/rules"
)

type SarifLog struct {
	Version $string$ `json:"version"`
	Schema  $string$ `json:"$schema"`
	Runs    []Run    `json:"runs"`
}

type Run struct {
	Tool    Tool     `json:"tool"`
	Results []Result `json:"results"`
}

type Tool struct {
	Driver Driver `json:"driver"`
}

type Driver struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Result struct {
	RuleID  string    `json:"ruleId"`
	Message Message   `json:"message"`
	Locs    []Location `json:"locations"`
}

type Message struct {
	Text string `json:"text"`
}

type Location struct {
	PhysicalLocation PhysicalLocation `json:"physicalLocation"`
}

type PhysicalLocation struct {
	ArtifactLocation ArtifactLocation `json:"artifactLocation"`
}

type ArtifactLocation struct {
	URI string `json:"uri"`
}

func WriteSarifReport(filePath string, findings []rules.Finding) error {
	results := make([]Result, 0, len(findings))

	for _, f := range findings {
		results = append(results, Result{
			RuleID:  f.RuleID,
			Message: Message{Text: f.Message},
			Locs: []Location{
				{
					PhysicalLocation: PhysicalLocation{
						ArtifactLocation: ArtifactLocation{URI: f.FilePath},
					},
				},
			},
		})
	}

	log := SarifLog{
		Version: "1.0.0",
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-1.0.0.json",
		Runs: []Run{
			{
				Tool: Tool{
					Driver: Driver{Name: "Raaya", Version: "0.1.0"},
				},
				Results: results,
			},
		},
	}

	data, err := json.MarshalIndent(log, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}
