package blastradius

import (
	"fmt"
	"sort"

	"raaya/pkg/graph"
)

// RiskLevel categorizes the severity of a computed blast radius score.
type RiskLevel string

const (
	RiskLow      RiskLevel = "LOW"
	RiskMedium   RiskLevel = "MEDIUM"
	RiskHigh     RiskLevel = "HIGH"
	RiskCritical RiskLevel = "CRITICAL"
)

// ImpactSummary provides a categorized breakdown of reachable entities and peak permission level.
type ImpactSummary struct {
	TotalExposedNodes int                    `json:"total_exposed_nodes"`
	NodesByType       map[graph.NodeType]int `json:"nodes_by_type"`
	HighestPermission graph.PermissionLevel  `json:"highest_permission"`
	HighRiskPaths     [][]string             `json:"high_risk_paths,omitempty"`
}

// BlastRadiusResult contains the calculated risk score, summary, and underlying propagation details.
type BlastRadiusResult struct {
	TargetNodeID string           `json:"target_node_id"`
	RiskLevel    RiskLevel        `json:"risk_level"`
	Score        float64          `json:"score"`
	Impact       *ReachableImpact `json:"impact"`
	Summary      ImpactSummary    `json:"summary"`
}

// Calculator evaluates security surface exposure using the capability propagation engine.
type Calculator struct {
	engine *CapabilityEngine
	graph  *graph.Graph
}

// NewCalculator creates a new Calculator instance wrapping the provided security graph.
func NewCalculator(g *graph.Graph) *Calculator {
	return &Calculator{
		engine: NewCapabilityEngine(g),
		graph:  g,
	}
}

// CalculateNodeImpact evaluates the blast radius if a specific target node is compromised.
func (c *Calculator) CalculateNodeImpact(nodeID string) (*BlastRadiusResult, error) {
	targetNode, exists := c.graph.Nodes[nodeID]
	if !exists {
		return nil, fmt.Errorf("node with ID '%s' not found in security graph", nodeID)
	}

	// 1. Delegate BFS graph traversal to propagation.go
	impact := c.engine.CalculateBlastRadius(nodeID)

	// 2. Aggregate counts and permissions
	summary := c.summarizeImpact(impact)

	// 3. Compute weighted blast radius score and risk classification
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

// CalculateGlobalImpact calculates blast radius metrics across every node in the graph.
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

// summarizeImpact organizes reachable nodes into type frequencies and detects the highest permission path.
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
		HighestPermission: highestPerm,
	}
}

// computeScore computes a normalized risk score (0.0 - 100.0) based on reachability and permissions.
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

	// Score = (Direct Reach * 3 + Transitive Reach * 1.5) * Node Base Weight * Permission Multiplier
	rawScore := (float64(len(impact.DirectNodes))*3.0 + float64(len(impact.TransitiveNodes))*1.5) * baseWeight * permMultiplier

	if rawScore > 100.0 {
		return 100.0
	}

	return rawScore
}

// determineRiskLevel assigns a categorical risk rating based on score and critical resource exposures.
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

// GetTopKHighRiskNodes isolates the top K highest-risk entities in the scan.
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
