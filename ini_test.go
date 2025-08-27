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

func createIniTestFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "config.ini")
	require.NoError(t, os.WriteFile(p, []byte(content), 0o666))
	return p
}

func TestINIResolver_Resolve(t *testing.T) {
	t.Run("Whole file", func(t *testing.T) {
		t.Parallel()
		r := &INIResolver{}

		content := `
[DEFAULT]
Key1=DefaultVal1

[SectionA]
Key2=SectionAVal2
Key3=SectionAVal3

[SectionB]
Key4=SectionBVal4
`
		p := createIniTestFile(t, content)

		val, err := r.Resolve(p)
		require.NoError(t, err)
		assert.Equal(t, strings.TrimSpace(content), val)
	})

	t.Run("Default section key", func(t *testing.T) {
		t.Parallel()
		r := &INIResolver{}

		content := `
[DEFAULT]
Key1=DefaultVal1

[SectionA]
Key2=SectionAVal2
`
		p := createIniTestFile(t, content)

		val, err := r.Resolve(p + "//Key1")
		require.NoError(t, err)
		assert.Equal(t, "DefaultVal1", val)
	})

	t.Run("Named section key", func(t *testing.T) {
		t.Parallel()
		r := &INIResolver{}

		content := `
[DEFAULT]
Key1=DefaultVal1

[SectionA]
Key2=SectionAVal2
Key3=SectionAVal3
`
		p := createIniTestFile(t, content)

		val, err := r.Resolve(p + "//SectionA.Key3")
		require.NoError(t, err)
		assert.Equal(t, "SectionAVal3", val)
	})

	t.Run("Missing key", func(t *testing.T) {
		t.Parallel()
		r := &INIResolver{}

		content := `
[DEFAULT]
Key1=DefaultVal1
`
		p := createIniTestFile(t, content)

		_, err := r.Resolve(p + "//Nope")
		require.Error(t, err)
	})

	t.Run("Missing section", func(t *testing.T) {
		t.Parallel()
		r := &INIResolver{}

		content := `
[DEFAULT]
Key1=DefaultVal1
`
		p := createIniTestFile(t, content)

		_, err := r.Resolve(p + "//Ghost.Key")
		require.Error(t, err)
	})

	t.Run("File not found", func(t *testing.T) {
		t.Parallel()
		r := &INIResolver{}
		_, err := r.Resolve(filepath.Join(t.TempDir(), "nonexistent.ini"))
		require.Error(t, err)
	})

	t.Run("Expands env vars in path", func(t *testing.T) {
		r := &INIResolver{}

		dir := t.TempDir()
		t.Setenv("DIR", dir)
		p := filepath.Join(dir, "cfg.ini")
		content := `
[DEFAULT]
X=42
`
		require.NoError(t, os.WriteFile(p, []byte(content), 0o666))

		val, err := r.Resolve("${DIR}/cfg.ini//X")
		require.NoError(t, err)
		assert.Equal(t, "42", val)
	})

	t.Run("INI scheme via default registry", func(t *testing.T) {
		t.Parallel()
		content := `
[DEFAULT]
Key1=V1

[Web]
Port=8080
`
		p := createIniTestFile(t, content)

		// default section
		got, err := ResolveVariable("ini:" + p + "//Key1")
		require.NoError(t, err)
		assert.Equal(t, "V1", got)

		// named section
		got, err = ResolveVariable("ini:" + p + "//Web.Port")
		require.NoError(t, err)
		assert.Equal(t, "8080", got)
	})

	t.Run("CRLF handling", func(t *testing.T) {
		t.Parallel()
		r := &INIResolver{}

		content := "[DEFAULT]\r\nA=1\r\n\r\n[Sec]\r\nB=2\r\n"
		p := createIniTestFile(t, content)

		val, err := r.Resolve(p + "//Sec.B")
		require.NoError(t, err)
		assert.Equal(t, "2", val)

		if runtime.GOOS == "windows" {
			all, err := r.Resolve(p)
			require.NoError(t, err)
			assert.Equal(t, strings.TrimSpace(content), all)
		}
	})
}
