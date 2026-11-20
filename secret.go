package resolver

import (
	"fmt"
	"os"
	"strings"
)

// SecretResolver reads a whole mounted secret file, removing trailing CR/LF only.
// Spaces, tabs, and internal newlines are preserved. Paths support environment
// variable expansion. Key selectors and Kubernetes API access are not supported.
type SecretResolver struct{}

// Resolve reads the path supplied after the secret: prefix.
func (r *SecretResolver) Resolve(value string) (string, error) {
	filename := os.ExpandEnv(value)
	if strings.TrimSpace(filename) == "" {
		return "", fmt.Errorf("%w: empty secret file path", ErrBadPath)
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", classifyFileError(fmt.Errorf("failed to read secret file %q: %w", filename, err))
	}
	return strings.TrimRight(string(data), "\r\n"), nil
}
