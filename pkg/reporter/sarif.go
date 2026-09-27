package reporter

import (
	"encoding/json"
	"io"

	"raaya/pkg/analysis/static"
)

type SARIFReport struct {
	Version string     `json:"$schema"`
	Schema  string     `json:"version"`
	Runs    []SARIFRun `json:"runs"`
}

type SARIFRun struct {
	Tool    SARIFTool    `json:"tool"`
	Results []SARIFResult `json:"results"`
}

type SARIFTool struct {
	Driver SARIFDriver `json:"driver"`
}

type SARIFDriver struct {
	Name           string      `json:"name"`
	InformationURI string      `json:"informationUri"`
	Rules          []SARIFRule `json:"rules"`
}

type SARIFRule struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	ShortDescription struct {
		Text string `json:"text"`
	} `json:"shortDescription"`
}

type SARIFResult struct {
	RuleID    string         `json:"ruleId"`
	Level     string         `json:"level"`
	Message   SARIFMessage   `json:"message"`
	Locations []SARIFLocation `json:"locations"`
}

type SARIFMessage struct {
	Text string `json:"text"`
}

type SARIFLocation struct {
	PhysicalLocation SARIFPhysicalLocation `json:"physicalLocation"`
}

type SARIFPhysicalLocation struct {
	ArtifactLocation SARIFArtifactLocation `json:"artifactLocation"`
	Region           SARIFRegion           `json:"region"`
}

type SARIFArtifactLocation struct {
	URI string `json:"uri"`
}

type SARIFRegion struct {
	StartLine   int `json:"startLine"`
	StartColumn int `json:"startColumn"`
}

func RenderSARIF(w io.Writer, findings []static.Finding) error {
	rulesMap := make(map[string]SARIFRule)
	var results []SARIFResult

	for _, f := range findings {
		if _, exists := rulesMap[f.RuleID]; !exists {
			rulesMap[f.RuleID] = SARIFRule{
				ID:   f.RuleID,
				Name: f.RuleName,
				ShortDescription: struct {
					Text string `json:"text"`
				}{Text: f.Message},
			}
		}

		level := "warning"
		if f.Severity == static.SeverityError {
			level = "error"
		}

		results = append(results, SARIFResult{
			RuleID: f.RuleID,
			Level:  level,
			Message: SARIFMessage{
				Text: f.Message + " | " + f.FixHint,
			},
			Locations: []SARIFLocation{
				{
					PhysicalLocation: SARIFPhysicalLocation{
						ArtifactLocation: SARIFArtifactLocation{
							URI: f.FilePath,
						},
						Region: SARIFRegion{
							StartLine:   f.Line,
							StartColumn: f.Col,
						},
					},
				},
			},
		})
	}

	var rules []SARIFRule
	for _, rule := range rulesMap {
		rules = append(rules, rule)
	}

	report := SARIFReport{
		Version: "https://json.schemastore.org/sarif-1.0.0.json",
		Schema:  "1.0.0",
		Runs: []SARIFRun{
			{
				Tool: SARIFTool{
					Driver: SARIFDriver{
						Name:           "Raaya Policy Engine",
						InformationURI: "https://github.com/raayadev/raaya",
						Rules:          rules,
					},
				},
				Results: results,
			},
		},
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}
