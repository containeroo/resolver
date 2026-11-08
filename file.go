package resolver

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"
)

// KeyValueFileResolver resolves a value from a file, optionally selecting a key.
// Format: "file:/path/file.txt//KEY" or "file:/path/file.txt" (entire file).
// FileWhitespacePolicy controls whitespace handling when reading a whole file.
// Key/value lookups retain their existing parsing semantics.
type FileWhitespacePolicy uint8

const (
	// FileTrimSpace preserves the original behavior and trims all surrounding whitespace.
	FileTrimSpace FileWhitespacePolicy = iota
	// FileTrimNewlines removes only trailing CR/LF characters.
	FileTrimNewlines
	// FilePreserveWhitespace returns the full file contents (apart from a UTF-8 BOM).
	FilePreserveWhitespace
)

type KeyValueFileResolver struct {
	// Whitespace chooses how to normalize entire-file contents.
	Whitespace FileWhitespacePolicy
}

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
	data, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file %q: %w", filePath, err)
	}
	content := stripBOM(string(data))
	switch f.Whitespace {
	case FileTrimNewlines:
		return strings.TrimRight(content, "\r\n"), nil
	case FilePreserveWhitespace:
		return content, nil
	default:
		return strings.TrimSpace(content), nil
	}
}
func searchKeyInFile(file *os.File, key string) (string, error) {
	scanner := bufio.NewScanner(file)
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
func parseKV(line string) (k, v string, ok bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", "", false
	}
	if rest, has := strings.CutPrefix(line, "export "); has {
		line = strings.TrimSpace(rest)
	}
	eq := strings.IndexByte(line, '=')
	if eq < 0 {
		return "", "", false
	}
	k = strings.TrimSpace(line[:eq])
	if k == "" {
		return "", "", false
	}
	val := strings.TrimSpace(line[eq+1:])
	val = cutInlineCommentUnquoted(val)
	if unq, okUnq := unquoteValue(val); okUnq {
		val = unq
	}
	return k, strings.TrimSpace(val), true
}
func cutInlineCommentUnquoted(s string) string {
	inSingle, inDouble := false, false
	seenSpace := true
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
func unquoteValue(s string) (string, bool) {
	n := len(s)
	if n >= 2 && s[0] == '"' && s[n-1] == '"' {
		return unescapeDoubleQuoted(s[1 : n-1]), true
	}
	if n >= 2 && s[0] == '\'' && s[n-1] == '\'' {
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
			b.WriteRune(r)
		}
		escape = false
	}
	if escape {
		b.WriteByte('\\')
	}
	return b.String()
}
func stripBOM(s string) string { return strings.TrimPrefix(s, "\uFEFF") }
