package bmssp

import (
	"fmt"
	"sort"

	"gonum.org/v1/gonum/graph"
)

// GonumConverter handles the conversion between a Gonum graph and a BMSSP graph.
// It maintains the mapping between Gonum's int64 IDs and BMSSP's int indices.
type GonumConverter struct {
	GonumToBMSSP map[int64]int
	BMSSPToGonum []int64
	Graph        *Graph
}

// NewGonumConverter creates a new converter and builds the BMSSP graph from the source.
// The source graph must be a weighted directed graph.
func NewGonumConverter(g graph.WeightedDirected) *GonumConverter {
	nodes := g.Nodes()
	var gonumIDs []int64

	// Collect all node IDs
	for nodes.Next() {
		gonumIDs = append(gonumIDs, nodes.Node().ID())
	}

	// Sort IDs for deterministic mapping
	sort.Slice(gonumIDs, func(i, j int) bool {
		return gonumIDs[i] < gonumIDs[j]
	})

	n := len(gonumIDs)
	gonumToBMSSP := make(map[int64]int, n)
	bmsspToGonum := make([]int64, n)

	// Create mapping
	for i, id := range gonumIDs {
		gonumToBMSSP[id] = i
		bmsspToGonum[i] = id
	}

	bmsspGraph := NewGraph(n)

	// Populate edges
	for i, uID := range bmsspToGonum {
		uNode := g.Node(uID)
		if uNode == nil {
			continue // Should not happen based on iteration above
		}

		toNodes := g.From(uID)
		for toNodes.Next() {
			vNode := toNodes.Node()
			vID := vNode.ID()

			vIdx, ok := gonumToBMSSP[vID]
			if !ok {
				continue
			}

			// Get weight
			weight, ok := g.Weight(uID, vID)
			if !ok {
				weight = 1.0
			}

			bmsspGraph.AddEdge(i, vIdx, weight)
		}
	}

	return &GonumConverter{
		GonumToBMSSP: gonumToBMSSP,
		BMSSPToGonum: bmsspToGonum,
		Graph:        bmsspGraph,
	}
}

// SolveGonum is a convenience function to solve SSSP on a Gonum graph using BMSSP.
func SolveGonum(g graph.WeightedDirected, sourceID, goalID int64) (float64, []int64, error) {
	converter := NewGonumConverter(g)

	srcIdx, ok := converter.GonumToBMSSP[sourceID]
	if !ok {
		return 0, nil, fmt.Errorf("source node %d not found in graph", sourceID)
	}

	goalIdx, ok := converter.GonumToBMSSP[goalID]
	if !ok {
		return 0, nil, fmt.Errorf("goal node %d not found in graph", goalID)
	}

	solver := NewSolver(converter.Graph)
	dist, pathIndices := solver.Solve(srcIdx, goalIdx)

	if pathIndices == nil {
		return dist, nil, nil // No path found, dist is Inf
	}

	// Convert path back to Gonum IDs
	pathIDs := make([]int64, len(pathIndices))
	for i, idx := range pathIndices {
		pathIDs[i] = converter.BMSSPToGonum[idx]
	}

	return dist, pathIDs, nil
}
