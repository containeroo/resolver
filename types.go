package resolver

import (
	"fmt"
	"strings"
	"sync"
)

// Resolver is the interface that different resolvers must implement.
// Each Resolver takes a value (with prefix already removed) and returns the resolved value or an error.
type Resolver interface {
	Resolve(value string) (string, error)
}

// Scheme prefixes (include trailing colon to make CutPrefix unambiguous).
const (
	envPrefix  string = "env:"
	filePrefix string = "file:"
	iniPrefix  string = "ini:"
	jsonPrefix string = "json:"
	tomlPrefix string = "toml:"
	yamlPrefix string = "yaml:"
)

// Registry holds an ordered set of (scheme -> Resolver) mappings.
// It is safe for concurrent use.
type Registry struct {
	mu      sync.RWMutex
	order   []string            // stable resolution order (schemes with trailing colon)
	backing map[string]Resolver // scheme -> resolver
}

// NewRegistry creates an empty registry.
func NewRegistry() *Registry {
	return &Registry{
		backing: make(map[string]Resolver),
	}
}

// NewDefaultRegistry returns a registry pre-populated with the built-in resolvers,
// in a stable, sensible order.
func NewDefaultRegistry() *Registry {
	r := NewRegistry()
	r.Register(envPrefix, &EnvResolver{})
	r.Register(jsonPrefix, &JSONResolver{})
	r.Register(yamlPrefix, &YAMLResolver{})
	r.Register(iniPrefix, &INIResolver{})
	r.Register(filePrefix, &KeyValueFileResolver{})
	r.Register(tomlPrefix, &TOMLResolver{})
	return r
}

// Register adds or replaces a resolver for a given scheme (e.g., "json:").
// If the scheme is new, it is appended to the end of the resolution order.
// Panics if scheme is empty or does not end with ":".
func (r *Registry) Register(scheme string, res Resolver) {
	if scheme == "" || !strings.HasSuffix(scheme, ":") {
		panic(fmt.Sprintf("resolver: scheme %q must end with colon", scheme))
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.backing[scheme]; !exists {
		r.order = append(r.order, scheme)
	}
	r.backing[scheme] = res
}

// Schemes returns a copy of the registered schemes in resolution order.
func (r *Registry) Schemes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, len(r.order))
	copy(out, r.order)
	return out
}

// ResolveVariable attempts to resolve `value` using the first matching scheme.
// If no known scheme prefix matches, ResolveVariable returns the input unchanged.
func (r *Registry) ResolveVariable(value string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, scheme := range r.order {
		if rest, ok := strings.CutPrefix(value, scheme); ok {
			resolver := r.backing[scheme]
			return resolver.Resolve(rest)
		}
	}
	// No scheme matched: return as-is (back-compat behavior).
	return value, nil
}

// ResolveSlice resolves each value with the registry using the same rules as
// ResolveVariable: known scheme => delegate; unknown scheme => return as-is.
// Returns a new slice in the same order. Fails fast on the first error.
func (r *Registry) ResolveSlice(values []string) ([]string, error) {
	out := make([]string, len(values))
	for i, v := range values {
		s, e := r.ResolveVariable(v)
		if e != nil {
			return nil, fmt.Errorf("resolve slice index %d (%q): %w", i, v, e)
		}
		out[i] = s
	}
	return out, nil
}

// ResolveSliceBestEffort resolves all values and returns the results plus a list of per-index errors.
// The output slice always has len(values). Callers can decide what to do with errors.
func (r *Registry) ResolveSliceBestEffort(values []string) (out []string, errs []error) {
	out = make([]string, len(values))
	errs = make([]error, 0, len(values)) // len 0, cap N

	for i, v := range values {
		s, err := r.ResolveVariable(v)
		if err != nil {
			errs = append(errs, fmt.Errorf("index %d (%q): %w", i, v, err))
		}
		out[i] = s // "" on error, pass-through or resolved on success
	}
	return out, errs
}
