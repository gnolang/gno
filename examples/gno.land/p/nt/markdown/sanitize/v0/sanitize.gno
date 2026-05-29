// Package sanitize provides input-cleaning primitives and safe-emit
// builders for each markdown lexical slot. Realm authors wrap user-
// supplied strings with these helpers before flowing them into rendered
// markdown output. Each helper targets one specific slot (link text,
// heading text, URL href, table cell, HTML attribute, fenced code block,
// blockquote, footnote definition, link-reference definition, etc.) and
// neutralizes the bytes that would otherwise let user content break out
// of that slot or inject new top-level structure.
//
// Pick the right helper from the table under "Picking the right helper"
// below, then wrap each user-supplied argument exactly once at the call
// site (see "The audit rule").
//
// # Wrap once
//
// Most escapers and safe-emit builders in this package are NOT
// idempotent — applying them twice re-escapes bytes the first pass
// added (`\*` becomes `\\\*`, `&amp;` becomes `&amp;amp;`, a fenced
// block gets re-fenced). Wrap each user-derived string with at most
// one sanitize.* call. Block and BlockRich are exceptions —
// idempotent by design — but the at-most-once rule is still the
// safest default. See the "Idempotence classes" enumeration below
// for the full breakdown.
//
// Some markdown-builder packages (e.g. p/moul/md) sanitize the args of
// specific helpers internally — see each builder's package doc for the
// per-helper contract. If the builder sanitizes for you, pass the raw
// user input; if it doesn't, wrap the input with the right sanitize.*
// helper at the call site.
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
//	multi-paragraph post body w/ rich block    BlockRich
//	  structure (headings, lists, tables, etc.)
//	multi-line blockquote (`> ` prefixed)      Blockquote
//	multi-line blockquote w/ rich block body   BlockquoteRich
//	[text](url "title")                        LinkTitle (title)
//	| cell |                                   TableCell
//	<gno-card caption="X">                     HTMLEscape
//	<h5>X</h5>                                 HTMLEscape
//	any URL going into ](X)                    URL
//	any image src going into (X)               ImageURL
//	`inline code` inside running prose         InlineCode
//	multi-line fenced code block               CodeBlock
//	multi-line fenced code with language tag   LanguageCodeBlock
//	[^name]: footnote body                     FootnoteDefinition
//	[label]: url "title" reference def         LinkReferenceDefinition
//	r/sys/users handle                         UserName       (validator)
//	g1.../gpub1... etc.                        BechString     (validator)
//	footnote / LRD label / {#id} anchor name   FootnoteLabel  (validator)
//	fenced-code language tag                   LanguageName   (validator)
//	prefix arg to md.Nested                    NestedPrefix   (validator)
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
//	  UserName, BechString, FootnoteLabel, LanguageName, NestedPrefix
//	  URL, ImageURL              (accept→identity; reject→"")
//	  Block                      (bracket walker treats \[/\] as ordinary;
//	                              line-leader escapes don't re-fire on
//	                              already-escaped `\#` etc.)
//	  BlockRich                  (TrimLeft/TrimRight + "\n\n" wrap is stable)
//
//	NOT idempotent — never wrap an already-sanitized string:
//	  InlineText, LinkTitle, TableCell   (re-escape backslashes)
//	  HTMLEscape                         (re-escapes `&` → `&amp;`)
//	  Blockquote, BlockquoteRich         (re-prefixes `> `, nesting the quote each pass)
//	  InlineCode, CodeBlock,
//	  LanguageCodeBlock                  (wrap with a fence — calling twice double-wraps)
//	  FootnoteDefinition,
//	  LinkReferenceDefinition            (compose Block/InlineText/URL internally —
//	                                      passing already-sanitized strings double-escapes)
//
//	CodeFence is pure: same inputs always give the same output.
//
// Validators (UserName / BechString / FootnoteLabel / LanguageName /
// NestedPrefix) return either the cleaned input verbatim or "". They
// never partially-sanitize: if the input doesn't match the slot's
// charset/shape, the answer is rejection.
//
// # Composition rules
//
// Direct sanitize use (when emitting markdown without a builder package):
//
//	out := "# " + sanitize.InlineText(userTitle) + "\n\n" +
//	       sanitize.Block(userBody)
//	out += sanitize.Blockquote(userQuote)
//	out += sanitize.LanguageCodeBlock(realmLang, userCode)
//
// Use with a builder package (e.g. p/moul/md): pass raw user input to
// the builder helpers that sanitize internally — do NOT pre-wrap with
// sanitize.*, or the input gets double-escaped (escapers are not
// idempotent). See the builder's package doc for the per-helper
// contract. For example, with p/moul/md:
//
//	md.Blockquote(userProse)                 // good — md.Blockquote sanitizes
//	md.LanguageCodeBlock(realmLang, userCode) // good — sanitizes both args
//	md.Link(userText, userURL)               // good — sanitizes both slots
//
//	md.Blockquote(sanitize.Block(userProse))         // BAD: double-wrap
//	md.Link(sanitize.InlineText(t), sanitize.URL(u)) // BAD: double-wrap
//
// Wrong (across all callers):
//
//	sanitize.InlineText(sanitize.InlineText(s))   double-wrap (re-escape)
//	sanitize.TableCell(sanitize.InlineText(s))    TableCell already calls InlineText
//	sanitize.URL(sanitize.InlineText(href))       inline-escape backslash-escapes `.` `-` `_`
//	                                              inside the URL, corrupting the host/path
//	sanitize.Blockquote(sanitize.Blockquote(s))   double-wrap — outer would escape the
//	                                              inner `> ` prefixes
//	sanitize.Block(sanitize.BlockRich(s))         double-sanitize — strict Block re-escapes
//	                                              the markers BlockRich preserved (headings,
//	                                              lists, tables); BlockRich's rich structure
//	                                              renders as literal text after Block escapes
//	                                              its line-leaders
//	sanitize.BlockRich(sanitize.Block(s))         pointless double-sanitize — Block already
//	                                              escaped every line-leader to `\#`/`\>`/etc.;
//	                                              BlockRich preserves the backslash escapes
//	                                              as visible artifacts in user prose
//	sanitize.Blockquote(sanitize.BlockRich(s))    double-sanitize — Blockquote's Block step
//	                                              re-escapes the markers BlockRich preserved
//	sanitize.BlockRich(sanitize.Blockquote(s))    nonsense — Blockquote already line-prefixed
//	                                              with `> `; BlockRich expects raw user content
//	sanitize.BlockquoteRich(sanitize.BlockRich(s)) double-wrap — Rich + Rich nests twice
//	sanitize.BlockRich(sanitize.TableCell(s))     wrong slot — use TableCell for cell content,
//	                                              BlockRich for multi-paragraph block content
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
//     blockquote, list, code fence, link-reference def, setext underline,
//     gnoweb extension delimiter, or GFM table row at document level.
//     Strict Block escapes the line-leading `|` of any GFM table row so
//     user content cannot inject `<table>`-shaped structure; permissive
//     BlockRich preserves table rows so authors can compose `<table>`
//     elements (gnoweb loads extension.Table per render_config.go).
//   - HTML block type 1-5 absorption: CommonMark §4.6 HTML block types 1
//     (`<script>`, `<pre>`, `<style>`, `<textarea>`), 2 (`<!--`), 3
//     (`<?`), 4 (`<!UPPER`), and 5 (`<![CDATA[`) do NOT close on a blank
//     line — they only close on a type-specific token (`</tag>`, `-->`,
//     `?>`, `>`, `]]>`) or EOF. Without a defense, user content opening
//     any of these would swallow realm chrome appended afterward. Both
//     Block and BlockRich line-escape the openers (prepend `\`) so the
//     block never opens; this defense is unconditional in both modes.
//     Types 6 and 7 close on a blank line, so BlockRich's `\n\n`
//     paragraph envelope already bounds them and no escape is needed.
//   - realm-discipline boundary (caller's responsibility, not enforced):
//     callers should emit realm chrome at flush-left column 0 around
//     `BlockRich(user)`. Indented chrome (4+ leading spaces, list-item
//     continuations, footnote-definition body, or an unclosed Type 1
//     HTML tag in realm chrome before the call) can extend across blank
//     lines into user content or vice versa. The sanitizer cannot
//     defend against malformed realm chrome — only against user input.
//   - footnote / link-reference namespace pollution: user content
//     containing `[^name]` or `[text][label]` syntax that would otherwise
//     resolve against realm-defined footnote definitions or link
//     reference definitions elsewhere on the page. Block escapes the
//     opening `[` in both shapes.
//   - reference-link / footnote-ref / shortcut-ref collisions:
//     `[text][label]`, `[^name]`, and bare `[label]` shortcut forms
//     are ALL neutralized by Block's bracket walk, which preserves
//     only inline `[text](url)` and `![alt](src)` syntax — everything
//     else has both `[` and `]` backslash-escaped, so the parser sees
//     literal text and can't resolve against realm-defined LRDs or
//     footnote definitions.
//   - multi-line LRD evasion: Block's walker recognises `[lab\nel]: url`
//     across newlines (single `\n` OK, blank line aborts) and strips
//     the whole region. `\]` inside the label is honored as an escaped
//     literal, so `[label\]: url` is NOT treated as an LRD (renders as
//     literal text).
//   - URL scheme abuse: javascript:, data:text/html, vbscript:, blob:,
//     protocol-relative //, mailto: with prefill phishing parameters.
//     Allowlist-only (URL / ImageURL).
//   - HTML attribute / element breakout: `"`, `<`, `>`, `&`, `'` inside
//     HTML lexical slots. Handled by HTMLEscape.
//   - CommonMark §2.3 NUL: replaced with U+FFFD by Block, InlineText,
//     LinkTitle, TableCell, HTMLEscape, InlineCode, CodeBlock, and
//     LanguageCodeBlock.
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
// # When to use Block vs BlockRich
//
// Both are safe sanitizers; both run identical realm-binding defenses.
// They differ in what user-authored block structure survives:
//
//   - Block — paragraph-shaped only. Escapes `#`, `>`, list markers,
//     `---`/`***`/`___` thematic breaks, and `===`/`---` setext
//     underlines. Use for leaf slots — footnote definition bodies,
//     table cells, blockquote bodies (Blockquote uses Block), single-
//     paragraph prose, any slot where richer structure has no benefit
//     or where richer structure could visually impersonate realm chrome.
//
//   - BlockRich — full-richness. Preserves user-authored headings,
//     lists, quotes, HR, setext. Use for user content the realm intends
//     to compose with full block-level structure, typically inside a
//     sandbox container (`<gno-card>`, `<gno-foreign>`) or a CSS-demoted
//     region. BlockRich's qualifying-setext defense prevents the
//     cross-boundary attack (user content reaching back to promote
//     realm chrome to a heading), but inner-heading visual containment
//     is the realm's CSS responsibility. gnoweb does not yet ship CSS
//     rules that demote headings inside sandbox containers — until they
//     land, BlockRich + sandbox renders inner headings at literal size.
//
// Do NOT compose Block and BlockRich in either direction. Pick one
// helper at the right level.
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
//     > [!NOTE] InlineText(title)
//     > Author: InlineText(name)
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
	s = markdown.NormalizeBreaks(s)       // CM §2.2 \r\n / \r → \n
	s = foldNewlinesAndSeparators(s, ' ') // \n + NEL + U+2028/U+2029 → space
	return markdown.EscapeInline(s)
}

// Block prepares user content for a top-level BLOCK markdown context
// where paragraphs, line breaks, code blocks, and other block structure
// should survive — but where the content must NOT be able to inject
// new top-level constructs (headings, lists, blockquotes,
// link-reference definitions, setext underlines, gnoweb extension
// delimiters, GFM table rows).
//
// Use for any multi-paragraph user-supplied prose that the realm
// concatenates into its rendered output:
//   - post bodies, comments, replies
//   - profile bios, About sections
//   - proposal descriptions, governance motions
//   - changelog entries, release notes
//
// What Block does with each kind of attacker input:
//
//	User attempt                                          | Block's response
//	------------------------------------------------------|----------------------------------------------------
//	  --- preserved verbatim ---                          |
//	[text](url) inline link, ![alt](src) image            | preserved verbatim
//	------------------------------------------------------|----------------------------------------------------
//	  --- escaped / stripped / folded ---                 |
//	# heading at line-start                               | escaped → literal `# heading`
//	> quoted at line-start                                | escaped → literal `>`
//	- item, * item, + item, 1. item at line-start         | escaped
//	---, ***, ___ (3+) at line-start                      | escaped
//	=== or --- on its own line after non-blank text       | escaped (no setext promotion of the line above)
//	<gno-card>, <gno-columns>, any <gno-…>/</gno-…> at    | escaped (wildcard match) → literal text
//	  line-start                                          |
//	| a | b | GFM table row (line-leading `|`)            | escaped → literal `| a | b |`
//	<!--, <script>, <pre>, <style>, <textarea>, <?…?>,    | escaped (\<…) → literal text;
//	  <!DOCTYPE…>, <![CDATA[…]]> at line-start            |   blocks goldmark from opening a
//	  (CM §4.6 HTML block types 1-5)                      |   blank-line-NON-terminating HTML block
//	[text][realm-label] ref-link USE                      | both bracket pairs escaped → \[text\]\[realm-label\]
//	[^name] footnote-ref                                  | both brackets escaped → \[^name\]
//	[label] bare shortcut-ref                             | both brackets escaped → \[label\]
//	[label]: url link-reference definition                | whole region stripped (incl. multi-line label
//	  (incl. [lab\nel]: url multi-line)                   |   `[lab\nel]: url` and any title continuation)
//	[label\]: url (backslash-escaped `]`)                 | NOT stripped; brackets escaped → paragraph text
//	code fence opened without close                       | autoclosed at end of input
//	NUL byte (\x00)                                       | replaced with U+FFFD
//	U+2028 / U+2029 / U+0085 (NEL)                        | folded to `\n`
//	bidi/zero-width controls                              | stripped
//
// COMPOSITION GOTCHA: Block's EOF fence-autoclose appends a final
// fence line. If you wrap Block's output with a line-prefixing
// builder like md.Blockquote (which prepends `> ` per line) or
// md.Nested, that closing fence becomes a prefixed line. The output
// is still safe (the fence still closes correctly) but may render
// awkwardly. If pixel-perfect output matters, strip a trailing blank
// fence line after Block.
//
// Why backslash and not a space for `<gno-…>` lines: gnoweb's
// extension parsers call `util.TrimLeftSpace` on the line before tag
// matching, which would strip a leading space and let the tag match
// anyway. A leading `\` survives the trim (only ASCII whitespace +
// form-feed are stripped) and is consumed by the inline escape phase
// before Type-7 HTML block detection can fire (Type-7 requires the
// first non-whitespace char to be `<`).
//
// Inline emphasis, code spans, inline links, and soft line breaks
// within a paragraph are PRESERVED — users can format. Pipes that
// are NOT at line-start stay literal so prose can still write things
// like `a | b`.
//
// Idempotent: Block(Block(s)) is byte-identical to Block(s). The
// bracket walker strips LRDs on the first pass; remaining `[`/`]`
// outside inline-link/image spans are escaped to `\[`/`\]`, and
// already-escaped brackets are preserved on subsequent passes
// (pass-2 backslash-parity tracking). Still, wrap each user-supplied
// string exactly once — chained sanitization adds no value and
// burns gas.
func Block(s string) string {
	s = markdown.NormalizeBreaks(s)
	s = markdown.StripBidiAndZeroWidth(s)
	s = replaceNULWithFFFD(s)
	return markdown.EscapeBlockHazards(s)
}

// BlockRich is the permissive counterpart of Block. Both are safe
// sanitizers — the distinction is what markdown structure survives:
//
//   - Block escapes line-leading block markers (`#`, `>`, `-`, `*`,
//     `+`, `1.`), thematic breaks (`---`/`***`/`___`), and setext
//     underlines (`===`/`---`). User content becomes paragraph-shaped.
//   - BlockRich PRESERVES all of those, so user content can compose
//     headings, lists, quotes, horizontal rules, and setext-styled
//     headings. Realm-binding defenses stay on (extension delimiters
//     `<gno-…>`, the bracket walker for link / LRD / ref /
//     footnote / shortcut, fence autoclose, NUL / bidi /
//     Unicode-separator folding). GFM table-row openers are
//     PRESERVED (see "Tables" below).
//
// Cross-paragraph safety: BlockRich's output begins with "\n\n"
// AND ends with "\n\n" — CM §4.8 blank lines on both sides — so
// user content is guaranteed to occupy its own paragraph(s),
// isolated from anything the realm emits before OR after. Symmetric
// isolation closes four distinct attacks:
//
//   - Cross-paragraph setext promotion (backward). User content
//     `body\n===\nmore` concatenated after realm chrome (no
//     trailing `\n`) would, without paragraph isolation, place
//     "chrome\nbody" in one paragraph; the `===` setext underline
//     would then promote that merged paragraph to H1, hijacking
//     realm chrome. The leading "\n\n" forces a paragraph break.
//
//   - Cross-paragraph GFM table promotion (backward). User content
//     beginning with `|---|---|` (a table delimiter row) would,
//     without a blank-line break, retroactively turn the preceding
//     realm line into a `<thead>`. Paragraph isolation prevents
//     the table-detection scan from crossing the boundary.
//
//   - Cross-paragraph GFM table promotion (forward). Realm chrome
//     appended immediately after BlockRich(user) that begins with
//     `|---|` would, without a trailing blank line, extend user's
//     last line into a table header and pull realm chrome into the
//     body row. The trailing "\n\n" prevents the merge.
//
//   - Lazy paragraph continuation (forward). Paragraph-shaped realm
//     chrome appended immediately after BlockRich(user) would, via
//     CM §5.2, merge into user's trailing paragraph and inherit any
//     block-level decoration it carries.
//
// First-line qualifying-setext escape (the
// `neuterLeadingSetextIfQualifying` pre-pass) remains in place as
// belt-and-suspenders: if the first non-blank line of user input
// matches the CM §4.3 setext-underline pattern (run of `=` or `-`
// with 0-3 leading spaces and only trailing whitespace), BlockRich
// inserts `\` before the first `=`/`-`. This is redundant given
// paragraph isolation but harmless and inexpensive.
//
// # Tables
//
// BlockRich preserves line-leading `|` so user content can
// compose GFM tables:
//
//	| Header A | Header B |
//	|----------|----------|
//	| cell a   | cell b   |
//
// renders as a real `<table>` element. Strict Block continues to
// escape line-leading `|` (each row becomes literal `\| a | b |`
// text). When the realm authors the table itself and inserts user
// content into a specific cell, use TableCell — NOT BlockRich —
// to sanitize that cell value.
//
// What attacker input produces what (full table, same rows as Block
// except where marked CHANGED):
//
//	User attempt                                  | BlockRich response
//	----------------------------------------------|--------------------------------------------------
//	  --- preserved (compose freely) ---          |
//	# heading at line-start                       | preserved [CHANGED from Block]
//	> quoted at line-start                        | preserved [CHANGED]
//	- item, * item, + item, 1. item               | preserved [CHANGED]
//	---, ***, ___ thematic break                  | preserved [CHANGED]
//	=== or --- setext underline                   | preserved when preceded by user text;
//	                                              | escaped (\===/\---) if the first non-blank
//	                                              | line of input [CHANGED]
//	| a | b | GFM table row (line-leading |)      | preserved → renders as <table> when followed by
//	                                              | a delimiter row [CHANGED]
//	[text](url), ![alt](src)                      | preserved verbatim [SAME]
//	----------------------------------------------|--------------------------------------------------
//	  --- escaped / stripped / folded ---         |
//	<gno-card>, any <gno-…>/</gno-…> at line-start| escaped (wildcard match) [SAME]
//	<!--, <script>, <pre>, <style>, <textarea>,   | escaped (\<…) [SAME] — Types 1-5 don't close
//	  <?…?>, <!DOCTYPE…>, <![CDATA[…]]>           |   on blank lines, so `\n\n` envelope
//	  at line-start (CM §4.6 HTML block types 1-5)|   doesn't isolate them; explicit escape
//	[text][realm-label] ref-link USE              | both pairs escaped [SAME]
//	[^name] footnote-ref                          | both brackets escaped [SAME]
//	[label] bare shortcut-ref                     | both brackets escaped [SAME]
//	[label]: url link-reference definition        | whole region stripped [SAME]
//	[label\]: url (escaped `]`)                   | not stripped; brackets escaped [SAME]
//	code fence opened without close               | autoclosed at end of input [SAME]
//	NUL byte (\x00)                               | replaced with U+FFFD [SAME]
//	U+2028 / U+2029 / U+0085 (NEL)                | folded to `\n` [SAME]
//	bidi/zero-width controls                      | stripped [SAME]
//
// Use BlockRich for user content the realm intends to compose with
// full block-level richness — typically inside a sandbox container
// (`<gno-card>`, `<gno-foreign>`) or a CSS-demoted region where inner
// headings render visually distinct from realm chrome. The realm
// must own the visual containment: concatenating BlockRich's output
// directly into a top-level page still lets the user write `# heading`
// at document level. BlockRich's cross-boundary setext defense prevents
// the worst case (reaching backwards into realm bytes), but visual
// containment of inner headings is the realm's CSS responsibility.
// gnoweb does not yet ship CSS rules that demote inner headings inside
// `<gno-card>` / `<gno-foreign>` — until those rules land, realms
// using BlockRich + a sandbox should be aware that inner headings
// render at their literal level.
//
// Idempotent: BlockRich(BlockRich(s)) is byte-identical to
// BlockRich(s). The TrimLeft-then-"\n\n"-prepend pattern strips
// any leading newlines and reapplies exactly two, so the leading
// shape is stable across passes; the qualifying-setext escape is
// stable (a line beginning with `\` no longer matches the setext
// pattern); and the bracket walker treats already-escaped
// `\[`/`\]` as ordinary bytes. Empty input (or input that strips
// to empty, e.g. a lone link-reference definition) returns "" —
// realm concatenation doesn't get a stray blank line.
// Still, wrap each user-supplied string exactly once — chained
// sanitization adds no value and burns gas.
//
// Realm-discipline boundary: BlockRich defends user input against
// every cross-paragraph attack listed above, but it CANNOT defend
// against malformed REALM chrome. Specifically, callers should emit
// realm chrome at flush-left column 0 around `BlockRich(user)`. If
// the realm chrome BEFORE the call contains an unclosed CM §4.6
// Type 1 HTML tag (`<script>`, `<pre>`, `<style>`, `<textarea>`),
// the `\n\n` envelope does NOT close it (Type 1 closes only on the
// matching close tag), and user-controlled `</tag>` content can
// then prematurely terminate it. Indented chrome (4+ leading
// spaces, list-item continuations, footnote-definition body) can
// likewise extend across the envelope into user content. Keep
// chrome flush-left and Type 1 tags closed within the chrome.
//
// PREVIEW: BlockquoteRich is currently the only in-tree caller of
// BlockRich; the API and the `"\n\n"` output shape may evolve once
// direct callers emerge.
func BlockRich(s string) string {
	s = markdown.NormalizeBreaks(s)
	s = markdown.StripBidiAndZeroWidth(s)
	s = replaceNULWithFFFD(s)
	// Fold Unicode separators (U+2028, U+2029, U+0085 NEL) to '\n'
	// BEFORE the setext-qualifying check. The native
	// EscapeBlockHazardsRich also folds them internally, but the Gno
	// helper below needs to see the folded form to correctly identify
	// the first non-blank line — otherwise an attacker can hide the
	// `===` setext underline behind a U+2028 / U+2029 / U+0085 and
	// reach back to promote realm chrome above it.
	s = foldSeparatorsToNewline(s)
	s = neuterLeadingSetextIfQualifying(s)
	s = markdown.EscapeBlockHazardsRich(s)
	// Ensure the output BOTH starts AND ends with "\n\n" — CM §4.8
	// blank lines on each side — so user content is GUARANTEED to
	// occupy its own paragraph(s), isolated from anything the realm
	// emits before OR after. Symmetric isolation closes four attacks:
	//
	//  Backward (closed by leading "\n\n"):
	//   1. Deeper-setext: user content `body\n===\nmore` concatenated
	//      after realm chrome (no trailing `\n`) would otherwise place
	//      "chrome\nbody" in one paragraph; the `===` setext underline
	//      would then promote that merged paragraph to H1, hijacking
	//      realm chrome.
	//   2. GFM table-row promotion: user content beginning with
	//      `|---|---|` (a table delimiter row) would, without a blank-
	//      line break, retroactively promote the preceding realm line
	//      into a `<thead>` cell.
	//
	//  Forward (closed by trailing "\n\n"):
	//   3. GFM table-row promotion in reverse: realm appending its own
	//      chrome immediately after BlockRich(user), where chrome
	//      starts with `|---|`, would extend user's last line into a
	//      table header and pull realm chrome into the body row.
	//   4. Lazy paragraph continuation: realm appending paragraph-
	//      shaped chrome immediately after BlockRich(user) would, via
	//      CM §5.2 lazy-continuation, merge into user's trailing
	//      paragraph and inherit any block-level decoration it carries.
	//
	// `neuterLeadingSetextIfQualifying` above is now belt-and-
	// suspenders for the first-line setext case: even if the blank-
	// line guarantee were somehow defeated by an exotic CM consumer,
	// the first-line escape still blocks the simplest setext shape.
	//
	// Empty post-escape result short-circuits to "" so realm
	// concatenation doesn't leak stray blank lines for trivially empty
	// inputs (e.g. a lone link-reference definition that strips
	// entirely).
	//
	// Idempotency: TrimLeft and TrimRight strip ALL leading/trailing
	// "\n"s, then the wrap adds exactly two on each side. Stable
	// across passes.
	s = strings.TrimLeft(s, "\n")
	s = strings.TrimRight(s, "\n")
	if s == "" {
		return ""
	}
	return "\n\n" + s + "\n\n"
}

// foldSeparatorsToNewline replaces U+0085 NEL (0xC2 0x85),
// U+2028 (0xE2 0x80 0xA8), and U+2029 (0xE2 0x80 0xA9) with '\n'.
// Leaves '\n' bytes alone. Used by BlockRich so the qualifying-setext
// pre-pass and the native both see the same line structure.
func foldSeparatorsToNewline(s string) string {
	// Cheap pre-check: only the 0xC2 / 0xE2 lead bytes can trigger.
	if !containsAnyByteForFold(s) {
		return s
	}
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); {
		c := s[i]
		if c == 0xC2 && i+1 < len(s) && s[i+1] == 0x85 {
			out = append(out, '\n')
			i += 2
			continue
		}
		if c == 0xE2 && i+2 < len(s) && s[i+1] == 0x80 && (s[i+2] == 0xA8 || s[i+2] == 0xA9) {
			out = append(out, '\n')
			i += 3
			continue
		}
		out = append(out, c)
		i++
	}
	return string(out)
}

func containsAnyByteForFold(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == 0xC2 || s[i] == 0xE2 {
			return true
		}
	}
	return false
}

// neuterLeadingSetextIfQualifying scans s for the first non-blank
// line. If that line matches the CommonMark §4.3 setext-underline
// pattern (0-3 leading spaces, then a run of all `=` or all `-`,
// then optional trailing whitespace, then `\n` or EOF), the function
// returns s with a `\` inserted before the first `=`/`-`. Otherwise
// returns s unchanged. The escape prevents a realm-emitted line above
// BlockRich's output from being retroactively promoted to a heading.
func neuterLeadingSetextIfQualifying(s string) string {
	pos := 0
	for pos < len(s) {
		// Walk to the first non-whitespace byte of the current line.
		lineStart := pos
		i := pos
		for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
			i++
		}
		if i >= len(s) || s[i] == '\n' {
			// Blank line; advance to next line.
			if i >= len(s) {
				return s
			}
			pos = i + 1
			continue
		}
		// First non-blank line. Check setext-underline shape.
		if i-lineStart > 3 {
			return s // 4+ leading spaces = indented code, not setext
		}
		c := s[i]
		if c != '=' && c != '-' {
			return s // not a setext underline candidate
		}
		j := i + 1
		for j < len(s) && s[j] == c {
			j++
		}
		for j < len(s) && (s[j] == ' ' || s[j] == '\t') {
			j++
		}
		if j < len(s) && s[j] != '\n' {
			return s // mixed content on the line — not setext
		}
		return s[:i] + "\\" + s[i:]
	}
	return s
}

// Blockquote wraps user content as a CommonMark blockquote: each line
// of the cleaned content gets a "> " prefix so the renderer displays
// it inside a `<blockquote>` element.
//
// Use for any multi-paragraph user-supplied text that the realm wants
// to render as a quotation: cited posts, attached responses, error
// snapshots that should visually stand out.
//
// The content is first cleaned by Block (bidi-strip, line-ending
// normalize, NUL→U+FFFD, bracket walker for link/image/LRD spans,
// block-marker escape, code-fence auto-close at EOF, Unicode-separator
// fold), then every resulting line — including blank lines — is
// prefixed with "> ". The
// user content can still use inline emphasis, code spans, and nested
// fenced code blocks inside the quote; what it cannot do is open new
// top-level structure (heading, list, blockquote, GFM table row, etc.)
// or escape the quote.
//
// Empty input returns "".
//
// Composition gotcha: Block's EOF code-fence auto-close (added when
// user content opens a ``` ``` ``` fence without closing it) becomes a
// "> ```" line at the end of the blockquote. Goldmark parses this
// correctly as the close of a fenced block inside the quote — the
// output is structurally safe — but the markdown source looks unusual
// to a human reviewer. If aesthetic output matters, ensure user
// content closes its own fences.
//
// Trailing newlines in input become quoted blank lines (`> \n`) rather
// than terminating the quote early — `Blockquote("foo\n\n")` emits
// `"> foo\n> \n"`. Use `strings.TrimRight(s, "\n")` before passing if
// you want the quote to end exactly at the last content line.
//
// Not idempotent (see package doc): wraps with `> ` per line, so
// calling twice double-wraps and the outer call's Block step escapes
// the inner `>` prefixes.
//
// Do NOT compose with BlockRich in either direction:
//   - Blockquote(BlockRich(s)) double-sanitizes: BlockRich preserves
//     `#`/`>`/etc., then Blockquote's Block step escapes them again.
//   - BlockRich(Blockquote(s)) doesn't make sense: Blockquote already
//     line-prefixed with `> `; BlockRich expects raw user content.
//
// For a quoted body that can contain headings, lists, nested quotes,
// or thematic breaks, use BlockquoteRich.
func Blockquote(text string) string {
	text = Block(text)
	if text == "" {
		return ""
	}
	text = strings.TrimSuffix(text, "\n")
	var sb strings.Builder
	for _, line := range strings.Split(text, "\n") {
		sb.WriteString("> ")
		sb.WriteString(line)
		sb.WriteByte('\n')
	}
	return sb.String()
}

// BlockquoteRich is the permissive counterpart of Blockquote. Both
// wrap user content as a CommonMark blockquote (each line prefixed
// with `> `), but they differ in what block-level structure inside
// the quote survives:
//
//   - Blockquote escapes line-leading block markers, so the quoted
//     body is paragraph-shaped — `# x` inside a Blockquote stays a
//     literal `#`.
//   - BlockquoteRich PRESERVES line-leading block markers, so the
//     quoted body can compose ATX headings, lists, thematic breaks,
//     nested blockquotes (`> > nested`), and other block-level
//     structure. Realm-binding defenses stay on (extension delimiters,
//     GFM table-row openers, bracket walker, fence autoclose,
//     NUL / bidi / Unicode-separator folding).
//
// Output shape — every non-empty result begins with "\n" and ends
// with "\n\n":
//
//   - Leading "\n" guarantees a clean blockquote opener even when the
//     realm concatenates `chrome + BlockquoteRich(user)` without its
//     own newline separator. Without the leading "\n", chrome ending
//     mid-line followed by "> quoted" would render `>` as literal
//     paragraph text instead of opening a blockquote.
//   - Trailing "\n\n" (blank line) cleanly ends the blockquote so a
//     realm appending `BlockquoteRich(user) + chrome` cannot pull
//     chrome bytes into the quote via CommonMark §5.2 lazy
//     continuation. Without the trailing blank line, paragraph chrome
//     immediately after BlockquoteRich would render inside the quote.
//   - BlockRich's own leading "\n\n" (paragraph-isolation blank line)
//     is stripped before line-prefixing — otherwise the output would
//     carry one or two redundant empty `> ` quoted lines at the top.
//     A single "\n" is then re-prepended at the BlockquoteRich
//     boundary so `chrome + BlockquoteRich(user)` still lands the
//     first `>` at column 0.
//   - The cross-boundary setext defense BlockRich provides is
//     redundant inside a blockquote: a setext underline inside `> `
//     content can only promote a line in the same blockquote, never
//     reach realm bytes (different CM container). BlockRich still
//     applies it, harmlessly.
//
// What attacker input produces what (rows that differ from
// Blockquote are marked CHANGED):
//
//	User attempt                                  | BlockquoteRich response
//	----------------------------------------------|------------------------------------------------
//	  --- preserved inside `> ` quote ---         |
//	# heading                                     | preserved as `> # heading` [CHANGED]
//	> nested quote                                | preserved as `> > nested quote` [CHANGED]
//	- item, * item, + item, 1. item               | preserved as `> - item` etc. [CHANGED]
//	---, ***, ___ thematic break                  | preserved [CHANGED]
//	=== or --- setext underline                   | preserved when preceded by user text;
//	                                              | escaped (\===/\---) if first non-blank
//	                                              | line of input [CHANGED]
//	| a | b | GFM table row (line-leading |)      | preserved → renders as <table> inside the
//	                                              | blockquote when followed by a delimiter row [CHANGED]
//	[text](url), ![alt](src)                      | preserved verbatim [SAME]
//	----------------------------------------------|------------------------------------------------
//	  --- escaped / stripped / folded ---         |
//	<gno-card>, any <gno-…>/</gno-…> at line-start| escaped (wildcard match) [SAME]
//	<!--, <script>, <pre>, <style>, <textarea>,   | escaped (\<…) [SAME] — CM §4.6 Types 1-5
//	  <?…?>, <!DOCTYPE…>, <![CDATA[…]]>           |   don't close on blank lines; without escape
//	  at line-start                               |   they would swallow chrome past the `> ` quote
//	[text][realm-label] ref-link USE              | both pairs escaped [SAME]
//	[^name] footnote-ref                          | both brackets escaped [SAME]
//	[label] bare shortcut-ref                     | both brackets escaped [SAME]
//	[label]: url link-reference definition        | whole region stripped [SAME]
//	code fence opened without close               | autoclosed at end of input [SAME]
//	NUL byte (\x00)                               | replaced with U+FFFD [SAME]
//	U+2028 / U+2029 / U+0085 (NEL)                | folded to `\n` [SAME]
//	bidi/zero-width controls                      | stripped [SAME]
//
// Use BlockquoteRich when the realm wants to render user content as
// a quotation that itself reads like authored markdown — the visual
// CSS containment of `<blockquote>` already demotes inner headings
// relative to realm chrome, so the "inner headings need a sandbox"
// caveat that applies to BlockRich at top level does not apply here.
//
// Not idempotent: like Blockquote, calling twice double-wraps —
// `BlockquoteRich(BlockquoteRich(s))` produces `> > content`,
// nesting the quote a level deeper each pass.
//
// Empty input (or input that reduces to nothing after BlockRich,
// e.g. a lone link-reference definition) returns "" — no blockquote
// is emitted and neither the leading "\n" nor the trailing "\n\n"
// shape applies.
func BlockquoteRich(text string) string {
	text = BlockRich(text)
	// BlockRich wraps user content with "\n\n" on each side for
	// cross-paragraph isolation. Inside a blockquote both wraps are
	// redundant — the `> ` marker creates the container boundary —
	// and they would line-prefix to two useless `> ` empty quoted
	// lines top and bottom. Strip ALL leading and trailing "\n"s so
	// the body starts and ends clean; this helper re-wraps with `\n`
	// + body + `\n\n` below.
	text = strings.TrimLeft(text, "\n")
	if text == "" {
		return ""
	}
	// Strip ALL trailing newlines so the loop produces exactly one
	// `> line` per content line, then append `\n\n` at the end so the
	// blockquote terminates cleanly (see "Output shape" above).
	text = strings.TrimRight(text, "\n")
	if text == "" {
		return ""
	}
	var sb strings.Builder
	// Leading "\n" so `chrome + BlockquoteRich(user)` cannot land the
	// first `>` mid-line.
	sb.WriteByte('\n')
	for _, line := range strings.Split(text, "\n") {
		sb.WriteString("> ")
		sb.WriteString(line)
		sb.WriteByte('\n')
	}
	// Trailing blank line so `BlockquoteRich(user) + chrome` cannot
	// pull chrome into the quote via lazy continuation.
	sb.WriteByte('\n')
	return sb.String()
}

// LinkTitle prepares user content for a CommonMark link-title or
// image-title slot — the optional quoted text after the URL in any of
// these forms:
//
//	[text](url "TITLE")
//	![alt](src "TITLE")
//	[label]: url "TITLE"
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
//	<gno-card type="..." caption="X">         attribute value
//	<gno-alert title="X">                     attribute value
//	<h5>X</h5>                                element body
//	<details><summary>X</summary>...          element body
//	<!-- X -->                                comment body (safe: `>`
//	                                          becomes `&gt;`, so user
//	                                          cannot inject `-->`)
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
//   - any URL WITHOUT a scheme — relative paths (`/path`, `./rel`,
//     `bare-path`), query-only (`?q=v`), fragment-only (`#anchor`).
//     A `:` appearing inside the URL (e.g. `/path:foo`, `?q=a:b`) is
//     NOT a scheme separator per RFC 3986 — only `:` immediately after
//     a leading `[a-zA-Z][a-zA-Z0-9+.-]*` counts.
//
// Rejected (have an unknown scheme):
//   - javascript:, data:, vbscript:, blob:, file:, etc.
//   - `//host/...` (protocol-relative — tracking-pixel vector)
//
// Returns "" if the URL is empty after trim or fails the allowlist.
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
	userNameFirstLo, userNameFirstHi           uint64
	userNameRestLo, userNameRestHi             uint64
	footnoteLabelFirstLo, footnoteLabelFirstHi uint64
	footnoteLabelRestLo, footnoteLabelRestHi   uint64
	langFirstLo, langFirstHi                   uint64
	langRestLo, langRestHi                     uint64
	bechHrpFirstLo, bechHrpFirstHi             uint64
	bechHrpRestLo, bechHrpRestHi               uint64
	bechDataFirstLo, bechDataFirstHi           uint64
	bechDataRestLo, bechDataRestHi             uint64
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

	// FootnoteLabel: [A-Za-z0-9_-] for both first and rest.
	for c := byte('A'); c <= 'Z'; c++ {
		setBit(&footnoteLabelFirstLo, &footnoteLabelFirstHi, c)
		setBit(&footnoteLabelRestLo, &footnoteLabelRestHi, c)
	}
	for c := byte('a'); c <= 'z'; c++ {
		setBit(&footnoteLabelFirstLo, &footnoteLabelFirstHi, c)
		setBit(&footnoteLabelRestLo, &footnoteLabelRestHi, c)
	}
	for c := byte('0'); c <= '9'; c++ {
		setBit(&footnoteLabelFirstLo, &footnoteLabelFirstHi, c)
		setBit(&footnoteLabelRestLo, &footnoteLabelRestHi, c)
	}
	for _, c := range []byte{'_', '-'} {
		setBit(&footnoteLabelFirstLo, &footnoteLabelFirstHi, c)
		setBit(&footnoteLabelRestLo, &footnoteLabelRestHi, c)
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

// FootnoteLabel validates an identifier used as a footnote name, link-
// reference-definition label, or {#id} anchor: ^[A-Za-z0-9_-]{1,64}$.
// Strips bidi/zero-width first. Returns s if valid, "" otherwise.
//
// Use for every shape where a markdown identifier is treated as an
// opaque key by the parser:
//
//   - footnote-definition labels:        [^FootnoteLabel(name)]: body
//   - footnote-reference labels:         see [^FootnoteLabel(name)]
//   - link-reference-definition labels:  [FootnoteLabel(label)]: url
//   - reference-link USE labels:         [text][FootnoteLabel(label)]
//   - goldmark auto-anchor {#id}:        # Heading {#FootnoteLabel(id)}
//
// The shared validator name reflects the shared charset and shared
// security goal — keep untrusted bytes out of any parser-managed
// identifier slot.
//
// On "" return, omit the footnote / LRD / anchor entirely rather than
// emitting it with raw user bytes.
func FootnoteLabel(s string) string {
	s = markdown.StripBidiAndZeroWidth(s)
	if markdown.MatchCharsetN(s, footnoteLabelFirstLo, footnoteLabelFirstHi, footnoteLabelRestLo, footnoteLabelRestHi, 1, 64) {
		return s
	}
	return ""
}

// LanguageName validates the language tag (a.k.a. "info string") for
// a fenced code block — the `go` in:
//
//	```go
//	fmt.Println("hi")
//	```
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
//	fence := sanitize.CodeFence(userCode, 3)
//	out += fence + "\n" + userCode + "\n" + fence + "\n"
//
// Typical minCount values:
//   - 1 for inline code spans (`x`)
//   - 3 for block fenced code (CommonMark §4.5 requires ≥3)
//
// `minCount < 1` is clamped to 1. Empty content returns
// strings.Repeat("`", max(minCount, 1)). Never panics.
//
// Most realms should reach for InlineCode / CodeBlock /
// LanguageCodeBlock below, which call CodeFence internally and emit
// the full code block for you. Call CodeFence directly only when
// you're rolling a custom fence emitter (e.g. a renderer that needs
// the fence length but emits the body differently).
func CodeFence(content string, minCount int) string {
	return markdown.CodeFence(content, minCount)
}

// InlineCode wraps user content as a CommonMark inline code span — the
// `code` in “ `code` “. Use for any user-derived token, identifier,
// or short literal that should render in monospace inside running
// prose: variable names, hashes, hex addresses, token symbols, error
// codes, package paths, transaction IDs.
//
// Inline code spans cannot span lines (a `\n` inside the content would
// end the span and leave the surrounding backticks as literal text),
// so all line breaks — CR / CRLF / LF, NEL (U+0085), U+2028, U+2029 —
// are folded to a single space. If you want each line of user content
// on its own row, use CodeBlock instead.
//
// Behavior:
//   - Bidi/zero-width controls are stripped (browsers honor bidi marks
//     inside `<code>`, so leaving them would let stored bytes display
//     as something different).
//   - NUL is replaced with U+FFFD.
//   - The wrapping fence is one backtick longer than the longest
//     backtick run in the content, so internal backticks can never
//     close the span prematurely.
//   - A single space pad is added on each side when content starts or
//     ends with “ ` “ or space, so leading/trailing backticks render
//     literally rather than fusing with the fence (the renderer
//     strips one space from each side per CommonMark spec).
//
// Empty input returns "" rather than a literal two-backtick string
// (which CommonMark parses as text, not as an empty code span). If
// you use InlineCode as link text and it returns "", omit the link
// entirely.
//
// Not idempotent (see package doc): wraps with a fence, so calling
// twice double-wraps.
func InlineCode(content string) string {
	content = markdown.StripBidiAndZeroWidth(content)
	content = markdown.NormalizeBreaks(content)
	content = foldNewlinesAndSeparators(content, ' ')
	content = replaceNULWithFFFD(content)
	if content == "" {
		return ""
	}
	fence := markdown.CodeFence(content, 1)
	pad := ""
	if content[0] == '`' || content[0] == ' ' ||
		content[len(content)-1] == '`' || content[len(content)-1] == ' ' {
		pad = " "
	}
	return fence + pad + content + pad + fence
}

// CodeBlock wraps user content as a CommonMark fenced code block.
// Use for any user-derived multi-line snippet that should render as a
// code block: log excerpts, JSON dumps, error backtraces, config
// snippets, posted code samples.
//
// Behavior:
//   - Bidi/zero-width controls are stripped.
//   - CR/CRLF line endings are normalized to LF; Unicode separators
//     (NEL U+0085, U+2028, U+2029) are folded to LF for line-count
//     consistency.
//   - NUL is replaced with U+FFFD per CM §2.3.
//   - The wrapping fence is at least 3 backticks (CM §4.5 minimum) and
//     sized to outscan internal backticks — an attacker cannot embed
//     a closing fence in the content.
//
// Empty content emits an empty fenced block ("```\n\n```\n"), which is
// valid CommonMark and renders as an empty `<pre><code></code></pre>`.
//
// Not idempotent (see package doc).
func CodeBlock(content string) string {
	content = markdown.StripBidiAndZeroWidth(content)
	content = markdown.NormalizeBreaks(content)
	content = foldNewlinesAndSeparators(content, '\n')
	content = replaceNULWithFFFD(content)
	fence := markdown.CodeFence(content, 3)
	return fence + "\n" + content + "\n" + fence + "\n"
}

// LanguageCodeBlock wraps user content as a fenced code block tagged
// with a programming-language hint (the "info string" after the
// opening fence, e.g. `go` in ```` ```go ````) so the renderer can
// apply syntax highlighting.
//
// An invalid `language` tag silently falls back to a tagless fence —
// the helper never returns an error or panics. If a realm author is
// debugging "why is my Go highlighting gone?", the input failed the
// language validator (charset ^[a-zA-Z0-9_+-]{1,32}$ after bidi-strip).
// This fallback exists because an unvalidated tag could contain a
// newline that injects content (e.g. a heading) onto what becomes the
// opening fence line.
//
// Content is cleaned exactly as in CodeBlock (bidi-strip, CR/CRLF
// normalize to LF, NEL/U+2028/U+2029 fold to LF, NUL→U+FFFD, fence
// sized to outscan internal backticks).
//
// Not idempotent (see package doc).
func LanguageCodeBlock(language, content string) string {
	content = markdown.StripBidiAndZeroWidth(content)
	content = markdown.NormalizeBreaks(content)
	content = foldNewlinesAndSeparators(content, '\n')
	content = replaceNULWithFFFD(content)
	fence := markdown.CodeFence(content, 3)
	lang := LanguageName(language) // "" on reject
	return fence + lang + "\n" + content + "\n" + fence + "\n"
}

// ----- Reference-style definitions -----

// FootnoteDefinition emits a GFM footnote definition — the
// `[^name]: body` form that introduces a footnote whose body is rendered
// in the page footer (or wherever the renderer chooses to place it).
// Other parts of the markdown reference the footnote by writing
// `[^name]` inline.
//
// Use for any realm-rendered footnote where the body text comes from
// user input. The realm picks the footnote name (passed as `name`,
// validated by FootnoteLabel — failure here returns ""); the user's
// content goes in `text`, which is sanitized via Block.
//
// Contract:
//   - `name`: passed raw, validated as a FootnoteLabel
//     (^[A-Za-z0-9_-]{1,64}$). Reject → return "".
//   - `text`: passed raw multi-paragraph user prose, cleaned via Block
//     (bidi-strip, line-ending normalize, LRD strip, block-marker
//     escape, ref-link USE escape, fence auto-close).
//
// Empty body → returns "" (a label without body is not a valid
// footnote definition; the markdown would parse as a paragraph
// containing the label).
//
// Output shape:
//
//	[^name]:
//	    line 1 of body
//	    line 2 of body
//	    ...
//
// The label sits on its own line and each body line gets a 4-space
// indent — the GFM continuation rule that keeps multi-paragraph body
// text bound to the footnote rather than detaching as a new paragraph.
//
// Not idempotent (see package doc): composes Block internally; passing
// already-sanitized body text double-escapes.
func FootnoteDefinition(name, text string) string {
	label := FootnoteLabel(name)
	if label == "" {
		return ""
	}
	body := strings.TrimRight(Block(text), "\n")
	if body == "" {
		return ""
	}
	var b strings.Builder
	b.WriteString("[^")
	b.WriteString(label)
	b.WriteString("]:\n")
	for _, line := range strings.Split(body, "\n") {
		if line == "" {
			b.WriteByte('\n')
		} else {
			b.WriteString("    ")
			b.WriteString(line)
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// LinkReferenceDefinition emits a CommonMark link reference definition
// (CM §4.7) — the `[label]: url "title"` form that other parts of the
// markdown reference by writing `[text][label]` or `[label]` (shortcut).
//
// Use for any realm-rendered LRD where the realm owns the label but
// any of the URL or title come from user input. The user content for
// the URL goes through URL (allowlist-based — reject → ""); the title
// goes through LinkTitle (escape).
//
// Contract:
//   - `label`: passed raw, validated as a FootnoteLabel
//     (^[A-Za-z0-9_-]{1,64}$). Realms should choose a namespaced label
//     using dashes (e.g. `r-myrealm-help`) so shortcut-reference
//     invocations from user content can't collide with bare prose
//     (`[help]`, `[click here]`). `/` is not in the FootnoteLabel
//     charset; reject → return "".
//   - `url`: passed raw, sanitized via URL. If URL rejects, the LRD is
//     skipped (return "").
//   - `title`: passed raw, sanitized via LinkTitle. Empty title → no
//     title clause emitted.
//
// The output is framed with leading and trailing blank lines so that
// the definition cannot accidentally fuse with adjacent paragraph
// content into a setext underline or a continuation line.
//
// Not idempotent (see package doc).
func LinkReferenceDefinition(label, url, title string) string {
	lbl := FootnoteLabel(label)
	if lbl == "" {
		return ""
	}
	safeURL := URL(url)
	if safeURL == "" {
		return ""
	}
	var b strings.Builder
	b.WriteString("\n\n[")
	b.WriteString(lbl)
	b.WriteString("]: ")
	b.WriteString(safeURL)
	if title != "" {
		b.WriteString(" \"")
		b.WriteString(LinkTitle(title))
		b.WriteString("\"")
	}
	b.WriteString("\n\n")
	return b.String()
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
	// Any URL with an unknown scheme (RFC 3986: `^[a-zA-Z][a-zA-Z0-9+.-]*:`)
	// is rejected — this blocks `javascript:`, `data:`, `vbscript:`, `blob:`,
	// and anything else not handled above. URLs without a scheme are
	// treated as relative and accepted (bare path, query-only, fragment).
	if hasURLScheme(s) {
		return false
	}
	return true
}

// hasURLScheme reports whether s begins with a scheme followed by ':'
// per RFC 3986 (^[a-zA-Z][a-zA-Z0-9+.-]*:). A `:` appearing later in
// the URL (e.g. `/path:foo` or `?q=a:b`) does not count.
func hasURLScheme(s string) bool {
	if len(s) == 0 {
		return false
	}
	c := s[0]
	if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')) {
		return false
	}
	for i := 1; i < len(s); i++ {
		c := s[i]
		if c == ':' {
			return true
		}
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '+' || c == '.' || c == '-') {
			return false
		}
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
