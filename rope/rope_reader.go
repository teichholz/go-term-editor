package rope

import (
	"io"
)

// A RopeReader provides an implementation of io.RopeReader for ropes.
type RopeReader struct {
	rope     Rope
	position int64
}


// Read implements the standard Read interface:
// it reads data from the rope, populating p, and returns
// the number of bytes actually read.
func (reader *RopeReader) Read(p []byte) (n int, err error) {
	n, err = reader.rope.ReadAt(p, reader.position)
	if err == nil {
		reader.position += int64(n)
	}
	return
}

func (rope Rope) Read(p []byte) (n int, err error) {
	return rope.ReadAt(p, 0)
}

func (rope Rope) OffsetReader(offset int) *RopeReader {
	return &RopeReader{rope: rope, position: int64(offset)}
}

// ReadAt implements the standard ReadAt interface:
// it reads len(p) bytes from offset off into p, and returns
// the number of bytes actually read. If n < len(p), err will
// explain the shortfall.
func (rope Rope) ReadAt(p []byte, off int64) (n int, err error) {
	o := int(off)
	for n < len(p) && o+n < rope.Length() {
		leaf, at := rope.leafForOffset(o + n)
		n += copy(p[n:], leaf.content[at:])
	}

	if n < len(p) {
		err = io.EOF
	}

	return
}
