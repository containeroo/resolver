package resolver

import (
	"errors"
	"testing"
)

func TestEmptyValuePolicy(t *testing.T) {
	t.Setenv("RESOLVER_TEST_EMPTY", "")
	t.Setenv("RESOLVER_TEST_BLANK", " 	 ")
	t.Setenv("RESOLVER_TEST_SET", " abc ")

	tests := []struct {
		name   string
		policy EmptyValuePolicy
		input  string
		want   string
		reject bool
	}{
		{"default allows empty env", AllowEmpty, "env:RESOLVER_TEST_EMPTY", "", false},
		{"default allows empty literal", AllowEmpty, "", "", false},
		{"reject empty env", RejectEmpty, "env:RESOLVER_TEST_EMPTY", "", true},
		{"reject blank only not configured", RejectEmpty, "env:RESOLVER_TEST_BLANK", " 	 ", false},
		{"reject blank env", RejectBlank, "env:RESOLVER_TEST_BLANK", "", true},
		{"reject blank literal", RejectBlank, "   ", "", true},
		{"preserve whitespace on valid value", RejectBlank, "env:RESOLVER_TEST_SET", " abc ", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewDefaultRegistry()
			r.SetEmptyValuePolicy(tt.policy)
			got, err := r.ResolveVariable(tt.input)
			if tt.reject {
				if !errors.Is(err, ErrEmptyValue) {
					t.Fatalf("want ErrEmptyValue, got %v", err)
				}
				return
			}
			if err != nil || got != tt.want {
				t.Fatalf("ResolveVariable()=(%q,%v), want (%q,nil)", got, err, tt.want)
			}
		})
	}
	r := NewDefaultRegistry()
	r.SetEmptyValuePolicy(RejectBlank)
	if _, err := r.ResolveVariable("env:RESOLVER_TEST_MISSING_139"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing env: %v", err)
	}
}
