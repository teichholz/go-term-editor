package BRope

import (
	"io"
	"testing"
)

func TestIO(t *testing.T) {
	r := NewRopeString("Hello, world!")
	var r2 Rope
	rw := RopeWriter{&r2}

	io.Copy(rw, r)

	if r.String() != r2.String() {
		t.Errorf("Expected %s, got %s", r.String(), r2.String())
	}
}