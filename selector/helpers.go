package selector

import (
	"fmt"
	"strconv"
	"strings"
)

// ParsePath splits a dotted path expression into tokens for Navigate.
// It treats dots ('.') as separators unless they occur inside a bracketed filter "[..]".
//
// Examples:
//
//	"server.host"                  → ["server", "host"]
//	"servers.0.host"               → ["servers", "0", "host"]
//	"servers.[name=example.org].ip" → ["servers", "[name=example.org]", "ip"]
//
// This allows array filters and nested fields to coexist without breaking on dots
// inside the filter expression.
func ParsePath(s string) []string {
	var out []string
	var buf []rune
	depth := 0 // bracket nesting depth

	for _, r := range s {
		switch r {
		case '[':
			depth++              // entering filter → disable splitting on dots
			buf = append(buf, r) // keep the rune
		case ']':
			if depth > 0 {
				depth-- // leaving filter
			}
			buf = append(buf, r)
		case '.':
			if depth == 0 {
				// split on dot only if not inside filter brackets
				out = append(out, string(buf))
				buf = buf[:0]
				continue
			}
			// inside filter → keep dot literal
			buf = append(buf, r)
		default:
			// normal character
			buf = append(buf, r)
		}
	}
	// flush the last token
	out = append(out, string(buf))
	return out
}

// isFilterToken reports whether tok looks like [key=value] (optional quotes around value).
func isFilterToken(tok string) bool {
	return strings.HasPrefix(tok, "[") && strings.HasSuffix(tok, "]") && strings.Contains(tok, "=")
}

// parseFilterToken parses [key=value] and returns key, value (unquoted).
func parseFilterToken(tok string) (string, string, error) {
	inner := strings.TrimSuffix(strings.TrimPrefix(tok, "["), "]")
	kv := strings.SplitN(inner, "=", 2)
	if len(kv) != 2 {
		return "", "", fmt.Errorf("invalid filter token %q", tok)
	}
	key := strings.TrimSpace(kv[0])
	val := strings.TrimSpace(kv[1])
	// Strip optional quotes
	if (strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"")) ||
		(strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'")) {
		val = strings.Trim(val, "\"'")
	}
	if key == "" {
		return "", "", fmt.Errorf("empty key in filter %q", tok)
	}
	return key, val, nil
}

// coerce tries int, float, then explicit bool ("true"/"false"); otherwise returns the raw string.
// Important: do NOT treat "1"/"0" as booleans, so numeric IDs match correctly.
func coerce(val string) any {
	if i, err := strconv.Atoi(val); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(val, 64); err == nil {
		return f
	}
	switch strings.ToLower(val) {
	case "true", "false":
		b, _ := strconv.ParseBool(val)
		return b
	}
	return val
}

// equalCoerced compares v (from YAML/JSON) with want (already coerced).
func equalCoerced(v any, want any) bool {
	switch w := want.(type) {
	case bool:
		if vb, ok := v.(bool); ok {
			return vb == w
		}
	case int:
		switch vv := v.(type) {
		case int:
			return vv == w
		case int64:
			return vv == int64(w)
		case float64:
			return int(vv) == w && float64(int(vv)) == vv
		}
	case float64:
		if vf, ok := v.(float64); ok {
			return vf == w
		}
	case string:
		if vs, ok := v.(string); ok {
			return vs == w
		}
	}
	// last resort: string compare
	return fmt.Sprint(v) == fmt.Sprint(want)
}
