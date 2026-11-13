package resolver

import (
	"errors"
	"fmt"
	"io/fs"
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

	if strings.TrimSpace(filePath) == "" {
		return "", fmt.Errorf("%w: empty file path", ErrBadPath)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", fmt.Errorf("%w: %s", ErrNotFound, filePath)
		}
		if errors.Is(err, fs.ErrPermission) {
			return "", fmt.Errorf("%w: %s", ErrForbidden, filePath)
		}
		return "", fmt.Errorf("failed to read TOML file %q: %w", filePath, err)
	}

	var content map[string]any
	if err := toml.Unmarshal(data, &content); err != nil {
		return "", fmt.Errorf("failed to parse TOML in %q: %w", filePath, err)
	}

	if keyPath == "" {
		return strings.TrimSpace(string(data)), nil
	}

	val, err := selector.Navigate(content, selector.ParsePath(keyPath))
	if err != nil {
		return "", fmt.Errorf("%w: key path %q in TOML %q: %v", ErrNotFound, keyPath, filePath, err)
	}

	if strVal, ok := val.(string); ok {
		return strVal, nil
	}

	// Tables can stand alone as documents. Other values need an assignment
	// so the encoder can format arrays, numbers and dates as TOML literals.
	var encoded strings.Builder
	encoder := toml.NewEncoder(&encoded)
	_, table := val.(map[string]any)
	if !table {
		val = map[string]any{"value": val}
		encoder.SetTablesInline(true)
	}
	if err := encoder.Encode(val); err != nil {
		return "", fmt.Errorf("failed to encode TOML value: %w", err)
	}
	result := strings.TrimSpace(encoded.String())
	if !table {
		result = strings.TrimPrefix(result, "value = ")
	}
	return result, nil
}
