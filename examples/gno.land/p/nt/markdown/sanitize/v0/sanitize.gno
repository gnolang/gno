// Package sanitize provides input-cleaning primitives that realm authors
// wrap around user-supplied strings before flowing them into markdown
// output. Each helper targets one specific lexical slot (link text,
// heading text, URL href, table cell, HTML attribute, etc.) and neutralizes
// the bytes that would otherwise let user content break out of that slot
// or inject new top-level structure.
//
// Pick the right helper from the table under "Picking the right helper"
// below, then wrap each user-supplied argument exactly once at the call
// site (see "The audit rule").
//
// # The audit rule
//
// Exactly one sanitize.* wrapper per user-derived argument, at the call
// site. Double-wrapping is a bug. Markdown-builder packages (e.g.
// p/moul/md) do NOT sanitize internally — wrapping is the caller's
// responsibility:
//
//	out += md.H2(sanitize.InlineText(post.Title))                       // good
//	out += md.H2(post.Title)                                            // BAD: raw user input
//	out += md.H2(sanitize.InlineText(sanitize.InlineText(post.Title)))  // BAD: double-wrap
//
// # Picking the right helper
//
// Match the helper to the slot the user content lands in:
//
//	slot                                       helper
//	-------------------------------------------------------------
//	[text](url)                                InlineText (text)
//	# Heading text                             InlineText
//	**bold** _italic_                          InlineText
//	![alt](src)                                InlineText (alt)
//	> [!NOTE] one-line title                   InlineText
//	multi-paragraph post body                  Block
//	[text](url "title")                        LinkTitle (title)
//	| cell |                                   TableCell
//	<gno-card caption="X">                     HTMLEscape
//	<h5>X</h5>                                 HTMLEscape
//	any URL going into ](X)                    URL
//	any image src going into (X)               ImageURL
//	r/sys/users handle                         UserName     (validator)
//	g1.../gpub1... etc.                        BechString   (validator)
//	footnote name in [^name]:                  FootnoteRef  (validator)
//	fenced-code language tag                   LanguageName (validator)
//	prefix arg to md.Nested                    NestedPrefix (validator)
//
// # Invariants
//
// All helpers in this package are panic-free for any string input and
// run in O(len(input)) time with bounded allocation.
//
// Idempotence classes:
//
//	Idempotent (calling twice == calling once):
//	  StripBidiAndZeroWidth, NormalizeBreaks
//	  UserName, BechString, FootnoteRef, LanguageName, NestedPrefix
//	  URL, ImageURL              (accept→identity; reject→"")
//
//	NOT idempotent — never wrap an already-sanitized string:
//	  InlineText, LinkTitle, TableCell   (re-escape backslashes)
//	  HTMLEscape                         (re-escapes `&` → `&amp;`)
//	  Block                              (re-strips LRDs, may double-escape line leaders)
//
//	CodeFence is pure: same inputs always give the same output.
//
// Validators (UserName / BechString / FootnoteRef / LanguageName /
// NestedPrefix) return either the cleaned input verbatim or "". They
// never partially-sanitize: if the input doesn't match the slot's
// charset/shape, the answer is rejection.
//
// # Composition rules
//
// Safe nestings:
//
//	md.Blockquote(sanitize.Block(userProse))         block, then line-prefix
//	md.LanguageCodeBlock(sanitize.LanguageName(lang), userCode)
//	md.Link(sanitize.InlineText(text), sanitize.URL(href))
//
// Wrong:
//
//	sanitize.InlineText(sanitize.InlineText(s))   double-wrap (re-escape)
//	sanitize.TableCell(sanitize.InlineText(s))    TableCell already calls InlineText
//	sanitize.URL(sanitize.InlineText(href))       inline-escape backslash-escapes `.` `-` `_`
//	                                              inside the URL, corrupting the host/path
//	sanitize.Block(md.Blockquote(...))            Block would re-escape the `>` chars
//	sanitize.TableCell(multiParagraphProse)       newlines fold to space silently; use a
//	                                              non-table layout for multi-paragraph text
//
// # Threat model
//
// Sanitizers in this package defend against:
//
//   - bidi/zero-width injection: invisible characters that make
//     displayed text disagree with stored bytes (e.g. an address `g1abc...`
//     that renders as `g1xyz...`, or a username that visually collides
//     with another). Stripped by StripBidiAndZeroWidth, which runs as
//     the first step of every text-shaped helper.
//   - line-ending homoglyphs: CR-only and Unicode separators
//     (U+0085 NEL, U+2028, U+2029) that some renderers treat as line
//     breaks. Folded uniformly.
//   - markdown-structure injection: user content opening a heading,
//     blockquote, list, code fence, table row, link-reference def,
//     setext underline, or gnoweb extension delimiter at document level.
//   - URL scheme abuse: javascript:, data:text/html, vbscript:, blob:,
//     protocol-relative //, mailto: with prefill phishing parameters.
//     Allowlist-only (URL / ImageURL).
//   - HTML attribute / element breakout: `"`, `<`, `>`, `&`, `'` inside
//     HTML lexical slots. Handled by HTMLEscape.
//   - CommonMark §2.3 NUL: replaced with U+FFFD by InlineText, LinkTitle,
//     TableCell, and HTMLEscape.
//   - code-fence leakage: a user-opened ``` ``` ``` fence that runs to EOF
//     with no closing fence, which would otherwise swallow every realm-
//     emitted line that follows. Block auto-closes any open fence at EOF.
//   - table-alignment drift: tabs inside table cells expanding to variable
//     widths (1-4 spaces depending on column position) and shifting cell
//     boundaries unpredictably. TableCell replaces tabs with single spaces.
//
// What this package does NOT do:
//
//   - It does not store state. Every helper is a pure function.
//   - It does not validate semantic correctness. sanitize.URL accepts
//     a syntactically valid https:// URL even if the host is malicious;
//     URL reputation is a separate layer.
//   - It does not enforce CSS containment. ImageURL admits data:image/*
//     URIs on the assumption that the deploying gnoweb instance caps
//     rendered image dimensions via CSS. Without that cap, a malicious
//     image can blow out the page layout or exhaust memory.
//   - It does not perform structural sandboxing of foreign markdown.
//     If a realm concatenates an opaque markdown blob returned from a
//     polymorphic interface (`someThing.Render()`), it needs a structural
//     sandbox primitive (e.g. a `<gno-card>` extension), not just leaf
//     sanitization.
//
// # Extending
//
// A new helper added to this package MUST:
//
//  1. Be panic-free for any string input.
//  2. Strip bidi+zero-width before any other transform (so display
//     equals storage end-to-end).
//  3. Declare its idempotence class in the table above.
//  4. Document the markdown / HTML lexical slot it targets.
//  5. Reject rather than partially-sanitize when input is structurally
//     invalid (return "" — never half-process an address or URL).
//  6. Pick exactly one of the two return-value contracts and stick to
//     it: escapers always return a transformed string and never reject
//     (any input is OK — the transformation makes it safe); validators
//     return the cleaned input verbatim on accept or "" on reject and
//     never half-process. Mixing the contracts within one helper is a
//     bug — callers can't reason about whether "" means "input was
//     already empty" or "input was rejected".
package sanitize

import (
	"chain/markdown"
	"html"
	"strings"
)

// ----- Re-exports of the public chain/markdown natives -----
//
// These are general-purpose data-hygiene primitives, not markdown-specific.
// The other helpers in this package call them internally, so realms emitting
// markdown rarely need to call them directly. Reach for these when you have
// a non-markdown use case — e.g. normalizing a username before storage,
// canonicalizing a search query, or stripping invisible characters from
// any user string that will be displayed or compared.

// StripBidiAndZeroWidth removes Unicode bidi controls and zero-width
// characters (U+200B-D, U+200E-F, U+202A-E, U+2066-9, U+FEFF) from s.
// Use it when storing or comparing user-supplied strings outside of a
// markdown context — for example, before saving a display name to state,
// or before hashing a search query. Idempotent: calling twice gives the
// same result.
//
// Thin wrapper over chain/markdown.StripBidiAndZeroWidth.
func StripBidiAndZeroWidth(s string) string {
	return markdown.StripBidiAndZeroWidth(s)
}

// NormalizeBreaks unifies CR-LF and lone CR to LF (CommonMark §2.2 line
// endings only — does NOT touch U+2028/U+2029). Use it when comparing
// or hashing user input that may have been authored on different
// platforms (Windows CRLF vs. Unix LF), so equivalent strings normalize
// to the same bytes. Idempotent.
//
// Thin wrapper over chain/markdown.NormalizeBreaks.
func NormalizeBreaks(s string) string {
	return markdown.NormalizeBreaks(s)
}

// ----- Escapers -----

// InlineText prepares an arbitrary user string for an INLINE markdown
// slot — anywhere the rendered output stays on a single line and lives
// inside a larger markdown construct.
//
// Use for:
//   - link text:        [InlineText(label)](url)
//   - heading text:     # InlineText(title)
//   - bold/italic body: **InlineText(name)**
//   - image alt text:   ![InlineText(alt)](src)
//   - single-line block-context slots:
//                       > [!NOTE] InlineText(title)
//                       > Author: InlineText(name)
//
// Multi-paragraph prose belongs in Block, not InlineText. InlineText
// folds every newline to a single space (so paragraph structure is
// erased) and escapes inline-active CommonMark punctuation:
//
//	\ * _ [ ] ( ) ~ > - + . ! ` # < &
//
// Two characters are intentionally NOT escaped:
//
//   - `|` — only meaningful in GFM table rows. Leaving it literal here
//     lets TableCell (which calls InlineText then escapes `|` itself)
//     avoid double-escaping pipes into `\\|`.
//   - `=` — only meaningful as a setext heading underline, which is a
//     line-level construct. Escaping `=` inline would mangle expressions
//     like `x = 1` for no benefit.
//
// Not idempotent (see package doc).
func InlineText(s string) string {
	s = markdown.StripBidiAndZeroWidth(s)
	s = markdown.NormalizeBreaks(s)        // CM §2.2 \r\n / \r → \n
	s = foldNewlinesAndSeparators(s, ' ')  // \n + NEL + U+2028/U+2029 → space
	return markdown.EscapeInline(s)
}

// Block prepares user content for a top-level BLOCK markdown context
// where paragraphs, line breaks, code blocks, and other block structure
// should survive — but where the content must NOT be able to inject
// new top-level constructs (headings, lists, blockquotes, tables,
// link-reference definitions).
//
// Use for any multi-paragraph user-supplied prose that the realm
// concatenates into its rendered output:
//   - post bodies, comments, replies
//   - profile bios, About sections
//   - proposal descriptions, governance motions
//   - changelog entries, release notes
//
// COMPOSITION GOTCHA: Block's EOF fence-autoclose (described below)
// appends a final fence line. If you wrap Block's output with a
// line-prefixing builder like md.Blockquote (which prepends `> ` per
// line) or md.Nested, that closing fence becomes a prefixed line. The
// output is still safe (the fence still closes correctly) but may
// render awkwardly. If pixel-perfect output matters, strip a trailing
// blank fence line after Block.
//
// Behavior worth knowing:
//
//   - Line-leading block markers (`#`, `>`, `-`, `*`, `+`, `1.`) are
//     backslash-escaped so user content can't open a new heading, list,
//     blockquote, or thematic break at the document level.
//   - Setext underlines — `===` or `---` on a line by itself following
//     a non-blank line, which CommonMark interprets as "promote the
//     preceding line to a heading" — are escaped so user prose can't
//     retroactively turn a realm-emitted line into a header.
//   - Link-reference DEFINITIONS — lines shaped `[label]: url [title]`
//     ("LRD" in CommonMark §4.7) — are stripped. User content cannot
//     define a reference that other parts of the page might resolve
//     against. (Multi-line LRDs with continuation titles strip only
//     the first line; the title continuation falls through as orphan
//     text, which renders as a normal paragraph — safe but visible.)
//   - Reference-link USES — `[text][label]` syntax — get the `][` pair
//     escaped to `]\[`, so user content cannot resolve against a
//     realm-defined `[label]: ...` reference elsewhere on the page.
//   - Extension delimiter lines (`<gno-card>`, `</gno-columns>`, `|||`,
//     and similar gnoweb structural markers) get a leading space,
//     neutralizing them as extension markers while preserving them as
//     visible text.
//   - Fenced code blocks (``` ``` ```, `~~~`) opened by user content
//     are PASSED THROUGH and the content inside the fence is left
//     unescaped for markdown purposes (otherwise users couldn't share
//     code at all). Note: the global bidi-strip and Unicode-separator
//     fold (U+2028/U+2029/U+0085) still apply inside the fence — only
//     the markdown-structure escapes (line-leader, setext, LRD, ref-
//     link, extension-delimiter) are skipped there. If the user opens
//     a fence and never closes it, Block appends a matching closing
//     fence at end of input — preventing the fence from "leaking"
//     into whatever realm chrome the realm concatenates after Block's
//     output.
//   - Unicode separators U+2028, U+2029, and U+0085 NEL are folded to
//     `\n` (CommonMark §2.2 doesn't recognize them, but some renderers
//     do — folding closes that ambiguity).
//   - Inline emphasis, code spans, inline links, GFM tables, and soft
//     line breaks within a paragraph are PRESERVED — users can format.
//
// Not idempotent (see package doc): Block strips LRDs in place, so a
// second call could behave differently if the first call's output
// reintroduces an LRD-shaped line. Wrap each user-supplied string
// exactly once.
func Block(s string) string {
	s = markdown.NormalizeBreaks(s)
	s = markdown.StripBidiAndZeroWidth(s)
	return markdown.EscapeBlockHazards(s)
}

// LinkTitle prepares user content for a CommonMark link-title or
// image-title slot — the optional quoted text after the URL in any of
// these forms:
//
//   [text](url "TITLE")
//   ![alt](src "TITLE")
//   [label]: url "TITLE"
//
// Escapes the inline-active set plus `"` and `'` (the title delimiters
// that aren't already in the inline set; `(` and `)` are), so the
// caller can choose any of the three title-quote styles safely.
//
// Pick the right helper for the slot — markdown title and HTML
// attribute share the look but use different escape rules:
//
//	[text](url "X")              → LinkTitle      (markdown title)
//	<a title="X">                → HTMLEscape     (HTML attribute)
//	<h5>X</h5>                   → HTMLEscape     (HTML element body)
//
// Swapping HTMLEscape for LinkTitle is wrong: HTML's `&amp;` written
// inside a markdown title renders as the literal characters `&amp;`.
// Swapping LinkTitle for HTMLEscape is wrong: markdown's `\"` survives
// into the rendered HTML as a literal backslash-quote.
//
// Not idempotent (see package doc).
func LinkTitle(s string) string {
	s = markdown.StripBidiAndZeroWidth(s)
	s = markdown.NormalizeBreaks(s)
	s = foldNewlinesAndSeparators(s, ' ')
	return markdown.EscapeTitle(s)
}

// TableCell prepares user content for a GFM table cell — the bytes
// between two `|` column delimiters in a table row like
// `| cell-a | cell-b | cell-c |`. An unescaped `|` inside cell
// content would open a new column, letting a malicious user shift
// every column to its right.
//
// On top of InlineText's behavior, TableCell:
//   - escapes `|` to `\|` so user content can't end the cell early.
//   - replaces tabs with single spaces. CommonMark expands tabs to
//     the next multiple-of-4 column boundary (variable 1-4 spaces),
//     which would shift the displayed cell-content width unpredictably
//     and confuse table alignment.
//
// Not idempotent (see package doc).
func TableCell(s string) string {
	s = InlineText(s)
	s = strings.ReplaceAll(s, "\t", " ")
	s = strings.ReplaceAll(s, "|", `\|`)
	return s
}

// HTMLEscape prepares user content for an HTML lexical slot inside
// markdown — covers attribute values, element bodies, and HTML
// comment bodies:
//
//   <gno-card type="..." caption="X">         attribute value
//   <gno-alert title="X">                     attribute value
//   <h5>X</h5>                                element body
//   <details><summary>X</summary>...          element body
//   <!-- X -->                                comment body (safe: `>`
//                                             becomes `&gt;`, so user
//                                             cannot inject `-->`)
//
// HTMLEscape escapes the union of attribute-breaking and body-breaking
// characters (`<`, `>`, `&`, `"`, `'`), so one function safely serves
// every HTML lexical context. Callers don't have to remember which
// subset to use for which slot.
//
// Pick the right helper — markdown title and HTML attribute share
// the look but use different escape rules:
//
//	[text](url "X")              → LinkTitle      (markdown title)
//	<span title="X">             → HTMLEscape     (HTML attribute)
//	<h5>X</h5>                   → HTMLEscape     (HTML element body)
//
// Swapping InlineText for HTMLEscape is wrong: markdown's backslash
// escapes survive into the rendered HTML as literal `\*`. Swapping
// LinkTitle for HTMLEscape is also wrong: `&amp;` written inside a
// markdown title renders as the literal characters `&amp;`.
//
// Not idempotent (see package doc): calling twice produces
// `&amp;` → `&amp;amp;`.
func HTMLEscape(s string) string {
	s = markdown.StripBidiAndZeroWidth(s)
	s = markdown.NormalizeBreaks(s)
	s = foldNewlinesAndSeparators(s, ' ')
	s = replaceNULWithFFFD(s)
	return html.EscapeString(s)
}

// ----- URL filters -----

// URL validates a URL for use as a link href, percent-encodes unsafe
// bytes, and rejects anything outside the allowlist of schemes.
//
// Allowlist:
//   - http, https
//   - mailto (rejected if contains ?body= or &body= — prefill phishing)
//   - schemeless relative URLs starting with /, ./, or ..
//     (rejects // protocol-relative — tracking-pixel vector)
//   - empty fragment-only URLs (#anchor)
//
// Returns "" if the URL is empty after trim or fails scheme allowlist.
func URL(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if !linkSchemeAllowed(s) {
		return ""
	}
	return markdown.PercentEncodeURL(s)
}

// ImageURL validates a URL for use as an image src. Kept separate from
// URL — not a parameterized variant — because the allowlist shapes
// differ qualitatively (data:image/* vs. mailto:) and a single boolean
// flag would invite callers to pass the wrong default.
//
// Allowlist:
//   - http, https
//   - schemeless relative URLs starting with /, ./, or ..
//     (rejects // protocol-relative — tracking-pixel vector)
//   - data:image/svg+xml, data:image/png, data:image/jpeg,
//     data:image/gif, data:image/webp
//
// Any other data: subtype is rejected — data:text/html etc. would
// render as inline HTML and execute embedded scripts.
//
// DEPLOYMENT PRECONDITION: data: URIs encode the bytes of the image
// directly into the markup, so a malicious sender can construct an
// image whose pixel dimensions are arbitrarily large at minimal byte
// cost. The deploying gnoweb instance MUST clamp rendered image
// dimensions via CSS (e.g. `max-width: 100%; max-height: <bound>`).
// Without that cap, a single image can blow out the page layout or
// exhaust the browser's memory.
//
// Returns "" if the URL is empty after trim or fails the allowlist.
func ImageURL(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if !imageSchemeAllowed(s) {
		return ""
	}
	return markdown.PercentEncodeURL(s)
}

// ----- Validators -----

// userNameCharsets builds the [2]uint64 bitmaps for the r/sys/users
// charset: first [a-z], rest [a-z0-9_-]. Initialized once at package
// init.
var (
	userNameFirstLo, userNameFirstHi uint64
	userNameRestLo, userNameRestHi   uint64
	footnoteFirstLo, footnoteFirstHi uint64
	footnoteRestLo, footnoteRestHi   uint64
	langFirstLo, langFirstHi         uint64
	langRestLo, langRestHi           uint64
	bechHrpFirstLo, bechHrpFirstHi   uint64
	bechHrpRestLo, bechHrpRestHi     uint64
	bechDataFirstLo, bechDataFirstHi uint64
	bechDataRestLo, bechDataRestHi   uint64
)

func init() {
	// UserName: first [a-z], rest [a-z0-9_-].
	for c := byte('a'); c <= 'z'; c++ {
		setBit(&userNameFirstLo, &userNameFirstHi, c)
		setBit(&userNameRestLo, &userNameRestHi, c)
	}
	for c := byte('0'); c <= '9'; c++ {
		setBit(&userNameRestLo, &userNameRestHi, c)
	}
	setBit(&userNameRestLo, &userNameRestHi, '_')
	setBit(&userNameRestLo, &userNameRestHi, '-')

	// FootnoteRef: [A-Za-z0-9_-] for both first and rest.
	for c := byte('A'); c <= 'Z'; c++ {
		setBit(&footnoteFirstLo, &footnoteFirstHi, c)
		setBit(&footnoteRestLo, &footnoteRestHi, c)
	}
	for c := byte('a'); c <= 'z'; c++ {
		setBit(&footnoteFirstLo, &footnoteFirstHi, c)
		setBit(&footnoteRestLo, &footnoteRestHi, c)
	}
	for c := byte('0'); c <= '9'; c++ {
		setBit(&footnoteFirstLo, &footnoteFirstHi, c)
		setBit(&footnoteRestLo, &footnoteRestHi, c)
	}
	for _, c := range []byte{'_', '-'} {
		setBit(&footnoteFirstLo, &footnoteFirstHi, c)
		setBit(&footnoteRestLo, &footnoteRestHi, c)
	}

	// LanguageName: [a-zA-Z0-9_+-] for both first and rest.
	for c := byte('A'); c <= 'Z'; c++ {
		setBit(&langFirstLo, &langFirstHi, c)
		setBit(&langRestLo, &langRestHi, c)
	}
	for c := byte('a'); c <= 'z'; c++ {
		setBit(&langFirstLo, &langFirstHi, c)
		setBit(&langRestLo, &langRestHi, c)
	}
	for c := byte('0'); c <= '9'; c++ {
		setBit(&langFirstLo, &langFirstHi, c)
		setBit(&langRestLo, &langRestHi, c)
	}
	for _, c := range []byte{'_', '+', '-'} {
		setBit(&langFirstLo, &langFirstHi, c)
		setBit(&langRestLo, &langRestHi, c)
	}

	// Bech HRP (when prefix==""): [a-z], 1-16 chars.
	for c := byte('a'); c <= 'z'; c++ {
		setBit(&bechHrpFirstLo, &bechHrpFirstHi, c)
		setBit(&bechHrpRestLo, &bechHrpRestHi, c)
	}

	// Bech data part: [a-z0-9], 6-90 chars.
	for c := byte('a'); c <= 'z'; c++ {
		setBit(&bechDataFirstLo, &bechDataFirstHi, c)
		setBit(&bechDataRestLo, &bechDataRestHi, c)
	}
	for c := byte('0'); c <= '9'; c++ {
		setBit(&bechDataFirstLo, &bechDataFirstHi, c)
		setBit(&bechDataRestLo, &bechDataRestHi, c)
	}
}

func setBit(lo, hi *uint64, c byte) {
	if c < 64 {
		*lo |= 1 << c
	} else {
		*hi |= 1 << (c - 64)
	}
}

// UserName validates the r/sys/users-registration charset:
// ^[a-z][a-z0-9]*([_-][a-z0-9]+)*$ length ≤ 64.
//
// The native MatchCharsetN enforces the leading-letter + tail-charset
// shape and length bound; this helper also performs the bidi-strip
// pre-pass. The "no consecutive [_-]" rule from r/sys/users is NOT
// enforced here (it's a registration-policy rule, not a sanitization
// concern — registrations go through r/sys/users itself).
//
// Returns the (bidi-stripped) input if valid, "" otherwise. On a ""
// return, do not emit the user-mention markup at all (e.g. skip the
// `[@user](/u/user)` link); falling back to the raw user-supplied
// string would defeat the validation.
func UserName(s string) string {
	s = markdown.StripBidiAndZeroWidth(s)
	if markdown.MatchCharsetN(s, userNameFirstLo, userNameFirstHi, userNameRestLo, userNameRestHi, 1, 64) {
		return s
	}
	return ""
}

// BechString validates a bech32-style address-like string.
//
// A bech32 string has the shape `<hrp>1<data>`: a human-readable
// prefix (HRP) that names the family (e.g. `g` for gno addresses,
// `gpub` for gno pubkeys, `cosmos` for cosmos addresses), the
// separator character `1`, then a data part carrying the encoded
// payload as lowercase alphanumerics.
//
// If prefix != "", requires s to start with prefix+"1" exactly, and the
// data part to match ^[a-z0-9]{6,90}$. Use this when you know the
// expected family:
//
//	sanitize.BechString(addr, "g")     // only g1...     (addresses)
//	sanitize.BechString(pk,   "gpub")  // only gpub1...  (pubkeys)
//
// If prefix == "", accepts any reasonable bech32 shape:
// ^[a-z]{1,16}1[a-z0-9]{6,90}$.
//
// Syntactic only — does NOT verify the bech32 checksum. Use a true
// bech32 decoder if you need that. Returns the cleaned input on
// accept, "" on reject; on "" return, do not emit the address-link
// markup (the user-supplied bytes have failed shape validation and
// should not appear unmodified in output).
func BechString(s, prefix string) string {
	s = markdown.StripBidiAndZeroWidth(s)
	if s == "" {
		return ""
	}
	if prefix != "" {
		// HRP must be lowercase ASCII letters.
		for i := 0; i < len(prefix); i++ {
			c := prefix[i]
			if c < 'a' || c > 'z' {
				return ""
			}
		}
		need := prefix + "1"
		if !strings.HasPrefix(s, need) {
			return ""
		}
		data := s[len(need):]
		if markdown.MatchCharsetN(data, bechDataFirstLo, bechDataFirstHi, bechDataRestLo, bechDataRestHi, 6, 90) {
			return s
		}
		return ""
	}
	// prefix == "" — accept any 1-16 char lowercase HRP, then '1', then data.
	sep := strings.IndexByte(s, '1')
	if sep < 1 || sep > 16 {
		return ""
	}
	hrp := s[:sep]
	if !markdown.MatchCharsetN(hrp, bechHrpFirstLo, bechHrpFirstHi, bechHrpRestLo, bechHrpRestHi, 1, 16) {
		return ""
	}
	data := s[sep+1:]
	if markdown.MatchCharsetN(data, bechDataFirstLo, bechDataFirstHi, bechDataRestLo, bechDataRestHi, 6, 90) {
		return s
	}
	return ""
}

// FootnoteRef validates a footnote name: ^[A-Za-z0-9_-]{1,64}$.
// Strips bidi/zero-width first. Returns s if valid, "" otherwise.
//
// The same charset matches goldmark's auto-anchor-ID extension {#id}
// (when enabled); use this helper for that slot too.
//
// On "" return, omit the footnote (or anchor) entirely rather than
// emitting it with raw user bytes — the validation exists to keep
// untrusted bytes out of the footnote-namespace key.
func FootnoteRef(s string) string {
	s = markdown.StripBidiAndZeroWidth(s)
	if markdown.MatchCharsetN(s, footnoteFirstLo, footnoteFirstHi, footnoteRestLo, footnoteRestHi, 1, 64) {
		return s
	}
	return ""
}

// LanguageName validates the language tag (a.k.a. "info string") for
// a fenced code block — the `go` in:
//
//   ```go
//   fmt.Println("hi")
//   ```
//
// Charset: ^[a-zA-Z0-9_+-]{1,32}$ — letters, digits, `_`, `+`, `-`,
// up to 32 bytes. Strips bidi/zero-width first.
//
// Returns the cleaned input if valid, "" otherwise. A "" return means
// the caller should emit a language-less fence (``` without a tag)
// rather than letting the user pick the syntax highlighter — which
// could otherwise be used to inject newlines or block markers into
// what becomes the opening fence line.
func LanguageName(s string) string {
	s = markdown.StripBidiAndZeroWidth(s)
	if markdown.MatchCharsetN(s, langFirstLo, langFirstHi, langRestLo, langRestHi, 1, 32) {
		return s
	}
	return ""
}

// NestedPrefix validates a prefix string for line-prefixing builders
// like md.Nested, which prepends `prefix` to every line of content
// to render the content as a nested/indented sub-block.
//
// Allowed: any string matching `^[ \t>]*$` — spaces, tabs, blockquote
// `>` chars only. Anything else (a `#`, a `-`, a letter) would let a
// caller turn benign sub-content into a heading, list, or paragraph
// at the wrong nesting level.
//
// Returns s if valid, "" otherwise. Strips bidi/zero-width first —
// otherwise an invisible character hidden inside a `>` prefix would
// be replicated on every nested content line, producing per-line
// display-vs-storage divergence.
//
// On "" return, fall back to a known-safe prefix literal (e.g.
// `"> "`) or skip the nesting entirely. Do not emit the raw
// user-supplied prefix.
func NestedPrefix(s string) string {
	s = markdown.StripBidiAndZeroWidth(s)
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c != ' ' && c != '\t' && c != '>' {
			return ""
		}
	}
	return s
}

// ----- Primitive -----

// CodeFence returns a string of backticks long enough to wrap content
// as a CommonMark fenced code block without the content's own backticks
// closing the fence prematurely.
//
// Returned length N = max(minCount, longestBacktickRunInContent + 1).
// Use N backticks both before and after the content:
//
//   fence := sanitize.CodeFence(userCode, 3)
//   out += fence + "\n" + userCode + "\n" + fence + "\n"
//
// Typical minCount values:
//   - 1 for inline code spans (`x`)
//   - 3 for block fenced code (CommonMark §4.5 requires ≥3)
//
// `minCount < 1` is clamped to 1. Empty content returns
// strings.Repeat("`", max(minCount, 1)). Never panics.
//
// Most realms should reach for `md.InlineCode` / `md.CodeBlock` /
// `md.LanguageCodeBlock` in p/moul/md, which call CodeFence internally
// and emit the full code block for you. Call CodeFence directly only
// when you're rolling a custom fence emitter (e.g. a renderer that
// needs the fence length but emits the body differently).
func CodeFence(content string, minCount int) string {
	return markdown.CodeFence(content, minCount)
}

// ----- internal helpers -----

// linkSchemeAllowed returns true if s passes the URL helper's scheme
// allowlist. See URL's doc for the policy.
func linkSchemeAllowed(s string) bool {
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		return true
	}
	if strings.HasPrefix(s, "mailto:") {
		// Reject prefill phishing via ?body= or &body=.
		if strings.Contains(s, "?body=") || strings.Contains(s, "&body=") {
			return false
		}
		return true
	}
	if strings.HasPrefix(s, "//") {
		// Protocol-relative — reject (tracking-pixel vector).
		return false
	}
	if strings.HasPrefix(s, "/") || strings.HasPrefix(s, "./") || strings.HasPrefix(s, "../") {
		return true
	}
	if strings.HasPrefix(s, "#") {
		// Fragment-only.
		return true
	}
	return false
}

// imageSchemeAllowed returns true if s passes the ImageURL helper's
// scheme allowlist. Tighter than linkSchemeAllowed: no mailto/tel,
// only data:image/<subset>.
func imageSchemeAllowed(s string) bool {
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		return true
	}
	if strings.HasPrefix(s, "//") {
		return false
	}
	if strings.HasPrefix(s, "/") || strings.HasPrefix(s, "./") || strings.HasPrefix(s, "../") {
		return true
	}
	if strings.HasPrefix(s, "data:") {
		// Only the curated image/* subset. CSS must enforce sizing.
		for _, p := range []string{
			"data:image/svg+xml",
			"data:image/png",
			"data:image/jpeg",
			"data:image/gif",
			"data:image/webp",
		} {
			if strings.HasPrefix(s, p) {
				return true
			}
		}
		return false
	}
	return false
}

// foldNewlinesAndSeparators replaces \n, U+0085 NEL, U+2028 LINE SEPARATOR,
// U+2029 PARAGRAPH SEPARATOR with the given replacement byte (typically
// space for inline-context helpers).
//
// NormalizeBreaks has already folded \r\n and \r to \n before this runs,
// so \n is the canonical break byte to substitute.
func foldNewlinesAndSeparators(s string, replacement byte) string {
	if !needsSeparatorFold(s) {
		return s
	}
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); {
		c := s[i]
		if c == '\n' {
			out = append(out, replacement)
			i++
			continue
		}
		// U+0085 NEL: 0xC2 0x85
		if c == 0xC2 && i+1 < len(s) && s[i+1] == 0x85 {
			out = append(out, replacement)
			i += 2
			continue
		}
		// U+2028 (0xE2 0x80 0xA8) or U+2029 (0xE2 0x80 0xA9)
		if c == 0xE2 && i+2 < len(s) && s[i+1] == 0x80 && (s[i+2] == 0xA8 || s[i+2] == 0xA9) {
			out = append(out, replacement)
			i += 3
			continue
		}
		out = append(out, c)
		i++
	}
	return string(out)
}

func needsSeparatorFold(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '\n' || c == 0xC2 || c == 0xE2 {
			return true
		}
	}
	return false
}

// replaceNULWithFFFD substitutes any NUL byte with the UTF-8 encoding
// of U+FFFD REPLACEMENT CHARACTER per CM §2.3.
func replaceNULWithFFFD(s string) string {
	if !strings.ContainsRune(s, 0) {
		return s
	}
	return strings.ReplaceAll(s, "\x00", "\ufffd")
}
