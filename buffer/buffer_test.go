package Buffer

import (
	BRope "main/brope"
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

func TestInsertChar(t *testing.T) {
	buffer := ExtendedBuffer{BRope.NewRopeString("foo")}
	buffer.Buffer = buffer.InsertChar(0, 3, '\n')
	expectString("foo\n", buffer.String(), t)
	buffer.Buffer = buffer.InsertChar(0, 4, 'b')
	expectString("foo\nb", buffer.String(), t)
	buffer.Buffer = buffer.InsertChar(1, 1, 'a')
	expectString("foo\nba", buffer.String(), t)
	buffer.Buffer = buffer.InsertChar(1, 2, 'r')
	expectString("foo\nbar", buffer.String(), t)
	buffer.Buffer = buffer.InsertChar(1, 3, '\n')
	expectString("foo\nbar\n", buffer.String(), t)
	buffer.Buffer = buffer.InsertChar(2, 0, 'b')
	expectString("foo\nbar\nb", buffer.String(), t)
}