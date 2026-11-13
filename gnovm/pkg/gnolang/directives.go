package gnolang

import "strings"

// IsDirectiveComment reports whether a comment's text is a compiler or tooling
// directive: "//line", "//extern", "//export", the "//tool:name" form that
// covers "//go:build", "//go:generate", "//go:embed" and the pragmas, or the
// block form "/*line ...*/".
//
// The rule mirrors the unexported go/ast.isDirective, plus the block form Go's
// scanner accepts anywhere (go/scanner takes a comment as a line directive when
// `lit[1] == '*' || offs == s.lineOffset` and the text after the opener begins
// with "line "). It is copied rather than called because go/ast does not export
// it and because callers need the rule to be stable rather than to track a
// toolchain upgrade.
//
// The argument is a comment token as go/scanner yields it, slashes included.
func IsDirectiveComment(text string) bool {
	if strings.HasPrefix(text, "/*") {
		// Only "line" has a block form; //tool:name directives are line
		// comments to Go.
		return strings.HasPrefix(text[2:], "line ")
	}
	return strings.HasPrefix(text, "//") && isDirectiveText(text[2:])
}

// isDirectiveText reports whether the text of a "//" comment, with the slashes
// removed, is a directive. A leading space disqualifies it, which is what keeps
// an ordinary "// see: below" comment from counting.
func isDirectiveText(c string) bool {
	if strings.HasPrefix(c, "line ") ||
		strings.HasPrefix(c, "extern ") ||
		strings.HasPrefix(c, "export ") {
		return true
	}
	// "//[a-z0-9]+:[a-z0-9]"
	colon := strings.Index(c, ":")
	if colon <= 0 || colon+1 >= len(c) {
		return false
	}
	for i := 0; i <= colon+1; i++ {
		if i == colon {
			continue
		}
		if b := c[i]; !('a' <= b && b <= 'z' || '0' <= b && b <= '9') {
			return false
		}
	}
	return true
}
