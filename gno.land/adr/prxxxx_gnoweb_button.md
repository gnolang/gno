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
  parser (400). It claims a token only if it is a well-formed
  `SelfClosingTagToken` named `gno-button` with a non-empty `href` and
  `label`. Anything else returns nil and goldmark's raw-HTML handling takes
  over (stripped in safe mode). It declines inside a link label, which would
  otherwise nest `<a>` elements.
- **Block parser** on `<` at priority 899, just ahead of the HTML block parser
  (900). A line holding only a tag is a CommonMark type-7 HTML block start, so
  without this a button alone on its line (or as a list item) would be
  swallowed, along with every following line up to the next blank one, before
  inline parsing ever ran. The block parser opens such a line as a paragraph
  by delegating to goldmark's own paragraph parser, so the result behaves like
  any paragraph (continuation lines, setext, paragraph transformers). It only
  checks the tag shape. Attribute validation stays in the inline parser, so a
  rejected button is stripped inline instead of taking the following lines
  with it.
- A parse attempt reads at most `MaxButtonTagLen` (2 KB) bytes and never
  past the next `<gno-button` prefix, and skips the tokenizer when that window
  holds no `/>`. Every prefix in a line is an attempt, so an unbounded read
  made a long line of unterminated tags quadratic (4000 tags, 96 KB: 1.17 s);
  bounded, it is linear (8 ms; `BenchmarkButtonHostileLine`). A longer tag,
  or one with `<gno-button` inside an attribute value, is not a button.
- The tag is tokenized with `x/net/html`, the tokenizer behind
  `ParseHTMLTokens`. `ParseHTMLTokens` itself is not called because it
  tokenizes the whole line and drops offsets, while an inline tag has to know
  how many bytes it consumed so text after it stays text. `parseButtonTag`
  reads only the first token and its raw length. A byte-prefix check
  (case-insensitive, like the other `gno-*` tags) runs first, so a `<` that is
  not a button costs no allocation (covered by a test).

### Rendering and safety

The parser emits a plain `*ast.Link`, as the mention parser does, with a
marker attribute holding the class. The link transformer turns it into a
`GnoLink`. The button therefore gets, unchanged: destination resolution,
link-type classification, `rel="noopener nofollow ugc"` on external links, the
external/internal/tx/user icons, and the dangerous-URL guard in
`renderGnoLink`. `renderGnoLink` gains a generic `gno:class` node attribute
(not settable from markdown) that it emits as the class.

The href is decoded once, as HTML decodes an attribute: the tokenizer resolves
its entities, and the parser re-escapes `&` and `\` before storing it, so the
link pipeline's `resolveDestination` gives the value back unchanged instead of
decoding it a second time. `href="?q=&amp;lt;"` therefore yields `?q=&lt;`,
the same destination `[x](?q=&amp;lt;)` yields, and a backslash in an href is
kept rather than read as a markdown escape.

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

`a.gno-button` in `06-blocks.css`, inside the realm/readme view block, built on
existing semantic tokens. Each variant sets three local properties (fill, text
on fill, outline text), so light and dark mode come from the token remaps
already in place. `warning` uses dark text on its light-yellow fill.

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
- Cost: each claimed tag allocates the `x/net/html` tokenizer buffer (~4 KB,
  ~46 allocations per button in `BenchmarkButton`, against an equivalent
  markdown link). A line holding only a button is tokenized twice (block shape
  check, then inline parse). Non-button `<` bytes cost nothing.
- A button inside a GFM table cell must write `|` in an attribute as
  `&#124;`: the table splits cells before inline parsing, and `\|` would keep
  its backslash because the label is raw text. Documented and tested.
- Sanitize (`chain/markdown`, used by `p/nt/markdown/sanitize`):
  `InlineText` escapes `<`; `Block`/`BlockRich` escape a `<gno-…>` line
  start. That escape used to put its backslash before the indent, which was
  enough for the block-level `gno-*` parsers but left an indented button live
  (the backslash only escaped a space). It now goes right before the `<`.
  A button in the middle of a line still passes through `Block` and renders:
  it is only a link, with the same URL checks as the `[text](url)` links
  `Block` already preserves, but a more prominent one. A sanitize fixture
  pins that behavior. The indent fix is a stopgap: the general fix is to
  escape every `<gno-button` in `Block`/`BlockRich` outside code spans, as
  `InlineText` escapes `<`, which is a separate change to the sanitize
  contract.
- `r/docs/markdown` documents the syntax with copyable examples.
