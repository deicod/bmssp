package bmssp

import (
	"math"
	"math/rand"
	"testing"
)

func TestBMSSP_ZeroWeightCycle(t *testing.T) {
	g := NewGraph(4)
	g.AddEdge(0, 1, 0)
	g.AddEdge(1, 2, 0)
	g.AddEdge(2, 1, 0)
	g.AddEdge(2, 3, 1)
	g.AddEdge(0, 3, 5)

	compareBMSSPToDijkstra(t, g, 0, 3)
}

func TestBMSSP_ParallelEdges(t *testing.T) {
	g := NewGraph(3)
	g.AddEdge(0, 1, 5)
	g.AddEdge(0, 1, 2)
	g.AddEdge(1, 2, 2)
	g.AddEdge(0, 2, 10)

	compareBMSSPToDijkstra(t, g, 0, 2)
}

func TestBMSSP_SingleNode(t *testing.T) {
	g := NewGraph(1)
	solver := NewSolver(g)
	solver.ForceBMSSP = true

	dist, path := solver.Solve(0, 0)
	if dist != 0 {
		t.Fatalf("expected distance 0, got %f", dist)
	}
	if len(path) != 1 || path[0] != 0 {
		t.Fatalf("expected path [0], got %v", path)
	}
}

func TestBMSSP_DenseGraph(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	n := 18
	g := NewGraph(n)
	for u := 0; u < n; u++ {
		for v := 0; v < n; v++ {
			if u == v {
				continue
			}
			if rng.Float64() < 0.7 {
				g.AddEdge(u, v, 1+rng.Float64()*4)
			}
		}
	}

	for i := 0; i < 10; i++ {
		source := rng.Intn(n)
		target := rng.Intn(n)
		compareBMSSPToDijkstra(t, g, source, target)
	}
}

func TestBMSSP_EqualWeights(t *testing.T) {
	rng := rand.New(rand.NewSource(11))
	n := 12
	g := NewGraph(n)
	for u := 0; u < n; u++ {
		for v := 0; v < n; v++ {
			if u == v {
				continue
			}
			if rng.Float64() < 0.4 {
				g.AddEdge(u, v, 3)
			}
		}
	}

	for i := 0; i < 8; i++ {
		source := rng.Intn(n)
		target := rng.Intn(n)
		compareBMSSPToDijkstra(t, g, source, target)
	}
}

func compareBMSSPToDijkstra(t *testing.T, g *Graph, source, target int) {
	t.Helper()

	solver := NewSolver(g)
	solver.ForceBMSSP = true
	bmDist, bmPath := solver.Solve(source, target)

	dDist, _ := Dijkstra(g, source, target)
	if math.IsInf(dDist, 1) {
		if !math.IsInf(bmDist, 1) || bmPath != nil {
			t.Fatalf("expected no path, got dist=%f path=%v", bmDist, bmPath)
		}
		return
	}

	if math.Abs(bmDist-dDist) > 1e-9 {
		t.Fatalf("distance mismatch: bm=%f dijkstra=%f", bmDist, dDist)
	}
	assertValidPath(t, g, source, target, bmDist, bmPath)
}
