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
- `variant` (optional): space-separated values from a whitelist
  (`outline`, `caution`, `warning`, `info`, `note`), matched case-insensitively
  and combinable (`caution outline`). Anything else is ignored.

There is no body, so there is nothing that could span lines, which answers the
first review point. The whole tag must fit on one line.

### Parsing

- **Inline parser** on `<` at priority 399, just ahead of goldmark's raw-HTML
  parser (400). It claims a tag only if it is a self-closing `gno-button`
  (ending in `/>`) with a non-empty `href` and a label that is not blank once
  its entities are decoded. Anything else returns nil and goldmark's raw-HTML
  handling takes over (stripped in safe mode).
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
side). The label is attribute text: its entities are decoded as HTML does;
then, since a button looks like first-party chrome, bidi and zero-width
characters (the set `sanitize` strips) and control characters are removed (a
line break or tab becomes a space), and the result is trimmed, so a label that
is blank or invisible after that is no label. It is escaped on output and
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

`a.gno-button` follows gnoweb's UI buttons instead of defining its own look.
In `06-blocks.css` it is grouped with `.b-btn` + `.b-btn--secondary` (default
look), with `.b-btn--ghost` (`outline`), and with the part of the ghost rule
every button shares (weight, focus ring, transition), so radius, padding,
gap, hover and focus come from one place. The realm-view rule only adapts it
to content (label wrapping, vertical rhythm, no hover underline) and maps the
variants to the semantic tokens the alerts use: `--s-color-bg-*-weak` fill,
`--s-color-text-*` text, `--s-color-border-*` border; light and dark come from
the token remaps already in place. An outline button's border takes its text
colour, since the border is its only shape, and the default outline uses
`--s-color-text-link-hover`, the link token that passes AA in dark. Every
variant passes WCAG AA for text in both themes; the measured ratios are in the
PR description.

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
    an already escaped tag keeps `Block` idempotent. Fenced code stays
    verbatim; the fence tracker's list-container gap is shared by every
    `gno-*` tag on master and is fixed separately.
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
- `r/docs/markdown` documents the syntax with copyable examples.
