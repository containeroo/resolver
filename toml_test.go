package resolver

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTOMLTestFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "config.toml")
	require.NoError(t, os.WriteFile(p, []byte(content), 0o666))
	return p
}

func TestTOMLResolver_Resolve(t *testing.T) {
	r := &TOMLResolver{}

	t.Run("Whole file", func(t *testing.T) {
		content := `emptyString = ""

[server]
host = "localhost"
port = 8080
[server.nested]
key = "value"

[[servers]]
host = "example.com"
port = 80

[[servers]]
host = "example.org"
port = 443

[nonString]
inner = true
`
		p := createTOMLTestFile(t, content)

		val, err := r.Resolve(p)
		require.NoError(t, err)
		assert.Equal(t, strings.TrimSpace(content), val)
	})

	t.Run("Top-level key", func(t *testing.T) {
		content := `
[server]
host = "localhost"
port = 8080
`
		p := createTOMLTestFile(t, content)

		val, err := r.Resolve(p + "//server.host")
		require.NoError(t, err)
		assert.Equal(t, "localhost", val)
	})

	t.Run("Nested key", func(t *testing.T) {
		content := `
[server]
host = "localhost"
[server.nested]
key = "value"
`
		p := createTOMLTestFile(t, content)

		val, err := r.Resolve(p + "//server.nested.key")
		require.NoError(t, err)
		assert.Equal(t, "value", val)
	})

	t.Run("Array index element", func(t *testing.T) {
		content := `
[[servers]]
host = "example.com"
port = 80

[[servers]]
host = "example.org"
port = 443
`
		p := createTOMLTestFile(t, content)

		val, err := r.Resolve(p + "//servers.1.host")
		require.NoError(t, err)
		assert.Equal(t, "example.org", val)
	})

	t.Run("Array filter element", func(t *testing.T) {
		// Ensures bracket-aware path parsing works (servers.[host=example.org].port)
		content := `
[[servers]]
host = "example.com"
port = 80

[[servers]]
host = "example.org"
port = 443
`
		p := createTOMLTestFile(t, content)

		val, err := r.Resolve(p + "//servers.[host=example.org].port")
		require.NoError(t, err)
		// Non-string → TOML-encoded back to string "443"
		assert.Equal(t, "443", val)
	})

	t.Run("Empty string value", func(t *testing.T) {
		content := `emptyString = ""`
		p := createTOMLTestFile(t, content)

		val, err := r.Resolve(p + "//emptyString")
		require.NoError(t, err)
		assert.Equal(t, "", val)
	})

	t.Run("Non-string value (encoded)", func(t *testing.T) {
		content := `
[nonString]
inner = true
`
		p := createTOMLTestFile(t, content)

		val, err := r.Resolve(p + "//nonString")
		require.NoError(t, err)
		// The resolver marshals non-strings back to TOML; a table encodes as "inner = true"
		assert.Equal(t, "inner = true", val)
	})

	t.Run("Missing key", func(t *testing.T) {
		content := `
[server]
host = "localhost"
`
		p := createTOMLTestFile(t, content)

		_, err := r.Resolve(p + "//server.missing")
		require.Error(t, err)
	})

	t.Run("Non-existing file", func(t *testing.T) {
		_, err := r.Resolve(filepath.Join(t.TempDir(), "nonexistent.toml"))
		require.Error(t, err)
	})

	t.Run("Invalid TOML", func(t *testing.T) {
		dir := t.TempDir()
		bad := filepath.Join(dir, "bad.toml")
		require.NoError(t, os.WriteFile(bad, []byte("= invalid"), 0o666))

		val, err := r.Resolve(bad)
		assert.Equal(t, "", val)
		require.Error(t, err)
	})
}
