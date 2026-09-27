package rules

import (
	"raaya/pkg/graph"
)

// Severity represents the impact level of a security finding
type Severity string

const (
	SeverityCritical Severity = "CRITICAL"
	SeverityHigh     Severity = "HIGH"
	SeverityMedium   Severity = "MEDIUM"
	SeverityLow      Severity = "LOW"
	SeverityInfo     Severity = "INFO"
)

// Finding represents an identified security issue within the graph
type Finding struct {
	RuleID           string   `json:"rule_id"`
	Severity         Severity `json:"severity"`
	Message          string   `json:"message"`
	AssetID          string   `json:"asset_id"`
	FilePath         string   `json:"file_path"`
	LineNumber       int      `json:"line_number,omitempty"`
	AutofixAvailable bool     `json:"autofix_available"`
}

// Rule defines the interface that all native security analysis checks must implement
type Rule interface {
	// ID returns the unique identifier for the rule (e.g., "RAAYA-001-SECRETS")
	ID() string

	// Evaluate analyzes the SecurityGraph and returns any detected findings
	Evaluate(sg *graph.SecurityGraph) []Finding
}

// Registry manages the collection of all active rules
type Registry struct {
	rules []Rule
}

// NewRegistry initializes a registry populated with default rules
func NewRegistry() *Registry {
	r := &Registry{
		rules: make([]Rule, 0),
	}
	
	// Register default rules
	r.Register(&SecretsRule{})
	r.Register(&UnauthenticatedEndpointRule{})

	return r
}

// Register adds a new rule to the evaluation registry
func (r *Registry) Register(rule Rule) {
	r.rules = append(r.rules, rule)
}

// Rules returns all registered rules
func (r *Registry) Rules() []Rule {
	return r.rules
}

// EvaluateAll runs all registered rules against the provided SecurityGraph
func (r *Registry) EvaluateAll(sg *graph.SecurityGraph) []Finding {
	var allFindings []Finding

	for _, rule := range r.rules {
		findings := rule.Evaluate(sg)
		allFindings = append(allFindings, findings...)
	}

	return allFindings
}
