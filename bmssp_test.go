package bmssp

import (
	"math"
	"math/rand"
	"testing"
)

func TestBMSSP(t *testing.T) {
	// Create a simple graph
	// 0 -> 1 (4)
	// 0 -> 2 (2)
	// 1 -> 2 (5)
	// 1 -> 3 (10)
	// 2 -> 3 (3)
	// 3 -> 4 (1)
	g := NewGraph(5)
	g.AddEdge(0, 1, 4)
	g.AddEdge(0, 2, 2)
	g.AddEdge(1, 2, 5)
	g.AddEdge(1, 3, 10)
	g.AddEdge(2, 3, 3)
	g.AddEdge(3, 4, 1)

	// Expected shortest path from 0 to 4:
	// 0 -> 2 (2)
	// 2 -> 3 (3) -> total 5
	// 3 -> 4 (1) -> total 6
	expectedDist := 6.0

	t.Run("Dijkstra Mode (Small Graph)", func(t *testing.T) {
		solver := NewSolver(g)
		dist, path := solver.Solve(0, 4)

		if dist != expectedDist {
			t.Errorf("Expected distance %f, got %f", expectedDist, dist)
		}
		assertValidPath(t, g, 0, 4, dist, path)
	})

	t.Run("BMSSP Mode (Forced)", func(t *testing.T) {
		solver := NewSolver(g)
		solver.ForceBMSSP = true
		dist, path := solver.Solve(0, 4)

		if dist != expectedDist {
			t.Errorf("Expected distance %f, got %f", expectedDist, dist)
		}
		assertValidPath(t, g, 0, 4, dist, path)
	})
}

func TestBMSSP_NoPath(t *testing.T) {
	g := NewGraph(3)
	g.AddEdge(0, 1, 1)
	// No edge to 2

	solver := NewSolver(g)
	solver.ForceBMSSP = true
	dist, path := solver.Solve(0, 2)

	if !math.IsInf(dist, 1) {
		t.Errorf("Expected infinite distance, got %f", dist)
	}
	if path != nil {
		t.Errorf("Expected nil path, got %v", path)
	}
}

func TestBMSSP_RandomGraphs(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	for i := 0; i < 25; i++ {
		n := 15 + rng.Intn(10)
		g := NewGraph(n)
		for u := 0; u < n; u++ {
			for v := 0; v < n; v++ {
				if u == v {
					continue
				}
				if rng.Float64() < 0.2 {
					g.AddEdge(u, v, 1+rng.Float64()*9)
				}
			}
		}

		source := rng.Intn(n)
		target := rng.Intn(n)

		solver := NewSolver(g)
		solver.ForceBMSSP = true
		bmDist, bmPath := solver.Solve(source, target)

		dDist, dPath := Dijkstra(g, source, target)

		if math.IsInf(dDist, 1) {
			if !math.IsInf(bmDist, 1) || bmPath != nil {
				t.Fatalf("Expected BMSSP to return no path, got dist=%f path=%v", bmDist, bmPath)
			}
			continue
		}

		if math.Abs(bmDist-dDist) > 1e-9 {
			t.Fatalf("BMSSP distance %f != Dijkstra distance %f", bmDist, dDist)
		}
		assertValidPath(t, g, source, target, bmDist, bmPath)
		assertValidPath(t, g, source, target, dDist, dPath)
	}
}

func assertValidPath(t *testing.T, g *Graph, source, target int, expected float64, path []int) {
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
	dist, ok := pathDistance(g, path)
	if !ok {
		t.Fatalf("Path contains invalid edge: %v", path)
	}
	if math.Abs(dist-expected) > 1e-9 {
		t.Fatalf("Path distance %f != expected %f", dist, expected)
	}
}

func pathDistance(g *Graph, path []int) (float64, bool) {
	if len(path) < 2 {
		return 0, true
	}
	total := 0.0
	for i := 0; i < len(path)-1; i++ {
		u := path[i]
		v := path[i+1]
		found := false
		minWeight := math.Inf(1)
		for _, edge := range g.Adj[u] {
			if edge.To == v && edge.Weight < minWeight {
				minWeight = edge.Weight
				found = true
			}
		}
		if !found {
			return 0, false
		}
		total += minWeight
	}
	return total, true
}
