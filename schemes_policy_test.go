package resolver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStrictSchemesAndLiteral(t *testing.T) {
	r := NewDefaultRegistry()
	r.SetUnknownSchemePolicy(ErrorOnUnknown)

	t.Run("misspelled scheme", func(t *testing.T) {
		t.Parallel()
		_, err := r.ResolveVariable("evn:TOKEN")
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("unknown URL scheme", func(t *testing.T) {
		t.Parallel()
		_, err := r.ResolveVariable("https://example.com")
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("unknown short scheme", func(t *testing.T) {
		t.Parallel()
		_, err := r.ResolveVariable("my:secret")
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("valid scheme punctuation", func(t *testing.T) {
		t.Parallel()
		_, err := r.ResolveVariable("X1+.-:")
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("multiple colons", func(t *testing.T) {
		t.Parallel()
		_, err := r.ResolveVariable("a:b:c")
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("space after colon", func(t *testing.T) {
		t.Parallel()
		const input = "message: after a space"
		got, err := r.ResolveVariable(input)
		require.NoError(t, err)
		assert.Equal(t, input, got)
	})

	t.Run("prose before colon", func(t *testing.T) {
		t.Parallel()
		const input = "some text: here"
		got, err := r.ResolveVariable(input)
		require.NoError(t, err)
		assert.Equal(t, input, got)
	})

	t.Run("leading colon", func(t *testing.T) {
		t.Parallel()
		const input = ":strange"
		got, err := r.ResolveVariable(input)
		require.NoError(t, err)
		assert.Equal(t, input, got)
	})

	t.Run("leading digit", func(t *testing.T) {
		t.Parallel()
		const input = "123:abc"
		got, err := r.ResolveVariable(input)
		require.NoError(t, err)
		assert.Equal(t, input, got)
	})

	t.Run("underscore", func(t *testing.T) {
		t.Parallel()
		const input = "a_b:c"
		got, err := r.ResolveVariable(input)
		require.NoError(t, err)
		assert.Equal(t, input, got)
	})

	t.Run("non-ASCII letter", func(t *testing.T) {
		t.Parallel()
		const input = "aé:b"
		got, err := r.ResolveVariable(input)
		require.NoError(t, err)
		assert.Equal(t, input, got)
	})

	t.Run("newline after colon", func(t *testing.T) {
		t.Parallel()
		const input = "a:\nnext line"
		got, err := r.ResolveVariable(input)
		require.NoError(t, err)
		assert.Equal(t, input, got)
	})

	t.Run("nonbreaking space after colon", func(t *testing.T) {
		t.Parallel()
		const input = "a:\u00a0word"
		got, err := r.ResolveVariable(input)
		require.NoError(t, err)
		assert.Equal(t, input, got)
	})

	t.Run("literal scheme", func(t *testing.T) {
		t.Parallel()
		got, err := r.ResolveVariable("literal:https://example.com")
		require.NoError(t, err)
		assert.Equal(t, "https://example.com", got)
	})

	t.Run("default passes unknown scheme through", func(t *testing.T) {
		t.Parallel()
		got, err := NewDefaultRegistry().ResolveVariable("unknown:secret")
		require.NoError(t, err)
		assert.Equal(t, "unknown:secret", got)
	})
}
