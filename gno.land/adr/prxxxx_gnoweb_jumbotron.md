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
- **Container block, like `ext_alert`.** The content is parsed as ordinary
  markdown children (headings, lists, links, alerts, code). Output is constant
  markup, `<section class="gno-panel">` … `</section>`; no source byte reaches
  the wrapper, and children render through the normal, escaped paths.
- **Empty attribute allowlist.** Any attribute, a self-closing form or
  trailing text makes the tag invalid instead of silently ignored, so a later
  attribute (a variant, say) cannot be misread by an older gnoweb. Tag names
  are case-insensitive, as for every gno-* tag.
- **Fail safe, never swallow.** A stray `</gno-panel>`, a malformed tag, a
  panel inside a panel, and an opener past the shared `MaxGnoNestDepth` cap
  all become an invalid leaf node rendered as
  `<!-- unexpected/invalid panel tag omitted -->` (the columns convention).
  Returning nil instead would hand the line to the type-7 HTML block parser,
  which swallows every line up to the next blank one. An unclosed panel is
  closed at EOF by goldmark.
- **Columns interplay.** Columns live at document level only, as before. A
  gno-columns tag inside a panel closes the panel without consuming the line,
  so an unclosed panel in a column cannot swallow the separator or the
  columns close tag.
- **Opaque children.** goldmark asks a container before its children, so the
  panel passes through lines owned by an open fenced code block (a code sample
  may show `</gno-panel>`) and by an open `<gno-foreign>` block. The second is
  a security property: without it, `</gno-panel>` in untrusted foreign bytes
  would close the host's panel and the rest of the foreign body would render
  as host markdown, outside the sandbox. Golden
  `hostile_foreign_close_in_panel` covers it and fails if the guard is
  removed.
- **Sandbox parity.** The `<gno-foreign>` inner instance loads the panel
  extension like columns and alerts, so a panel inside foreign content renders
  inside the sandbox.
- **Perf.** No regex; ordinary lines never reach the HTML tokenizer in `Continue` (prefix fast paths). See `BenchmarkPanel`.
- **Shared helper.** `trimForeignLine` became `trimTagLine` in `utils.go`, used by foreign and panel.

## Alternatives considered

- **Two tags (`<gno-jumbotron>` and `<gno-card>`).** Same parser, two names,
  two sets of docs; the context (in a column or not) already tells them apart.
- **Flat markers like `ext_columns`** (open/close as sibling nodes, wrapper
  emitted at render). Simpler parser, but the wrapper could then cross a
  column boundary and produce mismatched HTML; a container block keeps the
  HTML well-formed by construction.
- **A `class` / `variant` attribute.** Free-form classes are an injection and
  styling-abuse surface; a fixed variant list can be added later on top of the
  empty allowlist without breaking existing content.
- **Allowing nested panels.** Needs the outer `Continue` to defer to an open
  inner panel; no use case for a frame inside a frame.

## Consequences

- Realms and static pages get heroes and card grids without raw HTML.
- `p/nt/markdown/sanitize` already escapes every line-leading `<gno-…`, so
  user content cannot open or close a panel (golden
  `blockrich-gno-panel-escaped`); no sanitizer change.
- An HTML block inside a panel does not shield `</gno-panel>`; safe mode
  strips raw HTML anyway, and closing on the tag keeps the frame bounded.
- An unterminated fenced code block inside a panel runs to the end of the
  document, as it does at top level.
- Docs: a "Panels" section in `r/docs/markdown`, next to Columns.
