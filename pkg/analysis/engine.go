package analysis

import (
	"raaya/pkg/analysis/rules"
	"raaya/pkg/graph"
)

type Engine struct {
	ruleSet []Rule
}

type Rule interface {
	ID() string
	Evaluate(sg *graph.SecurityGraph) []rules.Finding
}

func NewEngine() *Engine {
	return &Engine{
		ruleSet: []Rule{
			&rules.HardcodedSecretRule{},
		},
	}
}

func (e *Engine) Run(sg *graph.SecurityGraph) []rules.Finding {
	var allFindings []rules.Finding
	for _, rule := range e.ruleSet {
		results := rule.Evaluate(sg)
		allFindings = append(allFindings, results...)
	}
	return allFindings
}
