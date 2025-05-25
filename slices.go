package resolver

// ResolveSlice resolves a list of strings using ResolveVariable.
// This is a convenience wrapper that applies ResolveVariable to each element.
//
// Parameters:
//   - in: a slice of strings to resolve.
//
// Returns:
//   - slice of resolved strings.
//   - error if any individual resolution fails.
func ResolveSlice(in []string) ([]string, error) {
	return MapWithError(in, ResolveVariable)
}

// MapWithError applies a transformation function that may return an error to each item in a slice.
// The operation stops on the first error.
//
// Parameters:
//   - in: the input slice of any type T.
//   - fn: a function that maps each T to U and may return an error.
//
// Returns:
//   - a slice of type U containing all successfully transformed items.
//   - error if transformation fails for any item.
func MapWithError[T, U any](in []T, fn func(T) (U, error)) ([]U, error) {
	out := make([]U, 0, len(in))
	for _, item := range in {
		v, err := fn(item)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}
