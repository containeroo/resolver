package resolver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/containeroo/resolver/selector"
)

// JSONResolver resolves a value by loading a JSON file and extracting a nested key.
// Format: "json:/path/file.json//key1.key2.keyN"
// If no key is provided, returns the whole JSON file as a string.
type JSONResolver struct{}

func (r *JSONResolver) Resolve(value string) (string, error) {
	filePath, keyPath := splitFileAndKey(value)
	filePath = os.ExpandEnv(filePath)

	if strings.TrimSpace(filePath) == "" {
		return "", fmt.Errorf("%w: empty file path", ErrBadPath)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", classifyFileError(fmt.Errorf("failed to read JSON file %q: %w", filePath, err))
	}

	if keyPath == "" {
		return strings.TrimSpace(string(data)), nil
	}

	var content map[string]any
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&content); err != nil {
		return "", fmt.Errorf("failed to parse JSON in %q: %w", filePath, err)
	}

	if len(bytes.Trim(data[decoder.InputOffset():], " \t\r\n")) != 0 {
		return "", fmt.Errorf("failed to parse JSON in %q: unexpected trailing data", filePath)
	}

	val, err := selector.Navigate(content, selector.ParsePath(keyPath))
	if err != nil {
		return "", fmt.Errorf("%w: key path %q in JSON %q: %v", ErrNotFound, keyPath, filePath, err)
	}

	if s, ok := val.(string); ok {
		return s, nil
	}
	jData, _ := json.Marshal(val)
	return string(jData), nil
}
