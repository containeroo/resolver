package resolver

import (
	"fmt"
	"os"
)

// Resolves a value from environment variables.
// The value after the prefix should be the name of the environment variable.
// Example:
// "env:MY_ENV_VAR"
// would return the value of the MY_ENV_VAR environment variable.
//
// If the variable is not set, an error is returned.
type EnvResolver struct{}

func (r *EnvResolver) Resolve(value string) (string, error) {
	res, found := os.LookupEnv(value)
	if !found {
		return "", fmt.Errorf("environment variable '%s' not found", value)
	}
	return res, nil
}
