package resolver

import "strings"

// looksLikeScheme recognizes an RFC 3986 scheme prefix at the beginning of a value.
// Colons in ordinary prose do not imply a scheme.
func looksLikeScheme(value string) bool {
	idx := strings.IndexByte(value, ':')
	if idx <= 0 {
		return false
	}
	// A colon followed by whitespace is prose, not a reference.
	if idx+1 < len(value) && (value[idx+1] == ' ' || value[idx+1] == '\t') {
		return false
	}
	for i := 0; i < idx; i++ {
		c := value[i]
		letter := c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
		if i == 0 {
			if !letter {
				return false
			}
			continue
		}
		if !letter && !(c >= '0' && c <= '9' || c == '+' || c == '-' || c == '.') {
			return false
		}
	}
	return true
}
