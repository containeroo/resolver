package selector

import (
	"fmt"
	"strconv"
)

// Navigate walks through a nested structure of maps and arrays using path tokens.
// Each element of `keys` is one segment of the path, typically produced by ParsePath.
//
// Supported key forms:
//   - Map key: "server" → looks up curr["server"]
//   - Array index: "0" → takes the 0th element of a slice
//   - Array filter: "[field=value]" → selects the first element of a slice where elem[field]==value
//
// Example paths (split into tokens before calling Navigate):
//
//	servers.[name=app].host → ["servers", "[name=app]", "host"]
//	servers.0.host           → ["servers", "0", "host"]
func Navigate(data any, keys []string) (any, error) {
	current := data
	for _, k := range keys {
		switch curr := current.(type) {

		case map[string]any:
			// Map lookup: require string key
			val, ok := curr[k]
			if !ok {
				return nil, fmt.Errorf("key %q not found", k)
			}
			current = val

		case []any:
			// Array filter form: [key=value]
			if isFilterToken(k) {
				fk, fvRaw, err := parseFilterToken(k)
				if err != nil {
					return nil, err
				}
				want := coerce(fvRaw) // coerce value to bool/int/float if possible

				found := false
				for _, elem := range curr {
					m, ok := elem.(map[string]any)
					if !ok {
						continue // skip if element is not a map
					}
					got, ok := m[fk]
					if !ok {
						continue // field not present
					}
					// Compare with coercion-aware equality
					if equalCoerced(got, want) {
						current = elem
						found = true
						break
					}
				}
				if !found {
					return nil, fmt.Errorf("no array element where %s=%v", fk, want)
				}
				continue // move to next key segment
			}

			// Array index form: must be parseable integer
			idx, err := strconv.Atoi(k)
			if err != nil {
				return nil, fmt.Errorf("%q is not a valid array index or filter", k)
			}
			if idx < 0 || idx >= len(curr) {
				return nil, fmt.Errorf("array index %d out of bounds", idx)
			}
			current = curr[idx]

		default:
			// Neither a map nor a slice → cannot descend further
			return nil, fmt.Errorf("path segment %q not found", k)
		}
	}
	return current, nil
}
