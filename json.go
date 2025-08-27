package resolver

import (
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

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read JSON file %q: %w", filePath, err)
	}

	if keyPath == "" {
		return strings.TrimSpace(string(data)), nil
	}

	var content map[string]any
	if err := json.Unmarshal(data, &content); err != nil {
		return "", fmt.Errorf("failed to parse JSON in %q: %w", filePath, err)
	}

	val, err := selector.Navigate(content, selector.ParsePath(keyPath))
	if err != nil {
		return "", fmt.Errorf("key path %q not found in JSON %q: %w", keyPath, filePath, err)
	}

	if s, ok := val.(string); ok {
		return s, nil
	}
	jData, _ := json.Marshal(val)
	return string(jData), nil
}
