# ADR: `<gno-frame>`, a bordered block for realm and static markdown

## Status

Proposed.

## Context

Realm authors and static pages (the mainnet home,
`misc/deployments/home-alias/home.mainnet.md`) render plain markdown plus the
gnoweb extensions. Production runs without `-html`, so a raw
`<div class="...">` is stripped. There was no way to draw a frame around a
block: neither a hero / call-to-action ("jumbotron"), nor a card in
a `<gno-columns>` grid (e.g. a 3x2 grid of apps, each with a bold link title
and a short description).

## Decision

Add one block extension, `<gno-frame>` … `</gno-frame>`, each tag alone on its
line, in `gno.land/pkg/gnoweb/markdown/ext_frame.go`.

- **One generic tag.** Standalone, a frame is a bordered block; alone in a
  column (`.gno-column > .gno-frame:only-child`) the CSS makes it a card that
  fills the column, so cards in a row share a height. Authors learn one tag,
  and the renderer has no mode to pick.
- **Name and style.** `gno-frame` names what the block draws: a thin border
  (`--s-border`) with rounded corners and padding, on a transparent
  background, so it reads the same in light and dark themes and around a
  hero or a grid card alike.
- **Flat markers, like `ext_columns`.** The open and close tags are one-line
  leaf nodes at document level, and the content between them is parsed as
  ordinary top-level blocks. goldmark therefore routes every line by its own
  rules: a `</gno-frame>` inside a fenced code block or a `<gno-foreign>` body
  stays content, and one that ends a blockquote or list item closes the frame.
  An AST transformer then moves the blocks after each open marker under it, up
  to the close marker (removed), a gno-columns marker that is not part of a
  grid closing inside the frame, or the end of the document. Output is constant markup, `<section class="gno-frame">` …
  `</section>`; no source byte reaches the wrapper.
- **Document level only.** A frame tag counts only when its parent is the
  document, like columns. Column content and a `<gno-foreign>` body are both
  document level; a blockquote, list item or alert is not. This keeps the
  `BlockRich` guarantee: the sanitizer escapes a `<gno-…` at line start but
  not after `> ` or `- `, so without this rule sanitized user content could
  draw a realm-styled frame inside a quote or a list.
- **Empty attribute allowlist.** Any attribute (on the open or the close
  tag) or a self-closing form makes the tag invalid instead of silently
  ignored, so a later attribute (a variant, say) cannot be misread by an older
  gnoweb. Text after the tag makes the line no tag line at all, as with
  gno-columns, so the paragraph keeps the text (`<gno-frame> hello` keeps
  `hello`). Tag names are case-insensitive, as for every gno-* tag.
  Tags are read with `scanGnoTag` (`utils.go`), the zero-alloc scanner shared
  with the gno-button branch; unlike the HTML tokenizer, it does not drop
  attributes from a close tag.
- **Fail safe, never swallow.** A stray `</gno-frame>`, a malformed tag, a
  tag below document level, a frame opened inside a frame (outside the card
  case below), and an opener past
  the shared `MaxGnoNestDepth` cap all become an invalid leaf rendered as
  `<!-- unexpected/invalid frame tag omitted -->` (the columns convention).
  Returning nil instead would hand the line to the type-7 HTML block parser,
  which swallows every line up to the next blank one. A frame left open is
  closed at the end of the document.
- **Columns interplay.** A frame holds complete grids: a `<gno-columns>`
  opened inside a frame, its separators and its `</gno-columns>` stay inside
  it, and a column of that grid may hold one frame at a time (a card; it
  ends at its close tag or the next columns tag). A card opener refused
  there (an attribute, or the depth cap) leaves its `</gno-frame>` an invalid
  leaf too, so that close does not end the outer frame (goldens
  `invalid_card_attrs_keeps_frame`, `invalid_card_at_depth_cap`). Any columns tag that would
  leave a grid half inside ends the frame instead: a separator or close of a
  grid opened before the frame (a frame in a column), and a `<gno-columns>`
  past the depth cap, so that grid still opens. A grid opened in a frame
  whose `</gno-frame>` comes before `</gno-columns>` is the other half case:
  the transformer ends the frame just before the grid, and the stray close
  renders as a comment. The frame parser runs just ahead of the columns
  parser (priority 499 vs 500), so when it ends a frame on a columns tag its
  depth is popped before the columns opener pushes. The columns parser marks
  an opener kept by a frame (`inFrame`), and the frame transformer runs after
  the columns one (501 vs 500), so a grid left open at EOF already has its
  close; the transformer then checks the grid closes before the frame's own
  close with one forward scan per grid, which never passes that close, so
  the work stays linear (`TestFrameGridScanLinear`).
  In CSS, a card grid at the edge of a frame drops its stacked-row margin,
  so the frame's padding alone spaces it.
- **HTML blocks.** One exception to "goldmark routes the lines": a type-6/7
  HTML block runs to the next blank line, so a `<div>` line right before
  `</gno-frame>` would swallow the close tag and stretch the frame over the
  page. The extension registers goldmark's own HTML block parser wrapped so
  that, while a frame is open, a frame tag line (open, close or invalid) or a
  gno-columns tag line also ends a
  document-level HTML block (with no frame open, the wrapper leaves the line
  to goldmark's own parser, so the HTML block checks run once), whatever its type (a `<!--` or `<script>` block
  is cut short too). Safe mode strips that HTML anyway. So a `<div>` right
  above a card's `<gno-frame>` leaves the card's tag to the frame parser
  (golden `html_block_before_card_open`). Only a tag alone on
  its line counts: `<div></gno-frame>` stays inside the HTML block
  (golden `html_same_line_close`).
- **Bare-CR line endings.** goldmark splits lines on `\n` only, so a page
  with `\r` line endings is one long line. A tag followed by anything, `\r`
  or ` \r` included, is then not treated as a tag line, and the text after
  it is kept (goldens `cr_only_*`).
- **Accessibility.** The wrapper is a `<section>` with no accessible name, so
  assistive tech exposes it as a generic group, not a landmark. That fits a
  visual frame around author content; naming it would mean pointing at a
  heading id, which the `<gno-foreign>` instance does not generate.
- **Sandbox parity.** The `<gno-foreign>` inner instance loads the frame
  extension like columns and alerts, so a frame inside foreign content
  renders inside the sandbox, and a `</gno-frame>` in untrusted foreign bytes
  never reaches the host's frame.
- **Forms (side fix).** `<gno-form>` now consumes its close with
  `AdvanceToEOL`, so a form inside a blockquote no longer opens a nested quote.
- **Perf.** No regex, and frame tags never reach the HTML tokenizer; one node
  per frame tag line, nothing on pages without frames, `<`-leading lines
  included (`BenchmarkFrame`,
  `TestParseFrameLineTagNoAlloc`).
- **Shared helper.** `trimForeignLine` became `trimTagLine` in `utils.go`,
  used by foreign and frame.

## Alternatives considered

- **`<gno-jumbotron>`.** The first name. It promises a large highlighted hero
  and reads wrong for a card in a grid.
- **`<gno-panel>`** (the name during review), with a filled surface
  background. A panel suggests a filled UI surface; the chosen style is a
  plain border on a transparent background, which `frame` describes.
- **`<gno-card>`.** Reserved: the sanitizer docs and `nestdepth.go` refer to
  a future opaque-body sandbox under that name.
- **Two tags (a hero and a card).** Same parser, two names, two sets of docs;
  the context (in a column or not) already tells them apart.
- **A container block, like `ext_alert`** (the first version of this PR). The
  content was parsed as the frame's children, and the frame's `Continue` was
  asked about every line before its children. To let a fence or foreign body
  keep a `</gno-frame>` line, it had to decide whether goldmark would route
  the line to that child, which meant copying goldmark's continuation rules
  for blockquotes, alerts and list items (`>` markers, list content offsets).
  That copy would have to follow goldmark and every new container. Flat
  markers let goldmark do the routing.
- **A `class` / `variant` attribute.** Free-form classes are an injection and
  styling-abuse surface; a fixed variant list can be added later on top of the
  empty allowlist without breaking existing content.
- **Allowing nested frames.** No use case for a frame directly inside a
  frame; the one nesting allowed is a card in a column of a framed grid.
- **Escaping gno-* tags after `>` / list markers in the sanitizer** instead
  of the document-level rule. That changes a Gno stdlib and the realm-side
  package, and every future gno-* tag would have to be checked against the
  same container shapes; the parser rule closes the hole where it opens.

## Consequences

- Realms and static pages get heroes and card grids without raw HTML.
- `p/nt/markdown/sanitize` escapes a line-leading `<gno-…` (any case, 1-3
  space indent), and after a `>` or list marker a frame tag stays an inert
  comment (goldens `blockrich-gno-frame-*`, `block-gno-frame-escaped`). No
  sanitizer change in this PR. Known gap on master, tracked in #6300: a
  fence indented under a list item makes the sanitizer
  skip the escape on the next unindented line, which goldmark parses at
  document level, so user content can open or close a frame there (the same
  shape already reaches `<gno-columns>` tags today).
- Frames cannot sit inside a list, a quote or an alert. Alerts, lists, quotes,
  code, forms and complete column grids (with one card per column at a time)
  inside a frame work.
- An unterminated fenced code block inside a frame runs to the end of the
  document, as it does at top level.
- `scanGnoTag` and its test are copied byte for byte from the gno-button
  branch; whichever PR lands second drops its copy.
- Docs: a "Frames" section in `r/docs/markdown`, next to Columns.
