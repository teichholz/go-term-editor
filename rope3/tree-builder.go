package BRope

import "slices"

// TreeBuilder A stack of partially built trees.
// These are kept in order of strictly descending height, and all vectors have a length less
// than MAX_CHILDREN and greater than zero.
// Each vector contains nodes of the same height.
//
// In addition, there is a balancing invariant: for each vector
// of length greater than one, all elements satisfy `is_ok_child`.
type TreeBuilder struct {
	stack [][]Node
}

func NewTreeBuilder() TreeBuilder {
	return TreeBuilder{[][]Node{}}
}

func (t *TreeBuilder) Push(toInsert Node) {
	for {
		if len(t.stack) == 0 {
			t.stack = append(t.stack, []Node{toInsert})
			break
		}

		lastLayerFirstNode := t.stack[len(t.stack)-1][0]
		h1, h2 := lastLayerFirstNode.Height(), toInsert.Height()
		switch {
		case h1 < h2:
			// toInsert height > last layer height
			toInsert = concat(t.Pop(), toInsert)
		case h1 > h2:
			t.stack = append(t.stack, []Node{toInsert})
			return
		case h1 == h2:
			// TODO rewrite using tos pointer?
			// tos = top of stack = last layer = last index of slice
			tos := &t.stack[len(t.stack)-1]
			lastLayerLastNode := (*tos)[len(*tos)-1]
			if lastLayerLastNode.isOkChild() && toInsert.isOkChild() {
				// simply append
				*tos = append(*tos, toInsert)
			} else if toInsert.Height() == 0 {
				// efficiently merge leaf nodes
				// cut off last node
				*tos = (*tos)[:len(*tos)-1]
				leaf1 := lastLayerLastNode.getLeaf()
				leaf2 := toInsert.getLeaf()
				if leaf1.Len()+leaf2.Len() <= MAX_LEAF {
					newRunes := slices.Concat(leaf1.Runes(), leaf2.Runes())
					newLeaf := StringLeaf{newRunes}
					*tos = append(*tos, NodeFromLeaf(newLeaf))
				} else {
					space := MAX_LEAF - leaf1.Len()
					left := StringLeaf{slices.Concat(leaf1.Runes(), leaf2.Runes()[:space])}
					right := StringLeaf{leaf2.Runes()[space:]}
					*tos = append(*tos, NodeFromLeaf(left))
					*tos = append(*tos, NodeFromLeaf(right))
				}
			} else {
				// not ok, not leafs. Try to make ok
				// cut off last node
				*tos = (*tos)[:len(*tos)-1]
				children1 := lastLayerLastNode.getChildren()
				children2 := toInsert.getChildren()
				nChildren := len(children1) + len(children2)
				if nChildren <= MAX_CHILDREN {
					node := NodeFromNodes(slices.Concat(children1, children2))
					*tos = append(*tos, node)
				} else {
					splitpoint := min(MAX_CHILDREN, nChildren-MIN_CHILDREN)
					all := slices.Concat(children1, children2)
					left := NodeFromNodes(slices.Clone(all[:splitpoint]))
					right := NodeFromNodes(slices.Clone(all[splitpoint:]))
					*tos = append(*tos, left)
					*tos = append(*tos, right)
				}
			}
			if len(*tos) < MAX_CHILDREN {
				return
			}
			toInsert = t.Pop()
		}

	}
}

func (t *TreeBuilder) PushSlice(node Node, iv Interval) {
	if iv.IsEmpty() {
		return
	}
	if iv == node.interval() {
		t.Push(node)
		return
	}
	switch node.NodeVal.(type) {
	case LeafNodeVal:
		t.PushLeafSlice(node.getLeaf(), iv)
	case InternalNodeVal:
		offset := 0
		for _, child := range node.getChildren() {
			if iv.IsBefore(offset) {
				break
			}
			childIv := child.interval()
			recIv := iv.Intersection(childIv.Translate(offset)).Translate(-offset)
			t.PushSlice(child, recIv)
			offset += childIv.Len()
		}
	default:
		panic("Unknown node type")
	}
}

func (t *TreeBuilder) PushLeaves(leaves []Leaf) {
	for _, leaf := range leaves {
		t.PushLeaf(leaf)
	}
}

func (t *TreeBuilder) PushLeaf(leaf Leaf) {
	t.Push(NodeFromLeaf(leaf))
}

func (t *TreeBuilder) PushLeafSlice(leaf Leaf, iv Interval) {
	t.PushLeaf(leaf.Slice(iv))
}

// Build the final tree.
//
// The tree is the concatenation of all the nodes and leaves that have been pushed
// on the builder, in order.
func (t *TreeBuilder) Build() Node {
	if len(t.stack) == 0 {
		panic("Empty stack")
	} else {
		// TODO Push is buggy, stack contains: f, foo, foobar
		lastNode := t.Pop()
		for len(t.stack) > 0 {
			lastNode = concat(t.Pop(), lastNode)
		}
		return lastNode
	}
}

func (t *TreeBuilder) Pop() Node {
	l := len(t.stack)
	if l == 0 {
		panic("Empty stack")
	}
	tos := t.stack[l-1]
	t.stack = t.stack[:l-1]
	if len(tos) == 1 {
		return tos[0]
	} else {
		// Invarinace enforced by Push
		return NodeFromNodes(tos)
	}
}
