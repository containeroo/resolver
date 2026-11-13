package resolver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEscapedExpressionsAcrossPasses(t *testing.T) {
	t.Setenv("RESOLVER_ESC_TEST", "secret")
	r := NewDefaultRegistry()

	t.Run("escaped expression", func(t *testing.T) {
		t.Parallel()
		got, err := r.ResolveString(`literal \${env:RESOLVER_ESC_TEST}`)
		require.NoError(t, err)
		assert.Equal(t, `literal ${env:RESOLVER_ESC_TEST}`, got)
	})

	t.Run("escaped and active expressions", func(t *testing.T) {
		t.Parallel()
		got, err := r.ResolveString(`literal \${env:RESOLVER_ESC_TEST} and ${env:RESOLVER_ESC_TEST}`)
		require.NoError(t, err)
		assert.Equal(t, `literal ${env:RESOLVER_ESC_TEST} and secret`, got)
	})

	t.Run("escaped expression at depth limit", func(t *testing.T) {
		t.Parallel()
		got, err := r.resolveStringDepth(`\${env:RESOLVER_ESC_TEST} ${env:RESOLVER_ESC_TEST}`, 1)
		require.NoError(t, err)
		assert.Equal(t, `${env:RESOLVER_ESC_TEST} secret`, got)
	})

	t.Run("recursive expression exceeds depth", func(t *testing.T) {
		t.Parallel()
		r := NewRegistry()
		r.Register("loop:", ResolverFunc(func(v string) (string, error) { return "${loop:" + v + "}", nil }))
		_, err := r.ResolveString("${loop:x}")
		assert.ErrorIs(t, err, ErrBadPath)
	})
}
