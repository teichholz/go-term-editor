package BRope

import "slices"

type Rope = Node

func (n Rope) String() string {
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

// Unsafe operation, gives a mutable view of the rope content
func (n Rope) runes() []rune {
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

// Invariance (Inverse): OffsetOfLine(LineOfOffset(offset)) == offset
// line [0, inf)
// \n \n \n \n \n
func (r Rope) OffsetOfLine(line int) int {
	maxLine := r.NodeInfo.newlines
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

// lines are separated by '\n', so '\n' is not part of the line
func (r Rope) GetLine(line int) []rune {
	if line < 0 || line > r.NodeInfo.newlines { return []rune{} }

	offset := r.OffsetOfLine(line)
	offset2 := r.OffsetOfLine(line+1)-1
	if line == r.NodeInfo.newlines { offset2 += 1 }
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

func (r Rope) LastCharInRow(row int) (col int) {
	line := r.GetLine(row)
	col = len(line) - 1
	return
}

func (r Rope) AppendChar(c rune) Rope {
	return r.Edit(IV(r.Length(), r.Length()), NewRope([]rune{c}))
}

func (r Rope) InsertChar(row, col int, c rune) Rope {
	offsetUntilRow := r.OffsetOfLine(row)
	i := offsetUntilRow + col
	return r.Edit(IV(i, i), NewRope([]rune{c}))
}

func (r Rope) DeleteAt(row, col int) Rope {
	offsetUntilRow := r.OffsetOfLine(row)
	i := offsetUntilRow + col
	return r.Edit(IV(i-1, i), NewRope([]rune{}))
}