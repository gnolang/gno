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

import "strings"

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
//  2. Any line whose trimmed content is exactly the sentinel tag
//     `<gno-foreign>` or `</gno-foreign>` is HTML-escaped (the leading
//     `<` becomes `&lt;`). The parser tokenizes line bytes literally,
//     so the escaped form is seen as text and cannot terminate the
//     outer block. Realm authors who want genuine nested gno-foreign
//     should compose Foreign calls explicitly: Foreign(Foreign(inner)).
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

	// Normalize line endings.
	body = strings.ReplaceAll(body, "\r\n", "\n")
	body = strings.ReplaceAll(body, "\r", "\n")

	// Mangle any line that would terminate the outer block.
	var b strings.Builder
	b.Grow(len(body) + len("\n\n<gno-foreign>\n</gno-foreign>\n\n") + len(label) + len(` label=""`))
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		switch trimSentinel(line) {
		case "<gno-foreign>":
			line = strings.Replace(line, "<gno-foreign>", "&lt;gno-foreign&gt;", 1)
		case "</gno-foreign>":
			line = strings.Replace(line, "</gno-foreign>", "&lt;/gno-foreign&gt;", 1)
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
	// Drop NUL and map other ASCII control characters to spaces.
	s = strings.Map(func(r rune) rune {
		if r == 0 {
			return -1
		}
		if r < 0x20 || r == 0x7f {
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
