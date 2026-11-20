package resolver

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// schemeName returns the RFC 3986 scheme at the start of value, or "" for a
// literal. Whitespace immediately after the colon identifies ordinary prose.
func schemeName(value string) string {
	name, rest, found := strings.Cut(value, ":")
	if !found || name == "" || !isASCIILetter(name[0]) {
		return ""
	}
	for i := 1; i < len(name); i++ {
		c := name[i]
		if !isASCIILetter(c) && !isASCIIDigit(c) && !isSchemePunctuation(c) {
			return ""
		}
	}
	first, _ := utf8.DecodeRuneInString(rest)
	if unicode.IsSpace(first) {
		return ""
	}
	return name
}

// isASCIILetter reports whether c is an ASCII letter.
func isASCIILetter(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

// isASCIIDigit reports whether c is an ASCII digit.
func isASCIIDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

// isSchemePunctuation reports whether c is '+', '-', or '.', as allowed in a URI scheme.
func isSchemePunctuation(c byte) bool {
	return c == '+' || c == '-' || c == '.'
}
