package bmssp

import "sort"

type frontierItem struct {
	Vertex int
	Label  Label
}

type block struct {
	items []frontierItem
	upper Label
	id    int
	prev  *block
	next  *block
	inD0  bool
}

func newBlock(items []frontierItem, id int, inD0 bool) *block {
	sort.Slice(items, func(i, j int) bool {
		return items[i].Label.Less(items[j].Label)
	})
	b := &block{
		items: items,
		id:    id,
		inD0:  inD0,
	}
	b.recomputeUpper()
	return b
}

func (b *block) recomputeUpper() {
	if len(b.items) == 0 {
		b.upper = infLabel()
		return
	}
	b.upper = b.items[len(b.items)-1].Label
}

type blockList struct {
	head *block
	tail *block
}

func (l *blockList) append(b *block) {
	b.prev = l.tail
	b.next = nil
	if l.tail != nil {
		l.tail.next = b
	} else {
		l.head = b
	}
	l.tail = b
}

func (l *blockList) insertBefore(ref *block, b *block) {
	if ref == nil {
		l.append(b)
		return
	}
	b.prev = ref.prev
	b.next = ref
	if ref.prev != nil {
		ref.prev.next = b
	} else {
		l.head = b
	}
	ref.prev = b
}

func (l *blockList) insertAfter(ref *block, b *block) {
	if ref == nil {
		l.append(b)
		return
	}
	b.next = ref.next
	b.prev = ref
	if ref.next != nil {
		ref.next.prev = b
	} else {
		l.tail = b
	}
	ref.next = b
}

func (l *blockList) remove(b *block) {
	if b.prev != nil {
		b.prev.next = b.next
	} else {
		l.head = b.next
	}
	if b.next != nil {
		b.next.prev = b.prev
	} else {
		l.tail = b.prev
	}
	b.prev = nil
	b.next = nil
}

func (l *blockList) prependBlocks(blocks []*block) {
	if len(blocks) == 0 {
		return
	}
	for i := 0; i < len(blocks); i++ {
		blocks[i].prev = nil
		blocks[i].next = nil
		if i > 0 {
			blocks[i-1].next = blocks[i]
			blocks[i].prev = blocks[i-1]
		}
	}
	first := blocks[0]
	last := blocks[len(blocks)-1]
	if l.head != nil {
		last.next = l.head
		l.head.prev = last
	} else {
		l.tail = last
	}
	l.head = first
}

func (l *blockList) prefix(limit int) ([]*block, int) {
	total := 0
	blocks := make([]*block, 0)
	for b := l.head; b != nil && total < limit; b = b.next {
		blocks = append(blocks, b)
		total += len(b.items)
	}
	return blocks, total
}

type Frontier struct {
	bound       Label
	limit       int
	d0          blockList
	d1          blockList
	index       blockIndex
	values      map[int]Label
	locations   map[int]*block
	nextBlockID int
}

func NewFrontier(limit int, bound Label) *Frontier {
	if limit < 1 {
		limit = 1
	}
	return &Frontier{
		bound:     bound,
		limit:     limit,
		index:     newBlockIndex(),
		values:    make(map[int]Label),
		locations: make(map[int]*block),
	}
}

func (f *Frontier) Insert(vertex int, label Label) {
	if !label.Less(f.bound) {
		return
	}
	if existing, ok := f.values[vertex]; ok {
		if !label.Less(existing) {
			return
		}
		f.remove(vertex)
	}

	item := frontierItem{Vertex: vertex, Label: label}

	if f.d1.head == nil {
		b := newBlock([]frontierItem{item}, f.nextID(), false)
		f.d1.append(b)
		f.index.Insert(b.upper, b.id, b)
		f.values[vertex] = label
		f.locations[vertex] = b
		return
	}

	target := f.index.LowerBound(label)
	if target == nil {
		b := newBlock([]frontierItem{item}, f.nextID(), false)
		f.d1.append(b)
		f.index.Insert(b.upper, b.id, b)
		f.values[vertex] = label
		f.locations[vertex] = b
		return
	}

	oldUpper := target.upper
	f.insertIntoBlock(target, item)
	f.values[vertex] = label
	f.locations[vertex] = target
	f.updateIndex(target, oldUpper)

	if len(target.items) > f.limit {
		f.splitBlock(target)
	}
}

func (f *Frontier) BatchPrepend(items []frontierItem) {
	if len(items) == 0 {
		return
	}

	filtered := make([]frontierItem, 0, len(items))
	for _, item := range items {
		if !item.Label.Less(f.bound) {
			continue
		}
		if existing, ok := f.values[item.Vertex]; ok {
			if !item.Label.Less(existing) {
				continue
			}
			f.remove(item.Vertex)
		}
		filtered = append(filtered, item)
	}
	if len(filtered) == 0 {
		return
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Label.Less(filtered[j].Label)
	})

	blocks := make([]*block, 0, (len(filtered)+f.limit-1)/f.limit)
	for i := 0; i < len(filtered); i += f.limit {
		end := i + f.limit
		if end > len(filtered) {
			end = len(filtered)
		}
		blockItems := append([]frontierItem(nil), filtered[i:end]...)
		b := newBlock(blockItems, f.nextID(), true)
		blocks = append(blocks, b)
		for _, item := range blockItems {
			f.values[item.Vertex] = item.Label
			f.locations[item.Vertex] = b
		}
	}
	f.d0.prependBlocks(blocks)
}

func (f *Frontier) Pull() (Label, []int) {
	if f.IsEmpty() {
		return f.bound, nil
	}

	result := make([]int, 0, f.limit)
	blocks0, size0 := f.d0.prefix(f.limit)
	blocks1, size1 := f.d1.prefix(f.limit)
	total := size0 + size1

	if total <= f.limit {
		f.removePrefix(&f.d0, blocks0, &result)
		f.removePrefix(&f.d1, blocks1, &result)
		return f.nextBound(), result
	}

	candidates := make([]frontierItem, 0, total)
	for _, b := range blocks0 {
		candidates = append(candidates, b.items...)
	}
	for _, b := range blocks1 {
		candidates = append(candidates, b.items...)
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Label.Less(candidates[j].Label)
	})
	cutoff := candidates[f.limit-1].Label

	f.removeUpToCutoff(&f.d0, blocks0, cutoff, &result)
	f.removeUpToCutoff(&f.d1, blocks1, cutoff, &result)

	if len(result) > f.limit {
		result = result[:f.limit]
	}

	return f.nextBound(), result
}

func (f *Frontier) IsEmpty() bool {
	return len(f.values) == 0
}

func (f *Frontier) insertIntoBlock(b *block, item frontierItem) {
	idx := sort.Search(len(b.items), func(i int) bool {
		return !b.items[i].Label.Less(item.Label)
	})
	b.items = append(b.items, frontierItem{})
	copy(b.items[idx+1:], b.items[idx:])
	b.items[idx] = item
	b.recomputeUpper()
}

func (f *Frontier) splitBlock(b *block) {
	if len(b.items) <= f.limit {
		return
	}
	oldUpper := b.upper
	mid := len(b.items) / 2
	rightItems := append([]frontierItem(nil), b.items[mid:]...)
	b.items = b.items[:mid]
	b.recomputeUpper()
	f.updateIndex(b, oldUpper)

	right := newBlock(rightItems, f.nextID(), false)
	f.d1.insertAfter(b, right)
	f.index.Insert(right.upper, right.id, right)
	for _, item := range right.items {
		f.locations[item.Vertex] = right
	}
}

func (f *Frontier) updateIndex(b *block, oldUpper Label) {
	if b.inD0 {
		return
	}
	if b.upper.Equal(oldUpper) {
		return
	}
	f.index.Delete(oldUpper, b.id)
	f.index.Insert(b.upper, b.id, b)
}

func (f *Frontier) remove(vertex int) {
	b, ok := f.locations[vertex]
	if !ok {
		return
	}
	label := f.values[vertex]
	delete(f.values, vertex)
	delete(f.locations, vertex)

	oldUpper := b.upper
	idx := sort.Search(len(b.items), func(i int) bool {
		return !b.items[i].Label.Less(label)
	})
	for idx < len(b.items) && b.items[idx].Vertex != vertex {
		idx++
	}
	if idx >= len(b.items) {
		return
	}
	copy(b.items[idx:], b.items[idx+1:])
	b.items = b.items[:len(b.items)-1]
	if len(b.items) == 0 {
		if b.inD0 {
			f.d0.remove(b)
		} else {
			f.d1.remove(b)
			f.index.Delete(oldUpper, b.id)
		}
		return
	}
	b.recomputeUpper()
	f.updateIndex(b, oldUpper)
}

func (f *Frontier) removePrefix(list *blockList, blocks []*block, result *[]int) {
	for _, b := range blocks {
		for _, item := range b.items {
			delete(f.values, item.Vertex)
			delete(f.locations, item.Vertex)
			*result = append(*result, item.Vertex)
		}
		if b.inD0 {
			list.remove(b)
		} else {
			list.remove(b)
			f.index.Delete(b.upper, b.id)
		}
	}
}

func (f *Frontier) removeUpToCutoff(list *blockList, blocks []*block, cutoff Label, result *[]int) {
	for _, b := range blocks {
		if len(b.items) == 0 {
			continue
		}
		idx := sort.Search(len(b.items), func(i int) bool {
			return cutoff.Less(b.items[i].Label)
		})
		if idx == 0 {
			continue
		}
		for i := 0; i < idx; i++ {
			item := b.items[i]
			delete(f.values, item.Vertex)
			delete(f.locations, item.Vertex)
			*result = append(*result, item.Vertex)
		}
		if idx >= len(b.items) {
			if b.inD0 {
				list.remove(b)
			} else {
				list.remove(b)
				f.index.Delete(b.upper, b.id)
			}
			continue
		}
		b.items = b.items[idx:]
	}
}

func (f *Frontier) nextBound() Label {
	bound := f.bound
	if f.d0.head != nil {
		bound = f.d0.head.items[0].Label
	}
	if f.d1.head != nil {
		candidate := f.d1.head.items[0].Label
		if f.d0.head == nil || candidate.Less(bound) {
			bound = candidate
		}
	}
	return bound
}

func (f *Frontier) nextID() int {
	id := f.nextBlockID
	f.nextBlockID++
	return id
}
