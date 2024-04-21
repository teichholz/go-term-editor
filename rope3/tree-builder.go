package BRope

// A stack of partially built trees. These are kept in order of
// strictly descending height, and all vectors have a length less
// than MAX_CHILDREN and greater than zero.
// Think of them as the layers of the b-tree being built.
//
// In addition, there is a balancing invariant: for each vector
// of length greater than one, all elements satisfy `is_ok_child`.
type TreeBuilder struct {
	stack [][]Node
}

func NewTreeBuilder() TreeBuilder {
	return TreeBuilder{[][]Node{}}
}

func (t *TreeBuilder) Push(node Node) {
	for {
		if len(t.stack) == 0 {
			t.stack = append(t.stack, []Node{node})
			break
		}

		lastNode := t.stack[len(t.stack) - 1][0]
		h1, h2 := lastNode.Height(), node.Height()
		switch {
		case h1 < h2:
			node = concat(t.Pop(), node)
		case h1 > h2:
			t.stack = append(t.stack, []Node{node})
			break
		case h1 == h2:
			lastLayer := t.stack[len(t.stack) - 1]
			lastNode = lastLayer[len(lastLayer) - 1]
			if lastNode.isOkChild() && node.isOkChild() {
				t.stack[len(t.stack) - 1] = append(lastLayer, node)
			} else if h1 == 0 {
				t.stack[len(t.stack) - 1][len(lastLayer) - 1] = mergeLeaves(lastNode, node)
			} else {

			}
			break;
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
				break;
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

func (t *TreeBuilder) Build() Node {
	if len(t.stack) == 0 {
		panic("Empty stack")
	} else {
		lastNode := t.Pop()
		for len(t.stack) > 0 {
			lastNode = concat(t.Pop(), lastNode)
		}
		return lastNode
	}
}

func (t *TreeBuilder) Pop() Node {
	l := len(t.stack)
	if (l == 0) { panic("Empty stack") }
	nodes := t.stack[l - 1]
	if len(nodes) == 1 {
		return nodes[0]
	} else {
		return NodeFromNodes(nodes)
	}
}