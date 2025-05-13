package resolver

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createTomlTestFile(t *testing.T) string {
	t.Helper()

	tempDir := t.TempDir()
	testFilePath := filepath.Join(tempDir, "config.toml")

	fileContent := `emptyString = ""

[server]
host = "localhost"
port = 8080
[server.nested]
key = "value"

[[servers]]
host = "example.com"
port = 80

[[servers]]
host = "example.org"
port = 443

[nonString]
inner = true
`

	err := os.WriteFile(testFilePath, []byte(fileContent), 0666)
	assert.NoError(t, err, "failed to create test TOML file")

	return testFilePath
}

func TestTOMLResolver_Resolve(t *testing.T) {
	t.Parallel()
	resolver := &TOMLResolver{}

	t.Run("Resolve entire file", func(t *testing.T) {
		t.Parallel()

		testFilePath := createTomlTestFile(t)
		val, err := resolver.Resolve(testFilePath)
		assert.NoError(t, err, "unexpected error resolving entire TOML file")

		expected := `emptyString = ""

[server]
host = "localhost"
port = 8080
[server.nested]
key = "value"

[[servers]]
host = "example.com"
port = 80

[[servers]]
host = "example.org"
port = 443

[nonString]
inner = true`
		assert.Equal(t, expected, val)
	})

	t.Run("Resolve top-level key", func(t *testing.T) {
		t.Parallel()

		testFilePath := createTomlTestFile(t)
		val, err := resolver.Resolve(testFilePath + "//server.host")
		assert.NoError(t, err, "unexpected error resolving top-level key")
		assert.Equal(t, "localhost", val)
	})

	t.Run("Resolve nested key", func(t *testing.T) {
		t.Parallel()

		testFilePath := createTomlTestFile(t)
		val, err := resolver.Resolve(testFilePath + "//server.nested.key")
		assert.NoError(t, err, "unexpected error resolving nested key")
		assert.Equal(t, "value", val)
	})

	t.Run("Resolve array element", func(t *testing.T) {
		t.Parallel()

		testFilePath := createTomlTestFile(t)
		val, err := resolver.Resolve(testFilePath + "//servers.1.host")
		assert.NoError(t, err, "unexpected error resolving array element")
		assert.Equal(t, "example.org", val)
	})

	t.Run("Resolve empty string key", func(t *testing.T) {
		t.Parallel()

		testFilePath := createTomlTestFile(t)
		val, err := resolver.Resolve(testFilePath + "//emptyString")
		assert.NoError(t, err, "unexpected error resolving empty string key")
		assert.Equal(t, "", val)
	})

	t.Run("Resolve non-string value", func(t *testing.T) {
		t.Parallel()

		testFilePath := createTomlTestFile(t)
		val, err := resolver.Resolve(testFilePath + "//nonString")
		assert.NoError(t, err, "unexpected error resolving non-string value")

		expected := `inner = true`
		assert.Equal(t, expected, val)
	})

	t.Run("Resolve missing key", func(t *testing.T) {
		t.Parallel()

		testFilePath := createTomlTestFile(t)
		_, err := resolver.Resolve(testFilePath + "//server.missing")
		assert.Error(t, err, "expected an error resolving a missing key, but got none")
	})

	t.Run("Resolve non-existing file", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		nonExistentFile := filepath.Join(tempDir, "nonexistent.toml")

		_, err := resolver.Resolve(nonExistentFile)
		assert.Error(t, err, "expected an error resolving a non-existing file, but got none")
	})

	t.Run("Invalid TOML", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		testFilePath := filepath.Join(tempDir, "bad.toml")

		invalid := "= invalid"
		err := os.WriteFile(testFilePath, []byte(invalid), 0666)
		assert.NoError(t, err)

		result, err := resolver.Resolve(testFilePath)
		assert.Equal(t, "", result)
		assert.Error(t, err, "expected parse error but got none")
	})
}
