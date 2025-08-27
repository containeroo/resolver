package resolver

import (
	"fmt"
	"os"
)

// EnvResolver resolves a value from environment variables.
// Format: "env:MY_ENV_VAR"
type EnvResolver struct{}

func (r *EnvResolver) Resolve(value string) (string, error) {
	res, found := os.LookupEnv(value)
	if !found {
		return "", fmt.Errorf("environment variable %q not found", value)
	}
	return res, nil
}
