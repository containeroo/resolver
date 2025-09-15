package resolver

// Package-level default registry and convenience functions.
// This preserves the original simple API while allowing advanced users
// to construct custom registries with NewRegistry/NewDefaultRegistry.
var defaultRegistry = NewDefaultRegistry()

// RegisterResolver adds or replaces a resolver in the default registry.
// scheme must include a trailing colon, e.g. "json:".
func RegisterResolver(scheme string, r Resolver) {
	defaultRegistry.Register(scheme, r)
}

// ResolveVariable attempts to resolve a variable string using a registered resolver
// from the default registry. If no known prefix is found, the original value is returned.
//
// Examples:
//
//	ResolveVariable("env:HOME")
//	ResolveVariable("json:/cfg/app.json//server.host")
//	ResolveVariable("yaml:${CONFIG}//servers.0.addr")
//	ResolveVariable("yaml:${CONFIG}//servers.[name=app].addr")
//	ResolveVariable("file:/etc/app.conf//USERNAME")
func ResolveVariable(value string) (string, error) {
	return defaultRegistry.ResolveVariable(value)
}

// ResolveSlice resolves each string in values using the default registry.
// It returns a new slice; the input is not modified. If any element fails
// to resolve, the function returns that error (strict mode).
func ResolveSlice(values []string) ([]string, error) {
	return defaultRegistry.ResolveSlice(values)
}

// ResolveSliceBestEffort resolves all values and returns the results plus a list of per-index errors.
// The output slice always has len(values). Callers can decide what to do with errors.
func ResolveSliceBestEffort(values []string) ([]string, []error) {
	return defaultRegistry.ResolveSliceBestEffort(values)
}

// ResolveString replaces ${...} tokens in s using the default registry.
func ResolveString(s string) (string, error) { return defaultRegistry.ResolveString(s) }

// DefaultRegistry returns the global default registry.
// Mutating it is safe for concurrent use.
func DefaultRegistry() *Registry {
	return defaultRegistry
}
