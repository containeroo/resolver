package resolver

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecretResolver(t *testing.T) {
	writeSecret := func(t *testing.T, content string) string {
		t.Helper()
		path := filepath.Join(t.TempDir(), "token")
		require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
		return path
	}
	t.Run("trims trailing LF", func(t *testing.T) {
		value, err := ResolveVariable("secret:" + writeSecret(t, "token\n"))
		require.NoError(t, err)
		assert.Equal(t, "token", value)
	})
	t.Run("trims trailing CRLF and repeated endings", func(t *testing.T) {
		value, err := ResolveVariable("secret:" + writeSecret(t, "token\r\n\r\n"))
		require.NoError(t, err)
		assert.Equal(t, "token", value)
	})
	t.Run("preserves spaces tabs internal newlines and BOM", func(t *testing.T) {
		content := "\uFEFF \tfirst\nsecond \t"
		value, err := ResolveVariable("secret:" + writeSecret(t, content+"\r\n"))
		require.NoError(t, err)
		assert.Equal(t, content, value)
	})
	t.Run("expands environment variables in path", func(t *testing.T) {
		t.Setenv("RESOLVER_SECRET_FILE", writeSecret(t, "token"))
		value, err := ResolveVariable("secret:${RESOLVER_SECRET_FILE}")
		require.NoError(t, err)
		assert.Equal(t, "token", value)
	})
	t.Run("rejects empty path", func(t *testing.T) {
		_, err := ResolveVariable("secret:")
		require.ErrorIs(t, err, ErrBadPath)
	})
	t.Run("classifies missing file", func(t *testing.T) {
		_, err := ResolveVariable("secret:" + filepath.Join(t.TempDir(), "missing"))
		require.ErrorIs(t, err, ErrNotFound)
	})
	t.Run("does not interpret key selectors", func(t *testing.T) {
		_, err := ResolveVariable("secret:" + writeSecret(t, "TOKEN=value") + "//TOKEN")
		require.Error(t, err)
	})
	t.Run("allows empty secret by default", func(t *testing.T) {
		value, err := ResolveVariable("secret:" + writeSecret(t, "\n"))
		require.NoError(t, err)
		assert.Empty(t, value)
	})
	t.Run("honors empty value policy", func(t *testing.T) {
		r := NewDefaultRegistry(WithEmptyValuePolicy(RejectEmpty))
		_, err := r.ResolveVariable("secret:" + writeSecret(t, "\n"))
		require.Error(t, err)
	})
	t.Run("is discoverable", func(t *testing.T) {
		assert.Contains(t, Prefixes(), "secret:")
		assert.True(t, CanResolve("secret:/run/secrets/token"))
	})
	t.Run("reads rotated secret on each call", func(t *testing.T) {
		filename := writeSecret(t, "old\n")
		_, err := ResolveVariable("secret:" + filename)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filename, []byte("new\n"), 0o600))
		value, err := ResolveVariable("secret:" + filename)
		require.NoError(t, err)
		assert.Equal(t, "new", value)
	})
}
