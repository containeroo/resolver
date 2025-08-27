package resolver

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// KeyValueFileResolver resolves a value by reading a key from a plain key=value text file.
// Format: "file:/path/file.txt//KEY" or "file:/path/file.txt" (entire file).
type KeyValueFileResolver struct{}

func (f *KeyValueFileResolver) Resolve(value string) (string, error) {
	filePath, keyPath := splitFileAndKey(value)
	filePath = os.ExpandEnv(filePath)

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
	return strings.TrimSpace(string(data)), nil
}

// searchKeyInFile searches for a specified key in a file and returns its associated value.
func searchKeyInFile(file *os.File, key string) (string, error) {
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Support "export KEY=VAL" and "KEY = VAL"
		if rest, ok := strings.CutPrefix(line, "export "); ok {
			line = strings.TrimSpace(rest)
		}
		pair := strings.SplitN(line, "=", 2)
		if len(pair) == 2 && strings.TrimSpace(pair[0]) == key {
			return strings.TrimSpace(pair[1]), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("failed scanning file %q: %w", file.Name(), err)
	}
	return "", fmt.Errorf("key %q not found in file %q", key, file.Name())
}
