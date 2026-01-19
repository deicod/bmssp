package bmssp

import (
	"fmt"
)

type Edge struct {
	To     int
	Weight float64
}

type Graph struct {
	Vertices int
	Edges    int
	Adj      [][]Edge
}

func NewGraph(vertices int) *Graph {
	if vertices < 0 {
		panic("Number of vertices cannot be negative")
	}
	return &Graph{
		Vertices: vertices,
		Edges:    0,
		Adj:      make([][]Edge, vertices),
	}
}

func (g *Graph) AddEdge(u, v int, weight float64) {
	if u < 0 || u >= g.Vertices || v < 0 || v >= g.Vertices {
		panic(fmt.Sprintf("Vertex index out of bounds: u=%d, v=%d, vertices=%d", u, v, g.Vertices))
	}
	g.Adj[u] = append(g.Adj[u], Edge{To: v, Weight: weight})
	g.Edges++
}
