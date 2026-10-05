package reporter

import (
	"io"
	"strings"

	"github.com/nareshgcv/raaya/pkg/analysis"
	"github.com/nareshgcv/raaya/pkg/analysis/rules"
)

type sarifLog struct {
	Schema  string     `json:"$schema"`
	Version string     `json:"version"`
	Runs    []sarifRun `json:"runs"`
}

type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results"`
}

type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

type sarifDriver struct {
	Name           string      `json:"name"`
	Version        string      `json:"version"`
	InformationURI string      `json:"informationUri"`
	Rules          []sarifRule `json:"rules"`
}

type sarifRule struct {
	ID                   string        `json:"id"`
	Name                 string        `json:"name,omitempty"`
	ShortDescription     sarifText     `json:"shortDescription"`
	DefaultConfiguration sarifLevelCfg `json:"defaultConfiguration"`
}

type sarifText struct {
	Text string `json:"text"`
}

type sarifLevelCfg struct {
	Level string `json:"level"`
}

type sarifResult struct {
	RuleID              string            `json:"ruleId"`
	Level               string            `json:"level"`
	Message             sarifText         `json:"message"`
	Locations           []sarifLocation   `json:"locations,omitempty"`
	PartialFingerprints map[string]string `json:"partialFingerprints"`
}

type sarifLocation struct {
	PhysicalLocation sarifPhysical `json:"physicalLocation"`
}

type sarifPhysical struct {
	ArtifactLocation sarifArtifact `json:"artifactLocation"`
	Region           sarifRegion   `json:"region"`
}

type sarifArtifact struct {
	URI string `json:"uri"`
}

type sarifRegion struct {
	StartLine int `json:"startLine"`
}

func sarifLevel(s analysis.Severity) string {
	switch s {
	case analysis.SevHigh:
		return "error"
	case analysis.SevMedium:
		return "warning"
	default:
		return "note"
	}
}

// SARIF writes findings as SARIF 1.0.0 for GitHub code scanning.
func SARIF(w io.Writer, findings []analysis.Finding, version string) error {
	ruleList := []sarifRule{}
	known := map[string]bool{}
	for _, r := range rules.All {
		known[r.ID] = true
		ruleList = append(ruleList, sarifRule{ID: r.ID, Name: r.Name, ShortDescription: sarifText{r.Description}, DefaultConfiguration: sarifLevelCfg{sarifLevel(r.Severity)}})
	}
	results := []sarifResult{}
	for _, f := range findings {
		if !known[f.RuleID] {
			known[f.RuleID] = true
			ruleList = append(ruleList, sarifRule{ID: f.RuleID, ShortDescription: sarifText{"Organisation policy"}, DefaultConfiguration: sarifLevelCfg{sarifLevel(f.Severity)}})
		}
		res := sarifResult{
			RuleID:              f.RuleID,
			Level:               sarifLevel(f.Severity),
			Message:             sarifText{f.Message},
			PartialFingerprints: map[string]string{"raayaFinding/v1": strings.ReplaceAll(f.Key(), "\x00", "|")},
		}
		if f.File != "" {
			res.Locations = []sarifLocation{{PhysicalLocation: sarifPhysical{
				ArtifactLocation: sarifArtifact{URI: f.File},
				Region:           sarifRegion{StartLine: max(f.Line, 1)},
			}}}
		}
		results = append(results, res)
	}
	return JSON(w, sarifLog{
		Schema:  "https://json.schemastore.org/sarif-1.0.0.json",
		Version: "1.0.0",
		Runs: []sarifRun{{
			Tool:    sarifTool{Driver: sarifDriver{Name: "raaya", Version: version, InformationURI: "https://github.com/nareshgcv/raaya", Rules: ruleList}},
			Results: results,
		}},
	})
}
