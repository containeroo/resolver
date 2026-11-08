package resolver

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWholeFileWhitespacePolicy(t *testing.T) {
	path := filepath.Join(t.TempDir(), "secret.txt")
	if err := os.WriteFile(path, []byte("  token  \r\n"), 0600); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		policy FileWhitespacePolicy
		want   string
	}{
		{FileTrimSpace, "token"},
		{FileTrimNewlines, "  token  "},
		{FilePreserveWhitespace, "  token  \r\n"},
	}
	for _, tt := range tests {
		r := &KeyValueFileResolver{Whitespace: tt.policy}
		got, err := r.Resolve(path)
		if err != nil || got != tt.want {
			t.Errorf("policy %v: (%q,%v), want %q", tt.policy, got, err, tt.want)
		}
	}
}
