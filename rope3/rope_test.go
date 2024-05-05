package BRope

import (
	"fmt"
	"runtime/debug"
	"testing"
	"unsafe"
)

func expectString(expected string, actual fmt.Stringer, t *testing.T) {
	if expected != actual.String() {
		t.Fatalf("expected '%v', got '%v'\nstacktrace: %s", expected, actual, debug.Stack())
	}
}

func expectRunes(expected []rune, actual []rune, t *testing.T) {
	if string(expected) != string(actual) {
		t.Fatalf("expected '%v', got '%v'\nstacktrace: %s", expected, actual, debug.Stack())
	}
}

func expectInt(a, b int, t *testing.T) {
	if a != b {
		t.Fatalf("expected %v, got %v\nstacktrace: %s", a, b, debug.Stack())
	}
}

func TestBigString(t * testing.T) {
	// TODO
}

func TestLineOffsets(t *testing.T) {
	rope1 := NewRopeString("foo\nbar\n")
	rope2 := NewRopeString("baz\nquux\n")
	rope := concat(rope1, rope2)

	expectInt(0, rope.OffsetOfLine(0), t)
	expectInt(0, rope.LineOfOffset(0), t)
	expectInt(4, rope.OffsetOfLine(1), t)
	expectInt(1, rope.LineOfOffset(4), t)
	expectInt(8, rope.OffsetOfLine(2), t)
	expectInt(2, rope.LineOfOffset(8), t)
	expectInt(12, rope.OffsetOfLine(3), t)
	expectInt(3, rope.LineOfOffset(12), t)

	rope3 := NewRopeString("")
	expectInt(0, rope3.OffsetOfLine(3), t)

	rope4 := NewRopeString("foo\n")
	expectInt(4, rope4.OffsetOfLine(1), t)
}

func TestLineOffsets2(t *testing.T) {
	rope := NewRopeString("\n\n\n")
	expectInt(0, rope.OffsetOfLine(0), t)
	expectInt(0, rope.LineOfOffset(0), t)
	expectInt(1, rope.OffsetOfLine(1), t)
	expectInt(1, rope.LineOfOffset(1), t)
	expectInt(2, rope.OffsetOfLine(2), t)
	expectInt(2, rope.LineOfOffset(2), t)
	expectInt(3, rope.OffsetOfLine(3), t)
	expectInt(2, rope.LineOfOffset(2), t)
}

func TestReplace(t *testing.T) {
	rope := NewRope([]rune("foobaz"))

	newrope := rope.Edit(Interval{3, 6}, NewRope([]rune("bar")))

	expectString("foobar", newrope, t)
}

func TestDelete(t *testing.T) {
	rope := NewRope([]rune("foooobar"))

	newrope := rope.Edit(Interval{1, 3}, EmptyRope())

	expectString("foobar", newrope, t)
}

func TestInsert(t *testing.T) {
	rope := NewRope([]rune("fbar"))
	toInsert := NewRope([]rune("oo"))

	newrope := rope.Edit(Interval{1, 1}, toInsert)

	expectString("foobar", newrope, t)
}

func TestGetLine(t *testing.T) {
	r := NewRopeString("\n\n\n")
	l := r.GetLine(0)
	expectRunes([]rune{}, l, t)
	l1 := r.GetLine(1)
	expectRunes([]rune{}, l1, t)
	l2 := r.GetLine(2)
	expectRunes([]rune{}, l2, t)
	l3 := r.GetLine(3)
	expectRunes([]rune{}, l3, t)
	// out of bounds
	l4 := r.GetLine(4)
	expectRunes([]rune{}, l4, t)
}

func TestGetLine2(t *testing.T) {
	r := NewRopeString("foo\nbar\nbaz")
	l := r.GetLine(0)
	expectRunes([]rune("foo"), l, t)
	l1 := r.GetLine(2)
	expectRunes([]rune("baz"), l1, t)
}

func TestWhatAreStrings(t *testing.T) {
	var a string = "a string"
	fmt.Print(a)
	var rs []rune = []rune{'h', 'a', 'l', 'l', 'o'}
	fmt.Println(rs)
	fmt.Println(string(rs))
	str2 := *(*string)(unsafe.Pointer(&rs))
	fmt.Println(str2)



	data := []byte("yellow submarine")
	str3 := *(*string)(unsafe.Pointer(&data))
	fmt.Println(str3)
	str4 := unsafe.String(&data[0], len(data))
	fmt.Println(str4)
	data[0] = 'g'
	fmt.Println(str4)

	// Strings are byte arrays, useful for efficient networking and byte sized operations
	// characters are runes, basically 4 byte unicode
	// conversions from []rune to string force a copy, because of the different data types / representations
}

func TestSlices(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	// increase capacity of slice
	slice2 := [][]int{slice}

	slice2[0] = append(slice, 6)
	fmt.Println(slice2)
	fmt.Println(slice)
}

func TestSlices2(t *testing.T) {
	slice := make([]int, 0, 10)
	slice = append(slice, 1)
	slice2 := [][]int{slice}
	slice2[0] = append(slice, 2)
	fmt.Println(slice2)
	fmt.Println(slice)
}

func TestSlices3(t *testing.T) {
	slice := make([]int, 0, 10)
	slice2 := [][]int{slice}
	tos := &slice2[0]
	*tos = append(*tos, 1)
	*tos = append(*tos, 2)
	fmt.Println(slice2)
	fmt.Println(*tos)
}

func TestSlices4(t *testing.T) {
	slice := make([]int, 0, 1)
	slice2 := [][]int{slice}
	tos := &slice2[0]
	*tos = append(*tos, 1)
	*tos = append(*tos, 2)
	*tos = append(*tos, 3)
	*tos = append(*tos, 4)
	*tos = append(*tos, 2)
	fmt.Println(slice2)
	fmt.Println(*tos)
}