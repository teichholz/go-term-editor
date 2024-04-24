package BRope

const (
	MIN_LEAF = 511
	MAX_LEAF = 1024
	MIN_CHILDREN = 4
	MAX_CHILDREN = 8
)

// The Rope data structure is a b-tree that store arbitratry length strings.
// It supports efficient insertion and deletion, no matter the position.
// The struct is immutable and supposed to be passed by value.

// Operations:
// Insert
// Replace
// Erase
// OffsetOfLine
// LineOfOffset

// TODO Features:
// Cursor with Metrics
// Parameterized b-tree

type Leaf interface {
	Len() int
	String() string
	IsOkChild() bool
	Slice(Interval) Leaf
	Interval() Interval
}

type StringLeaf struct {
	string
}

func (s StringLeaf) Len() int {
	return len(s.string)
}

func (s StringLeaf) String() string {
	return s.string
}

func (s StringLeaf) IsOkChild() bool {
	return len(s.string) >= MIN_LEAF
}

func (s StringLeaf) Slice(iv Interval) Leaf {
	return StringLeaf{s.String()[iv.Lo:iv.Hi]}
}

func (s StringLeaf) Interval() Interval {
	return Interval{0, s.Len()}
}

func NewString(s string) Node {
	leaf := StringLeaf{s}
	return NodeFromLeaf(leaf)
}

// Type of the nodes in the b-tree
type Node struct {
	*NodeBody
}

// The body of the node
type NodeBody struct {
	height, len int
	NodeInfo
	NodeVal
}

type NodeInfo struct {
	lineCount int
}

func (n NodeInfo) compute(leaf Leaf) {
	// TODO fix
	n.lineCount = leaf.Len()
}

func (n NodeInfo) accumulate(other NodeInfo) {
	n.lineCount = n.lineCount + other.lineCount
}

// NodeVal Type of node values.
// Each node is either an internal node or a leaf node.
type NodeVal interface {
	isNode()
}

type LeafNodeVal struct {
	Leaf
}
func (LeafNodeVal) isNode() {}

type InternalNodeVal struct {
	nodes []Node
}
func (InternalNodeVal) isNode() {}

func NodeFromLeaf(leaf Leaf) Node {
	info := NodeInfo{}
	len := leaf.Len()
	info.compute(leaf)
	return Node{&NodeBody{0, len, info, LeafNodeVal{leaf}}}
}

// NodeFromNodes Invariance: 1 <= len(nodes) <= MAX_CHILDREN
// Invariance: All nodes are same height
// Invariance: All the nodes must satisfy IsOkChild
func NodeFromNodes(nodes []Node) Node {
	if len(nodes) < 1 { panic("Invariance: 1 <= len(nodes) <= MAX_CHILDREN") }
	height := nodes[0].height
	len := 0
	info := NodeInfo{}
	for _, n := range nodes {
		if height + 1 != n.height + 1 { panic("Invariance: All nodes are same height") }
		if !n.isOkChild() { panic("Invariance: All nodes must satisfy isOkChild") }
		len += n.len
		info.accumulate(n.NodeInfo)
	}
	return Node{&NodeBody{height + 1, len, info, InternalNodeVal{nodes}}}
}

func (n Node) Len() int {
	return n.len
}

func (n Node) IsEmpty() bool {
	return n.Len() == 0
}

func (n Node) Height() int {
	return n.height
}

func (n Node) isLeaf() bool {
	return n.Height() == 0
}

func (n Node) interval() Interval {
	return Interval{0, n.Len()}
}

func (n Node) getChildren() []Node {
	internal, ok := n.NodeVal.(InternalNodeVal)
	if !ok {
		panic("Leaf node has no children")
	}

	return internal.nodes
}

func (n Node) getLeaf() Leaf {
	leaf, ok := n.NodeVal.(LeafNodeVal)
	if !ok {
		panic("Internal node has no leaf")
	}

	return leaf
}

func (n Node) isOkChild() bool {
	// type case on Nodfunc (n Node) isOkChild() bool {
    switch n.NodeVal.(type) {
    case InternalNodeVal:
        return len(n.getChildren()) >= MIN_CHILDREN
    case LeafNodeVal:
        return n.getLeaf().IsOkChild()
    default:
        panic("Unknown node type")
    }
}

// Always returns a new InternalNodeVal
func mergeNodes(children1 []Node, children2 []Node) Node {
	compoundSize := len(children1) + len(children2)
	all := append(children1[:len(children1):len(children1)], children2...)

	if compoundSize <= MAX_CHILDREN {
		return NodeFromNodes(all)
	} else {
		splitpoint := min(MAX_CHILDREN, compoundSize - MIN_CHILDREN)
		parentNodes := []Node{NodeFromNodes(all[:splitpoint]), NodeFromNodes(all[splitpoint:])}
		return NodeFromNodes(parentNodes)
	}
}

func mergeLeaves(rope1 Node, rope2 Node) Node {
	if (!rope1.isLeaf() || !rope2.isLeaf()) { panic("Both parameters must be a Leaf") }

	bothOk := rope1.getLeaf().IsOkChild() && rope2.getLeaf().IsOkChild()
	if bothOk {
		return NodeFromNodes([]Node{rope1, rope2})
	} else {
		leaf1 := rope1.getLeaf()
		leaf2 := rope2.getLeaf()
		if leaf1.Len() +  leaf2.Len() <= MAX_LEAF {
			new := StringLeaf{leaf1.String() + leaf2.String()}
			return NodeFromLeaf(new)
		} else {
			space := MAX_LEAF - leaf1.Len()
			new1 := StringLeaf{leaf1.String() + leaf2.String()[:space]}
			new2 := StringLeaf{leaf2.String()[space:]}
			return NodeFromNodes([]Node{NodeFromLeaf(new1), NodeFromLeaf(new2)})
		}
	}
}

// Concatenates two ropes (strings)
func concat(rope1 Node, rope2 Node) Node {
	h1 := rope1.Height()
	h2 := rope2.Height()

	switch {
	case h1 < h2:
		children2 := rope2.getChildren()
		if h1 == h2 - 1 && rope1.isOkChild() {
			mergeNodes([]Node{rope1}, children2)
		}
		newrope := concat(rope1, children2[0])
		if newrope.Height() == h2 - 1 {
			return mergeNodes([]Node{newrope}, children2[1:])
		} else {
			return mergeNodes(newrope.getChildren(), children2[1:])
		}
	case h1 > h2:
		children1 := rope1.getChildren()
		if h2 == h1 - 1 && rope2.isOkChild() {
			mergeNodes([]Node{rope1}, children1)
		}
		lasti := len(children1) - 1
		newrope := concat(children1[lasti], rope2)
		if newrope.Height() == h1 - 1 {
			return mergeNodes(children1[:lasti], []Node{newrope})
		} else {
			return mergeNodes(children1[:lasti], newrope.getChildren())
		}
	default:
		if rope1.isOkChild() && rope2.isOkChild() {
			return NodeFromNodes([]Node{rope1, rope2})
		}
		if h1 == 0 {
			return mergeLeaves(rope1, rope2)
		}
		return mergeNodes(rope1.getChildren(), rope2.getChildren())
	}
}

func (n Node) slice(iv Interval) Node {
	panic("Not implemented")
}

func (n Node) edit(iv Interval, node Node) Node {
	panic("Not implemented")
}

