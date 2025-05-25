package resolver

// Resolver is the interface that different resolvers must implement.
// Each Resolver takes a value (with prefix already removed) and returns the resolved value or an error.
type Resolver interface {
	Resolve(value string) (string, error)
}
