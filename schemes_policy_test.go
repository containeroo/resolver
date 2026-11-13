package resolver

import (
	"errors"
	"testing"
)

func TestStrictSchemesAndLiteral(t *testing.T) {
	r := NewDefaultRegistry()
	r.SetUnknownSchemePolicy(ErrorOnUnknown)
	for _, input := range []string{"evn:TOKEN", "https://example.com", "my:secret", "X1+.-:", "a:b:c"} {
		_, err := r.ResolveVariable(input)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("%q: got %v, want ErrNotFound", input, err)
		}
	}
	for _, input := range []string{"message: after a space", "some text: here", ":strange", "123:abc", "a_b:c", "aé:b", "a:\nnext line", "a:\u00a0word"} {
		got, err := r.ResolveVariable(input)
		if err != nil || got != input {
			t.Errorf("%q: got (%q,%v), want pass-through", input, got, err)
		}
	}
	got, err := r.ResolveVariable("literal:https://example.com")
	if err != nil || got != "https://example.com" {
		t.Fatalf("literal: %q %v", got, err)
	}
	r = NewDefaultRegistry()
	got, err = r.ResolveVariable("unknown:secret")
	if err != nil || got != "unknown:secret" {
		t.Fatalf("backward compatible pass through: %q %v", got, err)
	}
}
