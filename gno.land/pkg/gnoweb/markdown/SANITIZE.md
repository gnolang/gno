# Sanitize integration tests

Driver: [`sanitize_integration_test.go`](sanitize_integration_test.go).
Fixtures: [`golden/sanitize/*.txtar`](golden/sanitize/).

Each fixture exercises one helper from `p/nt/markdown/sanitize/v0`
(gno-side, on top of the `chain/markdown` natives) against one attacker
input, optionally substitutes the sanitized output into a surrounding
markdown template, renders that template through the production gnoweb
goldmark extension chain (`NewGnoExtension`), and locks in both the
sanitize output and the rendered HTML as goldens.

## File format

One `.txtar` per case. Three pieces:

1. A **directive comment block** before the first `-- section --` marker.
2. The **`-- input.md --` section** — the attacker bytes (required).
3. Optional **`-- output.md --`** (sanitize result) and
   **`-- output.html --`** (rendered HTML) — regenerated with the
   update flag, never hand-edited.

### Directives

| Directive | Required for | Meaning |
|---|---|---|
| `// MARKDOWNFUNC: <Name>` | every case | which sanitize helper to invoke |
| `// CONTEXT: <template>` | escapers / URL filters | surrounding markdown; the sanitized output is substituted at `%s`. `\n` in the template is a real newline. |
| `// ARGS: <Go literal>` | BechString / CodeFence / FootnoteDefinition / LanguageCodeBlock / LinkReferenceDefinition | extra argument(s): `"g"` (quoted string) for BechString prefix, FootnoteDefinition name, or LanguageCodeBlock language tag; `3` (int) for CodeFence minCount; `"url","title"` (two comma-separated quoted strings) for LinkReferenceDefinition |
| `// INPUT_ESCAPED` | optional | decode Go-string escapes in `input.md` (so you can author CR `\r`, NUL `\x00`, lone control bytes, `\u202E`, etc. — editors normalize line endings, this gets around that) |

Validators (`UserName`, `BechString`, `FootnoteLabel`, `LanguageName`,
`NestedPrefix`) take no `CONTEXT` directive — their output isn't
markdown, so there's nothing to render. The case checks only the
returned string against `output.md`.

### Why CONTEXT exists

Most sanitize outputs are **fragments**, not standalone markdown.
`InlineText("foo *bar*")` returns `foo \*bar\*` — meant to be placed
inside a `# `, `[ ]( )`, `**...**`, or table cell. Rendering it as a
top-level markdown document misses the whole point.

`CONTEXT` declares the surrounding markdown the realm would build:

```
// MARKDOWNFUNC: InlineText
// CONTEXT: # %s
-- input.md --
hello *world*
```

→ sanitize → `hello \*world\*`
→ substituted into context → `# hello \*world\*`
→ goldmark renders → `<h1>hello *world*</h1>`

This is the threat model: an attacker controls `%s`, the realm
controls everything else. The fixture verifies that the realm's
chrome plus the attacker's bytes can't combine into anything the
realm didn't intend.

A few helpers don't need `CONTEXT`:
- Validators (above) — output isn't markdown.
- `Block` — output IS a top-level markdown chunk; use `CONTEXT: %s`.

`CodeFence` is a special case: its output is the fence delimiter
(used twice). Author the template with two `%s` markers:
`// CONTEXT: %s\nuser code\n%s`.

## Update workflow

After authoring a new fixture or changing the sanitize implementation:

```
go test ./gno.land/pkg/gnoweb/markdown -run TestSanitizeIntegration -update-golden-tests
git diff gno.land/pkg/gnoweb/markdown/golden/sanitize/
```

Review the diff — each `output.md` / `output.html` change is a behavior
change you're locking in. Then commit.

Without `-update-golden-tests` the test runs in compare mode and fails
on any mismatch.

## Adding a case

Minimal skeleton:

```
// MARKDOWNFUNC: <helper>
// CONTEXT: <template with %s, if applicable>
// ARGS: <if applicable>
// INPUT_ESCAPED       (if you need CR / NUL / lone control bytes)
-- input.md --
<the attacker input>
```

Then `-update-golden-tests` to seed the output sections. Filename should
be `kebab-case-describing-what-it-tests.txtar`.

## Threat-surface coverage

| Helper | Cases | Threats covered |
|---|---|---|
| `InlineText` | 14 | bidi/ZWSP/NEL strip, CR-only fold, NUL→FFFD, backslash-escape-order, ampersand-entity, leading/trailing-`#` in ATX context, link-text bracket breakout, `=` and `\|` carve-outs |
| `Block` | 16 | heading/blockquote/list/thematic/setext injection, fence autoclose, LRD strip, ref-link USE collision, footnote-ref `[^` collision (basic + with preceding backslash, CM §2.4 parity), ext-delim (`<gno-card>`, `</gno-columns>`, `\|\|\|`), CR / U+2028 / U+2029 fold |
| `Blockquote` | 8 | basic, bidi strip, blank-line preservation, CR normalize, empty input, leading-marker escape, fence autoclose, LRD strip |
| `LinkTitle` | 4 | quote/apostrophe/paren delimiters, newline fold |
| `TableCell` | 2 | pipe escape, tab→space |
| `HTMLEscape` | 5 | attribute injection, element body, ampersand, comment context, `-->` terminator bypass |
| `URL` | 10 | `javascript:` (lowercase + mixed case), leading whitespace bypass, protocol-relative, `blob:`, `mailto:` `?body=`/`cc=`/`bcc=`, embedded CRLF, relative + fragment accept |
| `ImageURL` | 5 | `data:text/html` reject, `data:image/svg+xml` accept, `mailto:` / protocol-relative as image src |
| `UserName` | 4 | digit-first / uppercase reject, `_`/`-` accept, RLO-stripped-then-validated |
| `BechString` | 4 | `"g"` prefix, `""` any-prefix (`bc1...`), `"gpub"` prefix, wrong-prefix reject |
| `FootnoteLabel` | 2 | valid charset, space reject |
| `LanguageName` | 3 | `c++` accept, space reject, newline-injection reject |
| `NestedPrefix` | 3 | blockquote `>` accept, `##` reject, newline-in-prefix reject |
| `CodeFence` | 4 | grow-from-3-backticks, grow-from-4-backticks, min-clamp-zero (never-panic invariant) |
| `InlineCode` | 5 | basic, embedded backticks, multi-line fold, NUL, leading/trailing backtick padding |
| `CodeBlock` | 5 | basic, bidi strip, CR normalize, embedded-fence neutralization, empty content |
| `LanguageCodeBlock` | 3 | valid tag, rejected tag (silent fallback), embedded fence in body |
| `Link` | 4 | basic, rejected URL, javascript scheme, bracket breakout in text |
| `FootnoteDefinition` | 3 | basic body, multi-paragraph continuation indentation, rejected name suppresses output |
| `LinkReferenceDefinition` | 3 | basic label/url, with title, rejected URL suppresses output |

103 fixtures total. Grow the corpus by enumerating the threat surface
for each helper as new attacks/CVEs/audit findings surface — every
finding becomes a permanent regression test.

## Implementation notes

- The driver invokes the real `sanitize.X` gno helper through a per-case
  gno VM (constructed via `CacheWrap` over a shared base store, so cases
  stay isolated without re-loading the examples directory). This means
  the test corpus exercises the same code path as production realms —
  no Go reimplementation to drift against.
- HTML is rendered with `NewGnoExtension()` — no image validator, so
  cases that depend on validator state should be authored carefully.
  Add `WithImageValidator(...)` to the driver if a case requires it.
- The driver strips one trailing newline from each section (a txtar
  formatting artifact) before comparing. `INPUT_ESCAPED` decoding runs
  after that strip.
- `output.html` of a rejected URL is `<a href="">click</a>` — the link
  becomes inert, but the structural HTML survives. That's correct
  behavior, not a sanitization bypass.
