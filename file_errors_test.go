package resolver

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileErrorClassification(t *testing.T) {
	r := NewDefaultRegistry()
	dir := t.TempDir()
	for _, tc := range []struct {
		name            string
		cause, category error
	}{
		{"missing", fs.ErrNotExist, ErrNotFound},
		{"unreadable", fs.ErrPermission, ErrForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(dir, tc.name)
			if tc.cause == fs.ErrPermission {
				require.NoError(t, os.WriteFile(path, nil, 0000))
				if _, err := os.ReadFile(path); err == nil {
					t.Skip("current user can read mode-000 files")
				}
			}
			for _, prefix := range []string{"file:", "json:", "yaml:", "toml:", "ini:"} {
				_, err := r.ResolveVariable(prefix + path)
				require.ErrorIs(t, err, tc.cause, prefix)
				require.ErrorIs(t, err, tc.category, prefix)
				var pathErr *fs.PathError
				require.ErrorAs(t, err, &pathErr, prefix)
				require.Equal(t, path, pathErr.Path)
			}
		})
	}
}
