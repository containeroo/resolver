package resolver

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapWithError(t *testing.T) {
	t.Run("transforms all items", func(t *testing.T) {
		in := []string{"a", "b", "c"}
		out, err := MapWithError(in, func(s string) (string, error) {
			return s + "!", nil
		})
		assert.NoError(t, err)
		assert.Equal(t, []string{"a!", "b!", "c!"}, out)
	})

	t.Run("returns first error", func(t *testing.T) {
		in := []int{1, 2, 3}
		errBoom := errors.New("boom")
		out, err := MapWithError(in, func(i int) (string, error) {
			if i == 2 {
				return "", errBoom
			}
			return "ok", nil
		})
		assert.Nil(t, out)
		assert.ErrorIs(t, err, errBoom)
	})
}

func TestResolveSlice(t *testing.T) {
	t.Run("resolves env vars in list", func(t *testing.T) {
		t.Setenv("FOO", "abc")
		t.Setenv("BAR", "def")

		in := []string{"env:FOO", "env:BAR"}
		out, err := ResolveSlice(in)
		assert.NoError(t, err)
		assert.Equal(t, []string{"abc", "def"}, out)
	})

	t.Run("returns error on invalid key", func(t *testing.T) {
		in := []string{"env:FOO", "env:DOES_NOT_EXIST"}
		t.Setenv("FOO", "abc")

		out, err := ResolveSlice(in)
		assert.Nil(t, out)
		assert.Error(t, err)
	})
}
