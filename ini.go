package resolver

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/ini.v1"
)

// Resolves a value by loading an INI file and extracting a section.key pair.
// The value after the prefix should be in the format "path/to/file.ini//Section.Key"
// If no section is provided, the default section is used.
// If no key is provided, returns the entire INI file as a string.
// Example:
// "ini:/config/app.ini//Database.User"
// would load app.ini, locate the [Database] section, and return the value of User.
//
// Keys are navigated via "Section.Key" notation.
// If no key is provided (no "//" present), returns the entire INI file as string.
type INIResolver struct{}

func (r *INIResolver) Resolve(value string) (string, error) {
	filePath, keyPath := splitFileAndKey(value)
	filePath = os.ExpandEnv(filePath)

	cfg, err := ini.Load(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read INI file '%s': %w", filePath, err)
	}

	if keyPath == "" {
		// No key path means return the entire INI file
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to read INI file '%s': %w", filePath, err)
		}
		return strings.TrimSpace(string(data)), nil
	}

	// KeyPath can be "Section.Key" or just "Key" (default section)
	parts := strings.Split(keyPath, ".")
	var sectionName, keyName string
	if len(parts) == 1 {
		// No explicit section, default section assumed
		sectionName = "DEFAULT"
		keyName = parts[0]
	} else {
		sectionName = parts[0]
		keyName = strings.Join(parts[1:], ".")
	}

	section, err := cfg.GetSection(sectionName)
	if err != nil {
		return "", fmt.Errorf("section '%s' not found in INI '%s': %w", sectionName, filePath, err)
	}

	k := section.Key(keyName)
	if k == nil || k.String() == "" {
		return "", fmt.Errorf("key '%s' not found in section '%s' of INI '%s'", keyName, sectionName, filePath)
	}

	return k.String(), nil
}
