// Package foreign provides the realm-side helper for emitting the
// gno-foreign sandbox block. Realm authors wrap externally-built
// markdown (markdown returned by an interface method on a foreign
// realm, fetched from chain storage owned by another realm, etc.) in
// Foreign before flowing it into rendered output, so gnoweb renders
// the body inside its own goldmark sub-instance with structural
// extensions selectively loaded.
//
// The renderer-side contract lives in
// gno.land/pkg/gnoweb/markdown/ext_foreign.go. This helper produces
// bytes that satisfy the parser's opener requirements (CommonMark
// §4.6 Type-7 HTML block, no attribute fall-through) and neutralizes
// any literal sentinel lines in the body so the foreign markdown
// cannot terminate the outer block prematurely.
package foreign

import (
	"chain/markdown"
	"strings"
)

// Foreign wraps body in a `<gno-foreign>` ... `</gno-foreign>` sandbox
// block. The returned string is ready to concatenate into a larger
// markdown document.
//
// Three normalization steps apply to body:
//
//  1. \r\n and bare \r line endings are normalized to \n. The parser
//     uses byte-equal matching against the sentinel close tag, so
//     mixed line endings would otherwise change the match boundary.
//
//  2. Any line whose trimmed content looks like a gno-foreign tag
//     opener OR closer — bare (`<gno-foreign>`, `</gno-foreign>`) or
//     attribute-bearing (`<gno-foreign label="x">`, `</gno-foreign
//     attr="…">`, etc.) — is neutralized by HTML-escaping the leading
//     `<` to `&lt;`. The parser tokenizes line bytes literally, so the
//     escaped form is seen as text and cannot terminate the outer
//     block or open an unintended inner block.
//
//     Crucially, BOTH open-tag and close-tag attribute-bearing forms
//     are neutralized. The parser recognizes a bare `<gno-foreign>`
//     opener, a labeled `<gno-foreign label="x">` opener, and ANY
//     `</gno-foreign…>` closer (golang.org/x/net/html drops attrs on
//     end tags before our recognizer sees them, so attr-bearing
//     closers are sentinel-equivalent). Leaving any of those forms
//     un-neutralized in body bytes would let attacker-supplied
//     markdown adjust the parser's framing-depth counter and either
//     consume the helper's own close (capturing trailing realm
//     content into the sandbox) or close the outer block early
//     (escaping the sandbox entirely).
//
//     Realm authors who want genuine nested gno-foreign should
//     compose Foreign calls explicitly: Foreign(Foreign(inner)).
//
//  3. A leading and trailing blank line are emitted around the
//     opener / closer. CommonMark §4.6 forbids Type-7 HTML blocks
//     from interrupting a paragraph; without the blank line, an
//     opener following a non-blank line is absorbed into the
//     preceding paragraph instead of opening a sandbox.
//
// The renderer caps cross-family nesting at 4 levels and per-Convert
// foreign blocks at 100. Beyond those caps, the opener falls through
// to raw HTML and is stripped by the renderer's safe mode.
func Foreign(body string) string {
	return wrapForeign("", body)
}

// ForeignWithLabel wraps body like Foreign but emits an explicit
// `label="…"` attribute on the opener so the rendered sandbox carries
// a caller-supplied label (e.g., "Pulled from /r/foo") in place of
// the default "external content". The label is sanitized so it
// cannot inject HTML or break out of the attribute value:
//
//   - NUL bytes are dropped.
//   - Other control characters (U+0000–U+001F, U+007F) become spaces.
//   - `&`, `<`, `>`, and `"` are replaced with their HTML entities.
//   - Leading/trailing whitespace is trimmed.
//
// A label that is empty after sanitization behaves identically to
// Foreign (no attribute emitted; the renderer uses its default
// "external content" label).
func ForeignWithLabel(label, body string) string {
	return wrapForeign(label, body)
}

func wrapForeign(rawLabel, body string) string {
	label := sanitizeLabel(rawLabel)

	// Normalize line endings (CR/CRLF → LF). The parser matches the
	// sentinel close against \n-delimited lines, so mixed line endings
	// would otherwise shift the match boundary. CR/CRLF → LF ONLY: do
	// not fold Unicode separators here — they must stay verbatim in the
	// body so the inner renderer sees the foreign markdown unaltered.
	body = markdown.NormalizeBreaks(body)

	// Mangle any line that would terminate the outer block or open
	// an inner one. Covers bare and attribute-bearing forms of both
	// the opener and the closer (see step 2 in the package doc).
	var b strings.Builder
	// b accumulates only the body lines (the opener/closer envelope is
	// concatenated separately below), so len(body) is the exact size in
	// the common case. Sentinel lines that expand `<`→`&lt;` may force
	// one growth — rare enough not to pre-size for.
	b.Grow(len(body))
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		if isForeignSentinelLine(trimSentinel(line)) {
			// Escape just the leading `<` so the html tokenizer
			// sees this as text instead of a tag. Preserve any 0-3
			// leading spaces the parser's trim would have stripped.
			idx := strings.Index(line, "<")
			if idx >= 0 {
				line = line[:idx] + "&lt;" + line[idx+1:]
			}
		}
		b.WriteString(line)
		if i < len(lines)-1 {
			b.WriteByte('\n')
		}
	}

	opener := "<gno-foreign>"
	if label != "" {
		opener = `<gno-foreign label="` + label + `">`
	}
	return "\n\n" + opener + "\n" + b.String() + "\n</gno-foreign>\n\n"
}

// sanitizeLabel makes a user-supplied label safe to splice into an
// HTML attribute value on the gno-foreign opener line.
func sanitizeLabel(s string) string {
	// Strip bidi-override and zero-width controls FIRST — same ordering
	// as the sanitize package's HTMLEscape — so invisible reordering or
	// zero-width payloads can't survive into the rendered label.
	s = markdown.StripBidiAndZeroWidth(s)
	// Drop NUL; map other ASCII controls AND the Unicode line/paragraph
	// separators (U+2028, U+2029, U+0085 NEL) to spaces. The opener is a
	// single line, so any of these surviving in the label would either
	// add a control payload or, for the separators, render as a stray
	// line break inside the attribute.
	s = strings.Map(func(r rune) rune {
		if r == 0 {
			return -1
		}
		if r < 0x20 || r == 0x7f || r == 0x2028 || r == 0x2029 || r == 0x0085 {
			return ' '
		}
		return r
	}, s)
	// Escape `&` first so subsequent entity bytes don't get
	// re-escaped.
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return strings.TrimSpace(s)
}

// isForeignSentinelLine reports whether s (already trimmed via
// trimSentinel) looks like a gno-foreign opener or closer that the
// parser would recognize. Matches both bare forms (`<gno-foreign>`,
// `</gno-foreign>`) and attribute-bearing forms (`<gno-foreign attr…>`,
// `</gno-foreign attr…>`). Over-inclusive on purpose: even forms the
// parser would reject (e.g., unknown attribute keys on the opener)
// are mangled, since the only downside is that the user-visible
// literal text shows verbatim instead of being raw-HTML-stripped.
//
// The tag-name match is case-INSENSITIVE. goldmark's html.Tokenizer
// lowercases tag names before the renderer-side parser
// (parseForeignLineTag in ext_foreign.go) compares them, so the parser
// recognizes `<GNO-FOREIGN>`, `<Gno-Foreign>`, etc. as sentinels. This
// escaper MUST mirror that case-folding — otherwise a case-variant
// sentinel slips past here and reaches the (case-insensitive) parser,
// where it either closes the outer block early (escaping the sandbox)
// or opens a phantom inner block that consumes the helper's own close
// (capturing trailing realm content).
func isForeignSentinelLine(s string) bool {
	const (
		openPrefix  = "<gno-foreign"
		closePrefix = "</gno-foreign"
	)
	var rest string
	switch {
	case hasASCIIFoldPrefix(s, closePrefix):
		rest = s[len(closePrefix):]
	case hasASCIIFoldPrefix(s, openPrefix):
		rest = s[len(openPrefix):]
	default:
		return false
	}
	if len(rest) == 0 {
		return false // unterminated tag, parser wouldn't match either
	}
	// Tag name must end at '>' (bare form) or whitespace (attribute
	// form). Anything else (e.g., `<gno-foreign-other>`) is a
	// different tag — leave it alone.
	c := rest[0]
	return c == '>' || c == ' ' || c == '\t'
}

// hasASCIIFoldPrefix reports whether s begins with prefix, comparing
// ASCII letters case-insensitively. prefix must be lowercase ASCII;
// folding is ASCII-only on purpose (Unicode case folding would
// over-match, and the sentinel envelope is pure ASCII anyway).
func hasASCIIFoldPrefix(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	for i := 0; i < len(prefix); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		if c != prefix[i] {
			return false
		}
	}
	return true
}

// trimSentinel returns line with the leading 0-3 spaces and trailing
// ASCII whitespace that the parser's trimForeignLine strips. Mirrors
// the byte-level trim the parser performs so this helper detects the
// same sentinel match the parser would.
func trimSentinel(s string) string {
	i := 0
	for i < len(s) && i < 3 && s[i] == ' ' {
		i++
	}
	s = s[i:]
	for len(s) > 0 {
		c := s[len(s)-1]
		if c == ' ' || c == '\t' || c == '\n' || c == '\v' || c == '\f' || c == '\r' {
			s = s[:len(s)-1]
			continue
		}
		break
	}
	return s
}
