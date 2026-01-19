package bmssp

type labelHeap []frontierItem

func (h labelHeap) Len() int           { return len(h) }
func (h labelHeap) Less(i, j int) bool { return h[i].Label.Less(h[j].Label) }
func (h labelHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *labelHeap) Push(x interface{}) {
	*h = append(*h, x.(frontierItem))
}
func (h *labelHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[:n-1]
	return item
}
