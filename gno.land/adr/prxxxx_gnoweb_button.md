# ADR: `<gno-button />`, a self-closing link styled as a button

## Status

Proposed. Supersedes the closed PR #4481.

## Context

Realms want call-to-action links ("Read the blog", "Vote", "Delete") that
stand out from inline text. Markdown has no such construct, and raw HTML is
stripped by gnoweb.

#4481 (2025) added two syntaxes for it: a `{Label | options}(url)` shorthand
and a `<gno-button href="…" content="…" />` tag. The review asked for changes
(2025-07-17):

- the `<tag> text </tag>` shape should not be used for an inline element,
  since it cannot carry multi-line text and is "confusing and misleading";
  a self-closing `<tag />` was suggested instead;
- the tag should be parsed with `ParseHTMLTokens`, like the other `gno-*`
  tags, rather than by hand.

The PR then went stale and was closed in 2026-06. Since then the markdown
package has grown a consistent set of `gno-*` extensions (columns, forms,
alerts, foreign sandbox, mentions) with shared conventions: tags parsed with
`golang.org/x/net/html`, fail-closed fall-through to raw HTML (stripped in safe
mode), links routed through one link transformer that owns URL safety.

## Decision

One syntax: a self-closing inline tag with everything in attributes.

```markdown
<gno-button href="/r/gnoland/blog" label="Read the blog" variant="outline" />
```

- `href` (required): goes through the link extension like any markdown link.
- `label` (required): plain text, HTML-escaped, never parsed as markdown.
- `variant` (optional): space-separated values from a whitelist, `outline`
  plus the six alert kinds (`note`, `tip`, `caution`, `warning`, `success`,
  `info`), matched case-insensitively and combinable (`caution outline`).
  Anything else is ignored.

There is no body, so there is nothing that could span lines, which answers the
first review point. The whole tag must fit on one line.

### Parsing

- **Inline parser** on `<` at priority 399, just ahead of goldmark's raw-HTML
  parser (400). It claims a tag only if it is a self-closing `gno-button`
  (ending in `/>`) with an `href` and a label that are not blank once their
  entities are decoded. Anything else returns nil and goldmark's raw-HTML
  handling takes over (stripped in safe mode).
- **Escaped `<`.** The inline parser also returns nil when an odd run of
  backslashes precedes the `<` in the source. goldmark can carry a backslash
  escape over a line break (a line ending in `\` plus two spaces, or in
  `\\\`), eat the backslash `sanitize.Block` put before `<gno-button` and
  call the parser on the `<`. The sanitizer cannot mirror goldmark's line
  handling, so the parser checks the escape itself.
- **Links in links.** A button inside a link label behaves like a link there:
  goldmark keeps the inner link and turns the outer one into text, as
  CommonMark does for `[a [b](c) d](e)`, so no `<a>` is ever nested. A button
  inside brackets that never become a link (`[see <gno-button … />]`) is kept,
  which the earlier `IsInLinkLabel` guard lost. In an image's alt, goldmark
  renders children as text only, so the label becomes alt text. No guard or
  transformer is needed.
- **Alert titles.** An alert title renders as `<summary>`, which must not hold
  interactive content: a button there is reduced to its label text. A plain
  markdown link in a title still renders as a link, as on master; fixing that
  belongs to the alert extension.
- **Block parser** on `<` at priority 899, just ahead of the HTML block parser
  (900). A line holding only a tag is a CommonMark type-7 HTML block start, so
  without this a button alone on its line (or as a list item) would be
  swallowed, along with every following line up to the next blank one, before
  inline parsing ever ran. The block parser opens a line that starts with the
  tag as a paragraph by delegating to goldmark's own paragraph parser, so the
  result behaves like any paragraph (continuation lines, setext, paragraph
  transformers). It checks only the tag name (`hasGnoTagPrefix`), so even a
  tag too long or malformed for the scanner keeps the line out of an HTML
  block: a rejected button is stripped inline and the lines after it stay
  visible.
- **Tag scanner.** The inline parser reads the tag with `scanGnoTag` in `utils.go`:
  a hand-written scanner for one `<name …>` tag on one line, within a byte
  bound (`maxButtonTagLen`, 2 KB), that returns the tag's length and hands
  each attribute to a callback as raw bytes aliasing the source. It allocates
  nothing, valid tag or not (`BenchmarkParseButtonTag`: 0 allocs/op). Tag and
  attribute names are case-insensitive and the first occurrence of an
  attribute wins, as in HTML. `scanGnoTag`, `hasGnoTagPrefix` and the line
  parser (`gnoTagLineParser`) live in `utils.go`, identical on this branch and
  on the `<gno-icon />` PR, so both body-less inline tags use one
  implementation.
- **Bound.** Every `<gno-button` prefix in a line is a parse attempt. Reading
  to the end of the line made a long line of unterminated tags quadratic
  (4000 tags, 96 KB: 1.17 s). The scanner stops at the bound and at the next
  `<` outside a quoted value, so each attempt reads one tag and the line is
  linear (8 ms; `BenchmarkButtonHostileLine`). A longer tag is not a button.
- **Why not `ParseHTMLTokens`.** The review asked for it, and the first
  version of this PR used the same `x/net/html` tokenizer. It does not fit an
  inline tag: `ParseHTMLTokens` tokenizes a whole line and drops offsets,
  while an inline tag must know how many bytes it consumed so the text after
  it stays text; the tokenizer also allocates a 4 KB buffer per attempt and
  decodes attribute values, which the href must not get (see below). The
  block-level `gno-*` tags that own their line keep using `ParseHTMLTokens`.

### Rendering and safety

The parser emits a plain `*ast.Link`, as the mention parser does, with a
marker attribute holding the class. The link transformer turns it into a
`GnoLink`. The button therefore gets, unchanged: destination resolution,
link-type classification, `rel="noopener nofollow ugc"` on external links, the
external/internal/tx/user icons, and the dangerous-URL guard in
`renderGnoLink`. `renderGnoLink` gains a generic `gno:class` node attribute
(not settable from markdown) that it emits as the class.

The href is decoded once, exactly as a markdown link destination is: the
scanner keeps the raw bytes, and the link pipeline's `resolveDestination`
(backslash escapes, then entities) decodes them. `href="?q=&amp;lt;"` yields
`?q=&lt;` and `href="/r/a\_b"` yields `/r/a_b`, the same destinations
`[x](?q=&amp;lt;)` and `[x](/r/a\_b)` yield (a golden renders both side by
side); an href blank once decoded (`href="&#32;"`) is no href. The label's
entities are decoded the same way, named and numeric references ending in `;`
(`&not=` stays as written); then, since a button looks like first-party
chrome, bidi and zero-width characters (the set `sanitize` strips), format
characters (Cf) and control characters are removed (a line break or tab
becomes a space), and the result is trimmed. A label with no visible rune
left, only spaces or Hangul fillers (U+3164 and kin), is no label. It is escaped on output and
never parsed as markdown.

On top of that, the parser rejects outright (fall-through, stripped):
`javascript:`/`vbscript:`/`file:` after entity resolution, every `data:` URI
(goldmark lets `data:image/*` through for images), and any control byte (a
browser strips tab and newline from URLs, so `java&#x09;script:` would hide its
scheme from a prefix check). Unknown attributes (`onclick`, `class`, `style`)
are never emitted. `variant` values never reach the output: only
whitelisted classes are emitted, never the author's text.

The extension is **not loaded inside `<gno-foreign>`**. A button is first-party
call-to-action chrome, the same reasoning that keeps `gno-form` out of the
sandbox. Inside foreign content the tag stays raw HTML and is stripped.

### CSS

The button has its own `/* ===== BUTTON COMPONENT ===== */` block in
`06-blocks.css`, in the same `.c-realm-view, .c-readme-view` scope as the
FORM, COLUMNS, ALERT and FOREIGN blocks, named like them (`.gno-button`, and
`.gno-button-<variant>` as `.gno-alert-<variant>`). It shares no selector with
the UI buttons: an earlier version grouped it into the `.b-btn` rules, which
tied realm content to chrome styles. The UI buttons are unchanged by this
block; full-page captures of a `$help` page, an action page and a `$state`
page (light and dark) differ by 0 pixels before and after, and every
`.b-btn`'s computed style, hover included, is identical.

One base rule owns the shape, with the `.b-btn` silhouette (radius
`--s-rounded-sm`, padding `--g-space-1` / `--g-space-2`, gap `--g-space-1-5`,
1 px border, focus ring `--s-focus-ring`). Variants only set four local custom
properties the base rule consumes (fill, border, text, and the colour an
outline uses), all from semantic tokens; light and dark come from the token
remaps already in place. The default is the brand-green fill of the form
submit button (`--s-color-bg-brand-default`, `--s-color-border-brand-default`,
`--s-color-text-base`), so a content button does not read as grey chrome; the
six alert-kind variants use the alert tokens (`--s-color-bg-*-weak`,
`--s-color-text-*`, `--s-color-border-*`). One exception: `--s-color-text-tip`
is under AA in dark (2.94:1 on its fill, the alerts share that gap), so the
`tip` text is 70% of it mixed with `--s-color-text-primary`: 4.75:1 in dark,
still purple in light. An outline
button's border takes its text colour, since the border is its only shape; the
default outline is neutral (`--s-color-text-secondary`), a secondary button:
in dark the only green text that passes AA is the success token (the link
token gives 3.77:1), so a green default outline matched `success outline`. It
matches `note outline` in dark instead, the neutral alert kind. Hover thickens the border with a
ring instead of fading the fill: the form submit's `opacity: 0.9` dropped
`caution` in dark to 4.28:1. The link type icon is drawn in the button's text
colour, so it stays visible on every fill.

Measured text contrast (light / dark, identical at rest and on hover): default
6.27 / 6.27, caution 10.34 / 4.95, warning 7.13 / 9.59, info 8.13 / 6.35, note
17.47 / 8.77, tip 13.82 / 4.75, success 9.79 / 5.62; outline 7.10 / 10.26,
caution outline 10.98 / 6.23, warning outline 7.58 / 12.64, info outline 8.66
/ 7.92, note outline 19.91 / 10.28, tip outline 15.53 / 6.02, success outline
11.44 / 7.14. Every variant passes WCAG AA (4.5:1) in both themes. Outline
borders match their text (≥ 6:1); filled borders under 3:1 (warning light
1.67, note dark 2.71) are the alert tokens, the label identifying the button.

## Alternatives considered

- **Body tag `<gno-button href="…">Label</gno-button>`**: rejected in review.
  An inline element with a body invites multi-line content that CommonMark
  inline parsing cannot delimit reliably.
- **Shorthand `{Label | options}(url)`**: dropped. A second syntax doubles the
  surface to review and test. `{…}(…)` is ordinary text today, so claiming it
  would change existing pages. And its option separator is a pipe, which
  collides with GFM table cells.
- **Own renderer for the button node**: rejected. It would duplicate the link
  safety logic (resolution, scheme guard, rel, untrusted handling) that
  already lives in one place.
- **Inline parser only**: does not work for a button alone on its line or in a
  list item (type-7 HTML block, see above).
- **Claim invalid tags and emit an error comment** (like columns/forms): not
  needed. Falling through matches `gno-foreign`'s fail-closed approach and
  keeps one code path.

## Consequences

- Realms get a button with one obvious syntax. Invalid tags disappear silently
  in safe mode rather than rendering something unexpected.
- Cost: reading a tag allocates nothing. An accepted button costs about 16
  allocations more than the equivalent markdown link (`BenchmarkButton`): the
  link node, its class string and the label. A line starting with a button is
  scanned twice (block check, then inline parse), with no allocation.
- A button inside a GFM table cell must write `|` in an attribute as
  `&#124;`: the table splits cells before inline parsing, and `\|` would keep
  its backslash because the label is raw text. Documented and tested.
- Sanitize (`chain/markdown`, used by `p/nt/markdown/sanitize`):
  - `InlineText` escapes `<`, so no button.
  - `Block`/`BlockRich` escape the `<` of every `<gno-button` that is not
    already escaped, wherever it sits: line start, indented, mid-line, after
    a list, quote, heading or table marker, after `\f`, `\v`, a NBSP, CRLF
    or U+2028, inside emphasis. A line-start-only rule was bypassable by all
    of those, because the tag is inline. Code spans are not spared: whether
    a backtick opens one depends on goldmark's inline precedence (raw HTML,
    comments, link destinations and titles, spans across lines), and a
    line-scan guess that left the tag live was bypassed six ways. Inside a
    real code span the backslash shows, which is the accepted cost. Skipping
    an already escaped tag keeps `Block` idempotent. Fenced code is escaped
    too: the fence tracker can count a line as fenced while goldmark does
    not (a fence line inside an HTML block, or the list-container gap
    tracked in #6300), and such a line would render a live button. The
    backslash shows in real fenced code, the same cost as in a code span.
    The escape runs on the whole input before the bracket walker: added
    after it, the backslash broke a pointy link destination
    `[a](<gno-button x>)` the walker had kept as a link, and `[a]` bound to
    a realm reference definition.
  - The block-level `<gno-…>` line escape now puts its backslash after an
    indent of under 4 columns (right before the `<`); from 4 columns, where
    the line may be indented code, it keeps the line-start backslash it had.
  - **Button and icon differ on purpose.** The `<gno-icon />` extension
    exempts its tag from this escape, because an icon is an allowlisted
    glyph that carries no link, and escaping it only at line start would
    make it render or not depending on its position. A button is a link
    styled as first-party call to action, which user content should not be
    able to produce, so `<gno-button` stays escaped wherever the sanitizer
    looks today.
  - **Upgrade ordering.** `chain/markdown` is a native, so this escape only
    protects user content on a chain whose binary runs it. gnoweb renders
    buttons as soon as it is deployed; until the chain is upgraded,
    `sanitize.Block`, `BlockRich` and `Blockquote` output computed on chain
    lets `<gno-button` through, and gnoweb renders it as a button. The new
    escape also changes `Block` and `BlockRich` output on chain, bytes nodes
    agree on, so it ships as a MINOR coordinated upgrade (RELEASING.md).
    The gap is accepted rather than gated behind a flag: sanitized text can
    already carry a plain link to the same `$help` destination, so a button
    adds styling, not a capability, and the gap closes with the upgrade.
- `r/docs/markdown` documents the syntax with copyable examples.
