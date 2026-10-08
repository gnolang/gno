# ADR: `<gno-panel>`, a framed block for realm and static markdown

## Status

Proposed.

## Context

Realm authors and static pages (the mainnet home,
`misc/deployments/home-alias/home.mainnet.md`) render plain markdown plus the
gnoweb extensions. Production runs without `-html`, so a raw
`<div class="...">` is stripped. There was no way to draw a frame around a
block: neither a highlighted hero / call-to-action ("jumbotron"), nor a card in
a `<gno-columns>` grid (e.g. a 3x2 grid of apps, each with a bold link title
and a short description).

## Decision

Add one block extension, `<gno-panel>` … `</gno-panel>`, each tag alone on its
line, in `gno.land/pkg/gnoweb/markdown/ext_panel.go`.

- **One generic tag.** Standalone, a panel is the highlighted block; alone in a
  column (`.gno-column > .gno-panel:only-child`) the CSS makes it a card that
  fills the column, so cards in a row share a height. Authors learn one tag,
  and the renderer has no mode to pick.
- **Name.** `gno-card` is already reserved: the sanitizer docs and
  `nestdepth.go` refer to a future opaque-body sandbox under that name.
  `gno-jumbotron` reads wrong for a card in a grid. `gno-panel` covers both.
- **Flat markers, like `ext_columns`.** The open and close tags are one-line
  leaf nodes at document level, and the content between them is parsed as
  ordinary top-level blocks. goldmark therefore routes every line by its own
  rules: a `</gno-panel>` inside a fenced code block or a `<gno-foreign>` body
  stays content, and one that ends a blockquote or list item closes the panel.
  An AST transformer then moves the blocks after each open marker under it, up
  to the close marker (removed), the next gno-columns marker or the end of the
  document. Output is constant markup, `<section class="gno-panel">` …
  `</section>`; no source byte reaches the wrapper.
- **Document level only.** A panel tag counts only when its parent is the
  document, like columns. Column content and a `<gno-foreign>` body are both
  document level; a blockquote, list item or alert is not. This keeps the
  `BlockRich` guarantee: the sanitizer escapes a `<gno-…` at line start but
  not after `> ` or `- `, so without this rule sanitized user content could
  draw a realm-styled frame inside a quote or a list.
- **Empty attribute allowlist.** Any attribute (on the open or the close
  tag), a self-closing form or trailing text makes the tag invalid instead of
  silently ignored, so a later attribute (a variant, say) cannot be misread by
  an older gnoweb. Tag names are case-insensitive, as for every gno-* tag.
  Tags are read with `scanGnoTag` (`utils.go`), the zero-alloc scanner shared
  with the gno-button branch; unlike the HTML tokenizer, it does not drop
  attributes from a close tag.
- **Fail safe, never swallow.** A stray `</gno-panel>`, a malformed tag, a
  tag below document level, a panel opened inside a panel, and an opener past
  the shared `MaxGnoNestDepth` cap all become an invalid leaf rendered as
  `<!-- unexpected/invalid panel tag omitted -->` (the columns convention).
  Returning nil instead would hand the line to the type-7 HTML block parser,
  which swallows every line up to the next blank one. A panel left open is
  closed at the end of the document.
- **Columns interplay.** A gno-columns tag ends an open panel. The panel
  parser runs just ahead of the columns parser (priority 499 vs 500), pops
  the panel's depth, and leaves the line to columns, so the depth the columns
  opener sees is the same as if the panel had been closed explicitly.
- **HTML blocks.** One exception to "goldmark routes the lines": a type-6/7
  HTML block runs to the next blank line, so a `<div>` line right before
  `</gno-panel>` would swallow the close tag and stretch the panel over the
  page. The extension registers goldmark's own HTML block parser wrapped so
  that, while a panel is open, a line that ends the panel also ends a
  document-level HTML block, whatever its type (a `<!--` or `<script>` block
  is cut short too). Safe mode strips that HTML anyway. Only a tag alone on
  its line counts: `<div></gno-panel>` stays inside the HTML block
  (golden `html_same_line_close`).
- **Bare-CR line endings.** goldmark splits lines on `\n` only, so a page
  with `\r` line endings is one long line. A tag followed by `\r` is then
  not treated as a tag line, and the text after it is kept (golden
  `cr_only_line_endings`).
- **Accessibility.** The wrapper is a `<section>` with no accessible name, so
  assistive tech exposes it as a generic group, not a landmark. That fits a
  visual frame around author content; naming it would mean pointing at a
  heading id, which the `<gno-foreign>` instance does not generate.
- **Sandbox parity.** The `<gno-foreign>` inner instance loads the panel
  extension like columns and alerts, so a panel inside foreign content
  renders inside the sandbox, and a `</gno-panel>` in untrusted foreign bytes
  never reaches the host's panel.
- **Forms (side fix).** `<gno-form>` now consumes its close with
  `AdvanceToEOL`, so a form inside a blockquote no longer opens a nested quote.
- **Perf.** No regex, and panel tags never reach the HTML tokenizer; one node
  per panel tag line, nothing on pages without panels (`BenchmarkPanel`,
  `TestParsePanelLineTagNoAlloc`).
- **Shared helper.** `trimForeignLine` became `trimTagLine` in `utils.go`,
  used by foreign and panel.

## Alternatives considered

- **Two tags (`<gno-jumbotron>` and `<gno-card>`).** Same parser, two names,
  two sets of docs; the context (in a column or not) already tells them apart.
- **A container block, like `ext_alert`** (the first version of this PR). The
  content was parsed as the panel's children, and the panel's `Continue` was
  asked about every line before its children. To let a fence or foreign body
  keep a `</gno-panel>` line, it had to decide whether goldmark would route
  the line to that child, which meant copying goldmark's continuation rules
  for blockquotes, alerts and list items (`>` markers, list content offsets).
  That copy would have to follow goldmark and every new container. Flat
  markers let goldmark do the routing.
- **A `class` / `variant` attribute.** Free-form classes are an injection and
  styling-abuse surface; a fixed variant list can be added later on top of the
  empty allowlist without breaking existing content.
- **Allowing nested panels.** No use case for a frame inside a frame.
- **Escaping gno-* tags after `>` / list markers in the sanitizer** instead
  of the document-level rule. That changes a Gno stdlib and the realm-side
  package, and every future gno-* tag would have to be checked against the
  same container shapes; the parser rule closes the hole where it opens.

## Consequences

- Realms and static pages get heroes and card grids without raw HTML.
- `p/nt/markdown/sanitize` escapes a line-leading `<gno-…` (any case, 1-3
  space indent), and after a `>` or list marker a panel tag stays an inert
  comment (goldens `blockrich-gno-panel-*`, `block-gno-panel-escaped`). No
  sanitizer change in this PR. Known gap on master, fixed separately in
  `chain/markdown`: a fence indented under a list item makes the sanitizer
  skip the escape on the next unindented line, which goldmark parses at
  document level, so user content can open or close a panel there (the same
  shape already reaches `<gno-columns>` tags today).
- Panels cannot sit inside a list, a quote or an alert. Alerts, lists, quotes,
  code and forms inside a panel work.
- An unterminated fenced code block inside a panel runs to the end of the
  document, as it does at top level.
- `scanGnoTag` and its test are copied byte for byte from the gno-button
  branch; whichever PR lands second drops its copy.
- Docs: a "Panels" section in `r/docs/markdown`, next to Columns.
