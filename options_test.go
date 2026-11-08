package resolver

import (
	"errors"
	"testing"
)

func TestRegistryOptions(t *testing.T) {
	r := NewDefaultRegistry(WithEmptyValuePolicy(RejectBlank), WithUnknownSchemePolicy(ErrorOnUnknown), WithMaxDepth(2))
	if _, err := r.ResolveVariable("evn:VAR"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("strict: %v", err)
	}
	if _, err := r.ResolveVariable("literal:  "); !errors.Is(err, ErrEmptyValue) {
		t.Fatalf("empty: %v", err)
	}
	r.Register("a:", ResolverFunc(func(string) (string, error) { return "${b:x}", nil }))
	r.Register("b:", ResolverFunc(func(string) (string, error) { return "done", nil }))
	got, err := r.ResolveString("${a:x}")
	if err != nil || got != "done" {
		t.Fatalf("two passes: %q %v", got, err)
	}
	r2 := NewRegistry(WithMaxDepth(1))
	r2.Register("a:", ResolverFunc(func(string) (string, error) { return "${b:x}", nil }))
	r2.Register("b:", ResolverFunc(func(string) (string, error) { return "done", nil }))
	if _, err := r2.ResolveString("${a:x}"); !errors.Is(err, ErrBadPath) {
		t.Fatalf("depth: %v", err)
	}
	defer func() {
		if recover() == nil {
			t.Error("expected panic for invalid depth")
		}
	}()
	_ = NewRegistry(WithMaxDepth(0))
}
