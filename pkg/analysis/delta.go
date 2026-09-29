package analysis

import (
	"raaya/pkg/blastradius"
	"raaya/pkg/graph"
)

type SurfaceStatus string

const (
	SurfaceStable   SurfaceStatus = "STABLE"
	SurfaceExpanded SurfaceStatus = "EXPANDED"
	SurfaceReduced  SurfaceStatus = "REDUCED"
)

type GraphDiff struct {
	AddedNodes      []graph.Node                   `json:"added_nodes"`
	RemovedNodes    []graph.Node                   `json:"removed_nodes"`
	AddedPaths      []blastradius.ReachabilityPath `json:"added_paths"`
	EscalatedCaps  []graph.Capability             `json:"escalated_capabilities"`
	SecuritySurface SurfaceStatus                  `json:"security_surface_status"`
}

func ComputeDiff(base, target *graph.SecurityGraph) *GraphDiff {
	diff := &GraphDiff{
		AddedNodes:      make([]graph.Node, 0),
		RemovedNodes:    make([]graph.Node, 0),
		AddedPaths:      make([]blastradius.ReachabilityPath, 0),
		EscalatedCaps:  make([]graph.Capability, 0),
		SecuritySurface: SurfaceStable,
	}

	for id, node := range target.Nodes {
		if _, exists := base.Nodes[id]; !exists {
			diff.AddedNodes = append(diff.AddedNodes, node)
		}
	}
	for id, node := range base.Nodes {
		if _, exists := target.Nodes[id]; !exists {
			diff.RemovedNodes = append(diff.RemovedNodes, node)
		}
	}

	baseEngine := blastradius.NewReachabilityEngine(base)
	targetEngine := blastradius.NewReachabilityEngine(target)

	basePaths := baseEngine.ComputePaths()
	targetPaths := targetEngine.ComputePaths()

	baseMap := make(map[string]bool)
	for _, p := range basePaths {
		baseMap[p.AgentID+"->"+p.TargetID] = true
	}

	for _, tp := range targetPaths {
		key := tp.AgentID + "->" + tp.TargetID
		if !baseMap[key] {
			diff.AddedPaths = append(diff.AddedPaths, tp)
			diff.EscalatedCaps = append(diff.EscalatedCaps, tp.Capabilities...)
		}
	}

	if len(diff.AddedNodes) > 0 || len(diff.AddedPaths) > 0 {
		diff.SecuritySurface = SurfaceExpanded
	} else if len(diff.RemovedNodes) > 0 {
		diff.SecuritySurface = SurfaceReduced
	}

	return diff
}
