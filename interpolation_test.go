package resolver

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveString_Basics(t *testing.T) {
	r := NewRegistry()
	// Simple stub for easy visibility.
	r.Register("x:", ResolverFunc(func(v string) (string, error) { return "X(" + v + ")", nil }))

	t.Run("No tokens -> unchanged", func(t *testing.T) {
		in := "hello world"
		got, err := r.ResolveString(in)
		require.NoError(t, err)
		assert.Equal(t, in, got)
	})

	t.Run("Single token", func(t *testing.T) {
		got, err := r.ResolveString("start ${x:foo} end")
		require.NoError(t, err)
		assert.Equal(t, "start X(foo) end", got)
	})

	t.Run("Multiple & adjacent tokens", func(t *testing.T) {
		got, err := r.ResolveString("${x:a}${x:b}-${x:c}")
		require.NoError(t, err)
		assert.Equal(t, "X(a)X(b)-X(c)", got)
	})

	t.Run("Literal dollar (not a token)", func(t *testing.T) {
		got, err := r.ResolveString("cost is $$5 or $5")
		require.NoError(t, err)
		assert.Equal(t, "cost is $$5 or $5", got)
	})
}

func TestResolveString_Escapes(t *testing.T) {
	r := NewRegistry()
	r.Register("env:", ResolverFunc(func(v string) (string, error) { return "ENV(" + v + ")", nil }))

	t.Run(`\${ stays literal and is NOT expanded`, func(t *testing.T) {
		got, err := r.ResolveString(`literal \${env:USER} here`)
		require.NoError(t, err)
		assert.Equal(t, `literal ${env:USER} here`, got)
	})

	t.Run(`backslash kept when not before ${`, func(t *testing.T) {
		got, err := r.ResolveString(`path C:\temp $x`)
		require.NoError(t, err)
		assert.Equal(t, `path C:\temp $x`, got)
	})
}

func TestResolveString_Malformed(t *testing.T) {
	r := NewRegistry()

	t.Run("Empty token ${}", func(t *testing.T) {
		_, err := r.ResolveString("oops ${} here")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrBadPath)
	})

	t.Run("Missing closing brace", func(t *testing.T) {
		_, err := r.ResolveString("oops ${x:foo here")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrBadPath)
	})
}

func TestResolveString_UnknownSchemePolicy(t *testing.T) {
	t.Run("PassThrough (default): unknown scheme passes through", func(t *testing.T) {
		r := NewRegistry()
		got, err := r.ResolveString("v=${nosuch:foo}")
		require.NoError(t, err)
		assert.Equal(t, "v=nosuch:foo", got)
	})

	t.Run("ErrorOnUnknown: unknown scheme errors", func(t *testing.T) {
		r := NewRegistry()
		r.SetUnknownSchemePolicy(ErrorOnUnknown)
		_, err := r.ResolveString("v=${nosuch:foo}")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrNotFound)
	})
}

func TestResolveString_MultiPassAndDepth(t *testing.T) {
	t.Run("Recursive expansion across passes", func(t *testing.T) {
		r := NewRegistry()
		// a:foo -> "A ${b:bar} Z"
		r.Register("a:", ResolverFunc(func(v string) (string, error) { return "A ${b:bar} Z", nil }))
		// b:bar -> "B"
		r.Register("b:", ResolverFunc(func(v string) (string, error) { return "B", nil }))

		got, err := r.ResolveString("X ${a:foo} Y")
		require.NoError(t, err)
		assert.Equal(t, "X A B Z Y", got)
	})

	t.Run("Depth exceeded (self-reintroducing tokens)", func(t *testing.T) {
		r := NewRegistry()
		r.Register("loop:", ResolverFunc(func(v string) (string, error) { return "${loop:" + v + "}", nil }))

		_, err := r.ResolveString("begin ${loop:x} end")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrBadPath)
	})
}

func TestResolveString_InternalDepthHelper(t *testing.T) {
	r := NewRegistry()
	// a:any -> "${b:x}"
	r.Register("a:", ResolverFunc(func(v string) (string, error) { return "${b:x}", nil }))
	// b:x -> "OK"
	r.Register("b:", ResolverFunc(func(v string) (string, error) { return "OK", nil }))

	in := "s=${a:foo}"

	t.Run("Depth=1 fails (needs 2 passes)", func(t *testing.T) {
		_, err := r.resolveStringDepth(in, 1)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrBadPath)
	})

	t.Run("Depth=2 succeeds", func(t *testing.T) {
		got, err := r.resolveStringDepth(in, 2)
		require.NoError(t, err)
		assert.Equal(t, "s=OK", got)
	})
}

func TestResolveString_ErrPropagationFromResolvers(t *testing.T) {
	r := NewRegistry()
	r.Register("fail:", ResolverFunc(func(v string) (string, error) { return "", errors.New("boom") }))

	_, err := r.ResolveString("x=${fail:now}")
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "resolve ${fail:now}:"), "should prefix resolver errors with token context")
}
