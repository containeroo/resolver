package resolver

// LiteralResolver returns its input without interpreting any nested source prefix.
type LiteralResolver struct{}

// Resolve returns the original literal value.
func (LiteralResolver) Resolve(value string) (string, error) { return value, nil }
