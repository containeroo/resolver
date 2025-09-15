package resolver

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"gopkg.in/ini.v1"
)

// INIResolver resolves a value by loading an INI file and extracting a section.key pair.
// Format: "ini:/path/file.ini//Section.Key" or "ini:/path/file.ini//Key" (default section).
// If no key is provided, returns the entire INI file as a string.
type INIResolver struct{}

func (r *INIResolver) Resolve(value string) (string, error) {
	filePath, keyPath := splitFileAndKey(value)
	filePath = os.ExpandEnv(filePath)

	cfg, err := ini.Load(filePath)
	if err != nil {
		// ini.Load wraps os.Open errors; try to map to sentinels.
		if errors.Is(err, fs.ErrNotExist) {
			return "", fmt.Errorf("%w: %s", ErrNotFound, filePath)
		}
		if errors.Is(err, fs.ErrPermission) {
			return "", fmt.Errorf("%w: %s", ErrForbidden, filePath)
		}
		return "", fmt.Errorf("failed to read INI file %q: %w", filePath, err)
	}

	if keyPath == "" {
		// No key path means return the entire INI file
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to read INI file %q: %w", filePath, err)
		}
		return strings.TrimSpace(string(data)), nil
	}

	// KeyPath can be "Section.Key" or just "Key" (default section)
	parts := strings.Split(keyPath, ".")
	var sectionName, keyName string
	if len(parts) == 1 {
		sectionName = "DEFAULT"
		keyName = parts[0]
	} else {
		sectionName = parts[0]
		keyName = strings.Join(parts[1:], ".")
	}
	if strings.TrimSpace(keyName) == "" {
		return "", fmt.Errorf("%w: empty key in %q", ErrBadPath, keyPath)
	}

	section, err := cfg.GetSection(sectionName)
	if err != nil {
		return "", fmt.Errorf("%w: section %q in %q", ErrNotFound, sectionName, filePath)
	}

	k, err := section.GetKey(keyName)
	if err != nil {
		return "", fmt.Errorf("%w: key %q in section %q of %q", ErrNotFound, keyName, sectionName, filePath)
	}
	return k.String(), nil
}
