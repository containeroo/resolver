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

	t.Run("Quoted values, escapes, and inline comments", func(t *testing.T) {
		r := &KeyValueFileResolver{}

		// Use a raw string literal so we don't double-escape.
		content := `
# full-line comment
A=foo  # trailing comment
B="value with spaces"
C='single # not a comment'
D="has # inside"  # after-quote comment
E="line1\nline2"  # escaped newline
F="quote:\" and slash\\"
G='it\'s ok'
HASH_NO_SPACE=foo#bar
HASH_SPACE=foo #bar
JUSTEXPORT
export
Q="trail\\"
U="a\z b"
T="a\tb\rc"
`
		p := createKeyValueTestFile(t, content)

		val, err := r.Resolve(p + "//A")
		require.NoError(t, err)
		assert.Equal(t, "foo", val)

		val, err = r.Resolve(p + "//B")
		require.NoError(t, err)
		assert.Equal(t, "value with spaces", val)

		val, err = r.Resolve(p + "//C")
		require.NoError(t, err)
		assert.Equal(t, "single # not a comment", val)

		val, err = r.Resolve(p + "//D")
		require.NoError(t, err)
		assert.Equal(t, "has # inside", val)

		val, err = r.Resolve(p + "//E")
		require.NoError(t, err)
		assert.Equal(t, "line1\nline2", val)

		val, err = r.Resolve(p + "//F")
		require.NoError(t, err)
		assert.Equal(t, "quote:\" and slash\\", val)

		val, err = r.Resolve(p + "//G")
		require.NoError(t, err)
		assert.Equal(t, "it's ok", val)

		// '#' immediately after value without whitespace should NOT start a comment.
		val, err = r.Resolve(p + "//HASH_NO_SPACE")
		require.NoError(t, err)
		assert.Equal(t, "foo#bar", val)

		// '#' after at least one whitespace IS a comment delimiter.
		val, err = r.Resolve(p + "//HASH_SPACE")
		require.NoError(t, err)
		assert.Equal(t, "foo", val)

		// Lines without '=' or with only 'export' should be ignored (don't crash, don't match).
		_, err = r.Resolve(p + "//JUSTEXPORT")
		require.Error(t, err)
		_, err = r.Resolve(p + "//export")
		require.Error(t, err)

		// Trailing backslash inside double quotes should be preserved by unescapeDoubleQuoted.
		val, err = r.Resolve(p + "//Q")
		require.NoError(t, err)
		assert.Equal(t, `trail\`, val)

		// Unknown escape sequences drop the backslash but keep the char.
		val, err = r.Resolve(p + "//U")
		require.NoError(t, err)
		assert.Equal(t, "a"+"z b", val) // backslash removed before 'z'

		// \t and \r are handled.
		val, err = r.Resolve(p + "//T")
		require.NoError(t, err)
		assert.Equal(t, "a\tb\rc", val)
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

	t.Run("CRLF", func(t *testing.T) {
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

	t.Run("Whole file strips BOM", func(t *testing.T) {
		r := &KeyValueFileResolver{}
		// Note: BOM at the start of the file should be removed for whole-file reads.
		content := "\uFEFFA=1\nB=2\n"
		p := createKeyValueTestFile(t, content)

		all, err := r.Resolve(p)
		require.NoError(t, err)
		assert.Equal(t, "A=1\nB=2", all)
	})

	t.Run("Scanner error path (token too long)", func(t *testing.T) {
		r := &KeyValueFileResolver{}

		// Construct a file with a huge line > 1MB to exceed Scanner max token size (we set max to 1MB).
		var b strings.Builder
		b.WriteString("A=1\n")                          // small line first
		b.WriteString(strings.Repeat("X", 2*1024*1024)) // 2MB single line, no '='
		b.WriteByte('\n')

		p := createKeyValueTestFile(t, b.String())

		// Ask for a key that doesn't exist so we force the scan to traverse the huge line
		// and trigger ErrTooLong.
		_, err := r.Resolve(p + "//ZZZ")
		require.Error(t, err, "expected scanner to report ErrTooLong for oversized token")
	})
}
