package resolver

import (
	"fmt"
	"os"
	"strings"

	"github.com/containeroo/resolver/selector"
	"github.com/pelletier/go-toml/v2"
)

// TOMLResolver resolves a value by loading a TOML file and extracting a nested key.
// Format: "toml:/path/file.toml//key1.key2.keyN"
// If no key is provided, returns the entire TOML file as a string.
type TOMLResolver struct{}

func (r *TOMLResolver) Resolve(value string) (string, error) {
	filePath, keyPath := splitFileAndKey(value)
	filePath = os.ExpandEnv(filePath)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read TOML file %q: %w", filePath, err)
	}

	// Validate TOML syntax by decoding
	var validationTarget struct{}
	if err := toml.Unmarshal(data, &validationTarget); err != nil {
		return "", fmt.Errorf("failed to parse TOML in %q: %w", filePath, err)
	}

	// Decode into navigable structure
	var content map[string]any
	if err := toml.Unmarshal(data, &content); err != nil {
		return "", fmt.Errorf("failed to parse TOML in %q: %w", filePath, err)
	}

	if keyPath == "" {
		return strings.TrimSpace(string(data)), nil
	}

	val, err := selector.Navigate(content, selector.ParsePath(keyPath))
	if err != nil {
		return "", fmt.Errorf("key path %q not found in TOML %q: %w", keyPath, filePath, err)
	}

	if strVal, ok := val.(string); ok {
		return strVal, nil
	}

	tomlVal, err := toml.Marshal(val)
	if err != nil {
		return "", fmt.Errorf("failed to encode TOML value: %w", err)
	}

	return strings.TrimSpace(string(tomlVal)), nil
}
