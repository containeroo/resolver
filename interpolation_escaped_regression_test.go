package resolver

import (
	"errors"
	"testing"
)

func TestEscapedExpressionsAcrossPasses(t *testing.T) {
	t.Setenv("RESOLVER_ESC_TEST", "secret")
	r := NewDefaultRegistry()
	cases := []struct{ in, want string }{
		{`literal \${env:RESOLVER_ESC_TEST}`, `literal ${env:RESOLVER_ESC_TEST}`},
		{`literal \${env:RESOLVER_ESC_TEST} and ${env:RESOLVER_ESC_TEST}`, `literal ${env:RESOLVER_ESC_TEST} and secret`},
	}
	for _, tt := range cases {
		got, err := r.ResolveString(tt.in)
		if err != nil || got != tt.want {
			t.Errorf("ResolveString(%q) = (%q,%v), want %q", tt.in, got, err, tt.want)
		}
	}
	// An escaped expression must not count against the depth limit.
	got, err := r.resolveStringDepth(`\${env:RESOLVER_ESC_TEST} ${env:RESOLVER_ESC_TEST}`, 1)
	if err != nil || got != `${env:RESOLVER_ESC_TEST} secret` {
		t.Fatalf("one pass: (%q,%v)", got, err)
	}
	r.Register("loop:", ResolverFunc(func(v string) (string, error) { return "${loop:" + v + "}", nil }))
	if _, err = r.ResolveString("${loop:x}"); !errors.Is(err, ErrBadPath) {
		t.Fatalf("expected depth error, got %v", err)
	}
}
