package resolver

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/containeroo/resolver/selector"
	"gopkg.in/yaml.v3"
)

// YAMLResolver resolves a value by loading a YAML file and extracting a nested key.
// Format: "yaml:/path/file.yaml//key1.key2.keyN".
// If no key is provided, returns the whole YAML file as a string.
type YAMLResolver struct{}

func (r *YAMLResolver) Resolve(value string) (string, error) {
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
		return "", fmt.Errorf("failed to read YAML file %q: %w", filePath, err)
	}

	// Parse YAML into a generic structure (map[string]any / []any / scalars).
	var content any
	if err := yaml.Unmarshal(data, &content); err != nil {
		return "", fmt.Errorf("failed to parse YAML in %q: %w", filePath, err)
	}

	// Normalize to map[string]any at the root so selector can navigate uniformly.
	contentMap, err := convertToMapStringInterface(content)
	if err != nil {
		return "", fmt.Errorf("failed to process YAML %q: %w", filePath, err)
	}

	// No key → return the entire file (trimmed).
	if keyPath == "" {
		return strings.TrimSpace(string(data)), nil
	}

	// Bracket-aware path splitting (supports servers.[host=example.org].port).
	tokens := selector.ParsePath(keyPath)
	// Walk the structure using selector.
	val, err := selector.Navigate(contentMap, tokens)
	if err != nil {
		return "", fmt.Errorf("%w: key path %q in YAML %q: %v", ErrNotFound, keyPath, filePath, err)
	}

	// Strings are returned as-is; non-strings are re-encoded as YAML (trimmed).
	if s, ok := val.(string); ok {
		return s, nil
	}
	yData, _ := yaml.Marshal(val)
	return strings.TrimSpace(string(yData)), nil
}

// convertToMapStringInterface converts arbitrary YAML-parsed data into map[string]any at the root
// and recursively ensures maps/slices contain only map[string]any / []any / scalars.
func convertToMapStringInterface(val any) (map[string]any, error) {
	switch v := val.(type) {
	case map[string]any:
		for k, vv := range v {
			conv, err := convertValue(vv)
			if err != nil {
				return nil, err
			}
			v[k] = conv
		}
		return v, nil
	default:
		// If the root isn’t a map, return an empty map so navigation will fail cleanly.
		return map[string]any{}, nil
	}
}

func convertValue(val any) (any, error) {
	switch vv := val.(type) {
	case map[string]any:
		for k, v := range vv {
			conv, err := convertValue(v)
			if err != nil {
				return nil, err
			}
			vv[k] = conv
		}
		return vv, nil
	case []any:
		for i, elem := range vv {
			conv, err := convertValue(elem)
			if err != nil {
				return nil, err
			}
			vv[i] = conv
		}
		return vv, nil
	default:
		return vv, nil
	}
}
