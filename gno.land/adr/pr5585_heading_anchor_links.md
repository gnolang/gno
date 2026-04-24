# ADR: Heading Anchor Links in gnoweb

## Status

Accepted

## Context

Issue [#5579](https://github.com/gnolang/gno/issues/5579): Rendered realm/readme headings get auto-generated IDs via `parser.WithAutoHeadingID()` but are not themselves clickable. The URL hash does not update on click, and there is no way to copy a stable link to a section without going through the ToC sidebar.

## Decision

Add a custom `headingRenderer` in the `GnoExtension` that overrides goldmark's default heading renderer. The renderer picks one of two modes per heading:

1. **Wrap mode** (default): wrap the heading text in `<a class="heading-anchor" href="#id">…</a>`. Clicking anywhere on the heading text sets `window.location.hash`. No `aria-label` is set — the anchor's accessible name falls back to the wrapped heading text so screen-reader heading navigation keeps announcing the actual title.

2. **Sibling mode** (fallback): when the heading AST contains a `Link` or `AutoLink` descendant, wrapping would emit nested `<a>` tags (invalid per the HTML spec). In that case the renderer emits `<a class="heading-anchor" href="#id" aria-hidden="true"></a>` as a sibling after the heading text. The sibling anchor has no visible indicator — it is reachable by keyboard (Tab) and exposed to the DOM for programmatic access, but mouse users in this case cannot trigger the hash via click. The inline link inside the heading remains fully functional.

The renderer also balances its own tags: if the heading has no usable `id` (e.g. the extension is used without `parser.WithAutoHeadingID()`), no `<a>` is emitted on entry *or* exit — no stray `</a>` is produced.

## Alternatives Considered

1. **Always wrap heading text in `<a href="#id">`**: simpler, but breaks on headings containing inline links — nested `<a>` is invalid HTML and browsers auto-close the outer anchor.
2. **Sibling-only empty anchor for every heading**: accessibility-safe but requires hovering to reveal the `§` indicator; less discoverable than click-anywhere-on-text.
3. **goldmark-anchor external extension**: external dependency; our custom renderer is minimal (same pattern as existing `GnoExtension` renderers) and keeps the dependency tree unchanged.
4. **JavaScript-only approach**: could update `window.location.hash` on click, but requires extra UI to expose a copy-pasteable link.
5. **`aria-label="Link to this section"` on the wrapping anchor**: overrides the anchor's accessible name and, by extension, the heading's accessible name — screen-reader heading navigation would announce every heading as "Link to this section" instead of the title. Rejected.

## Consequences

- Headings with auto-generated IDs are clickable: either the whole heading text (common case), or the hover-§ indicator (when the heading contains an inline link).
- The golden test suite gained `parser.WithAutoHeadingID()` in the test setup to match production; existing fixtures gained `id` attributes and the anchor markup.
- The extension requires `parser.WithAutoHeadingID()` to emit anchors. Without it, headings render plain — no stray `</a>`.
- No new external dependencies.
