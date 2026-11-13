package resolver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistryOptions(t *testing.T) {
	r := NewDefaultRegistry(WithEmptyValuePolicy(RejectBlank), WithUnknownSchemePolicy(ErrorOnUnknown), WithMaxDepth(2))

	t.Run("unknown scheme", func(t *testing.T) {
		t.Parallel()
		_, err := r.ResolveVariable("evn:VAR")
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("blank value", func(t *testing.T) {
		t.Parallel()
		_, err := r.ResolveVariable("literal:  ")
		assert.ErrorIs(t, err, ErrEmptyValue)
	})

	t.Run("two passes", func(t *testing.T) {
		t.Parallel()
		r.Register("a:", ResolverFunc(func(string) (string, error) { return "${b:x}", nil }))
		r.Register("b:", ResolverFunc(func(string) (string, error) { return "done", nil }))
		got, err := r.ResolveString("${a:x}")
		require.NoError(t, err)
		assert.Equal(t, "done", got)
	})

	t.Run("one pass", func(t *testing.T) {
		t.Parallel()
		r2 := NewRegistry(WithMaxDepth(1))
		r2.Register("a:", ResolverFunc(func(string) (string, error) { return "${b:x}", nil }))
		r2.Register("b:", ResolverFunc(func(string) (string, error) { return "done", nil }))
		_, err := r2.ResolveString("${a:x}")
		assert.ErrorIs(t, err, ErrBadPath)
	})

	t.Run("invalid depth", func(t *testing.T) {
		t.Parallel()
		assert.Panics(t, func() { NewRegistry(WithMaxDepth(0)) })
	})
}
