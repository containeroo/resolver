package resolver

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
)

// KeyValueFileResolver resolves a value by reading a key from a plain key=value text file.
// Format: "file:/path/file.txt//KEY" or "file:/path/file.txt" (entire file).
type KeyValueFileResolver struct{}

func (f *KeyValueFileResolver) Resolve(value string) (string, error) {
	filePath, keyPath := splitFileAndKey(value)
	filePath = os.ExpandEnv(filePath)

	if strings.TrimSpace(filePath) == "" {
		return "", fmt.Errorf("%w: empty file path", ErrBadPath)
	}
	if keyPath == "" && strings.HasSuffix(value, "//") {
		return "", fmt.Errorf("%w: empty key after // in %q", ErrBadPath, value)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file %q: %w", filePath, err)
	}
	defer file.Close() // nolint:errcheck

	if keyPath != "" {
		return searchKeyInFile(file, keyPath)
	}

	// No key specified, read the whole file
	data, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file %q: %w", filePath, err)
	}
	return strings.TrimSpace(stripBOM(string(data))), nil
}

// searchKeyInFile searches for a specified key in a file and returns its associated value.
func searchKeyInFile(file *os.File, key string) (string, error) {
	scanner := bufio.NewScanner(file)
	// Bump max token size to handle unusually long lines.
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		k, v, ok := parseKV(line)
		if !ok {
			continue
		}
		if k == key {
			return v, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed scanning file %q: %w", file.Name(), err)
	}
	return "", fmt.Errorf("%w: key %q in %q", ErrNotFound, key, file.Name())
}

// parseKV parses a single line of the form:
//
//	[export ]KEY = VALUE[# inline comment]
//
// It returns k, v and ok=true if a key/value was found. It supports:
//   - optional "export " prefix
//   - spaces around '='
//   - single/double quoted values (quotes are stripped)
//   - inline comments starting with an unquoted '#' that is preceded by whitespace
//     (e.g., `VALUE  # comment`). '#' inside quotes is preserved.
func parseKV(line string) (k, v string, ok bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", "", false
	}
	if rest, has := strings.CutPrefix(line, "export "); has {
		line = strings.TrimSpace(rest)
	}
	// Find first '='; key is left, value is right.
	eq := strings.IndexByte(line, '=')
	if eq < 0 {
		return "", "", false
	}
	k = strings.TrimSpace(line[:eq])
	if k == "" {
		return "", "", false
	}
	val := strings.TrimSpace(line[eq+1:])

	// Remove inline comments that start with an unquoted '#' with whitespace before it.
	val = cutInlineCommentUnquoted(val)

	// Strip surrounding quotes and unescape if double-quoted.
	if unq, okUnq := unquoteValue(val); okUnq {
		val = unq
	}
	return k, strings.TrimSpace(val), true
}

// cutInlineCommentUnquoted trims any trailing comment that begins with an unquoted '#' that
// is preceded by at least one whitespace character. '#' inside quotes is ignored.
func cutInlineCommentUnquoted(s string) string {
	inSingle, inDouble := false, false
	seenSpace := true // treat leading '#' as comment as well
	for i, r := range s {
		switch r {
		case '\'':
			if !inDouble {
				inSingle = !inSingle
			}
		case '"':
			if !inSingle {
				inDouble = !inDouble
			}
		case '#':
			if !inSingle && !inDouble && seenSpace {
				return strings.TrimSpace(s[:i])
			}
		}
		seenSpace = unicode.IsSpace(r)
	}
	return strings.TrimSpace(s)
}

// unquoteValue removes matching single or double quotes around s.
// For double quotes, it processes common escape sequences: \n \r \t \\ \" \'
// Returns (value, true) if quotes were stripped or (s, true) if not quoted.
// Returns (s, false) only if quotes are unmatched (should not happen with trimmed lines).
func unquoteValue(s string) (string, bool) {
	n := len(s)
	if n >= 2 && s[0] == '"' && s[n-1] == '"' {
		return unescapeDoubleQuoted(s[1 : n-1]), true
	}
	if n >= 2 && s[0] == '\'' && s[n-1] == '\'' {
		// Single-quoted: treat content mostly literally; unescape \' minimally.
		return strings.ReplaceAll(s[1:n-1], `\'`, `'`), true
	}
	return s, true
}

func unescapeDoubleQuoted(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	escape := false
	for _, r := range s {
		if !escape {
			if r == '\\' {
				escape = true
				continue
			}
			b.WriteRune(r)
			continue
		}
		switch r {
		case 'n':
			b.WriteByte('\n')
		case 'r':
			b.WriteByte('\r')
		case 't':
			b.WriteByte('\t')
		case '\\':
			b.WriteByte('\\')
		case '"':
			b.WriteByte('"')
		case '\'':
			b.WriteByte('\'')
		default:
			// Unknown escape: keep the character as-is.
			b.WriteRune(r)
		}
		escape = false
	}
	if escape {
		// Trailing backslash - keep it.
		b.WriteByte('\\')
	}
	return b.String()
}

// stripBOM removes a UTF-8 BOM if present.
func stripBOM(s string) string {
	const bom = "\uFEFF"
	return strings.TrimPrefix(s, bom)
}
