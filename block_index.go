package bmssp

type blockIndex struct {
	root *blockNode
	seed uint32
}

type blockNode struct {
	label    Label
	id       int
	block    *block
	priority uint32
	left     *blockNode
	right    *blockNode
}

func newBlockIndex() blockIndex {
	return blockIndex{seed: 1}
}

func (t *blockIndex) Insert(label Label, id int, b *block) {
	node := &blockNode{
		label:    label,
		id:       id,
		block:    b,
		priority: t.nextPriority(),
	}
	t.root = insertNode(t.root, node)
}

func (t *blockIndex) Delete(label Label, id int) {
	t.root = deleteNode(t.root, label, id)
}

func (t *blockIndex) LowerBound(label Label) *block {
	node := t.root
	var best *blockNode
	for node != nil {
		if keyLess(node.label, node.id, label, -1) {
			node = node.right
		} else {
			best = node
			node = node.left
		}
	}
	if best == nil {
		return nil
	}
	return best.block
}

func (t *blockIndex) nextPriority() uint32 {
	t.seed = t.seed*1664525 + 1013904223
	return t.seed
}

func keyLess(a Label, aid int, b Label, bid int) bool {
	if a.Less(b) {
		return true
	}
	if b.Less(a) {
		return false
	}
	return aid < bid
}

func insertNode(root *blockNode, node *blockNode) *blockNode {
	if root == nil {
		return node
	}
	if keyLess(node.label, node.id, root.label, root.id) {
		root.left = insertNode(root.left, node)
		if root.left.priority < root.priority {
			root = rotateRight(root)
		}
		return root
	}
	root.right = insertNode(root.right, node)
	if root.right.priority < root.priority {
		root = rotateLeft(root)
	}
	return root
}

func deleteNode(root *blockNode, label Label, id int) *blockNode {
	if root == nil {
		return nil
	}
	if keyLess(label, id, root.label, root.id) {
		root.left = deleteNode(root.left, label, id)
		return root
	}
	if keyLess(root.label, root.id, label, id) {
		root.right = deleteNode(root.right, label, id)
		return root
	}
	return mergeNodes(root.left, root.right)
}

func mergeNodes(left, right *blockNode) *blockNode {
	if left == nil {
		return right
	}
	if right == nil {
		return left
	}
	if left.priority < right.priority {
		left.right = mergeNodes(left.right, right)
		return left
	}
	right.left = mergeNodes(left, right.left)
	return right
}

func rotateRight(y *blockNode) *blockNode {
	x := y.left
	t2 := x.right
	x.right = y
	y.left = t2
	return x
}

func rotateLeft(x *blockNode) *blockNode {
	y := x.right
	t2 := y.left
	y.left = x
	x.right = t2
	return y
}
