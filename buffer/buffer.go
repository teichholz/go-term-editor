package Buffer

import (
	"io"
	BRope "main/rope3"
)

// undo: history of ropes
// redo: history of undone ropes
// multiple cursor

type Text interface {
	io.Reader
}

type Buffer[C Text] interface {
	Text
	Edit(iv BRope.Interval, rope C) C
	GetLine(y int) []rune
	OffsetOfLine(row int) (pos int)
	LineOfOffset(offset int) (row int)
	Length() int
	LineCount() int
	String() string
}

type ExtendedBuffer struct {
	Buffer[BRope.Rope]
}

func (buf ExtendedBuffer) LastCharInRow(row int) (col int) {
	line := buf.GetLine(row)
	col = len(line) - 1
	return
}

func (buf ExtendedBuffer) AppendChar(c rune) BRope.Rope {
	return buf.Buffer.Edit(BRope.IV(buf.Length(), buf.Length()), BRope.NewRope([]rune{c}))
}

func (buf ExtendedBuffer) InsertChar(row, col int, c rune) BRope.Rope {
	offsetUntilRow := buf.Buffer.OffsetOfLine(row)
	i := offsetUntilRow + col
	return buf.Buffer.Edit(BRope.IV(i, i), BRope.NewRope([]rune{c}))
}

func (buf ExtendedBuffer) DeleteAt(row, col int) BRope.Rope {
	offsetUntilRow := buf.Buffer.OffsetOfLine(row)
	i := offsetUntilRow + col
	return buf.Buffer.Edit(BRope.IV(i-1, i), BRope.NewRope([]rune{}))
}