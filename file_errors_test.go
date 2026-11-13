package resolver

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileErrorClassification(t *testing.T) {
	r := NewDefaultRegistry()
	check := func(t *testing.T, prefix string, cause, category error) {
		t.Helper()
		path := filepath.Join(t.TempDir(), "config")
		if cause == fs.ErrPermission {
			require.NoError(t, os.WriteFile(path, nil, 0000))
			if _, err := os.ReadFile(path); err == nil {
				t.Skip("current user can read mode-000 files")
			}
		}

		_, err := r.ResolveVariable(prefix + path)
		assert.ErrorIs(t, err, cause, prefix)
		assert.ErrorIs(t, err, category, prefix)
		var pathErr *fs.PathError
		require.ErrorAs(t, err, &pathErr, prefix)
		assert.Equal(t, path, pathErr.Path)
	}

	t.Run("file missing", func(t *testing.T) {
		t.Parallel()
		check(t, "file:", fs.ErrNotExist, ErrNotFound)
	})

	t.Run("json missing", func(t *testing.T) {
		t.Parallel()
		check(t, "json:", fs.ErrNotExist, ErrNotFound)
	})

	t.Run("yaml missing", func(t *testing.T) {
		t.Parallel()
		check(t, "yaml:", fs.ErrNotExist, ErrNotFound)
	})

	t.Run("toml missing", func(t *testing.T) {
		t.Parallel()
		check(t, "toml:", fs.ErrNotExist, ErrNotFound)
	})

	t.Run("ini missing", func(t *testing.T) {
		t.Parallel()
		check(t, "ini:", fs.ErrNotExist, ErrNotFound)
	})

	t.Run("file unreadable", func(t *testing.T) {
		t.Parallel()
		check(t, "file:", fs.ErrPermission, ErrForbidden)
	})

	t.Run("json unreadable", func(t *testing.T) {
		t.Parallel()
		check(t, "json:", fs.ErrPermission, ErrForbidden)
	})

	t.Run("yaml unreadable", func(t *testing.T) {
		t.Parallel()
		check(t, "yaml:", fs.ErrPermission, ErrForbidden)
	})

	t.Run("toml unreadable", func(t *testing.T) {
		t.Parallel()
		check(t, "toml:", fs.ErrPermission, ErrForbidden)
	})

	t.Run("ini unreadable", func(t *testing.T) {
		t.Parallel()
		check(t, "ini:", fs.ErrPermission, ErrForbidden)
	})
}
