package resolver

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// Resolves a value by reading a key from a plain key=value text file.
// The value after the prefix should be in the format "path/to/file.txt//Key"
// If no key is provided, returns the entire file as a string.
// Example:
// "file:/config/app.txt//USERNAME"
// would search for a line like "USERNAME = alice" and return "alice".
//
// Lines are matched by exact key name before the equals sign.
// If no key is provided (no "//" present), returns the entire file as string.
type KeyValueFileResolver struct{}

func (f *KeyValueFileResolver) Resolve(value string) (string, error) {
	filePath, keyPath := splitFileAndKey(value)
	filePath = os.ExpandEnv(filePath)

	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file '%s'. %v", filePath, err)
	}
	defer file.Close() // nolint:errcheck

	if keyPath != "" {
		return searchKeyInFile(file, keyPath)
	}

	// No key specified, read the whole file
	data, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file '%s'. %v", filePath, err)
	}
	return strings.TrimSpace(string(data)), nil
}

// searchKeyInFile searches for a specified key in a file and returns its associated value.
func searchKeyInFile(file *os.File, key string) (string, error) {
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		pair := strings.SplitN(line, "=", 2)
		if len(pair) == 2 && strings.TrimSpace(pair[0]) == key {
			return strings.TrimSpace(pair[1]), nil
		}
	}
	return "", fmt.Errorf("key '%s' not found in file '%s'", key, file.Name())
}
