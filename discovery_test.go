package resolver

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistryPrefixes(t *testing.T) {
	t.Run("empty registry", func(t *testing.T) {
		assert.Empty(t, NewRegistry().Prefixes())
	})
	t.Run("sorted independent copy includes custom registrations", func(t *testing.T) {
		r := NewRegistry()
		r.Register("z:", ResolverFunc(func(v string) (string, error) { return v, nil }))
		r.Register("a:", ResolverFunc(func(v string) (string, error) { return v, nil }))
		prefixes := r.Prefixes()
		assert.Equal(t, []string{"a:", "z:"}, prefixes)
		prefixes[0] = "changed:"
		assert.Equal(t, []string{"a:", "z:"}, r.Prefixes())
		assert.Equal(t, []string{"z:", "a:"}, r.Schemes())
	})
	t.Run("replacement does not duplicate prefix", func(t *testing.T) {
		r := NewRegistry()
		r.Register("custom:", &EnvResolver{})
		r.Register("custom:", &KeyValueFileResolver{})
		assert.Equal(t, []string{"custom:"}, r.Prefixes())
	})
}

func TestRegistryCanResolve(t *testing.T) {
	t.Run("recognizes builtins without reading resources", func(t *testing.T) {
		r := NewDefaultRegistry()
		assert.True(t, r.CanResolve("env:RESOLVER_NONEXISTENT_VARIABLE"))
		assert.True(t, r.CanResolve("file:/nonexistent/secret"))
		assert.True(t, r.CanResolve("env:"))
	})
	t.Run("custom resolver is not invoked", func(t *testing.T) {
		r := NewRegistry()
		assert.False(t, r.CanResolve("custom:value"))
		r.Register("custom:", ResolverFunc(func(string) (string, error) {
			t.Fatal("discovery must not invoke resolver")
			return "", nil
		}))
		assert.True(t, r.CanResolve("custom:value"))
		assert.False(t, NewRegistry().CanResolve("custom:value"))
	})
	t.Run("matching is case sensitive and anchored", func(t *testing.T) {
		r := NewDefaultRegistry()
		assert.False(t, r.CanResolve("ENV:NAME"))
		assert.False(t, r.CanResolve(" env:NAME"))
		assert.False(t, r.CanResolve("Bearer env:NAME"))
		assert.False(t, r.CanResolve("https://example.com"))
		assert.False(t, r.CanResolve(""))
	})
	t.Run("overlapping prefixes match resolution", func(t *testing.T) {
		r := NewRegistry()
		r.Register("custom:", ResolverFunc(func(v string) (string, error) { return "first:" + v, nil }))
		r.Register("custom:nested:", ResolverFunc(func(v string) (string, error) { return "second:" + v, nil }))
		assert.True(t, r.CanResolve("custom:nested:value"))
		value, err := r.ResolveVariable("custom:nested:value")
		require.NoError(t, err)
		assert.Equal(t, "first:nested:value", value)
	})
}

func TestDefaultRegistryDiscovery(t *testing.T) {
	prefix := "discovery-test-custom:"
	RegisterResolver(prefix, ResolverFunc(func(v string) (string, error) { return v, nil }))
	assert.Contains(t, Prefixes(), prefix)
	assert.True(t, CanResolve(prefix+"value"))
	assert.NotContains(t, NewDefaultRegistry().Prefixes(), prefix)
}

func TestMountedSecretNewlines(t *testing.T) {
	t.Run("trims trailing newlines and preserves spaces", func(t *testing.T) {
		filename := filepath.Join(t.TempDir(), "token")
		require.NoError(t, os.WriteFile(filename, []byte(" secret with spaces \r\n\n"), 0o600))
		r := NewDefaultRegistry()
		r.Register("file:", &KeyValueFileResolver{Whitespace: FileTrimNewlines})
		value, err := r.ResolveVariable("file:" + filename)
		require.NoError(t, err)
		assert.Equal(t, " secret with spaces ", value)
	})
}
