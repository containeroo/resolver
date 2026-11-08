package resolver

// Option customizes a Registry at construction time.
type Option func(*Registry)

// WithEmptyValuePolicy controls whether resolved values may be empty.
func WithEmptyValuePolicy(p EmptyValuePolicy) Option {
	return func(r *Registry) { r.empty = p }
}

// WithUnknownSchemePolicy controls how unregistered schemes are handled.
func WithUnknownSchemePolicy(p UnknownSchemePolicy) Option {
	return func(r *Registry) { r.unknown = p }
}

// WithMaxDepth sets the maximum number of interpolation passes.
// A nonpositive limit is invalid and panics during construction.
func WithMaxDepth(n int) Option {
	if n <= 0 {
		panic("resolver: interpolation depth must be positive")
	}
	return func(r *Registry) { r.maxDepth = n }
}
