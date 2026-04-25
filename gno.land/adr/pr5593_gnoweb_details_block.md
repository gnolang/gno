# ADR: Neutral `:::details` collapsible block in gnoweb markdown

## Context

Since #4171, gnoweb renders `> [!INFO]- Title` blockquote-style syntax as a
collapsible `<details>` element. That path always carries alert chrome — a
coloured left border, a semantic icon, and a mandatory alert type (INFO,
NOTE, TIP, SUCCESS, WARNING, CAUTION). Realm authors who want to fold plain
content (logs, JSON dumps, changelogs, FAQ answers) have no way to do so
without the alert visuals.

Issue #5578 asks for a neutral collapsible block with no alert styling, no
icon, and no type requirement.

## Decision

Add a new Goldmark extension `ExtDetails` to
`gno.land/pkg/gnoweb/markdown/ext_details.go` that recognises a
pandoc-style fenced container block:

```
:::details Summary text
arbitrary **markdown**
:::
```

- Opening fence: a line starting with `:::details`. An optional `[open]`
  flag directly after `details` (e.g. `:::details[open]`) makes the block
  render with the HTML `open` attribute. Anything after a single space is
  treated as the summary, parsed as inline markdown.
- Closing fence: a line containing exactly `:::`. If the fence is missing,
  the block closes at the end of the document (matching CommonMark fenced
  code behaviour).
- Rendered HTML:
  `<details class="gno-details" [open]><summary><svg><use href="#ico-arrow"></use></svg>…inline
  summary…</summary><div>…block content…</div></details>`. The chevron SVG
  references the bundled `#ico-arrow` symbol; the body is wrapped in a
  `<div>` to anchor padding without affecting margin collapse on the first
  child paragraph.
- When the opening fence has no summary, the renderer emits a default
  `<summary><svg/>Details</summary>` so the chevron and body wrapper still
  apply consistently.

The extension is wired into `GnoExtension.Extend` next to `ExtAlerts`. It
uses the same block-parser priority (799) as the alert parser.

The docs realm (`examples/gno.land/r/docs/markdown/markdown.gno`) gains a
new "Collapsible blocks" section demonstrating the syntax alongside the
existing Alerts documentation.

## Alternatives Considered

- **Extend the existing alert syntax with a `neutral` / `plain` type.**
  Rejected: overloads alert semantics, pulls neutral blocks through the
  alert renderer's icon/summary pipeline, and forces CSS branching. The
  goal is *no* alert chrome, not a new alert variant.
- **Raw HTML `<details>`/`<summary>`.** Already possible but verbose and
  does not play well with markdown parsing inside the summary. Realm
  authors asked for a markdown-native form.
- **GitHub-style `<details>`/`<summary>` inside a blockquote.** Same
  verbosity problem, and still parses as HTML rather than markdown.
- **Reuse the `|||`/`<gno-columns>` custom-tag approach from the columns
  extension.** Pandoc-style `:::` fences are more conventional for fenced
  divs and easier to type; the columns extension is tag-oriented because
  it tracks multiple separator nodes, which does not apply here.

## Consequences

- **Positive:** Realm authors can now fold neutral content without opting
  in to alert styling. Markdown stays readable; output is a single, well
  understood HTML element.
- **Positive:** Implementation is isolated to a single file plus a one-
  line registration in `ext.go`; alert behaviour and existing markdown
  fixtures are untouched.
- **Trade-off:** `:::` fences introduce a new namespace in gnoweb
  markdown. Future extensions (`:::columns`, `:::warning`, …) should pick
  a consistent grammar — the parser rejects `:::something` that does not
  match `:::details`, falling through to default paragraph rendering, so
  adding new `:::<name>` blocks later is non-breaking.
- **Styling shipped.** Project styles for `.gno-details` are added to
  `gno.land/pkg/gnoweb/frontend/css/06-blocks.css`, scoped under
  `.c-realm-view, .c-readme-view`: bordered, rounded, surface-coloured
  card with a chevron summary that rotates between `-90deg` (closed) and
  `0` (open). The shared `#ico-arrow` symbol is reused so no new SVG asset
  is bundled.
