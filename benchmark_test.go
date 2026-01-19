package bmssp

import (
	"math/rand"
	"testing"
)

const (
	benchNodes = 1000
	benchEdges = 4000
	benchPairs = 256
)

type pair struct {
	source int
	target int
}

func BenchmarkDijkstraSparse(b *testing.B) {
	g := makeSparseGraph(benchNodes, benchEdges, 1)
	pairs := makePairs(benchNodes, benchPairs, 2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := pairs[i%len(pairs)]
		Dijkstra(g, p.source, p.target)
	}
}

func BenchmarkBMSSPSparse(b *testing.B) {
	g := makeSparseGraph(benchNodes, benchEdges, 3)
	pairs := makePairs(benchNodes, benchPairs, 4)
	solver := NewSolver(g)
	solver.ForceBMSSP = true

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := pairs[i%len(pairs)]
		solver.Solve(p.source, p.target)
	}
}

func BenchmarkBMSSPSparseCore(b *testing.B) {
	g := makeSparseGraph(benchNodes, benchEdges, 5)
	pairs := makePairs(benchNodes, benchPairs, 6)

	transform := NewConstantDegreeGraph(g)
	solver := NewSolver(transform.Graph)
	mapped := make([]pair, len(pairs))
	for i, p := range pairs {
		mapped[i] = pair{
			source: transform.OrigToNew[p.source],
			target: transform.OrigToNew[p.target],
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := mapped[i%len(mapped)]
		solver.solveBMSSP(p.source, p.target)
	}
}

func makeSparseGraph(n, m int, seed int64) *Graph {
	if n < 1 {
		return NewGraph(0)
	}
	g := NewGraph(n)
	rng := rand.New(rand.NewSource(seed))

	for i := 0; i < m; i++ {
		u := rng.Intn(n)
		v := rng.Intn(n - 1)
		if v >= u {
			v++
		}
		weight := 1 + rng.Float64()*9
		g.AddEdge(u, v, weight)
	}
	return g
}

func makePairs(n, count int, seed int64) []pair {
	if count < 1 {
		count = 1
	}
	rng := rand.New(rand.NewSource(seed))
	pairs := make([]pair, count)
	for i := 0; i < count; i++ {
		source := rng.Intn(n)
		target := rng.Intn(n)
		pairs[i] = pair{source: source, target: target}
	}
	return pairs
}
