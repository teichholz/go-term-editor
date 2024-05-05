package BRope

import (
	"slices"
)

const (
	MIN_LEAF     = 511
	MAX_LEAF     = 1024
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
	Runes() []rune
	IsOkChild() bool
	Slice(Interval) Leaf
	Interval() Interval
	Copy() Leaf
}

type StringLeaf struct {
	vec []rune
}

func (s StringLeaf) Len() int {
	return len(s.vec)
}

func (s StringLeaf) Runes() []rune {
	return s.vec
}

func (s StringLeaf) IsOkChild() bool {
	return s.Len() >= MIN_LEAF
}

func (s StringLeaf) Slice(iv Interval) Leaf {
	return StringLeaf{s.vec[iv.Lo:iv.Hi]}
}

func (s StringLeaf) Copy() Leaf {
	return StringLeaf{slices.Clone(s.Runes())}
}

func (s StringLeaf) Interval() Interval {
	return Interval{0, s.Len()}
}

func NewRope(s []rune) Node {
	leaf := StringLeaf{s}
	return NodeFromLeaf(leaf)
}

func NewRopeString(s string) Node {
	leaf := StringLeaf{[]rune(s)}
	return NodeFromLeaf(leaf)
}

func EmptyRope() Node {
	return NodeFromLeaf(StringLeaf{[]rune{}})
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
	len int
	// amount of '\n' in a string
	newlines int
}

func (n *NodeInfo) compute(leaf Leaf) {
	n.len = leaf.Len()
	for _, char := range leaf.Runes() {
		if char == '\n' {
			n.newlines++
		}
	}
}

func (n *NodeInfo) accumulate(other NodeInfo) {
	n.len = n.len + other.len
	n.newlines = n.newlines + other.newlines
}

// Type of node values.
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

// Invariance: 1 <= len(nodes) <= MAX_CHILDREN
// Invariance: All nodes are same height
// Invariance: All the nodes must satisfy IsOkChild
func NodeFromNodes(nodes []Node) Node {
	if len(nodes) < 1 || len(nodes) > MAX_CHILDREN {
		panic("Invariance: 1 <= len(nodes) <= MAX_CHILDREN")
	}
	height := nodes[0].height
	len := nodes[0].len
	info := NodeInfo{}
	for _, n := range nodes[1:] {
		if height != n.height {
			panic("Invariance: All nodes are same height")
		}
		if !n.isLeaf() && !n.isOkChild() {
			panic("Invariance: All nodes must satisfy isOkChild")
		}
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
	all := slices.Concat(children1, children2)

	if compoundSize <= MAX_CHILDREN {
		return NodeFromNodes(all)
	} else {
		splitpoint := min(MAX_CHILDREN, compoundSize-MIN_CHILDREN)
		// TODO is this safe?
		parentNodes := []Node{NodeFromNodes(all[:splitpoint]), NodeFromNodes(all[splitpoint:])}
		return NodeFromNodes(parentNodes)
	}
}

func mergeLeaves(rope1 Node, rope2 Node) Node {
	if !rope1.isLeaf() || !rope2.isLeaf() {
		panic("Both parameters must be a Leaf")
	}

	bothOk := rope1.getLeaf().IsOkChild() && rope2.getLeaf().IsOkChild()
	if bothOk {
		return NodeFromNodes([]Node{rope1, rope2})
	} else {
		leaf1 := rope1.getLeaf()
		leaf2 := rope2.getLeaf()
		if leaf1.Len()+leaf2.Len() <= MAX_LEAF {
			// TODO currently always copy for safety. Later one could use context on write to also mutate in place, if only one referen to node
			newRunes := slices.Concat(slices.Clone(leaf1.Runes()), slices.Clone(leaf2.Runes()))
			newLeaf := StringLeaf{newRunes}
			return NodeFromLeaf(newLeaf)
		} else {
			space := MAX_LEAF - leaf1.Len()
			nv := slices.Concat(slices.Clone(leaf1.Runes()), slices.Clone(leaf2.Runes()[:space]))
			new1 := StringLeaf{nv}
			new2 := StringLeaf{slices.Clone(leaf2.Runes()[space:])}
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
		// recursion base
		if h1 == h2-1 && rope1.isOkChild() {
			mergeNodes([]Node{rope1}, children2)
		}
		newrope := concat(rope1, children2[0])
		if newrope.Height() == h2-1 {
			return mergeNodes([]Node{newrope}, children2[1:])
		} else {
			return mergeNodes(newrope.getChildren(), children2[1:])
		}
	case h1 > h2:
		children1 := rope1.getChildren()
		// recursion base
		if h2 == h1-1 && rope2.isOkChild() {
			mergeNodes([]Node{rope1}, children1)
		}
		lasti := len(children1) - 1
		newrope := concat(children1[lasti], rope2)
		if newrope.Height() == h1-1 {
			return mergeNodes(children1[:lasti], []Node{newrope})
		} else {
			return mergeNodes(children1[:lasti], newrope.getChildren())
		}
	// case h1 == h2
	default:
		// isOkChild checks for min size of leafs as well, so we prefer to merge leafs
		if rope1.isOkChild() && rope2.isOkChild() {
			return NodeFromNodes([]Node{rope1, rope2})
		}
		if h1 == 0 {
			return mergeLeaves(rope1, rope2)
		}
		return mergeNodes(rope1.getChildren(), rope2.getChildren())
	}
}

// slice or subseq
func (n Node) slice(iv Interval) Node {
	builder := NewTreeBuilder()
	builder.PushSlice(n, iv)
	return builder.Build()
}

func (n Node) Edit(iv Interval, toInsert Node) Node {
	b := NewTreeBuilder()
	selfIv := n.interval()
	prefix := selfIv.Prefix(iv)
	suffix := selfIv.Suffix(iv)
	b.PushSlice(n, prefix)
	b.Push(toInsert)
	b.PushSlice(n, suffix)
	return b.Build()
}

func (n Node) copy() Node {
	if n.isLeaf() {
		return NodeFromLeaf(n.getLeaf().Copy())
	}

	children := n.getChildren()
	copied := make([]Node, len(children))
	for i, child := range children {
		copied[i] = child.copy()
	}
	return NodeFromNodes(copied)
}

func (n Node) String() string {
	if n.isLeaf() {
		return string(n.getLeaf().Runes())
	}

	children := n.getChildren()
	str := ""
	for _, child := range children {
		str += child.String()
	}
	return str
}

// Unsafe operation, gives a mutable view of the line
func (n Node) runes() []rune {
	if n.isLeaf() {
		return n.getLeaf().Runes()
	}

	children := n.getChildren()
	rs := []rune{}
	for _, child := range children {
		rs = slices.Concat(rs, child.runes())
	}
	return rs
}

type Rope = Node

// Invariance (Inverse): OffsetOfLine(LineOfOffset(offset)) == offset
// line [0, inf)
func (r Rope) OffsetOfLine(line int) int {
	maxLine := r.NodeInfo.newlines + 1
	if line > maxLine {
		// maybe return error value if offset is not contained in the rope
		return r.NodeInfo.len
	}

	cur := r
	len := 0
	for cur.Height() > 0 {
		children := cur.getChildren()
		for _, child := range children {
			// note that lineCount == line is in the next leaf, since we start with 0
			// 2 '\n', lines in [0, 2], line 2 could be between this and the next leaf, but it does not matter, we know the offset
			if child.NodeInfo.newlines >= line {
				cur = child
				break
			}
			line -= child.NodeInfo.newlines
			len += child.Len()
		}
	}

	slice := cur.getLeaf().Runes()
	if line > 0 && slices.Index(slice, '\n') == -1 {
		// line is bigger than the amount of newlines in the leaf
		// maybe return error value if offset is not contained in the rope
		return len + cur.NodeInfo.len
	}

	for line > 0 {
		pos := slices.Index(slice, '\n')
		if (pos == -1) { panic("OffsetOfLine: Expected slice to contain newline") }
		len += pos+1 // there are pos character before the position
		line--
		slice = slice[pos+1:]
	}
	return len
}

// Invariance (Inverse): LineOfOffset(OffsetOfLine(offset)) == offset
// offset [0, inf)
func (r Rope) LineOfOffset(offset int) int {
	cur := r
	lines := 0

	for cur.Height() > 0 {
		children := r.getChildren()
		for _, child := range children {
			// len 2 = offset in [0, 1]
			if child.NodeInfo.len > offset {
				cur = child
				break;
			}
			lines += child.NodeInfo.newlines
			offset -= child.NodeInfo.len
		}
	}

	newlines := 0
	for _, char := range cur.getLeaf().Runes()[:offset] {
		if char == '\n' {
			newlines++
		}
	}

	return lines + newlines
}

func (r Rope) GetLine(line int) []rune {
	if (line < 0 || line > r.LineCount()) { panic("GetLine: line must be in [0, linecount]") }

	offset := r.OffsetOfLine(line)
	offset2 := r.OffsetOfLine(line+1)
	b := NewTreeBuilder()
	b.PushSlice(r, IV(offset, offset2))

	return b.Build().runes()
}

func (r Rope) Length() int {
	return r.NodeInfo.len
}

func (r Rope) LineCount() int {
	return r.NodeInfo.newlines + 1
}