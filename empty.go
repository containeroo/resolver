package resolver

import "strings"

// validateEmptyValue applies the registry's empty-value policy without exposing the value.
func validateEmptyValue(value string, p EmptyValuePolicy) (string, error) {
	switch p {
	case RejectEmpty:
		if value == "" {
			return "", ErrEmptyValue
		}
	case RejectBlank:
		if strings.TrimSpace(value) == "" {
			return "", ErrEmptyValue
		}
	}
	return value, nil
}
