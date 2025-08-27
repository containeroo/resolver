package resolver

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createKeyValueTestFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "app.txt")
	require.NoError(t, os.WriteFile(p, []byte(content), 0o666))
	return p
}

func TestKeyValueFileResolver_Resolve(t *testing.T) {
	t.Run("Whole file", func(t *testing.T) {
		r := &KeyValueFileResolver{}

		content := "Key1 = Value1\nKey2=Value2\nKey with spaces =  TrimmedValue\n"
		p := createKeyValueTestFile(t, content)

		val, err := r.Resolve(p)
		require.NoError(t, err)
		assert.Equal(t, strings.TrimSpace(content), val)
	})

	t.Run("Specific keys", func(t *testing.T) {
		r := &KeyValueFileResolver{}

		content := `
# comment
export KEY2 =  Value2
Key1=Value1
Key with spaces =  TrimmedValue
WITH_EQUALS=foo=bar=baz
`
		p := createKeyValueTestFile(t, content)

		// plain key
		val, err := r.Resolve(p + "//Key1")
		require.NoError(t, err)
		assert.Equal(t, "Value1", val)

		// key with spaces
		val, err = r.Resolve(p + "//Key with spaces")
		require.NoError(t, err)
		assert.Equal(t, "TrimmedValue", val)

		// export form, extra spaces around '='
		val, err = r.Resolve(p + "//KEY2")
		require.NoError(t, err)
		assert.Equal(t, "Value2", val)

		// value containing '='
		val, err = r.Resolve(p + "//WITH_EQUALS")
		require.NoError(t, err)
		assert.Equal(t, "foo=bar=baz", val)
	})

	t.Run("Missing key", func(t *testing.T) {
		r := &KeyValueFileResolver{}
		p := createKeyValueTestFile(t, "A=1\n")

		_, err := r.Resolve(p + "//NOPE")
		require.Error(t, err)
	})

	t.Run("File not found", func(t *testing.T) {
		r := &KeyValueFileResolver{}
		_, err := r.Resolve(filepath.Join(t.TempDir(), "nope.txt"))
		require.Error(t, err)
	})

	t.Run("Expands env vars in path", func(t *testing.T) {
		r := &KeyValueFileResolver{}

		dir := t.TempDir()
		t.Setenv("DIR", dir)
		p := filepath.Join(dir, "env.txt")
		require.NoError(t, os.WriteFile(p, []byte("X=42\n"), 0o666))

		// exercise os.ExpandEnv on the path
		val, err := r.Resolve("${DIR}/env.txt//X")
		require.NoError(t, err)
		assert.Equal(t, "42", val)
	})

	t.Run("File scheme", func(t *testing.T) {
		content := "A=1\n"
		p := createKeyValueTestFile(t, content)

		got, err := ResolveVariable("file:" + p + "//A")
		require.NoError(t, err)
		assert.Equal(t, "1", got)
	})

	t.Run("CRFL", func(t *testing.T) {
		// Ensure Windows CRLF is handled by TrimSpace.
		r := &KeyValueFileResolver{}
		content := "A=1\r\nB=2\r\n"
		p := createKeyValueTestFile(t, content)

		val, err := r.Resolve(p + "//B")
		require.NoError(t, err)
		assert.Equal(t, "2", val)

		if runtime.GOOS == "windows" {
			// Whole file should trim trailing CRLF as well
			all, err := r.Resolve(p)
			require.NoError(t, err)
			assert.Equal(t, "A=1\r\nB=2", all)
		}
	})
}
