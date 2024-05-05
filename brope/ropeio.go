package BRope

import (
	"io"
)

// Rope -> Buffer
// This implementation is probably not the most efficient, but suffices for now.
func (r Rope) Read(p []byte) (n int, err error) {
	bs := []byte(r.String())

	n  += copy(p, bs)
	err = io.EOF

	return
}

type RopeWriter struct {
	*Rope
}

// Buffer -> Rope
func (r *Rope) Write(p []byte) (n int, err error) {
	// Read from the rope into p.
	rs := []rune(string(p))

	if r.NodeBody == nil {
		*r = NewRope(rs)
	} else {
		*r = concat(*r, NewRope(rs))
	}

	// Read from the rope into p.
	return len(rs), nil
}

func (r *Rope) Close() error { return nil }
