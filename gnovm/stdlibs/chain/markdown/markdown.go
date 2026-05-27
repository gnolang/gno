// Package markdown is the Go-side implementation of the chain/markdown
// gno stdlib. See markdown.gno for the public contract.
package markdown

import (
	"strings"
	"unicode/utf8"
)

// ---------- StripBidiAndZeroWidth ----------

func StripBidiAndZeroWidth(s string) string {
	// Fast path: all stripped codepoints lie in U+200B..U+FEFF, all of
	// which require a 3-byte UTF-8 sequence whose lead byte is 0xE2 or
	// 0xEF. Check the bytes directly (strings.ContainsAny is rune-based
	// and would mis-handle these continuation-expecting bytes).
	if !containsAnyByte(s, 0xe2, 0xef) {
		return s
	}
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); {
		r, sz := utf8.DecodeRuneInString(s[i:])
		if isBidiOrZeroWidth(r) {
			i += sz
			continue
		}
		out = append(out, s[i:i+sz]...)
		i += sz
	}
	return string(out)
}

func isBidiOrZeroWidth(r rune) bool {
	switch {
	case r >= 0x200B && r <= 0x200F: // ZWSP, ZWNJ, ZWJ, LRM, RLM
		return true
	case r >= 0x202A && r <= 0x202E: // LRE, RLE, PDF, LRO, RLO
		return true
	case r >= 0x2066 && r <= 0x2069: // LRI, RLI, FSI, PDI
		return true
	case r == 0xFEFF: // BOM / ZWNBSP
		return true
	}
	return false
}

// ---------- NormalizeBreaks ----------

func NormalizeBreaks(s string) string {
	if !strings.ContainsAny(s, "\r") {
		return s
	}
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '\r' {
			out = append(out, '\n')
			if i+1 < len(s) && s[i+1] == '\n' {
				i++ // skip the \n in \r\n
			}
			continue
		}
		out = append(out, s[i])
	}
	return string(out)
}

// ---------- EscapeInline / EscapeTitle ----------

var inlineEscapeSet [128]bool
var titleEscapeSet [128]bool

func init() {
	for _, c := range []byte{'\\', '*', '_', '[', ']', '(', ')', '~', '>', '-', '+', '.', '!', '`', '#', '<', '&'} {
		inlineEscapeSet[c] = true
	}
	titleEscapeSet = inlineEscapeSet
	titleEscapeSet['"'] = true
	titleEscapeSet['\''] = true
}

func EscapeInline(s string) string { return escapeWithSet(s, &inlineEscapeSet) }
func EscapeTitle(s string) string  { return escapeWithSet(s, &titleEscapeSet) }

func escapeWithSet(s string, set *[128]bool) string {
	// Pre-scan: skip allocation if no byte needs escaping and no NUL.
	work := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == 0 || (c < 128 && set[c]) {
			work = true
			break
		}
	}
	if !work {
		return s
	}
	out := make([]byte, 0, len(s)+8)
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == 0 {
			out = append(out, "\xef\xbf\xbd"...) // U+FFFD
			continue
		}
		if c < 128 && set[c] {
			out = append(out, '\\')
		}
		out = append(out, c)
	}
	return string(out)
}

// ---------- PercentEncodeURL ----------

func PercentEncodeURL(s string) string {
	work := false
	for i := 0; i < len(s); i++ {
		if needsPercentEncode(s, i) {
			work = true
			break
		}
	}
	if !work {
		return s
	}
	out := make([]byte, 0, len(s)+16)
	for i := 0; i < len(s); i++ {
		if needsPercentEncode(s, i) {
			c := s[i]
			out = append(out, '%', hexDigit(c>>4), hexDigit(c&0xF))
			continue
		}
		out = append(out, s[i])
	}
	return string(out)
}

func needsPercentEncode(s string, i int) bool {
	c := s[i]
	if c <= 0x20 || c == 0x7F || c >= 0x80 {
		return true
	}
	switch c {
	case '"', '\'', '(', ')', '<', '>', '\\', '`', '{', '|', '}', '^':
		return true
	case '%':
		// Bare % (not followed by two hex digits) gets encoded.
		if i+2 >= len(s) || !isHex(s[i+1]) || !isHex(s[i+2]) {
			return true
		}
	}
	return false
}

func isHex(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'A' && c <= 'F') || (c >= 'a' && c <= 'f')
}

func hexDigit(n byte) byte {
	if n < 10 {
		return '0' + n
	}
	return 'A' + (n - 10)
}

// ---------- MatchCharsetN ----------

func MatchCharsetN(s string, firstLo, firstHi, restLo, restHi uint64, minLen, maxLen int) bool {
	n := len(s)
	if n < minLen || n > maxLen {
		return false
	}
	if n == 0 {
		return minLen == 0
	}
	if !inBitmap(s[0], firstLo, firstHi) {
		return false
	}
	for i := 1; i < n; i++ {
		if !inBitmap(s[i], restLo, restHi) {
			return false
		}
	}
	return true
}

func inBitmap(c byte, lo, hi uint64) bool {
	if c >= 128 {
		return false
	}
	if c < 64 {
		return (lo>>uint(c))&1 != 0
	}
	return (hi>>uint(c-64))&1 != 0
}

// ---------- CodeFence ----------

func CodeFence(content string, minCount int) string {
	if minCount < 1 {
		minCount = 1
	}
	longest, cur := 0, 0
	for i := 0; i < len(content); i++ {
		if content[i] == '`' {
			cur++
			if cur > longest {
				longest = cur
			}
		} else {
			cur = 0
		}
	}
	n := longest + 1
	if n < minCount {
		n = minCount
	}
	out := make([]byte, n)
	for i := range out {
		out[i] = '`'
	}
	return string(out)
}

// ---------- EscapeBlockHazards ----------

func EscapeBlockHazards(s string) string {
	if s == "" {
		return s
	}
	// Fold Unicode separators U+2028, U+2029, U+0085 NEL to \n before
	// line splitting. These are not CM §2.2 line endings, so NormalizeBreaks
	// doesn't touch them; in block context we treat them as paragraph-
	// internal breaks.
	s = foldUnicodeSeparators(s)

	var out strings.Builder
	out.Grow(len(s) + 16)

	lines := strings.Split(s, "\n")
	trailingNewline := strings.HasSuffix(s, "\n")
	if trailingNewline {
		lines = lines[:len(lines)-1] // strip artifact of split
	}

	var (
		inFence       bool
		fenceChar     byte
		fenceLen      int
		prevNonBlank  bool
	)

	for idx, line := range lines {
		writeNL := idx < len(lines)-1 || trailingNewline

		if inFence {
			out.WriteString(line)
			if writeNL {
				out.WriteByte('\n')
			}
			if isCloseFence(line, fenceChar, fenceLen) {
				inFence = false
			}
			prevNonBlank = line != ""
			continue
		}

		// LRD definition: strip entirely.
		if isLRDDefinition(line) {
			prevNonBlank = false
			continue
		}

		// Extension delimiter lines: prefix with backslash so the line's
		// opening `<` becomes a CM §2.4 inline escape. A leading space
		// would be ineffective because gnoweb's extension parsers call
		// util.TrimLeftSpace before tag matching (see ext_columns.go,
		// ext_alert.go etc.) — the space gets stripped and the parser
		// sees the bare tag. Backslash survives util.TrimLeftSpace
		// (which only strips ASCII whitespace and form-feed) and
		// goldmark's Type-7 HTML block detection (which requires the
		// first non-whitespace char to be `<`, not `\`).
		if isExtDelimiter(line) {
			out.WriteByte('\\')
			out.WriteString(escapeRefLinkUse(line))
			if writeNL {
				out.WriteByte('\n')
			}
			prevNonBlank = true
			continue
		}

		// Setext underline (only if previous line was non-blank).
		if prevNonBlank && isSetextUnderline(line) {
			out.WriteByte('\\')
			out.WriteString(line)
			if writeNL {
				out.WriteByte('\n')
			}
			prevNonBlank = true
			continue
		}

		// Block markers / fence open.
		escaped, fc, fl := escapeLineLeader(line)
		escaped = escapeRefLinkUse(escaped)
		out.WriteString(escaped)
		if writeNL {
			out.WriteByte('\n')
		}
		if fc != 0 {
			inFence = true
			fenceChar = fc
			fenceLen = fl
		}
		prevNonBlank = line != ""
	}

	// Auto-close any open code fence at end-of-input.
	if inFence {
		if out.Len() > 0 && out.String()[out.Len()-1] != '\n' {
			out.WriteByte('\n')
		}
		for i := 0; i < fenceLen; i++ {
			out.WriteByte(fenceChar)
		}
		out.WriteByte('\n')
	}

	return out.String()
}

// foldUnicodeSeparators replaces U+2028, U+2029 (3-byte UTF-8 starting
// with 0xE2 0x80) and U+0085 NEL (2-byte UTF-8 0xC2 0x85) with '\n'.
func foldUnicodeSeparators(s string) string {
	if !containsAnyByte(s, 0xe2, 0xc2) {
		return s
	}
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); {
		r, sz := utf8.DecodeRuneInString(s[i:])
		if r == 0x2028 || r == 0x2029 || r == 0x0085 {
			out = append(out, '\n')
			i += sz
			continue
		}
		out = append(out, s[i:i+sz]...)
		i += sz
	}
	return string(out)
}

// isLRDDefinition matches lines shaped like `^ {0,3}\[[^\]]+\]:\s+\S`.
// Multi-line LRDs with continuation titles are not handled (the
// first-line strip is enough for the security goal — the LRD body
// becomes orphaned text, which renders as a normal paragraph).
func isLRDDefinition(line string) bool {
	i := 0
	for i < len(line) && i < 3 && line[i] == ' ' {
		i++
	}
	if i >= len(line) || line[i] != '[' {
		return false
	}
	i++
	closeBracket := -1
	for j := i; j < len(line); j++ {
		if line[j] == ']' {
			closeBracket = j
			break
		}
	}
	if closeBracket < 0 || closeBracket == i || closeBracket+1 >= len(line) {
		return false
	}
	if line[closeBracket+1] != ':' {
		return false
	}
	rest := line[closeBracket+2:]
	// require at least one whitespace then a non-whitespace
	j := 0
	for j < len(rest) && (rest[j] == ' ' || rest[j] == '\t') {
		j++
	}
	if j == 0 || j >= len(rest) {
		return false
	}
	return true
}

// isExtDelimiter recognises any gnoweb structural-extension delimiter
// shape via the open/close `<gno-...>` / `</gno-...>` prefixes. The
// wildcard is intentionally over-permissive: a malformed line like
// `<gno-x asdf` still trips it, which is fine — the only effect is
// that the line gets a leading backslash prepended (rendered as a
// literal `<` per CM §2.4). Future extensions (`<gno-card>`,
// `<gno-foreign>`, anything later) auto-cover without needing a
// sanitize-side update.
//
// Bare `|||` (the legacy `<gno-columns>` shorthand) is intentionally
// NOT matched here — the shorthand has been removed from the columns
// parser, so user content writing `|||` is now harmless paragraph
// text and doesn't need neutralisation.
func isExtDelimiter(line string) bool {
	trim := strings.TrimLeft(line, " \t")
	return strings.HasPrefix(trim, "<gno-") || strings.HasPrefix(trim, "</gno-")
}

// isSetextUnderline reports whether line is shaped like ^ {0,3}=+[ \t]*$
// or ^ {0,3}-+[ \t]*$. Only valid following a non-blank line.
func isSetextUnderline(line string) bool {
	i := 0
	for i < len(line) && i < 3 && line[i] == ' ' {
		i++
	}
	if i >= len(line) {
		return false
	}
	marker := line[i]
	if marker != '=' && marker != '-' {
		return false
	}
	for i < len(line) && line[i] == marker {
		i++
	}
	for i < len(line) {
		if line[i] != ' ' && line[i] != '\t' {
			return false
		}
		i++
	}
	return true
}

// escapeLineLeader escapes a single line-leading block marker by
// prefixing it with \. Returns the escaped line. If the line opens
// a code fence, also returns the fence char and length (>=3).
func escapeLineLeader(line string) (string, byte, int) {
	i := 0
	for i < len(line) && i < 3 && line[i] == ' ' {
		i++
	}
	if i >= len(line) {
		return line, 0, 0
	}
	c := line[i]
	switch c {
	case '#':
		// ATX heading: # to ###### followed by space, EOL, or tab
		j := i
		for j < len(line) && j-i < 6 && line[j] == '#' {
			j++
		}
		if j-i >= 1 && j-i <= 6 && (j == len(line) || line[j] == ' ' || line[j] == '\t') {
			return line[:i] + "\\" + line[i:], 0, 0
		}
	case '>':
		return line[:i] + "\\" + line[i:], 0, 0
	case '-', '*', '_':
		// Thematic break: 3+ of -, *, or _ optionally separated by spaces.
		if isThematicBreak(line, i) {
			return line[:i] + "\\" + line[i:], 0, 0
		}
		// Bullet list marker: - or * followed by space.
		if (c == '-' || c == '*') && i+1 < len(line) && (line[i+1] == ' ' || line[i+1] == '\t') {
			return line[:i] + "\\" + line[i:], 0, 0
		}
	case '+':
		if i+1 < len(line) && (line[i+1] == ' ' || line[i+1] == '\t') {
			return line[:i] + "\\" + line[i:], 0, 0
		}
	case '|':
		// GFM table-row line leader. gnoweb's NewGnoExtension does
		// not load the GFM Table extension today, so a line-leading
		// `|` is harmless there — but Block aims to be a portable
		// defensive primitive: any goldmark consumer that DOES
		// enable tables could otherwise have user-content table
		// rows injected at document level (e.g. an empty 2-cell row
		// `|||` slotted into an existing table). Prepend `\` so the
		// line renders as literal text in every renderer.
		return line[:i] + "\\" + line[i:], 0, 0
	case '`', '~':
		// Fenced code block: 3+ of ` or ~. Code fences are legitimate
		// user content (users want to share code); do NOT escape them.
		// Just mark the fence as open so the EOF auto-close can fire if
		// the user never closes it.
		j := i
		for j < len(line) && line[j] == c {
			j++
		}
		if j-i >= 3 {
			return line, c, j - i
		}
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		// Ordered list marker: 1-9 digits followed by . or ) and space.
		j := i
		for j < len(line) && j-i < 9 && line[j] >= '0' && line[j] <= '9' {
			j++
		}
		if j < len(line) && (line[j] == '.' || line[j] == ')') &&
			j+1 < len(line) && (line[j+1] == ' ' || line[j+1] == '\t') {
			// Escape the digit run by prefixing the delimiter with \.
			return line[:j] + "\\" + line[j:], 0, 0
		}
	}
	return line, 0, 0
}

func isThematicBreak(line string, start int) bool {
	c := line[start]
	if c != '-' && c != '*' && c != '_' {
		return false
	}
	count := 0
	for i := start; i < len(line); i++ {
		if line[i] == c {
			count++
		} else if line[i] != ' ' && line[i] != '\t' {
			return false
		}
	}
	return count >= 3
}

// isCloseFence reports whether line is a valid close fence for the
// current open fence (same char, length >= open, optional trailing
// whitespace, leading whitespace <= 3).
func isCloseFence(line string, fenceChar byte, fenceLen int) bool {
	i := 0
	for i < len(line) && i < 3 && line[i] == ' ' {
		i++
	}
	count := 0
	for i < len(line) && line[i] == fenceChar {
		count++
		i++
	}
	if count < fenceLen {
		return false
	}
	for i < len(line) {
		if line[i] != ' ' && line[i] != '\t' {
			return false
		}
		i++
	}
	return true
}

// containsAnyByte reports whether s contains any of the given bytes.
// Byte-level (not rune-level) — required for UTF-8 lead-byte fast paths.
func containsAnyByte(s string, bs ...byte) bool {
	for i := 0; i < len(s); i++ {
		for _, b := range bs {
			if s[i] == b {
				return true
			}
		}
	}
	return false
}

// escapeRefLinkUse neutralizes reference-link uses [text][label],
// [text][], and footnote-ref invocations [^name] by escaping the
// offending [ — prevents user content from invoking realm-defined
// LRDs through label collision or footnote-definitions through
// footnote-namespace pollution.
//
// CM §2.4 governs backslash escapes: backslashes themselves can be
// escaped, and the parser counts pairs. So if the input is `\[^x]`,
// a naive `ReplaceAll("[^", "\\[^")` would yield `\\[^x]` which
// goldmark reads as a literal backslash followed by `[^x]` — the
// footnote ref survives. We avoid that by tracking the running count
// of immediately-preceding backslashes and, if it is odd, prepending
// one more `\` before the escape so the total parity stays odd.
func escapeRefLinkUse(line string) string {
	if !strings.Contains(line, "][") && !strings.Contains(line, "[^") {
		return line
	}
	var b strings.Builder
	b.Grow(len(line) + 16)
	trailingBackslashes := 0
	for i := 0; i < len(line); i++ {
		c := line[i]
		needsEscape := c == '[' &&
			((i+1 < len(line) && line[i+1] == '^') ||
				(i > 0 && line[i-1] == ']'))
		if needsEscape {
			if trailingBackslashes%2 == 1 {
				b.WriteByte('\\')
				trailingBackslashes++
			}
			b.WriteByte('\\')
			trailingBackslashes++
		}
		b.WriteByte(c)
		if c == '\\' {
			trailingBackslashes++
		} else {
			trailingBackslashes = 0
		}
	}
	return b.String()
}
