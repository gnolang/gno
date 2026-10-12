# ADR: `<gno-details>`, a neutral collapsible block in gnoweb markdown

## Context

Since #4171, gnoweb renders `> [!INFO]- Title` as a collapsible `<details>`
element. That path always carries alert chrome: a coloured left border, a
semantic icon and a mandatory alert type. Realm authors who want to fold plain
content, such as logs, JSON dumps, changelogs or FAQ answers, have no way to do
so without the alert visuals.

Issue #5578 asks for a neutral collapsible block with no alert styling, no
icon and no type requirement.

## Decision

Add one block extension, `<gno-details>` … `</gno-details>`, each tag alone on
its line, in `gno.land/pkg/gnoweb/markdown/ext_details.go`.

```
<gno-details summary="Logs" open>
arbitrary **markdown**
</gno-details>
```

- **Attributes.** `summary` sets the label and `open` renders the block
  expanded. Any other attribute, a valued `open`, a repeated attribute, an
  attribute on the close tag or a self-closing form makes the line an invalid
  tag, rendered as an HTML comment. A missing `summary` renders as `Details`.
- **Plain-text summary.** The summary is escaped and written inside a
  `<span>`, so no markup in it reaches the page and the summary row's flex
  layout treats it as one item beside the chevron.
- **Flat markers, like `ext_frame`.** Each tag is a one-line leaf. An AST
  transformer moves the blocks between an open marker and its close marker
  under the open one. A line therefore belongs to the innermost block that
  takes it: a close tag inside a code fence, an HTML block or a
  `<gno-foreign>` body stays that block's text.
- **Pairing.** A close tag pairs with the innermost opener under the same
  parent, so nested blocks close inner first. A close tag in another
  container, such as a list item, never closes a block opened outside it and
  renders as a comment. An unclosed block runs to the end of its parent.
- **Any level.** Unlike a frame, a details block may open in a list item, a
  blockquote, an alert or another details block. It never spans a gno-columns
  or gno-frame marker; it ends before one.
- **Depth cap.** An opener takes a level of the shared gno-* cap in
  `nestdepth.go` and is refused past it; its own close tag is then consumed as
  a comment, so it cannot close an accepted block.
- **Sanitizer.** `sanitize.Block` and `sanitize.BlockRich` already escape any
  line opening on `<gno-` or `</gno-`, so user content can neither open a
  block around the realm's own content nor close the realm's block early.
- **Registration.** The extension is registered in `GnoExtension.Extend` and
  in `buildInnerForeignMarkdown`, so a block also renders inside
  `<gno-foreign>`.

The docs realm, `examples/quarantined/gno.land/r/docs/markdown/markdown.gno`,
gains a "Collapsible blocks" section beside the Alerts documentation.

## Alternatives Considered

- **Pandoc-style `:::details` fences.** Rejected after review. A fenced
  container decides whether a line closes it before any block inside it sees
  the line, so a `:::` in a code sample or a foreign body closed the block,
  and the sanitizer passed `:::` lines through untouched.
- **Keep `:::` and teach the sanitizer about it.** Rejected: it changes
  `isExtDelimiter` in the `chain/markdown` stdlib, whose single `<gno-` rule
  covers every gno-* extension, and leaves the line-ownership problem above.
- **A `neutral` alert type.** Rejected: it overloads alert semantics and pulls
  neutral blocks through the alert's icon and summary pipeline.
- **Raw HTML `<details>`.** Production gnoweb strips raw HTML.
- **Inline markdown in the summary.** Left out: an attribute value carries no
  markdown, and a separate summary line would add a second marker.

## Consequences

- Realm authors fold neutral content without alert styling, with the same
  tag grammar as `<gno-columns>`, `<gno-frame>` and `<gno-foreign>`.
- An HTML block runs to the next blank line, so a close tag directly after an
  HTML line is part of that block; a blank line before the tag closes it.
- Styles for `.gno-details` live in
  `gno.land/pkg/gnoweb/frontend/css/06-blocks.css`, scoped under
  `.c-realm-view, .c-readme-view`, and reuse the bundled `#ico-arrow` symbol.
