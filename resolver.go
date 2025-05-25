package resolver

import (
	"strings"
)

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
	filePrefix: &KeyValueFileResolver{},
	tomlPrefix: &TOMLResolver{},
}

// ResolveVariable attempts to resolve a variable string using a registered resolver.
// The function checks for known prefixes such as "env:", "json:", or "file:" and delegates
// to the corresponding resolver. If no known prefix is found, the original value is returned.
//
// Parameters:
//   - value: the input string, possibly prefixed with a resolver keyword.
//
// Returns:
//   - resolved string if resolved successfully or the original string if no prefix matches.
//   - error if resolution fails.
func ResolveVariable(value string) (string, error) {
	for prefix, resolver := range resolvers {
		if strings.HasPrefix(value, prefix) {
			return resolver.Resolve(strings.TrimPrefix(value, prefix))
		}
	}
	return value, nil
}
