package resolver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUtils(t *testing.T) {
	t.Parallel()

	t.Run("NoDelimiter", func(t *testing.T) {
		t.Parallel()
		file, key := splitFileAndKey("path/to/file")
		assert.Equal(t, "path/to/file", file)
		assert.Equal(t, "", key)
	})

	t.Run("DelimiterStart", func(t *testing.T) {
		t.Parallel()
		file, key := splitFileAndKey("//key")
		assert.Equal(t, "", file)
		assert.Equal(t, "key", key)
	})

	t.Run("DelimiterEnd", func(t *testing.T) {
		t.Parallel()
		file, key := splitFileAndKey("path/to/file//")
		assert.Equal(t, "path/to/file", file)
		assert.Equal(t, "", key)
	})

	t.Run("DelimiterMiddle", func(t *testing.T) {
		t.Parallel()
		file, key := splitFileAndKey("path/to/file//somekey")
		assert.Equal(t, "path/to/file", file)
		assert.Equal(t, "somekey", key)
	})

	t.Run("MultipleDelimiters", func(t *testing.T) {
		t.Parallel()
		file, key := splitFileAndKey("path//to//file//key")
		assert.Equal(t, "path//to//file", file)
		assert.Equal(t, "key", key)
	})
}
