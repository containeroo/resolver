package resolver

import (
	"fmt"
	"strings"
)

// ResolveString replaces ${...} tokens in s using the registry (max 8 passes).
// Use \${ to emit a literal ${. A bare '$' not followed by '{' is literal.
// Malformed tokens (missing '}' or empty ${}) return ErrBadPath.
func (r *Registry) ResolveString(s string) (string, error) {
	return r.resolveStringDepth(s, 8)
}

// resolveStringDepth performs up to maxDepth interpolation passes.
// Each pass scans left-to-right, replacing tokens found in that pass.
func (r *Registry) resolveStringDepth(s string, maxDepth int) (string, error) {
	out := s

	for range maxDepth {
		var b strings.Builder
		b.Grow(len(out))
		expanded := false // set to true only when a ${...} token is expanded

		for p := 0; p < len(out); {
			dollarRel := strings.IndexByte(out[p:], '$')
			if dollarRel < 0 {
				// no more '$' -> write tail and finish this pass
				b.WriteString(out[p:])
				break
			}
			dollar := p + dollarRel

			// \${ -> emit "${" (drop the backslash); do NOT mark expanded
			if isEscapedDollarBrace(out, p, dollar) {
				b.WriteString(out[p : dollar-1]) // exclude the backslash
				b.WriteString("${")
				p = dollar + 2 // skip "\${"
				continue
			}

			// write up to '$'
			b.WriteString(out[p:dollar])

			// not a token → literal '$'
			if !isTokenStart(out, dollar) {
				b.WriteByte('$')
				p = dollar + 1
				continue
			}

			// ${...} token bounds & validation
			start, end, err := tokenBounds(out, dollar)
			if err != nil {
				return "", err
			}
			token := out[start:end]

			// resolve token
			val, err := r.ResolveVariable(token)
			if err != nil {
				return "", fmt.Errorf("resolve ${%s}: %w", token, err)
			}

			b.WriteString(val)
			p = end + 1
			expanded = true
		}

		// If no ${...} expanded (only literals/escapes handled), return the built string.
		if !expanded {
			return b.String(), nil
		}
		out = b.String()
	}

	// Max depth reached. If tokens remain, it's a cycle or too-deep nesting.
	if strings.Contains(out, "${") {
		return "", fmt.Errorf("%w: interpolation depth exceeded", ErrBadPath)
	}
	return out, nil
}

// isEscapedDollarBrace reports whether out has "\${" with '\' immediately before '$'.
func isEscapedDollarBrace(out string, p, dollar int) bool {
	return dollar > p && out[dollar-1] == '\\' && // escaped backslash
		dollar+1 < len(out) && // avoid out-of-bounds
		out[dollar+1] == '{' // '${' immediately after '\'
}

// isTokenStart reports whether "$" at index dollar begins a "${...}" token.
func isTokenStart(out string, dollar int) bool {
	return dollar+1 < len(out) && // avoid out-of-bounds
		out[dollar+1] == '{' // '${' immediately after '$'
}

// tokenBounds returns [start,end) of the token contents inside "${...}" and validates it.
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
