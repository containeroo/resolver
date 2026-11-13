package resolver

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWholeFileWhitespacePolicy(t *testing.T) {
	path := filepath.Join(t.TempDir(), "secret.txt")
	require.NoError(t, os.WriteFile(path, []byte("  token  \r\n"), 0600))

	t.Run("trim surrounding space", func(t *testing.T) {
		t.Parallel()
		r := &KeyValueFileResolver{Whitespace: FileTrimSpace}
		got, err := r.Resolve(path)
		require.NoError(t, err)
		assert.Equal(t, "token", got)
	})

	t.Run("trim trailing newlines", func(t *testing.T) {
		t.Parallel()
		r := &KeyValueFileResolver{Whitespace: FileTrimNewlines}
		got, err := r.Resolve(path)
		require.NoError(t, err)
		assert.Equal(t, "  token  ", got)
	})

	t.Run("preserve whitespace", func(t *testing.T) {
		t.Parallel()
		r := &KeyValueFileResolver{Whitespace: FilePreserveWhitespace}
		got, err := r.Resolve(path)
		require.NoError(t, err)
		assert.Equal(t, "  token  \r\n", got)
	})
}
