package resolver

import (
	"fmt"
	"strings"
	"sync"
)

// ResolverFunc adapts a plain function to the Resolver interface.
type ResolverFunc func(string) (string, error)

// Resolve implements Resolver by invoking the function.
func (f ResolverFunc) Resolve(v string) (string, error) { return f(v) }

// Resolver is implemented by all scheme resolvers; the input has the scheme stripped.
type Resolver interface {
	Resolve(string) (string, error)
}

// UnknownSchemePolicy controls how unknown scheme prefixes are handled.
type UnknownSchemePolicy int

const (
	// PassThrough returns the original value unchanged if the scheme is unknown.
	PassThrough UnknownSchemePolicy = iota
	// ErrorOnUnknown returns ErrNotFound for unknown-looking values (contain a ':').
	ErrorOnUnknown
)

// Scheme prefixes (include trailing colon so CutPrefix is unambiguous).
const (
	envPrefix  string = "env:"
	filePrefix string = "file:"
	iniPrefix  string = "ini:"
	jsonPrefix string = "json:"
	tomlPrefix string = "toml:"
	yamlPrefix string = "yaml:"
)

// Registry holds an ordered set of (scheme -> Resolver) mappings; it is concurrency-safe.
type Registry struct {
	mu      sync.RWMutex        // guards all fields below
	order   []string            // stable resolution order (schemes incl. trailing ':')
	backing map[string]Resolver // scheme -> resolver
	unknown UnknownSchemePolicy // policy for unknown schemes
}

// NewRegistry creates an empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		backing: make(map[string]Resolver),
	}
}

// NewDefaultRegistry returns a Registry with built-in resolvers pre-registered.
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

// Register adds or replaces a resolver for a scheme (e.g., "json:") and preserves order.
// Panics if scheme is empty or missing the trailing ":".
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

// SetUnknownSchemePolicy sets the policy for handling unknown scheme prefixes.
func (r *Registry) SetUnknownSchemePolicy(p UnknownSchemePolicy) {
	r.mu.Lock()
	r.unknown = p
	r.mu.Unlock()
}

// Schemes returns the registered schemes in resolution order.
func (r *Registry) Schemes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, len(r.order))
	copy(out, r.order)
	return out
}

// ResolveVariable resolves value using the first matching scheme; unknown handling is policy-driven.
func (r *Registry) ResolveVariable(value string) (string, error) {
	r.mu.RLock()
	for _, scheme := range r.order {
		if rest, ok := strings.CutPrefix(value, scheme); ok {
			res := r.backing[scheme]
			r.mu.RUnlock()
			return res.Resolve(rest)
		}
	}
	p := r.unknown
	r.mu.RUnlock()

	// If configured to be strict and the string looks like "scheme:...", treat as unknown.
	if p == ErrorOnUnknown && strings.Contains(value, ":") {
		return "", fmt.Errorf("%w: %q", ErrNotFound, value)
	}
	// Pass-through (back-compat behavior).
	return value, nil
}

// ResolveSlice resolves each value with the same rules as ResolveVariable (strict, fail-fast).
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

// ResolveSliceBestEffort resolves all values and returns outputs plus one error per failed index.
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
