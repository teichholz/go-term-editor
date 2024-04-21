package Buffer

import (
	"io"
	"main/rope"
	Util "main/util"
)

// undo: history of ropes
// redo: history of undone ropes
// multiple cursor

type Text interface {
	io.Reader
}

type Buffer[C Text] interface {
	Text
	WriteChar(pos int, c rune) C
	OffsetOfLine(row int) (pos int)
	LineOfOffset(offset int) (row int)
	DeleteChar(pos int) C
	Length() int
	String() string
	GetLine(y int) string
	LineCount() int
}

type ExtendedBuffer struct {
	Buffer[rope.Rope]
}

func (buf ExtendedBuffer) LastNonWhitespaceChar(row int) int {
	line := buf.GetLine(row)
	for i := len(line) - 1; i >= 0; i-- {
		if rune(line[i]) != 0 && !Util.IsWhitespace(rune(line[i])) {
			return i
		}
	}

	return 0
}

func (buf ExtendedBuffer) AppendChar(c rune) rope.Rope {
	return buf.Buffer.WriteChar(buf.Buffer.Length(), c)
}

func (buf ExtendedBuffer) InsertChar(row, col int, c rune) rope.Rope {
	offsetUntilRow := buf.Buffer.OffsetOfLine(row)
	return buf.Buffer.WriteChar(offsetUntilRow+col, c)
}

func (buf ExtendedBuffer) DeleteAt(row, col int) rope.Rope {
	offsetUntilRow := buf.Buffer.OffsetOfLine(row)
	return buf.Buffer.DeleteChar(offsetUntilRow+col)
}