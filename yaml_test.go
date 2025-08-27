package resolver

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createYAMLTestFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(p, []byte(content), 0o666))
	return p
}

func TestYAMLResolver_Resolve(t *testing.T) {
	r := &YAMLResolver{}

	t.Run("Whole file", func(t *testing.T) {
		content := `server:
  host: localhost
  port: 8080
  nested:
    key: value
servers:
  - host: example.com
    port: 80
  - host: example.org
    port: 443
emptyString: ""
nonString:
  inner: true
`
		p := createYAMLTestFile(t, content)

		val, err := r.Resolve(p)
		require.NoError(t, err)
		assert.Equal(t, strings.TrimSpace(content), val)
	})

	t.Run("Top-level key", func(t *testing.T) {
		content := `server:
  host: localhost
  port: 8080
`
		p := createYAMLTestFile(t, content)

		val, err := r.Resolve(p + "//server.host")
		require.NoError(t, err)
		assert.Equal(t, "localhost", val)
	})

	t.Run("Nested key", func(t *testing.T) {
		content := `server:
  nested:
    key: value
`
		p := createYAMLTestFile(t, content)

		val, err := r.Resolve(p + "//server.nested.key")
		require.NoError(t, err)
		assert.Equal(t, "value", val)
	})

	t.Run("Array index element", func(t *testing.T) {
		content := `servers:
  - host: example.com
    port: 80
  - host: example.org
    port: 443
`
		p := createYAMLTestFile(t, content)

		val, err := r.Resolve(p + "//servers.1.host")
		require.NoError(t, err)
		assert.Equal(t, "example.org", val)
	})

	t.Run("Array filter element", func(t *testing.T) {
		// Requires bracket-aware path splitting in yaml.go:
		// selector.Navigate(contentMap, selector.ParsePath(keyPath))
		content := `servers:
  - host: example.com
    port: 80
  - host: example.org
    port: 443
`
		p := createYAMLTestFile(t, content)

		val, err := r.Resolve(p + "//servers.[host=example.org].port")
		require.NoError(t, err)
		// YAML numbers remain numbers; resolver marshals non-strings to YAML and trims.
		// For a scalar 443 that becomes "443".
		assert.Equal(t, "443", val)
	})

	t.Run("Empty string value", func(t *testing.T) {
		content := `emptyString: ""`
		p := createYAMLTestFile(t, content)

		val, err := r.Resolve(p + "//emptyString")
		require.NoError(t, err)
		assert.Equal(t, "", val)
	})

	t.Run("Non-string value (encoded)", func(t *testing.T) {
		content := `nonString:
  inner: true
`
		p := createYAMLTestFile(t, content)

		val, err := r.Resolve(p + "//nonString")
		require.NoError(t, err)
		// Object is marshaled back to YAML, then trimmed.
		assert.Equal(t, "inner: true", val)
	})

	t.Run("Missing key", func(t *testing.T) {
		content := `server:
  nested:
    key: value
`
		p := createYAMLTestFile(t, content)

		_, err := r.Resolve(p + "//server.nested.missingKey")
		require.Error(t, err)
	})

	t.Run("Non-existing file", func(t *testing.T) {
		_, err := r.Resolve(filepath.Join(t.TempDir(), "nonexistent.yaml"))
		require.Error(t, err)
	})

	t.Run("Invalid YAML", func(t *testing.T) {
		p := createYAMLTestFile(t, `key: "unclosed string`)
		_, err := r.Resolve(p)
		require.Error(t, err)
		// Be resilient to upstream error message wording; check key fragments.
		msg := err.Error()
		assert.Contains(t, msg, "failed to parse YAML in")
		assert.Contains(t, msg, p)
	})
}
