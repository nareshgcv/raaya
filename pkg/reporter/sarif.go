package reporter

import (
	"encoding/json"

	"github.com/raaya/pkg/analysis"
)

type SARIFLog struct {
	Version string `json:"version"`
	Schema  string `json:"$schema"`
	Runs    []Run  `json:"runs"`
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
	RuleID  string  `json:"ruleId"`
	Message Message `json:"message"`
}

type Message struct {
	Text string `json:"text"`
}

func GenerateSARIF(evalResult *analysis.EvaluationResult) ([]byte, error) {
	results := make([]Result, 0)
	for _, v := range evalResult.Violations {
		results = append(results, Result{
			RuleID:  "APE001",
			Message: Message{Text: v},
		})
	}

	log := SARIFLog{
		Version: "2.1.0",
		Schema:  "https://schemastore.azurewebsites.net/schemas/json/sarif-2.1.0-rtm.5.json",
		Runs: []Run{
			{
				Tool: Tool{
					Driver: Driver{
						Name:    "ape",
						Version: "0.1.0",
					},
				},
				Results: results,
			},
		},
	}

	return json.MarshalIndent(log, "", "  ")
}
