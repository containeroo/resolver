package resolver

import (
	"fmt"
	"os"
	"strings"
)

// EnvResolver resolves values from environment variables.
// Format: "env:MY_ENV_VAR".
type EnvResolver struct{}

// Resolve returns the environment variable value or a typed error (ErrBadPath / ErrNotFound).
func (r *EnvResolver) Resolve(value string) (string, error) {
	v := strings.TrimSpace(value)
	if v == "" {
		return "", fmt.Errorf("%w: empty environment variable name", ErrBadPath)
	}
	res, found := os.LookupEnv(v)
	if !found {
		return "", fmt.Errorf("%w: env %q", ErrNotFound, v)
	}
	return res, nil
}
