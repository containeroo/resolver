package resolver

import (
	"strings"
)

// Resolver is the interface that different resolvers must implement.
// Each Resolver takes a value (with prefix already removed) and returns the resolved value or an error.
type Resolver interface {
	Resolve(value string) (string, error)
}

// Prefixes for different resolvers
const (
	envPrefix  string = "env:"
	filePrefix string = "file:"
	iniPrefix  string = "ini:"
	jsonPrefix string = "json:"
	tomlPrefix string = "toml:"
	yamlPrefix string = "yaml:"
)

// Global registry of resolvers
var resolvers = map[string]Resolver{
	envPrefix:  &EnvResolver{},
	jsonPrefix: &JSONResolver{},
	yamlPrefix: &YAMLResolver{},
	iniPrefix:  &INIResolver{},
	filePrefix: &INIResolver{},
	tomlPrefix: &TOMLResolver{},
}

// ResolveVariable attempts to resolve the given value by checking for known prefixes.
// If no known prefix is found, it returns the value as-is.
func ResolveVariable(value string) (string, error) {
	for prefix, resolver := range resolvers {
		if strings.HasPrefix(value, prefix) {
			return resolver.Resolve(strings.TrimPrefix(value, prefix))
		}
	}
	return value, nil
}
