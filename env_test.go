package resolver

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestEnvResolver_Resolve(t *testing.T) {
	t.Setenv("TEST_ENV_VAR", "test_value")
	t.Setenv("EMPTY_ENV_VAR", "")
	r := &EnvResolver{}

	t.Run("existing variable", func(t *testing.T) {
		t.Parallel()
		got, err := r.Resolve("TEST_ENV_VAR")
		require.NoError(t, err)
		assert.Equal(t, "test_value", got)
	})

	t.Run("empty variable", func(t *testing.T) {
		t.Parallel()
		got, err := r.Resolve("EMPTY_ENV_VAR")
		require.NoError(t, err)
		assert.Equal(t, "", got)
	})

	t.Run("missing variable returns error", func(t *testing.T) {
		t.Parallel()
		// do not set MISSING_ENV_VAR
		_, err := r.Resolve("MISSING_ENV_VAR")
		assert.ErrorIs(t, err, ErrNotFound)
	})
}

func TestDefaultRegistry_EnvScheme(t *testing.T) {
	t.Setenv("FOO", "bar")
	t.Run("resolves environment variable", func(t *testing.T) {
		t.Parallel()
		// Uses the package-level default registry via ResolveVariable.
		got, err := ResolveVariable("env:FOO")
		require.NoError(t, err)
		assert.Equal(t, "bar", got)
	})
}

func TestDefaultRegistry_NoMatchingSchemePassThrough(t *testing.T) {
	t.Run("passes literal through", func(t *testing.T) {
		t.Parallel()
		// No prefix -> should return the input unchanged
		const in = "plain-literal"
		got, err := ResolveVariable(in)
		require.NoError(t, err)
		assert.Equal(t, in, got)
	})
}
