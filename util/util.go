package Util

import "unicode"


func IsWhitespace(r rune) bool {
	return unicode.IsSpace(r)
}