package resolver

import (
	"testing"
)

func TestEnvResolver_Resolve(t *testing.T) {
	r := &EnvResolver{}

	t.Run("existing variable", func(t *testing.T) {
		t.Setenv("TEST_ENV_VAR", "test_value")
		got, err := r.Resolve("TEST_ENV_VAR")
		if err != nil {
			t.Fatalf("Resolve(TEST_ENV_VAR) unexpected error: %v", err)
		}
		if got != "test_value" {
			t.Fatalf("Resolve(TEST_ENV_VAR) = %q, want %q", got, "test_value")
		}
	})

	t.Run("empty variable", func(t *testing.T) {
		t.Setenv("EMPTY_ENV_VAR", "")
		got, err := r.Resolve("EMPTY_ENV_VAR")
		if err != nil {
			t.Fatalf("Resolve(EMPTY_ENV_VAR) unexpected error: %v", err)
		}
		if got != "" {
			t.Fatalf("Resolve(EMPTY_ENV_VAR) = %q, want empty string", got)
		}
	})

	t.Run("missing variable returns error", func(t *testing.T) {
		// do not set MISSING_ENV_VAR
		if _, err := r.Resolve("MISSING_ENV_VAR"); err == nil {
			t.Fatalf("Resolve(MISSING_ENV_VAR) expected error, got nil")
		}
	})
}

func TestDefaultRegistry_EnvScheme(t *testing.T) {
	t.Setenv("FOO", "bar")

	// Uses the package-level default registry via ResolveVariable.
	got, err := ResolveVariable("env:FOO")
	if err != nil {
		t.Fatalf("ResolveVariable(env:FOO) unexpected error: %v", err)
	}
	if got != "bar" {
		t.Fatalf("ResolveVariable(env:FOO) = %q, want %q", got, "bar")
	}
}

func TestDefaultRegistry_NoMatchingSchemePassThrough(t *testing.T) {
	// No prefix -> should return the input unchanged
	const in = "plain-literal"
	got, err := ResolveVariable(in)
	if err != nil {
		t.Fatalf("ResolveVariable(%q) unexpected error: %v", in, err)
	}
	if got != in {
		t.Fatalf("ResolveVariable(%q) = %q, want %q", in, got, in)
	}
}
