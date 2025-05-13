package resolver

import (
	"fmt"
	"os"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// Resolves a value by loading a TOML file and extracting a nested key.
// The value after the prefix should be in the format "path/to/file.toml//key1.key2.keyN"
// If no key is provided, returns the entire TOML file as a string.
// Example:
// "toml:/config/app.toml//server.host"
// would load app.toml, parse it as TOML, and then return the value at server.host.
//
// Keys are navigated via dot notation.
// If no key is provided (no "//" present), returns the entire TOML file as string.
type TOMLResolver struct{}

func (r *TOMLResolver) Resolve(value string) (string, error) {
	filePath, keyPath := splitFileAndKey(value)
	filePath = os.ExpandEnv(filePath)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read TOML file '%s': %w", filePath, err)
	}

	// Validate TOML syntax by decoding into a dummy struct
	var validationTarget struct{}
	if err := toml.Unmarshal(data, &validationTarget); err != nil {
		return "", fmt.Errorf("failed to parse TOML in '%s': %w", filePath, err)
	}

	// Decode into navigable structure
	var content map[string]any
	if err := toml.Unmarshal(data, &content); err != nil {
		return "", fmt.Errorf("failed to parse TOML in '%s': %w", filePath, err)
	}

	if keyPath == "" {
		return strings.TrimSpace(string(data)), nil
	}

	val, err := navigateData(content, strings.Split(keyPath, "."))
	if err != nil {
		return "", fmt.Errorf("key path '%s' not found in TOML '%s': %w", keyPath, filePath, err)
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
