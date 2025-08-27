package resolver

import "strings"

// splitFileAndKey splits a value by "//" to separate file path and key path.
func splitFileAndKey(value string) (string, string) {
	const keyDelim = "//"
	idx := strings.LastIndex(value, keyDelim)
	if idx == -1 {
		return value, ""
	}
	return value[:idx], value[idx+len(keyDelim):]
}
