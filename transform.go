package bmssp

import "sort"

type Transformation struct {
	Graph      *Graph
	OrigToNew  []int
	NewToOrig  []int
}

func NewConstantDegreeGraph(g *Graph) *Transformation {
	n := g.Vertices
	if n == 0 {
		return &Transformation{
			Graph:     NewGraph(0),
			OrigToNew: nil,
			NewToOrig: nil,
		}
	}

	weights := make([]map[int]float64, n)
	neighbors := make([]map[int]struct{}, n)
	for i := 0; i < n; i++ {
		neighbors[i] = make(map[int]struct{})
	}

	for u, edges := range g.Adj {
		for _, edge := range edges {
			if weights[u] == nil {
				weights[u] = make(map[int]float64)
			}
			if prev, ok := weights[u][edge.To]; !ok || edge.Weight < prev {
				weights[u][edge.To] = edge.Weight
			}
			neighbors[u][edge.To] = struct{}{}
			neighbors[edge.To][u] = struct{}{}
		}
	}

	neighborList := make([][]int, n)
	for v := 0; v < n; v++ {
		if len(neighbors[v]) == 0 {
			continue
		}
		list := make([]int, 0, len(neighbors[v]))
		for w := range neighbors[v] {
			list = append(list, w)
		}
		sort.Ints(list)
		neighborList[v] = list
	}

	indexMap := make([]map[int]int, n)
	origToNew := make([]int, n)
	newToOrig := make([]int, 0, n)
	next := 0

	for v := 0; v < n; v++ {
		if len(neighborList[v]) == 0 {
			indexMap[v] = map[int]int{v: next}
			origToNew[v] = next
			newToOrig = append(newToOrig, v)
			next++
			continue
		}
		indexMap[v] = make(map[int]int, len(neighborList[v]))
		for i, w := range neighborList[v] {
			indexMap[v][w] = next
			if i == 0 {
				origToNew[v] = next
			}
			newToOrig = append(newToOrig, v)
			next++
		}
	}

	tg := NewGraph(next)
	for v := 0; v < n; v++ {
		list := neighborList[v]
		if len(list) <= 1 {
			continue
		}
		for i := 0; i < len(list); i++ {
			from := indexMap[v][list[i]]
			to := indexMap[v][list[(i+1)%len(list)]]
			tg.AddEdge(from, to, 0)
		}
	}

	for u := 0; u < n; u++ {
		if weights[u] == nil {
			continue
		}
		for v, w := range weights[u] {
			uIdx, ok := indexMap[u][v]
			if !ok {
				continue
			}
			vIdx, ok := indexMap[v][u]
			if !ok {
				continue
			}
			tg.AddEdge(uIdx, vIdx, w)
		}
	}

	return &Transformation{
		Graph:     tg,
		OrigToNew: origToNew,
		NewToOrig: newToOrig,
	}
}

func (t *Transformation) MapPath(path []int) []int {
	if len(path) == 0 {
		return nil
	}
	out := make([]int, 0, len(path))
	last := -1
	for _, node := range path {
		if node < 0 || node >= len(t.NewToOrig) {
			continue
		}
		orig := t.NewToOrig[node]
		if orig != last {
			out = append(out, orig)
			last = orig
		}
	}
	return out
}
