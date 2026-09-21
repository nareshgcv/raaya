package analysis

import (
	"context"
	"fmt"

	"github.com/raaya/pkg/graph"
)

// Severity indicates the risk level of a rule violation.
type Severity string

const (
	SeverityCritical Severity = "CRITICAL"
	SeverityHigh     Severity = "HIGH"
	SeverityMedium   Severity = "MEDIUM"
	SeverityLow      Severity = "LOW"
	SeverityInfo     Severity = "INFO"
)

// Rule defines the interface that all security analysis rules must satisfy.
type Rule interface {
	// ID returns a unique identifier for the rule (e.g., "SEC-001").
	ID() string
	// Name returns a human-readable title for the rule.
	Name() string
	// Description details what security risk or misconfiguration this rule checks for.
	Description() string
	// Severity returns the risk level if this rule is violated.
	Severity() Severity
	// Evaluate runs the rule against the provided SecurityGraph and returns violations found.
	Evaluate(ctx context.Context, sg *graph.SecurityGraph) ([]RuleViolation, error)
}

// RuleViolation represents an instance where a security rule was breached.
type RuleViolation struct {
	RuleID      string                 `json:"rule_id"`
	RuleName    string                 `json:"rule_name"`
	Severity    Severity               `json:"severity"`
	ResourceID  string                 `json:"resource_id,omitempty"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// RuleRegistry manages and executes the collection of active security rules.
type RuleRegistry struct {
	rules map[string]Rule
}

// NewRuleRegistry initializes a new RuleRegistry instance.
func NewRuleRegistry() *RuleRegistry {
	return &RuleRegistry{
		rules: make(map[string]Rule),
	}
}

// Register adds a new rule to the registry.
func (r *RuleRegistry) Register(rule Rule) error {
	if rule == nil {
		return fmt.Errorf("cannot register nil rule")
	}
	if _, exists := r.rules[rule.ID()]; exists {
		return fmt.Errorf("rule with ID %q already registered", rule.ID())
	}
	r.rules[rule.ID()] = rule
	return nil
}

// Get retrieves a registered rule by its ID.
func (r *RuleRegistry) Get(id string) (Rule, bool) {
	rule, ok := r.rules[id]
	return rule, ok
}

// List returns all currently registered rules.
func (r *RuleRegistry) List() []Rule {
	list := make([]Rule, 0, len(r.rules))
	for _, rule := range r.rules {
		list = append(list, rule)
	}
	return list
}

// EvaluateAll executes all registered rules against the SecurityGraph and collects violations.
func (r *RuleRegistry) EvaluateAll(ctx context.Context, sg *graph.SecurityGraph) ([]RuleViolation, error) {
	var allViolations []RuleViolation

	for _, rule := range r.rules {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		violations, err := rule.Evaluate(ctx, sg)
		if err != nil {
			return nil, fmt.Errorf("failed evaluating rule %s: %w", rule.ID(), err)
		}
		allViolations = append(allViolations, violations...)
	}

	return allViolations, nil
}
