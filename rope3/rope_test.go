package BRope

import (
	"fmt"
	"testing"
	"unsafe"
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