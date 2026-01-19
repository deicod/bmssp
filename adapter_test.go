package bmssp

import (
	"math"
	"testing"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

func TestSolveGonum(t *testing.T) {
	// Create a generic directed weighted graph using simple.WeightedDirectedGraph
	g := simple.NewWeightedDirectedGraph(0, 0)

	// Add nodes: 0, 1, 2, 3, 4 with IDs 10, 20, 30, 40, 50 (sparse/arbitrary IDs)

	// simple.NewNode ID strategy starts at 0 or maxID+1.
	// Let's add nodes specifically if we want specific IDs, but simple.NewNode is simpler.
	// Actually simple.WeightedDirectedGraph.AddNode accepts a Node. 
	// To customize IDs, we can implement our own Node or just rely on autogen.
	// Let's try to make IDs non-0-indexed to test mapping.
	
	nodes := make([]simple.Node, 5)
	for i := 0; i < 5; i++ {
		// Create nodes with ID = (i+1)*10
		nodes[i] = simple.Node((i + 1) * 10)
		g.AddNode(nodes[i])
	}

	// Edges similar to previous test:
	// 10 -> 20 (4)
	// 10 -> 30 (2)
	// 20 -> 30 (5)
	// 20 -> 40 (10)
	// 30 -> 40 (3)
	// 40 -> 50 (1)

	g.SetWeightedEdge(g.NewWeightedEdge(nodes[0], nodes[1], 4)) // 10->20
	g.SetWeightedEdge(g.NewWeightedEdge(nodes[0], nodes[2], 2)) // 10->30
	g.SetWeightedEdge(g.NewWeightedEdge(nodes[1], nodes[2], 5)) // 20->30
	g.SetWeightedEdge(g.NewWeightedEdge(nodes[1], nodes[3], 10)) // 20->40
	g.SetWeightedEdge(g.NewWeightedEdge(nodes[2], nodes[3], 3)) // 30->40
	g.SetWeightedEdge(g.NewWeightedEdge(nodes[3], nodes[4], 1)) // 40->50

	// Path: 10 -> 30 -> 40 -> 50, Distance: 2 + 3 + 1 = 6
	expectedDist := 6.0

	dist, path, err := SolveGonum(g, 10, 50)
	if err != nil {
		t.Fatalf("SolveGonum failed: %v", err)
	}

	if dist != expectedDist {
		t.Errorf("Expected distance %f, got %f", expectedDist, dist)
	}

	assertValidGonumPath(t, g, 10, 50, dist, path)
}

func TestSolveGonum_NoPath(t *testing.T) {
	g := simple.NewWeightedDirectedGraph(0, 0)
	n1 := simple.Node(1)
	n2 := simple.Node(2)
	n3 := simple.Node(3)
	
	g.AddNode(n1)
	g.AddNode(n2)
	g.AddNode(n3)
	
	g.SetWeightedEdge(g.NewWeightedEdge(n1, n2, 5))
	// No path to n3
	
	dist, path, err := SolveGonum(g, 1, 3)
	if err != nil {
		t.Fatalf("SolveGonum error: %v", err)
	}
	
	if !math.IsInf(dist, 1) {
		t.Errorf("Expected Inf distance, got %f", dist)
	}
	if path != nil {
		t.Errorf("Expected nil path, got %v", path)
	}
}

func TestSolveGonum_InvalidID(t *testing.T) {
	g := simple.NewWeightedDirectedGraph(0, 0)
	n1 := simple.Node(1)
	g.AddNode(n1)
	
	_, _, err := SolveGonum(g, 1, 999)
	if err == nil {
		t.Errorf("Expected error for invalid goal ID, got nil")
	}
}

func assertValidGonumPath(t *testing.T, g graph.WeightedDirected, source, target int64, expected float64, path []int64) {
	t.Helper()
	if path == nil {
		t.Fatalf("Expected path, got nil")
	}
	if len(path) == 0 {
		t.Fatalf("Expected non-empty path")
	}
	if path[0] != source {
		t.Fatalf("Path does not start at source: got %d want %d", path[0], source)
	}
	if path[len(path)-1] != target {
		t.Fatalf("Path does not end at target: got %d want %d", path[len(path)-1], target)
	}
	dist, ok := gonumPathDistance(g, path)
	if !ok {
		t.Fatalf("Path contains invalid edge: %v", path)
	}
	if math.Abs(dist-expected) > 1e-9 {
		t.Fatalf("Path distance %f != expected %f", dist, expected)
	}
}

func gonumPathDistance(g graph.WeightedDirected, path []int64) (float64, bool) {
	if len(path) < 2 {
		return 0, true
	}
	total := 0.0
	for i := 0; i < len(path)-1; i++ {
		u := path[i]
		v := path[i+1]
		weight, ok := g.Weight(u, v)
		if !ok {
			return 0, false
		}
		total += weight
	}
	return total, true
}
