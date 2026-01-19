package bmssp

import (
	"math"
	"math/rand"
	"testing"
)

func TestConstantDegreeTransformation(t *testing.T) {
	rng := rand.New(rand.NewSource(99))
	for iter := 0; iter < 20; iter++ {
		n := 6 + rng.Intn(6)
		g := NewGraph(n)
		for u := 0; u < n; u++ {
			for v := 0; v < n; v++ {
				if u == v {
					continue
				}
				if rng.Float64() < 0.3 {
					g.AddEdge(u, v, 1+rng.Float64()*9)
				}
			}
		}

		transform := NewConstantDegreeGraph(g)
		tg := transform.Graph
		indeg := make([]int, tg.Vertices)
		for u, edges := range tg.Adj {
			if len(edges) > 2 {
				t.Fatalf("outdegree > 2 at node %d: %d", u, len(edges))
			}
			for _, edge := range edges {
				indeg[edge.To]++
			}
		}
		for v := 0; v < tg.Vertices; v++ {
			if indeg[v] > 2 {
				t.Fatalf("indegree > 2 at node %d: %d", v, indeg[v])
			}
		}

		for trial := 0; trial < 5; trial++ {
			source := rng.Intn(n)
			target := rng.Intn(n)
			distOrig, _ := Dijkstra(g, source, target)
			distTrans, pathTrans := Dijkstra(tg, transform.OrigToNew[source], transform.OrigToNew[target])

			if math.IsInf(distOrig, 1) {
				if !math.IsInf(distTrans, 1) {
					t.Fatalf("expected Inf in transformed graph, got %f", distTrans)
				}
				continue
			}

			if math.Abs(distOrig-distTrans) > 1e-9 {
				t.Fatalf("distance mismatch: orig=%f trans=%f", distOrig, distTrans)
			}

			mapped := transform.MapPath(pathTrans)
			if len(mapped) == 0 {
				t.Fatalf("mapped path is empty")
			}
			if mapped[0] != source || mapped[len(mapped)-1] != target {
				t.Fatalf("mapped path endpoints mismatch: got %v", mapped)
			}
			pathDist, ok := pathDistance(g, mapped)
			if !ok {
				t.Fatalf("mapped path invalid in original graph: %v", mapped)
			}
			if math.Abs(pathDist-distOrig) > 1e-9 {
				t.Fatalf("mapped path distance mismatch: %f vs %f", pathDist, distOrig)
			}
		}
	}
}
