package BRope

// cow.go implements copy-on-write semantics for the rope.

import (
	"sync"
)

const (
	DefaultFreeListSize = 32
)

var (
	nilNodes    = make([]Node, 16)
)

// FreeList represents a free list of btree nodes. By default each
// BTree has its own FreeList, but multiple BTrees can share the same
// FreeList.
// Two Btrees using the same freelist are safe for concurrent write access.
type FreeList struct {
	mu       sync.Mutex
	freelist []*Node
}

// NewFreeList creates a new free list.
// size is the maximum size of the returned free list.
func NewFreeList(size int) *FreeList {
	return &FreeList{freelist: make([]*Node, 0, size)}
}

func (f *FreeList) newNode() (n *Node) {
	f.mu.Lock()
	index := len(f.freelist) - 1
	if index < 0 {
		f.mu.Unlock()
		return new(Node)
	}
	n = f.freelist[index]
	f.freelist[index] = nil
	f.freelist = f.freelist[:index]
	f.mu.Unlock()
	return
}

// freeNode adds the given node to the list, returning true if it was added
// and false if it was discarded.
func (f *FreeList) freeNode(n *Node) (out bool) {
	f.mu.Lock()
	if len(f.freelist) < cap(f.freelist) {
		f.freelist = append(f.freelist, n)
		out = true
	}
	f.mu.Unlock()
	return
}

// copyOnWriteContext pointers determine node ownership... a tree with a write
// context equivalent to a node's write context is allowed to modify that node.
// A tree whose write context does not match a node's is not allowed to modify
// it, and must create a new, writable copy (IE: it's a Clone).
//
// When doing any write operation, we maintain the invariant that the current
// node's context is equal to the context of the tree that requested the write.
// We do this by, before we descend into any node, creating a copy with the
// correct context if the contexts don't match.
//
// Since the node we're currently visiting on any write has the requesting
// tree's context, that node is modifiable in place.  Children of that node may
// not share context, but before we descend into them, we'll make a mutable
// copy.
type copyOnWriteContext struct {
	freelist *FreeList
}

func (c *copyOnWriteContext) newNode() (n *Node) {
	n = c.freelist.newNode()
	n.cow = c
	return
}
func (n *Node) mutableFor(cow *copyOnWriteContext) *Node {
	if n.cow == cow {
		return n
	}

	if n.isLeaf() {
		newLeaf := n.getLeaf().Copy()
		newNode := NodeFromLeaf(newLeaf)
		newNode.cow = cow;
		return &newNode
	}

	out := cow.newNode()
	oi := out.NodeVal.(InternalNodeVal)
	ni := n.NodeVal.(InternalNodeVal)

	if cap(oi.nodes) >= len(ni.nodes) {
		oi.nodes = oi.nodes[:len(ni.nodes)]
	} else {
		oi.nodes = make(Nodes, len(ni.nodes), cap(ni.nodes))
	}
	copy(oi.nodes, ni.nodes)

	return out
}

type freeType int

const (
	ftFreelistFull freeType = iota // node was freed (available for GC, not stored in freelist)
	ftStored                       // node was stored in the freelist for later use
	ftNotOwned                     // node was ignored by COW, since it's owned by another one
)

// freeNode frees a node within a given COW context, if it's owned by that
// context.  It returns what happened to the node (see freeType const
// documentation).
func (c *copyOnWriteContext) freeNode(n *Node) freeType {
	if n.cow == c {
		// clear to allow GC
		switch n.NodeVal.(type) {
		case InternalNodeVal:
			i := n.NodeVal.(InternalNodeVal)
			(&i.nodes).truncate(0)
		case LeafNodeVal:
			l := n.NodeVal.(LeafNodeVal)
			switch l.Leaf.(type) {
			case StringLeaf:
				// no op
			default:
				panic("Unknown Leaf type")
			}
		default:
			panic("Unknown NodeVal type")
		}

		n.cow = nil
		if c.freelist.freeNode(n) {
			return ftStored
		} else {
			return ftFreelistFull
		}
	} else {
		return ftNotOwned
	}
}


func (s *Nodes) truncate(index int) {
	var toClear Nodes
	*s, toClear = (*s)[:index], (*s)[index:]
	for len(toClear) > 0 {
		toClear = toClear[copy(toClear, nilNodes):]
	}
}