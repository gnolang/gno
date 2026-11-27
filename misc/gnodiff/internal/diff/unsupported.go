package diff

import (
	"regexp"
	"strings"
)

// Unsupported reports why src is outside the language subset this fuzzer
// compares. The empty string means the source is in scope (it may still
// fail to type-check). The check is a token scan, used to keep the native
// fuzzer from filing a "divergence" for every mutant that introduces a
// channel or a goroutine — those are known non-goals, listed in the README.
func Unsupported(src string) string {
	stripped := stripCommentsAndStrings(src)
	for _, p := range unsupportedPatterns {
		if p.re.MatchString(stripped) {
			return p.why
		}
	}
	for _, name := range []string{"min", "max", "clear"} {
		call := regexp.MustCompile(`\b` + name + `\s*\(`)
		decl := regexp.MustCompile(`\bfunc\s+` + name + `\b`)
		if call.MatchString(stripped) && !decl.MatchString(stripped) {
			return name + " builtin is not in the Gno subset (added after Go 1.17)"
		}
	}
	return ""
}

type pattern struct {
	re  *regexp.Regexp
	why string
}

var unsupportedPatterns = []pattern{
	{regexp.MustCompile(`\bchan\b`), "channels are not in the Gno subset"},
	{regexp.MustCompile(`\bselect\b`), "select is not in the Gno subset"},
	{regexp.MustCompile(`\bgo\b`), "goroutines are not in the Gno subset"},
	{regexp.MustCompile(`\bunsafe\b`), "unsafe is not in the Gno subset"},
	{regexp.MustCompile(`\bcomplex(?:64|128)\b`), "complex numbers are not in the Gno subset"},
	{regexp.MustCompile(`\bimport\b`), "imports are out of scope (no shared stdlib subset yet)"},
	{regexp.MustCompile(`\bfunc\s+[A-Za-z_]\w*\[`), "type parameters are not in the Gno subset"},
	{regexp.MustCompile(`\btype\s+[A-Za-z_]\w*\[`), "type parameters are not in the Gno subset"},
	{regexp.MustCompile(`\brange\s+[0-9]`), "integer range (Go 1.22) is not targeted by this fuzzer"},
}

// stripCommentsAndStrings blanks comments and string contents so keyword
// checks do not fire on prose or on data.
func stripCommentsAndStrings(src string) string {
	var b strings.Builder
	b.Grow(len(src))
	inLine, inBlock, inStr, inRaw, escape := false, false, false, false, false
	for i := 0; i < len(src); i++ {
		c := src[i]
		var next byte
		if i+1 < len(src) {
			next = src[i+1]
		}
		switch {
		case inLine:
			if c == '\n' {
				inLine = false
				b.WriteByte('\n')
			}
		case inBlock:
			if c == '*' && next == '/' {
				inBlock = false
				i++
			}
		case inRaw:
			if c == '`' {
				inRaw = false
			}
		case inStr:
			if escape {
				escape = false
				continue
			}
			if c == '\\' {
				escape = true
				continue
			}
			if c == '"' {
				inStr = false
			}
		case c == '/' && next == '/':
			inLine = true
			i++
		case c == '/' && next == '*':
			inBlock = true
			i++
		case c == '"':
			inStr = true
		case c == '`':
			inRaw = true
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}
