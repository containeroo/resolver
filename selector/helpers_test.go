package selector

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePath(t *testing.T) {
	t.Parallel()

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		got := ParsePath("")
		assert.Equal(t, []string{""}, got)
	})

	t.Run("single", func(t *testing.T) {
		t.Parallel()
		got := ParsePath("server")
		assert.Equal(t, []string{"server"}, got)
	})

	t.Run("dot path", func(t *testing.T) {
		t.Parallel()
		got := ParsePath("server.host")
		assert.Equal(t, []string{"server", "host"}, got)
	})

	t.Run("array index", func(t *testing.T) {
		t.Parallel()
		got := ParsePath("servers.0.host")
		assert.Equal(t, []string{"servers", "0", "host"}, got)
	})

	t.Run("filter simple", func(t *testing.T) {
		t.Parallel()
		got := ParsePath("servers.[name=api].host")
		assert.Equal(t, []string{"servers", "[name=api]", "host"}, got)
	})

	t.Run("filter with dots", func(t *testing.T) {
		t.Parallel()
		got := ParsePath("servers.[host=example.org].port")
		assert.Equal(t, []string{"servers", "[host=example.org]", "port"}, got)
	})
}

func TestIsFilterToken(t *testing.T) {
	t.Parallel()

	t.Run("plain filter", func(t *testing.T) {
		t.Parallel()
		assert.True(t, isFilterToken("[k=v]"))
	})

	t.Run("quoted value", func(t *testing.T) {
		t.Parallel()
		assert.True(t, isFilterToken("[k=\"v\"]"))
	})

	t.Run("no brackets", func(t *testing.T) {
		t.Parallel()
		assert.False(t, isFilterToken("k=v"))
	})

	t.Run("missing equals", func(t *testing.T) {
		t.Parallel()
		assert.False(t, isFilterToken("[kv]"))
	})
}

func TestParseFilterToken(t *testing.T) {
	t.Parallel()

	t.Run("simple", func(t *testing.T) {
		t.Parallel()
		k, v, err := parseFilterToken("[k=v]")
		require.NoError(t, err)
		assert.Equal(t, "k", k)
		assert.Equal(t, "v", v)
	})

	t.Run("quoted value", func(t *testing.T) {
		t.Parallel()
		k, v, err := parseFilterToken("[k=\"v.with.dots\"]")
		require.NoError(t, err)
		assert.Equal(t, "k", k)
		assert.Equal(t, "v.with.dots", v)
	})

	t.Run("invalid", func(t *testing.T) {
		t.Parallel()
		_, _, err := parseFilterToken("[kv]")
		require.Error(t, err)
	})
}

func TestCoerce(t *testing.T) {
	t.Parallel()

	t.Run("bool true", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, true, coerce("true"))
	})

	t.Run("bool false", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, false, coerce("false"))
	})

	t.Run("integer", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, 42, coerce("42"))
	})

	t.Run("negative integer", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, -7, coerce("-7"))
	})

	t.Run("float", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, 3.14, coerce("3.14"))
	})

	t.Run("float in scientific notation", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, 1.2e3, coerce("1.2e3"))
	})

	t.Run("string fallback", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "hello", coerce("hello"))
	})

	t.Run("numeric string does not become bool", func(t *testing.T) {
		t.Parallel()
		// Ensure "1" and "0" are treated as numbers, not bool
		assert.Equal(t, 1, coerce("1"))
		assert.Equal(t, 0, coerce("0"))
	})
}

func TestEqualCoerced(t *testing.T) {
	t.Parallel()

	t.Run("bool equal", func(t *testing.T) {
		t.Parallel()
		assert.True(t, equalCoerced(true, true))
	})

	t.Run("bool not equal", func(t *testing.T) {
		t.Parallel()
		assert.False(t, equalCoerced(false, true))
	})

	t.Run("int and int64 equal", func(t *testing.T) {
		t.Parallel()
		assert.True(t, equalCoerced(int64(5), 5))
	})

	t.Run("float not equal int", func(t *testing.T) {
		t.Parallel()
		assert.False(t, equalCoerced(5.1, 5))
	})

	t.Run("string equal", func(t *testing.T) {
		t.Parallel()
		assert.True(t, equalCoerced("x", "x"))
	})

	t.Run("string not equal", func(t *testing.T) {
		t.Parallel()
		assert.False(t, equalCoerced("x", "y"))
	})
}
