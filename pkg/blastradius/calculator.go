package blastradius

import (
	"fmt"
	"sort"

	"raaya/pkg/graph"
)

// RiskLevel categorizes the severity of a computed blast radius.
type RiskLevel string

const (
	RiskLow      RiskLevel = "LOW"
	RiskMedium   RiskLevel = "MEDIUM"
	RiskHigh     RiskLevel = "HIGH"
	RiskCritical RiskLevel = "CRITICAL"
)

// ImpactSummary provides a breakdown of affected entities for security reporting.
type ImpactSummary struct {
	TotalExposedNodes int                       `json:"total_exposed_nodes"`
	NodesByType       map[graph.NodeType]int   `json:"nodes_by_type"`
	HighestPermission  graph.PermissionLevel    `json:"highest_permission"`
	HighRiskPaths     [][]string                `json:"high_risk_paths,omitempty"`
}

// BlastRadiusResult holds the complete risk valuation for a targeted node.
type BlastRadiusResult struct {
	TargetNodeID string          `json:"target_node_id"`
	RiskLevel    RiskLevel       `json:"risk_level"`
	Score        float64         `json:"score"`
	Impact       *ReachableImpact `json:"impact"`
	Summary      ImpactSummary   `json:"summary"`
}

// Calculator handles impact calculations and security surface risk scoring.
type Calculator struct {
	engine *CapabilityEngine
	graph  *graph.Graph
}

// NewCalculator initializes a new blast radius calculator instance.
func NewCalculator(g *graph.Graph) *Calculator {
	return &Calculator{
		engine: NewCapabilityEngine(g),
		graph:  g,
	}
}

// CalculateNodeImpact evaluates the damage potential if the target node is compromised.
func (c *Calculator) CalculateNodeImpact(nodeID string) (*BlastRadiusResult, error) {
	targetNode, exists := c.graph.Nodes[nodeID]
	if !exists {
		return nil, fmt.Errorf("node with ID '%s' not found in security graph", nodeID)
	}

	// 1. Run BFS capability propagation
	impact := c.engine.CalculateBlastRadius(nodeID)

	// 2. Aggregate impact stats across direct & transitive exposures
	summary := c.summarizeImpact(impact)

	// 3. Compute weighted blast radius score & determine risk level
	score := c.computeScore(targetNode, impact, summary)
	riskLevel := c.determineRiskLevel(score, summary)

	return &BlastRadiusResult{
		TargetNodeID: nodeID,
		RiskLevel:    riskLevel,
		Score:        score,
		Impact:       impact,
		Summary:      summary,
	}, nil
}

// CalculateGlobalImpact runs blast radius calculations across all nodes in the graph.
func (c *Calculator) CalculateGlobalImpact() map[string]*BlastRadiusResult {
	results := make(map[string]*BlastRadiusResult)

	for id := range c.graph.Nodes {
		res, err := c.CalculateNodeImpact(id)
		if err == nil {
			results[id] = res
		}
	}

	return results
}

// summarizeImpact organizes reachable nodes into structured counts and permissions.
func (c *Calculator) summarizeImpact(impact *ReachableImpact) ImpactSummary {
	nodesByType := make(map[graph.NodeType]int)
	allReachable := append([]*graph.Node{}, impact.DirectNodes...)
	allReachable = append(allReachable, impact.TransitiveNodes...)

	highestPerm := graph.PermRead
	permWeights := map[graph.PermissionLevel]int{
		graph.PermRead:    1,
		graph.PermExecute: 2,
		graph.PermWrite:   3,
		graph.PermAdmin:   4,
	}

	for _, node := range allReachable {
		nodesByType[node.Type]++
		for _, perm := range node.Permissions {
			if permWeights[perm] > permWeights[highestPerm] {
				highestPerm = perm
			}
		}
	}

	return ImpactSummary{
		TotalExposedNodes: len(allReachable),
		NodesByType:       nodesByType,
		HighestPermission:  highestPerm,
	}
}

// computeScore calculates a numerical risk score (0.0 - 100.0) based on reachable capabilities.
func (c *Calculator) computeScore(target *graph.Node, impact *ReachableImpact, summary ImpactSummary) float64 {
	baseWeight := 1.0
	switch target.Type {
	case graph.NodeAgent:
		baseWeight = 2.5
	case graph.NodeMCPServer:
		baseWeight = 2.0
	case graph.NodeTool:
		baseWeight = 1.5
	case graph.NodeResource:
		baseWeight = 1.0
	}

	// Permission multiplier
	permMultiplier := 1.0
	switch summary.HighestPermission {
	case graph.PermAdmin:
		permMultiplier = 3.0
	case graph.PermWrite:
		permMultiplier = 2.0
	case graph.PermExecute:
		permMultiplier = 1.5
	case graph.PermRead:
		permMultiplier = 1.0
	}

	// Score = (Direct Reach * 3 + Transitive Reach * 1.5) * Node Weight * Permission Multiplier
	rawScore := (float64(len(impact.DirectNodes))*3.0 + float64(len(impact.TransitiveNodes))*1.5) * baseWeight * permMultiplier

	// Cap maximum score to 100.0
	if rawScore > 100.0 {
		return 100.0
	}

	return rawScore
}

// determineRiskLevel maps numerical scores and exposed asset criticalities to risk thresholds.
func (c *Calculator) determineRiskLevel(score float64, summary ImpactSummary) RiskLevel {
	if summary.HighestPermission == graph.PermAdmin || summary.NodesByType[graph.NodeResource] > 5 || score >= 75.0 {
		return RiskCritical
	}
	if summary.HighestPermission == graph.PermWrite || score >= 45.0 {
		return RiskHigh
	}
	if score >= 20.0 {
		return RiskMedium
	}
	return RiskLow
}

// GetTopKHighRiskNodes returns the top K nodes with the highest blast radius scores.
func GetTopKHighRiskNodes(results map[string]*BlastRadiusResult, k int) []*BlastRadiusResult {
	list := make([]*BlastRadiusResult, 0, len(results))
	for _, res := range results {
		list = append(list, res)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].Score > list[j].Score
	})

	if len(list) > k {
		return list[:k]
	}
	return list
}
