package resolver

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createJSONTestFile(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	p := filepath.Join(dir, "config.json")

	// Nested objects and arrays; also an empty string and a non-string value.
	content := `{
  "server": {
    "host": "localhost",
    "port": 8080,
    "nested": {
      "key": "value"
    }
  },
  "servers": [
    { "host": "example.com", "port": 80 },
    { "host": "example.org", "port": 443 }
  ],
  "emptyString": "",
  "nonString": { "inner": true }
}`
	require.NoError(t, os.WriteFile(p, []byte(content), 0o666), "failed to create test JSON file")
	return p
}

func TestJSONResolver_Resolve(t *testing.T) {
	t.Run("Whole file", func(t *testing.T) {
		t.Parallel()
		r := &JSONResolver{}
		p := createJSONTestFile(t)

		val, err := r.Resolve(p)
		require.NoError(t, err)

		expected := `{
  "server": {
    "host": "localhost",
    "port": 8080,
    "nested": {
      "key": "value"
    }
  },
  "servers": [
    { "host": "example.com", "port": 80 },
    { "host": "example.org", "port": 443 }
  ],
  "emptyString": "",
  "nonString": { "inner": true }
}`
		assert.Equal(t, expected, val)
	})

	t.Run("Top level key", func(t *testing.T) {
		t.Parallel()
		r := &JSONResolver{}
		p := createJSONTestFile(t)

		val, err := r.Resolve(p + "//server.host")
		require.NoError(t, err)
		assert.Equal(t, "localhost", val)
	})

	t.Run("Nested key", func(t *testing.T) {
		t.Parallel()
		r := &JSONResolver{}
		p := createJSONTestFile(t)

		val, err := r.Resolve(p + "//server.nested.key")
		require.NoError(t, err)
		assert.Equal(t, "value", val)
	})

	t.Run("Array index", func(t *testing.T) {
		t.Parallel()
		r := &JSONResolver{}
		p := createJSONTestFile(t)

		val, err := r.Resolve(p + "//servers.1.host")
		require.NoError(t, err)
		assert.Equal(t, "example.org", val)
	})

	t.Run("Array filter", func(t *testing.T) {
		t.Parallel()
		// Uses the selector filter form: [key=value]
		r := &JSONResolver{}
		p := createJSONTestFile(t)

		val, err := r.Resolve(p + "//servers.[host=example.org].port")
		require.NoError(t, err)
		// Non-string values are JSON-encoded on return; 443 -> "443"
		assert.Equal(t, "443", val)
	})

	t.Run("Empty string value", func(t *testing.T) {
		t.Parallel()
		r := &JSONResolver{}
		p := createJSONTestFile(t)

		val, err := r.Resolve(p + "//emptyString")
		require.NoError(t, err)
		assert.Equal(t, "", val)
	})

	t.Run("Non-string value encoded", func(t *testing.T) {
		t.Parallel()
		r := &JSONResolver{}
		p := createJSONTestFile(t)

		val, err := r.Resolve(p + "//nonString")
		require.NoError(t, err)
		assert.Equal(t, `{"inner":true}`, val)
	})

	t.Run("Missing key", func(t *testing.T) {
		t.Parallel()
		r := &JSONResolver{}
		p := createJSONTestFile(t)

		_, err := r.Resolve(p + "//server.nested.missingKey")
		require.Error(t, err)
	})

	t.Run("File not found", func(t *testing.T) {
		t.Parallel()
		r := &JSONResolver{}
		_, err := r.Resolve(filepath.Join(t.TempDir(), "nonexistent.json"))
		require.Error(t, err)
	})
}

func TestJSONNumberPrecision(t *testing.T) {
	p := filepath.Join(t.TempDir(), "ids.json")
	data := `{"rows":[{"id":18446744073709551616},{"id":18446744073709551617}]}`
	require.NoError(t, os.WriteFile(p, []byte(data), 0600))
	r := &JSONResolver{}

	t.Run("index preserves integer", func(t *testing.T) {
		t.Parallel()
		got, err := r.Resolve(p + "//rows.1.id")
		require.NoError(t, err)
		assert.Equal(t, "18446744073709551617", got)
	})

	t.Run("integer filter", func(t *testing.T) {
		t.Parallel()
		got, err := r.Resolve(p + "//rows.[id=18446744073709551617].id")
		require.NoError(t, err)
		assert.Equal(t, "18446744073709551617", got)
	})

	t.Run("decimal filter", func(t *testing.T) {
		t.Parallel()
		got, err := r.Resolve(p + "//rows.[id=18446744073709551617.0].id")
		require.NoError(t, err)
		assert.Equal(t, "18446744073709551617", got)
	})

	t.Run("second document", func(t *testing.T) {
		t.Parallel()
		invalid := filepath.Join(t.TempDir(), "invalid.json")
		require.NoError(t, os.WriteFile(invalid, []byte(data+" {}"), 0600))
		_, err := r.Resolve(invalid + "//rows")
		require.Error(t, err)
	})

	t.Run("trailing text", func(t *testing.T) {
		t.Parallel()
		invalid := filepath.Join(t.TempDir(), "invalid.json")
		require.NoError(t, os.WriteFile(invalid, []byte(data+" garbage"), 0600))
		_, err := r.Resolve(invalid + "//rows")
		require.Error(t, err)
	})

	t.Run("trailing nonbreaking space", func(t *testing.T) {
		t.Parallel()
		invalid := filepath.Join(t.TempDir(), "invalid.json")
		require.NoError(t, os.WriteFile(invalid, []byte(data+"\u00a0"), 0600))
		_, err := r.Resolve(invalid + "//rows")
		require.Error(t, err)
	})
}
