package resolver

import (
	"fmt"
	"strings"
)

// ResolveString replaces ${...} tokens in s using the registry (max 8 passes).
// Use \${ to emit a literal ${. A bare '$' not followed by '{' is literal.
// Malformed tokens (missing '}' or empty ${}) return ErrBadPath.
func (r *Registry) ResolveString(s string) (string, error) {
	r.mu.RLock()
	depth := r.maxDepth
	r.mu.RUnlock()
	return r.resolveStringDepth(s, depth)
}

func (r *Registry) resolveStringDepth(s string, maxDepth int) (string, error) {
	out := s
	for range maxDepth {
		var b strings.Builder
		b.Grow(len(out))
		expanded := false
		for p := 0; p < len(out); {
			dollarRel := strings.IndexByte(out[p:], '$')
			if dollarRel < 0 {
				b.WriteString(out[p:])
				break
			}
			dollar := p + dollarRel
			if isEscapedDollarBrace(out, p, dollar) {
				// Keep the escape until all passes finish so it is never expanded.
				b.WriteString(out[p:dollar])
				b.WriteString("${")
				p = dollar + 2
				continue
			}
			b.WriteString(out[p:dollar])
			if !isTokenStart(out, dollar) {
				b.WriteByte('$')
				p = dollar + 1
				continue
			}
			start, end, err := tokenBounds(out, dollar)
			if err != nil {
				return "", err
			}
			token := out[start:end]
			val, err := r.ResolveVariable(token)
			if err != nil {
				return "", fmt.Errorf("resolve ${%s}: %w", token, err)
			}
			b.WriteString(val)
			p = end + 1
			expanded = true
		}
		if !expanded {
			return unescapeExpressions(b.String()), nil
		}
		out = b.String()
	}
	if containsUnescapedExpression(out) {
		return "", fmt.Errorf("%w: interpolation depth exceeded", ErrBadPath)
	}
	return unescapeExpressions(out), nil
}

func isEscapedDollarBrace(out string, p, dollar int) bool {
	return dollar > p && out[dollar-1] == '\\' && dollar+1 < len(out) && out[dollar+1] == '{'
}
func isTokenStart(out string, dollar int) bool {
	return dollar+1 < len(out) && out[dollar+1] == '{'
}
func tokenBounds(out string, dollar int) (start, end int, err error) {
	start = dollar + 2
	closeRel := strings.IndexByte(out[start:], '}')
	if closeRel < 0 {
		return 0, 0, fmt.Errorf("%w: missing closing '}' at offset %d", ErrBadPath, dollar)
	}
	end = start + closeRel
	if strings.TrimSpace(out[start:end]) == "" {
		return 0, 0, fmt.Errorf("%w: empty ${} at offset %d", ErrBadPath, dollar)
	}
	return start, end, nil
}

// containsUnescapedExpression checks only active interpolation tokens.
func containsUnescapedExpression(s string) bool {
	for i := 0; i < len(s)-1; i++ {
		if s[i] == '$' && s[i+1] == '{' && (i == 0 || s[i-1] != '\\') {
			return true
		}
	}
	return false
}

// unescapeExpressions removes the escape only when expansion is complete.
func unescapeExpressions(s string) string { return strings.ReplaceAll(s, `\${`, `${`) }
