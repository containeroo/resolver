package resolver

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubResolver helps test custom resolvers.
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

// countingResolver helps verify short-circuiting behavior in strict mode.
type countingResolver struct {
	prefix string
	count  int
}

func (c *countingResolver) Resolve(v string) (string, error) {
	c.count++
	return c.prefix + v, nil
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

func TestResolveSlice(t *testing.T) {
	t.Run("Empty slice returns empty", func(t *testing.T) {
		got, err := ResolveSlice(nil)
		require.NoError(t, err)
		assert.NotNil(t, got)
		assert.Len(t, got, 0)
	})

	t.Run("Mixed literals, unknown scheme pass-through, and successful resolutions", func(t *testing.T) {
		// env
		t.Setenv("HELLO", "world")

		// custom resolvers
		RegisterResolver("sliceok:", &stubResolver{}) // returns "stub:<rest>"

		in := []string{
			"sliceok:a",    // resolved by stub
			"just-literal", // unchanged
			"unknown:zzz",  // unknown → unchanged
			"env:HELLO",    // resolved by EnvResolver
			"sliceok:b",    // resolved by stub
		}
		got, err := ResolveSlice(in)
		require.NoError(t, err)

		assert.Equal(t, []string{
			"stub:a",
			"just-literal",
			"unknown:zzz",
			"world",
			"stub:b",
		}, got)
	})

	t.Run("Stops on first error (strict)", func(t *testing.T) {
		ok := &countingResolver{prefix: "ok:"}
		RegisterResolver("sliceok:", ok)

		wantErr := errors.New("boom")
		RegisterResolver("sliceerr:", &stubResolver{err: wantErr})

		// Should stop at index 1, not resolving the last element.
		in := []string{"sliceok:a", "sliceerr:oops", "sliceok:b"}

		got, err := ResolveSlice(in)
		require.Error(t, err)
		assert.ErrorIs(t, err, wantErr)

		// Verify short-circuit: only the first "sliceok:" call should have happened.
		assert.Equal(t, 1, ok.count, "should stop after the first error")
		assert.Nil(t, got, "strict resolver should not return partial results")
	})
}

func TestResolveSliceBestEffort(t *testing.T) {
	t.Run("Empty slice returns empty outputs and no errors", func(t *testing.T) {
		out, errs := ResolveSliceBestEffort(nil)
		assert.NotNil(t, out)
		assert.Len(t, out, 0)
		assert.Len(t, errs, 0)
	})

	t.Run("Collects errors but resolves the rest; unknown schemes pass through", func(t *testing.T) {
		ok := &countingResolver{prefix: "ok:"}
		RegisterResolver("sliceok:", ok)

		wantErr := errors.New("kaput")
		RegisterResolver("sliceerr:", &stubResolver{err: wantErr})

		in := []string{
			"sliceok:a",    // success
			"sliceerr:x",   // error
			"unknown:zzz",  // unknown → unchanged
			"sliceok:b",    // success
			"just-literal", // unchanged
			"sliceerr:y",   // error
		}

		out, errs := ResolveSliceBestEffort(in)

		// Output length always equals input length.
		require.Len(t, out, len(in))

		// Successes & pass-throughs are correct.
		assert.Equal(t, "ok:a", out[0])
		assert.Equal(t, "unknown:zzz", out[2])
		assert.Equal(t, "ok:b", out[3])
		assert.Equal(t, "just-literal", out[4])

		// We don't assert what out[1] / out[5] contain, since implementations may
		// choose "" or the original unresolved value on error. We only assert errors.
		require.Len(t, errs, 2)
		assert.ErrorIs(t, errs[0], wantErr)
		assert.ErrorIs(t, errs[1], wantErr)

		// Best-effort should process all resolvable items.
		assert.Equal(t, 2, ok.count, "both sliceok:* entries should be resolved")
	})
}
