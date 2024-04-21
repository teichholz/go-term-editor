// The rope package provides an immutable, value-oriented Rope.
// Ropes allow large sequences of text to be manipulated efficiently.
package rope

import (
	"bytes"
	"strings"
)

const (
	maxDepth    = 64
	maxLeafSize = 1024
)



// A Rope is a data structure for storing long runs of text.
// Ropes are persistent: there is no way to modify an existing rope.
// Instead, all operations return a new rope with the requested changes.
//
// This persistence makes it easy to store old versions of a Rope just by holding on to old roots.
type Rope struct {
	content                              string
	length, height                       int
	// 0-based line coverage of the rope
	LineStart, LineEnd                   int
	MaxLine, FirstBoundary, LastBoundary int

	left, right *Rope
}

// Return a new empty rope.
func New() Rope {
	return Rope{}
}

// Return a new rope with the contents of string s.
func NewString(s string) Rope {
	if strings.Contains(s, "\n") {
		maxLine, firstBoundary, lastBoundary := 0, 0, 0
		split := strings.Split(s, "\n")
		lines := len(split)
		for _, line := range split {
			maxLine = max(maxLine, len(line))
		}
		firstBoundary = len(split[0])
		lastBoundary = len(split[len(split)-1])
		return Rope{content: s, length: len(s), LineStart: 0, LineEnd: lines - 1, MaxLine: maxLine, FirstBoundary: firstBoundary, LastBoundary: lastBoundary}
	} else {
		return Rope{content: s, length: len(s)}
	}
}

// Notice that all of the methods take and return ropes by value.
// This is slightly less efficient than if we'd done pointers, but it
// seems cleaner from a "persistent data structure" point of view.
func (rope Rope) concat(other Rope) Rope {
	switch {
	case rope.length == 0:
		return other
	case other.length == 0:
		return rope
	case rope.length+other.length <= maxLeafSize:
		return NewString(rope.String() + other.String())
	default:
		height := rope.height
		if other.height > height {
			height = other.height
		}
		other.LineStart = rope.LineEnd + 1
		other.LineEnd = rope.LineEnd + other.LineEnd
		return Rope{
			length:    rope.length + other.length,
			LineStart: rope.LineStart,
			LineEnd:   other.LineEnd,
			height:    height + 1,
			left:      &rope,
			right:     &other,
		}
	}
}

// Return a new rope that is the concatenation of this rope and the other rope.
func (rope Rope) Append(other Rope) Rope {
	return rope.concat(other).rebalanceIfNeeded()
}

// Return a new rope that is the concatenation of this rope and string s.
func (rope Rope) AppendString(other string) Rope {
	return rope.Append(NewString(other))
}

// Return a new rope with length bytes at offset deleted.
func (rope Rope) Delete(offset, length int) Rope {
	if length == 0 || offset == rope.length {
		return rope
	}

	left, right := rope.Split(offset)
	_, newRight := right.Split(length)
	return left.Append(newRight)
}

// Returns true if this rope is equal to other.
func (rope Rope) Equal(other Rope) bool {
	if rope == other {
		return true
	}

	if rope.length != other.length {
		return false
	}

	for i := 0; i < rope.length; i += maxLeafSize {
		if !bytes.Equal(rope.Slice(i, i+maxLeafSize), other.Slice(i, i+maxLeafSize)) {
			return false
		}
	}

	return true
}

// Return a new rope with the contents of other inserted at the given index.
func (rope Rope) Insert(at int, other Rope) Rope {
	switch at {
	case 0:
		return other.Append(rope)
	case rope.length:
		return rope.Append(other)
	default:
		left, right := rope.Split(at)
		return left.concat(other).Append(right)
	}
}

// Return a new rope with the contents of string other inserted at the given index.
func (rope Rope) InsertString(at int, other string) Rope {
	return rope.Insert(at, NewString(other))
}

// Return the length of the rope in bytes.
func (rope Rope) Length() int {
	return rope.length
}

// Return a new version of this rope that is balanced for better performance.
// Generally speaking, this will be invoked automatically during the course of other operations and
// thus only needs to be called if you know you'll be generating a lot of unbalanced ropes.
func (rope Rope) Rebalance() Rope {
	if rope.isBalanced() {
		return rope
	}

	var leaves []Rope
	rope.walk(func(node Rope) {
		leaves = append(leaves, node)
	})

	return merge(leaves, 0, len(leaves))
}

// Return the bytes in [a, b)
func (rope Rope) Slice(a, b int) []byte {
	p := make([]byte, b-a)
	n, _ := rope.ReadAt(p, int64(a))
	return p[:n]
}

// Returns two new ropes, one containing the content to the left of the given index and the other the content to the right.
func (rope Rope) Split(at int) (Rope, Rope) {
	switch {
	case rope.isLeaf():
		return NewString(rope.content[0:at]), NewString(rope.content[at:])

	case at == 0:
		return Rope{}, rope

	case at == rope.length:
		return rope, Rope{}

	case at < rope.left.length:
		left, right := rope.left.Split(at)
		return left, right.Append(*rope.right)

	case at > rope.left.length:
		left, right := rope.right.Split(at - rope.left.length)
		return rope.left.Append(left), right

	default:
		return *rope.left, *rope.right
	}
}

// Return the contents of the rope as a string.
func (rope Rope) String() string {
	if rope.isLeaf() {
		return rope.content
	}

	var builder strings.Builder
	rope.walk(func(node Rope) {
		builder.WriteString(node.content)
	})

	return builder.String()
}

func (rope Rope) isBalanced() bool {
	switch {
	case rope.isLeaf():
		return true
	case rope.height >= len(fibonacci)-2:
		return false
	default:
		return fibonacci[rope.height+2] <= rope.length
	}
}

func (rope Rope) isLeaf() bool {
	return rope.left == nil
}

func (rope Rope) leafForOffset(at int) (Rope, int) {
	switch {
	case rope.isLeaf():
		return rope, at
	case at < rope.left.length:
		return rope.left.leafForOffset(at)
	default:
		return rope.right.leafForOffset(at - rope.left.length)
	}
}

func (rope Rope) rebalanceIfNeeded() Rope {
	if rope.isBalanced() || abs(rope.left.height-rope.right.height) < maxDepth {
		return rope
	}

	return rope.Rebalance()
}

func (rope Rope) hasBoundaries() bool {
	return rope.MaxLine > 0
}

func (rope Rope) walk(callback func(Rope)) {
	if rope.isLeaf() {
		callback(rope)
	} else {
		rope.left.walk(callback)
		rope.right.walk(callback)
	}
}

func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

func merge(leaves []Rope, start, end int) Rope {
	length := end - start
	switch length {
	case 1:
		return leaves[start]
	case 2:
		return leaves[start].concat(leaves[start+1])
	default:
		mid := start + length/2
		return merge(leaves, start, mid).concat(merge(leaves, mid, end))
	}
}

func (r Rope) WriteChar(pos int, c rune) Rope {
	return r.InsertString(pos, string(c))
}

func (r Rope) DeleteChar(pos int) Rope {
	return r.Delete(pos-1, 1)
}

func (r Rope) GetLine(y int) (line string) {
	if r.isLeaf() {
		if (y > r.LineEnd) { return "" }

		lines := strings.Split(r.String(), "\n")
		return lines[y-r.LineStart]
	} else {
		if y >= r.left.LineStart && y <= r.left.LineEnd {
			return r.left.GetLine(y)
		} else {
			return r.right.GetLine(y)
		}
	}
}

// OffsetOfLine returns the offset of the end of the given line.
// Note that the offset is 1 plus the position of the newline character.
func (r Rope) OffsetOfLine(row int) (pos int) {
	if row <= 0 { return r.LineStart }
	if row > r.LineEnd { return r.length }

	return r.internalOffsetOfLine(row, 1)
}

func (r Rope) internalOffsetOfLine(row, offset int) (pos int) {
	if r.isLeaf() {
		lines := strings.Split(r.String(), "\n")
		// if len(lines) == 1 { return 2 + offset  }
		return len(strings.Join(lines[:row], "\n")) + offset
	} else {
		if row >= r.left.LineStart && row <= r.left.LineEnd {
			return r.left.internalOffsetOfLine(row-r.left.LineStart, offset)
		} else {
			return r.right.internalOffsetOfLine(row-r.left.LineEnd, offset+r.left.length)
		}
	}
}

// LineOfOffset returns the line number of the given offset.
// Note that LineOfOffset is the inverse of OffsetOfLine and vice versa.
func (r Rope) LineOfOffset(offset int) (row int) {
	if offset <= 0 { return r.LineStart }
	if offset >= r.length { return r.LineEnd }

	return r.internalLineOfOffset(offset, 0)
}

func (r Rope) internalLineOfOffset(offset, accLines int) (row int) {
	if r.isLeaf() {
		lines := strings.Count(r.String()[:offset], "\n")
		return lines + accLines
	} else {
		if offset <= r.left.length {
			return r.left.internalLineOfOffset(offset, accLines)
		} else {
			return r.right.internalLineOfOffset(offset-r.left.length, accLines+r.left.LineEnd)
		}
	}
}

func (r Rope) LineCount() int {
	return r.LineEnd + 1
}

func (r Rope) Underlying() Rope {
	return r
}

func Map[T any](rope *Rope, f func(*Rope) T) chan T {
	ch := make(chan T)
	go traverse(rope, rope, ch, f)
	return ch
}

func traverse[T any](root *Rope, cur *Rope, ch chan T, f func(*Rope) T) {
	if cur.isLeaf() {
		ch <- f(cur)
	} else {
		traverse(root, cur.left, ch, f)
		traverse(root, cur.right, ch, f)
	}

	if root.LineEnd == cur.LineEnd {
		close(ch)
	}
}

type Reducer[T any, R any] func(R, T) R

func Reduce[T any, R any](ch chan T, acc Reducer[T, R], initial R) R {
	result := initial
	for v := range ch {
		result = acc(result, v)
	}
	return result
}

var fibonacci []int

func init() {
	// The heurstic for whether a rope is balanced depends on the Fibonacci sequence;
	// we initialize the table of Fibonacci numbers here.
	first := 0
	second := 1

	for c := 0; c < maxDepth+3; c++ {
		next := 0
		if c <= 1 {
			next = c
		} else {
			next = first + second
			first = second
			second = next
		}
		fibonacci = append(fibonacci, next)
	}
}
