package resolver

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubResolver struct {
	out  string
	err  error
	last string
}

func (s *stubResolver) Resolve(v string) (string, error) {
	s.last = v
	if s.err != nil {
		return "", s.err
	}
	if s.out != "" {
		return s.out, nil
	}
	return "stub:" + v, nil
}

func TestResolveVariable(t *testing.T) {
	t.Run("PassThroug no prefix", func(t *testing.T) {
		const in = "just-a-literal"
		got, err := ResolveVariable(in)
		require.NoError(t, err)
		assert.Equal(t, in, got)
	})

	t.Run("Unknown scheme", func(t *testing.T) {
		const in = "unknown:foo"
		got, err := ResolveVariable(in)
		require.NoError(t, err)
		// No registered "unknown:" scheme → unchanged
		assert.Equal(t, in, got)
	})

	t.Run("Env", func(t *testing.T) {
		t.Setenv("HELLO", "world")
		got, err := ResolveVariable("env:HELLO")
		require.NoError(t, err)
		assert.Equal(t, "world", got)
	})

	t.Run("Register and resolve custom scheme", func(t *testing.T) {
		stub := &stubResolver{}
		RegisterResolver("test1:", stub)

		got, err := ResolveVariable("test1:abc/def")
		require.NoError(t, err)
		assert.Equal(t, "stub:abc/def", got)
		assert.Equal(t, "abc/def", stub.last, "resolver should receive value without the scheme prefix")
	})

	t.Run("Register replaces existing resolver", func(t *testing.T) {
		stub1 := &stubResolver{out: "first"}
		stub2 := &stubResolver{out: "second"}
		scheme := "testreplace:"

		RegisterResolver(scheme, stub1)
		got1, err := ResolveVariable(scheme + "x")
		require.NoError(t, err)
		assert.Equal(t, "first", got1)

		// Replace with a new resolver
		RegisterResolver(scheme, stub2)
		got2, err := ResolveVariable(scheme + "x")
		require.NoError(t, err)
		assert.Equal(t, "second", got2)
	})
}

func TestDefaultRegistry(t *testing.T) {
	t.Run("Register invalid scheme", func(t *testing.T) {
		// Missing trailing colon → must panic per contract
		assert.Panics(t, func() {
			RegisterResolver("bad", &stubResolver{})
		})
		assert.Panics(t, func() {
			RegisterResolver("", &stubResolver{})
		})
	})

	t.Run("Custom resolver error propagates", func(t *testing.T) {
		wantErr := errors.New("boom")
		stub := &stubResolver{err: wantErr}
		RegisterResolver("testerr:", stub)

		_, err := ResolveVariable("testerr:anything")
		require.Error(t, err)
		assert.ErrorIs(t, err, wantErr)
	})

	t.Run("Returns singleton pointer", func(t *testing.T) {
		reg := DefaultRegistry()
		require.NotNil(t, reg)
		// Calling again should yield the same pointer
		reg2 := DefaultRegistry()
		assert.Same(t, reg, reg2)
	})
}
