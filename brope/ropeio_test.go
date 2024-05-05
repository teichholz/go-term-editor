package BRope

import (
	"io"
	"testing"
)

func TestIO(t *testing.T) {
	r := NewRopeString("Hello, world!")
	var r2 Rope

	io.Copy(&r2, r)

	if r.String() != r2.String() {
		t.Errorf("Expected %s, got %s", r.String(), r2.String())
	}
}