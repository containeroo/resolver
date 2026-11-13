package resolver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmptyValuePolicy(t *testing.T) {
	t.Setenv("RESOLVER_TEST_EMPTY", "")
	t.Setenv("RESOLVER_TEST_BLANK", " 	 ")
	t.Setenv("RESOLVER_TEST_SET", " abc ")

	t.Run("default allows empty env", func(t *testing.T) {
		t.Parallel()
		r := NewDefaultRegistry()
		r.SetEmptyValuePolicy(AllowEmpty)
		got, err := r.ResolveVariable("env:RESOLVER_TEST_EMPTY")
		require.NoError(t, err)
		assert.Equal(t, "", got)
	})

	t.Run("default allows empty literal", func(t *testing.T) {
		t.Parallel()
		r := NewDefaultRegistry()
		r.SetEmptyValuePolicy(AllowEmpty)
		got, err := r.ResolveVariable("")
		require.NoError(t, err)
		assert.Equal(t, "", got)
	})

	t.Run("reject empty env", func(t *testing.T) {
		t.Parallel()
		r := NewDefaultRegistry()
		r.SetEmptyValuePolicy(RejectEmpty)
		_, err := r.ResolveVariable("env:RESOLVER_TEST_EMPTY")
		assert.ErrorIs(t, err, ErrEmptyValue)
	})

	t.Run("reject blank only not configured", func(t *testing.T) {
		t.Parallel()
		r := NewDefaultRegistry()
		r.SetEmptyValuePolicy(RejectEmpty)
		got, err := r.ResolveVariable("env:RESOLVER_TEST_BLANK")
		require.NoError(t, err)
		assert.Equal(t, " 	 ", got)
	})

	t.Run("reject blank env", func(t *testing.T) {
		t.Parallel()
		r := NewDefaultRegistry()
		r.SetEmptyValuePolicy(RejectBlank)
		_, err := r.ResolveVariable("env:RESOLVER_TEST_BLANK")
		assert.ErrorIs(t, err, ErrEmptyValue)
	})

	t.Run("reject blank literal", func(t *testing.T) {
		t.Parallel()
		r := NewDefaultRegistry()
		r.SetEmptyValuePolicy(RejectBlank)
		_, err := r.ResolveVariable("   ")
		assert.ErrorIs(t, err, ErrEmptyValue)
	})

	t.Run("preserve whitespace on valid value", func(t *testing.T) {
		t.Parallel()
		r := NewDefaultRegistry()
		r.SetEmptyValuePolicy(RejectBlank)
		got, err := r.ResolveVariable("env:RESOLVER_TEST_SET")
		require.NoError(t, err)
		assert.Equal(t, " abc ", got)
	})

	t.Run("missing environment variable", func(t *testing.T) {
		t.Parallel()
		r := NewDefaultRegistry(WithEmptyValuePolicy(RejectBlank))
		_, err := r.ResolveVariable("env:RESOLVER_TEST_MISSING_139")
		assert.ErrorIs(t, err, ErrNotFound)
	})
}
