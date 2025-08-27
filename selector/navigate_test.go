package selector

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNavigate(t *testing.T) {
	t.Parallel()

	// Test fixture: nested map/arrays with mixed types
	data := map[string]any{
		"server": map[string]any{
			"host": "localhost",
			"port": 8080,
			"nested": map[string]any{
				"key": "value",
			},
		},
		"servers": []any{
			map[string]any{"name": "web", "host": "example.com", "port": 80, "enabled": true, "id": 1},
			map[string]any{"name": "api", "host": "example.org", "port": 443, "enabled": false, "id": 2},
			// non-map element to ensure filter logic skips safely
			"not-a-map",
		},
		"nums": []any{10, 20, 30},
		"leaf": "done",
	}

	t.Run("map key", func(t *testing.T) {
		t.Parallel()
		val, err := Navigate(data, ParsePath("server.host"))
		require.NoError(t, err)
		assert.Equal(t, "localhost", val)
	})

	t.Run("nested map key", func(t *testing.T) {
		t.Parallel()
		val, err := Navigate(data, ParsePath("server.nested.key"))
		require.NoError(t, err)
		assert.Equal(t, "value", val)
	})

	t.Run("array index", func(t *testing.T) {
		t.Parallel()
		val, err := Navigate(data, ParsePath("servers.1.host"))
		require.NoError(t, err)
		assert.Equal(t, "example.org", val)
	})

	t.Run("array filter by string", func(t *testing.T) {
		t.Parallel()
		val, err := Navigate(data, ParsePath("servers.[name=api].port"))
		require.NoError(t, err)
		assert.Equal(t, 443, val)
	})

	t.Run("array filter by int (with coercion)", func(t *testing.T) {
		t.Parallel()
		val, err := Navigate(data, ParsePath("servers.[id=2].host"))
		require.NoError(t, err)
		assert.Equal(t, "example.org", val)
	})

	t.Run("array filter by bool (with coercion)", func(t *testing.T) {
		t.Parallel()
		val, err := Navigate(data, ParsePath("servers.[enabled=true].name"))
		require.NoError(t, err)
		assert.Equal(t, "web", val)
	})

	t.Run("array filter with dots in value", func(t *testing.T) {
		t.Parallel()
		// augment a local copy to avoid data races across parallel tests
		local := map[string]any{
			"items": []any{
				map[string]any{"key": "a.b.c", "val": 1},
				map[string]any{"key": "x.y", "val": 2},
			},
		}
		val, err := Navigate(local, ParsePath("items.[key=a.b.c].val"))
		require.NoError(t, err)
		assert.Equal(t, 1, val)
	})

	t.Run("non-string return (number)", func(t *testing.T) {
		t.Parallel()
		val, err := Navigate(data, ParsePath("server.port"))
		require.NoError(t, err)
		assert.Equal(t, 8080, val)
	})

	t.Run("missing map key error", func(t *testing.T) {
		t.Parallel()
		_, err := Navigate(data, ParsePath("server.nope"))
		require.Error(t, err)
	})

	t.Run("invalid array index token", func(t *testing.T) {
		t.Parallel()
		_, err := Navigate(data, ParsePath("servers.one.host"))
		require.Error(t, err)
	})

	t.Run("array index out of bounds", func(t *testing.T) {
		t.Parallel()
		_, err := Navigate(data, ParsePath("servers.99.host"))
		require.Error(t, err)
	})

	t.Run("array filter no match", func(t *testing.T) {
		t.Parallel()
		_, err := Navigate(data, ParsePath("servers.[name=missing].host"))
		require.Error(t, err)
	})

	t.Run("descending through non-container", func(t *testing.T) {
		t.Parallel()
		_, err := Navigate(data, ParsePath("leaf.next"))
		require.Error(t, err)
	})

	t.Run("non-map element skipped safely in filter", func(t *testing.T) {
		t.Parallel()
		// Should still find the matching map element even with a non-map present.
		val, err := Navigate(data, ParsePath("servers.[id=1].host"))
		require.NoError(t, err)
		assert.Equal(t, "example.com", val)
	})
}
