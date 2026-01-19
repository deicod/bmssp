package bmssp

import (
	"container/heap"
	"math"
)

const (
	kExponent = 1.0 / 3.0
	tExponent = 2.0 / 3.0
)

type Solver struct {
	Graph        *Graph
	N            int
	K            int
	T            int
	Levels       int
	Distances    []float64
	Hops         []int
	Predecessors []int
	ForceBMSSP   bool
}

func NewSolver(graph *Graph) *Solver {
	n := graph.Vertices
	k, t := computeParameters(n)
	levels := computeLevels(n, t)

	return &Solver{
		Graph:        graph,
		N:            n,
		K:            k,
		T:            t,
		Levels:       levels,
		Distances:    make([]float64, n),
		Hops:         make([]int, n),
		Predecessors: make([]int, n),
	}
}

func (s *Solver) Solve(source, goal int) (float64, []int) {
	if source < 0 || source >= s.N || goal < 0 || goal >= s.N {
		return math.Inf(1), nil
	}

	if s.N < 1000 && !s.ForceBMSSP {
		return Dijkstra(s.Graph, source, goal)
	}

	transform := NewConstantDegreeGraph(s.Graph)
	internal := NewSolver(transform.Graph)
	dist, path := internal.solveBMSSP(transform.OrigToNew[source], transform.OrigToNew[goal])
	if math.IsInf(dist, 1) || path == nil {
		return math.Inf(1), nil
	}
	mapped := transform.MapPath(path)
	if len(mapped) == 0 {
		return math.Inf(1), nil
	}
	return dist, mapped
}

func (s *Solver) solveBMSSP(source, goal int) (float64, []int) {
	s.resetState()
	s.Distances[source] = 0
	s.Hops[source] = 0

	s.bmssp(s.Levels, infLabel(), []int{source})

	if math.IsInf(s.Distances[goal], 1) {
		return math.Inf(1), nil
	}

	return s.Distances[goal], s.reconstructPath(source, goal)
}

func computeParameters(n int) (int, int) {
	if n <= 1 {
		return 1, 1
	}
	logn := math.Log2(float64(n))
	k := int(math.Floor(math.Pow(logn, kExponent)))
	t := int(math.Floor(math.Pow(logn, tExponent)))
	if k < 1 {
		k = 1
	}
	if t < 1 {
		t = 1
	}
	return k, t
}

func computeLevels(n, t int) int {
	if n <= 1 {
		return 0
	}
	if t < 1 {
		t = 1
	}
	return int(math.Ceil(math.Log2(float64(n)) / float64(t)))
}

func (s *Solver) resetState() {
	for i := 0; i < s.N; i++ {
		s.Distances[i] = math.Inf(1)
		s.Hops[i] = maxInt
		s.Predecessors[i] = -1
	}
}

func (s *Solver) label(v int) Label {
	return Label{
		Dist:   s.Distances[v],
		Hops:   s.Hops[v],
		Vertex: v,
	}
}

func (s *Solver) bmssp(level int, bound Label, frontier []int) (Label, []int) {
	if len(frontier) == 0 {
		return bound, nil
	}

	if level <= 0 {
		return s.baseCase(bound, frontier)
	}

	pivots, workingSet := s.findPivots(bound, frontier)

	blockSize := s.blockSize(level - 1)
	ds := NewFrontier(blockSize, bound)
	for _, pivot := range pivots {
		ds.Insert(pivot, s.label(pivot))
	}

	limit := s.threshold(level)
	uSet := make(map[int]struct{})
	uList := make([]int, 0, len(workingSet))
	lastBound := bound

	for len(uSet) < limit && !ds.IsEmpty() {
		subBound, subset := ds.Pull()
		if len(subset) == 0 {
			continue
		}

		subPrime, subResult := s.bmssp(level-1, subBound, subset)
		lastBound = subPrime
		addUnique(uSet, &uList, subResult)

		batch := make([]frontierItem, 0)
		for _, u := range subResult {
			for _, edge := range s.Graph.Adj[u] {
				v := edge.To
				if !s.relaxEdge(u, v, edge.Weight) {
					continue
				}
				label := s.label(v)
				if labelInRange(label, subBound, bound) {
					ds.Insert(v, label)
				} else if labelInRange(label, subPrime, subBound) {
					batch = append(batch, frontierItem{Vertex: v, Label: label})
				}
			}
		}

		for _, v := range subset {
			label := s.label(v)
			if labelInRange(label, subPrime, subBound) {
				batch = append(batch, frontierItem{Vertex: v, Label: label})
			}
		}

		if len(batch) > 0 {
			ds.BatchPrepend(batch)
		}
	}

	partial := !ds.IsEmpty() && len(uSet) >= limit
	resultBound := bound
	if partial {
		resultBound = lastBound
		if !resultBound.Less(bound) {
			resultBound = bound
		}
	}

	for _, v := range workingSet {
		if s.label(v).Less(resultBound) {
			addUnique(uSet, &uList, []int{v})
		}
	}

	return resultBound, uList
}

func (s *Solver) baseCase(bound Label, sources []int) (Label, []int) {
	if len(sources) == 0 {
		return bound, nil
	}

	pq := &labelHeap{}
	heap.Init(pq)

	for _, start := range sources {
		label := s.label(start)
		if label.Less(bound) {
			heap.Push(pq, frontierItem{Vertex: start, Label: label})
		}
	}

	visitedSet := make(map[int]struct{})
	visited := make([]int, 0, s.K+1)

	for pq.Len() > 0 && len(visited) < s.K+1 {
		item := heap.Pop(pq).(frontierItem)
		u := item.Vertex
		if !item.Label.Equal(s.label(u)) {
			continue
		}
		if _, ok := visitedSet[u]; ok {
			continue
		}
		visitedSet[u] = struct{}{}
		visited = append(visited, u)

		for _, edge := range s.Graph.Adj[u] {
			v := edge.To
			if !s.relaxEdge(u, v, edge.Weight) {
				continue
			}
			label := s.label(v)
			if label.Less(bound) {
				heap.Push(pq, frontierItem{Vertex: v, Label: label})
			}
		}
	}

	if len(visited) <= s.K {
		return bound, visited
	}

	maxLabel := s.label(visited[0])
	for i := 1; i < len(visited); i++ {
		label := s.label(visited[i])
		if maxLabel.Less(label) {
			maxLabel = label
		}
	}

	result := make([]int, 0, len(visited)-1)
	for _, v := range visited {
		if s.label(v).Less(maxLabel) {
			result = append(result, v)
		}
	}

	return maxLabel, result
}

func (s *Solver) findPivots(bound Label, frontier []int) ([]int, []int) {
	if len(frontier) == 0 {
		return nil, nil
	}

	workingSet := make(map[int]struct{}, len(frontier))
	for _, v := range frontier {
		workingSet[v] = struct{}{}
	}

	current := make([]int, 0, len(frontier))
	for _, v := range frontier {
		if s.label(v).Less(bound) {
			current = append(current, v)
		}
	}

	limit := safeMul(s.K, len(frontier))
	for i := 0; i < s.K; i++ {
		if len(current) == 0 {
			break
		}

		nextSet := make(map[int]struct{})
		for _, u := range current {
			if !s.label(u).Less(bound) {
				continue
			}
			for _, edge := range s.Graph.Adj[u] {
				v := edge.To
				if !s.relaxEdge(u, v, edge.Weight) {
					continue
				}
				if s.label(v).Less(bound) {
					if _, ok := workingSet[v]; !ok {
						workingSet[v] = struct{}{}
						nextSet[v] = struct{}{}
					}
				}
			}
		}

		if len(workingSet) > limit {
			return frontier, setToSlice(workingSet)
		}

		current = current[:0]
		for v := range nextSet {
			current = append(current, v)
		}
	}

	children := make(map[int][]int, len(workingSet))
	for v := range workingSet {
		pred := s.Predecessors[v]
		if pred == -1 {
			continue
		}
		if _, ok := workingSet[pred]; ok {
			children[pred] = append(children[pred], v)
		}
	}

	subtreeSizes := make(map[int]int, len(children))
	var dfs func(int) int
	dfs = func(u int) int {
		if size, ok := subtreeSizes[u]; ok {
			return size
		}
		size := 1
		for _, child := range children[u] {
			size += dfs(child)
		}
		subtreeSizes[u] = size
		return size
	}

	pivots := make([]int, 0, len(frontier))
	for _, root := range frontier {
		if _, ok := workingSet[root]; !ok {
			continue
		}
		if dfs(root) >= s.K {
			pivots = append(pivots, root)
		}
	}

	return pivots, setToSlice(workingSet)
}

func (s *Solver) relaxEdge(u, v int, weight float64) bool {
	if math.IsInf(s.Distances[u], 1) {
		return false
	}

	newDist := s.Distances[u] + weight
	newHops := s.Hops[u]
	if newHops < maxInt {
		newHops++
	}

	if newDist < s.Distances[v] {
		s.Distances[v] = newDist
		s.Hops[v] = newHops
		s.Predecessors[v] = u
		return true
	}
	if newDist > s.Distances[v] {
		return false
	}

	if newHops < s.Hops[v] {
		s.Distances[v] = newDist
		s.Hops[v] = newHops
		s.Predecessors[v] = u
		return true
	}
	if newHops > s.Hops[v] {
		return false
	}

	if s.Predecessors[v] == -1 || u < s.Predecessors[v] {
		s.Distances[v] = newDist
		s.Hops[v] = newHops
		s.Predecessors[v] = u
		return true
	}
	if u == s.Predecessors[v] {
		return true
	}

	return false
}

func (s *Solver) blockSize(level int) int {
	if level <= 0 {
		return 1
	}
	exp := level * s.T
	return pow2(exp)
}

func (s *Solver) threshold(level int) int {
	if level <= 0 {
		return s.K
	}
	exp := level * s.T
	return safeMul(s.K, pow2(exp))
}

func pow2(exp int) int {
	if exp <= 0 {
		return 1
	}
	value := 1
	for i := 0; i < exp; i++ {
		if value > maxInt/2 {
			return maxInt
		}
		value *= 2
	}
	return value
}

func safeMul(a, b int) int {
	if a == 0 || b == 0 {
		return 0
	}
	if a > maxInt/b {
		return maxInt
	}
	return a * b
}

func labelInRange(label, low, high Label) bool {
	return !label.Less(low) && label.Less(high)
}

func addUnique(set map[int]struct{}, list *[]int, vertices []int) {
	for _, v := range vertices {
		if _, ok := set[v]; ok {
			continue
		}
		set[v] = struct{}{}
		*list = append(*list, v)
	}
}

func setToSlice(set map[int]struct{}) []int {
	result := make([]int, 0, len(set))
	for v := range set {
		result = append(result, v)
	}
	return result
}

func (s *Solver) reconstructPath(source, goal int) []int {
	path := make([]int, 0, 16)
	curr := goal
	for curr != -1 {
		path = append(path, curr)
		if curr == source {
			break
		}
		curr = s.Predecessors[curr]
	}
	if len(path) == 0 || path[len(path)-1] != source {
		return nil
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path
}

type dijkstraItem struct {
	Vertex   int
	Distance float64
	Index    int
}

type dijkstraQueue []*dijkstraItem

func (pq dijkstraQueue) Len() int { return len(pq) }
func (pq dijkstraQueue) Less(i, j int) bool {
	return pq[i].Distance < pq[j].Distance
}
func (pq dijkstraQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}
func (pq *dijkstraQueue) Push(x interface{}) {
	item := x.(*dijkstraItem)
	item.Index = len(*pq)
	*pq = append(*pq, item)
}
func (pq *dijkstraQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.Index = -1
	*pq = old[0 : n-1]
	return item
}

func Dijkstra(g *Graph, source, goal int) (float64, []int) {
	n := g.Vertices
	dist := make([]float64, n)
	prev := make([]int, n)
	for i := 0; i < n; i++ {
		dist[i] = math.Inf(1)
		prev[i] = -1
	}
	dist[source] = 0

	pq := &dijkstraQueue{}
	heap.Init(pq)
	heap.Push(pq, &dijkstraItem{Vertex: source, Distance: 0})

	for pq.Len() > 0 {
		item := heap.Pop(pq).(*dijkstraItem)
		u := item.Vertex

		if item.Distance > dist[u] {
			continue
		}

		if u == goal {
			break
		}

		for _, edge := range g.Adj[u] {
			v := edge.To
			newDist := dist[u] + edge.Weight
			if newDist < dist[v] {
				dist[v] = newDist
				prev[v] = u
				heap.Push(pq, &dijkstraItem{Vertex: v, Distance: newDist})
			}
		}
	}

	if math.IsInf(dist[goal], 1) {
		return math.Inf(1), nil
	}

	path := make([]int, 0, 16)
	curr := goal
	for curr != -1 {
		path = append(path, curr)
		if curr == source {
			break
		}
		curr = prev[curr]
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return dist[goal], path
}
