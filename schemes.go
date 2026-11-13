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
		if !isASCIILetter(c) && !(c >= '0' && c <= '9' || c == '+' || c == '-' || c == '.') {
			return ""
		}
	}
	first, _ := utf8.DecodeRuneInString(rest)
	if unicode.IsSpace(first) {
		return ""
	}
	return name
}

func isASCIILetter(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}
