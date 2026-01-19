package bmssp

import "math"

var maxInt = int(^uint(0) >> 1)

type Label struct {
	Dist   float64
	Hops   int
	Vertex int
}

func infLabel() Label {
	return Label{
		Dist:   math.Inf(1),
		Hops:   maxInt,
		Vertex: maxInt,
	}
}

func (a Label) Less(b Label) bool {
	if a.Dist != b.Dist {
		return a.Dist < b.Dist
	}
	if a.Hops != b.Hops {
		return a.Hops < b.Hops
	}
	return a.Vertex < b.Vertex
}

func (a Label) Equal(b Label) bool {
	return a.Dist == b.Dist && a.Hops == b.Hops && a.Vertex == b.Vertex
}

func (a Label) LessOrEqual(b Label) bool {
	return a.Less(b) || a.Equal(b)
}
