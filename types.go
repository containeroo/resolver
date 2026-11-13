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
	PassThrough UnknownSchemePolicy = iota
	ErrorOnUnknown
)

const (
	envPrefix     string = "env:"
	filePrefix    string = "file:"
	iniPrefix     string = "ini:"
	jsonPrefix    string = "json:"
	tomlPrefix    string = "toml:"
	yamlPrefix    string = "yaml:"
	literalPrefix string = "literal:"
)

// EmptyValuePolicy controls whether zero-length or whitespace-only values are accepted.
type EmptyValuePolicy uint8

const (
	// AllowEmpty keeps the legacy behavior and permits all values.
	AllowEmpty EmptyValuePolicy = iota
	// RejectEmpty rejects only zero-length values.
	RejectEmpty
	// RejectBlank rejects empty or whitespace-only values.
	RejectBlank
)

// Registry holds an ordered set of (scheme -> Resolver) mappings; it is concurrency-safe.
type Registry struct {
	mu       sync.RWMutex
	order    []string
	backing  map[string]Resolver
	unknown  UnknownSchemePolicy
	empty    EmptyValuePolicy
	maxDepth int
}

// NewRegistry creates an empty Registry.
func NewRegistry(options ...Option) *Registry {
	r := &Registry{backing: make(map[string]Resolver), maxDepth: 8}
	for _, option := range options {
		if option != nil {
			option(r)
		}
	}
	return r
}

// NewDefaultRegistry returns a Registry with built-in resolvers pre-registered.
func NewDefaultRegistry(options ...Option) *Registry {
	r := NewRegistry(options...)
	r.Register(envPrefix, &EnvResolver{})
	r.Register(jsonPrefix, &JSONResolver{})
	r.Register(yamlPrefix, &YAMLResolver{})
	r.Register(iniPrefix, &INIResolver{})
	r.Register(filePrefix, &KeyValueFileResolver{})
	r.Register(tomlPrefix, &TOMLResolver{})
	r.Register(literalPrefix, &LiteralResolver{})
	return r
}

// Register adds or replaces a resolver in the registry.
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

// SetUnknownSchemePolicy sets the policy for unknown scheme prefixes.
func (r *Registry) SetUnknownSchemePolicy(p UnknownSchemePolicy) {
	r.mu.Lock()
	r.unknown = p
	r.mu.Unlock()
}

// SetEmptyValuePolicy updates validation of all resolved and literal values.
func (r *Registry) SetEmptyValuePolicy(p EmptyValuePolicy) {
	r.mu.Lock()
	r.empty = p
	r.mu.Unlock()
}

// Schemes returns the registered schemes in registration order.
func (r *Registry) Schemes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, len(r.order))
	copy(out, r.order)
	return out
}

// ResolveVariable resolves value using the first matching scheme.
func (r *Registry) ResolveVariable(value string) (string, error) {
	r.mu.RLock()
	empty := r.empty
	for _, scheme := range r.order {
		if rest, ok := strings.CutPrefix(value, scheme); ok {
			res := r.backing[scheme]
			r.mu.RUnlock()
			resolved, err := res.Resolve(rest)
			if err != nil {
				return "", err
			}
			return validateEmptyValue(resolved, empty)
		}
	}
	p := r.unknown
	r.mu.RUnlock()
	if p == ErrorOnUnknown {
		if name := schemeName(value); name != "" {
			return "", fmt.Errorf("%w: unknown scheme %q", ErrNotFound, name)
		}
	}
	return validateEmptyValue(value, empty)
}

// ResolveSlice resolves every string or fails at the first error.
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

// ResolveSliceBestEffort resolves every string and returns per-index errors.
func (r *Registry) ResolveSliceBestEffort(values []string) (out []string, errs []error) {
	out = make([]string, len(values))
	errs = make([]error, 0, len(values))
	for i, v := range values {
		s, err := r.ResolveVariable(v)
		if err != nil {
			errs = append(errs, fmt.Errorf("index %d (%q): %w", i, v, err))
		}
		out[i] = s
	}
	return out, errs
}
