package BRope

import (
	"testing"
)

func expectString(a, b string, t *testing.T) {
	if a != b {
		t.Fatalf("expected '%v', got '%v'", a, b)
	}
}

func expectInt(a, b int, t *testing.T) {
	if a != b {
		t.Fatalf("expected %v, got %v", a, b)
	}
}

/* func TestString(t *testing.T) {
	str := "Hallo, Welt!"
	ref := &str
	slice := (*ref)[0:5]

	t.Log(&str)
	t.Log(ref)
	t.Log(&slice)
	expectInt(1, 2, t)
} */