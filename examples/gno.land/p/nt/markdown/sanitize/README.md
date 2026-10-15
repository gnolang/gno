> **v0 - Unaudited**
> This is an initial version of this package that has not yet been formally audited.
> A fully audited version will be published as a subsequent release.
> Use in production at your own risk.

# `sanitize` - Markdown input sanitizers

Input-cleaning primitives and safe-emit builders, one per markdown lexical slot. Wrap a user-supplied string with the matching helper before flowing it into rendered markdown, so user content cannot break out of its slot or inject new top-level structure (a heading, table, code fence, link-reference definition, HTML block, or invisible bidi/zero-width spoof).

## Usage

```go
import "gno.land/p/nt/markdown/sanitize/v0"

out := "# " + sanitize.InlineText(userTitle) + "\n\n" +
    sanitize.Block(userBody)
out += sanitize.Blockquote(userQuote)
out += sanitize.LanguageCodeBlock(realmLang, userCode)
```

## Two rules

1. **Wrap once.** Most helpers are *not* idempotent: a second pass re-escapes the bytes the first added (`\*` becomes `\\\*`, `&amp;` becomes `&amp;amp;`, a fenced block gets re-fenced). Wrap each user-derived string with at most one `sanitize.*` call. If a builder package (e.g. `p/moul/md`) already sanitizes an argument, pass the raw input, do not pre-wrap.
2. **Right helper per slot.** Match the helper to the slot the content lands in.

## Picking the right helper

| Slot | Helper |
|---|---|
| `[text](url)`, `# Heading`, `**bold**`, `![alt]`, alert title | `InlineText` |
| Multi-paragraph body (paragraph-only) | `Block` |
| Multi-paragraph body with rich structure (headings, lists, tables) | `BlockRich` |
| Multi-line blockquote | `Blockquote` / `BlockquoteRich` |
| `[text](url "title")` | `LinkTitle` |
| Table cell | `TableCell` |
| Inside an HTML tag/attribute (`<gno-card caption="X">`) | `HTMLEscape` |
| Any link URL / image src | `URL` / `ImageURL` |
| Inline / fenced code, with or without a language tag | `InlineCode` / `CodeBlock` / `LanguageCodeBlock` |
| Footnote body / link-reference definition | `FootnoteDefinition` / `LinkReferenceDefinition` |
| Validate a handle / bech32 address / label / language / nest prefix | `UserName` / `BechString` / `FootnoteLabel` / `LanguageName` / `NestedPrefix` |

## Escapers vs validators

- **Escapers** always return a transformed, safe string and never reject: any input is acceptable because the transformation makes it safe.
- **Validators** (`UserName`, `BechString`, `FootnoteLabel`, `LanguageName`, `NestedPrefix`) return the cleaned input verbatim on accept, or `""` on reject. They never half-process, so `""` unambiguously means rejected (or empty input).

## `Block` vs `BlockRich`

Both run identical realm-binding defenses; they differ in what user structure survives.

- **`Block`** — paragraph-shaped only. Escapes `#`, `>`, list markers, thematic breaks, and setext underlines. Use for leaf slots and any content that must not visually impersonate realm chrome.
- **`BlockRich`** — preserves user headings, lists, quotes, and tables. Use for content the realm intends to render with full block structure, typically inside a sandbox container (`<gno-card>`, [`<gno-foreign>`](../../foreign/v0)). Inner-heading visual containment is the realm's CSS responsibility.

Do not compose the two in either direction; pick one at the right level.

## API

Escapers (always return a safe, transformed string; never reject):

```go
func InlineText(s string) string
func Block(s string) string
func BlockRich(s string) string
func Blockquote(text string) string
func BlockquoteRich(text string) string
func LinkTitle(s string) string
func TableCell(s string) string
func HTMLEscape(s string) string
func URL(s string) string
func ImageURL(s string) string
func InlineCode(content string) string
func CodeBlock(content string) string
func LanguageCodeBlock(language, content string) string
func CodeFence(content string, minCount int) string // raw fence builder for custom emitters
func FootnoteDefinition(name, text string) string
func LinkReferenceDefinition(label, url, title string) string
```

Validators (return the cleaned input verbatim, or `""` on reject):

```go
func UserName(s string) string
func BechString(s, prefix string) string
func FootnoteLabel(s string) string
func LanguageName(s string) string
func NestedPrefix(s string) string
```

Low-level normalizers (rarely needed directly; the helpers above call them):

```go
func StripBidiAndZeroWidth(s string) string
func NormalizeBreaks(s string) string
```

## Threat model

Helpers defend against bidi/zero-width injection, line-ending homoglyphs, markdown-structure injection, CommonMark HTML-block absorption (types 1-5 that do not close on a blank line), footnote / link-reference namespace pollution, URL scheme abuse (`javascript:`, `data:text/html`, protocol-relative), unclosed code-fence leakage, and table-alignment drift.

Out of scope: no state, no URL reputation, no CSS containment, and no structural sandboxing of opaque foreign blobs (use [`foreign`](../../foreign/v0) for that).

## Notes

- Every helper is a pure function, panic-free for any string input, and runs in `O(len(input))` with bounded allocation.
- Every text-shaped helper strips bidi/zero-width characters (`Block` and `BlockRich` normalize line breaks first, then strip), so displayed text always matches stored bytes.
