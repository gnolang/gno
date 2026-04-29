# ADR: Heading Anchor Links in gnoweb

## Context

Issue [#5579](https://github.com/gnolang/gno/issues/5579): Rendered realm/readme headings get auto-generated IDs via `parser.WithAutoHeadingID()` but are not themselves clickable. The URL hash does not update on click, and there is no way to copy a stable link to a section without going through the ToC sidebar.

## Decision

Add an AST transformer + inline node + node renderer to `GnoExtension`:

- The transformer walks every heading and groups its children. Each contiguous run of non-link inline children is moved under a synthetic `headingAnchorNode` carrying the heading's id.
- The renderer emits `<a class="heading-anchor" href="#id">…</a>` for each `headingAnchorNode`. Inline links inside the heading are left untouched and rendered by their existing renderers.
- The default goldmark heading renderer continues to emit `<h2 id="…">…</h2>`; no override is needed.

Result: clicking on any non-link text in a heading updates `window.location.hash`, while clicks on inline links still navigate to their destination. Nested `<a>` is impossible by construction — the transformer never wraps a link.

No `aria-label` is set on the heading-anchor. Its accessible name falls back to the wrapped text, so screen-reader heading navigation keeps announcing the actual title.

## Alternatives Considered

1. **Always wrap heading text in a single `<a href="#id">`**: simpler, but breaks on headings containing inline links — nested `<a>` is invalid HTML and browsers auto-close the outer anchor.
2. **Sibling-only empty anchor for every heading**: accessibility-safe but requires hovering to reveal a `§` / `#` indicator; less discoverable than click-anywhere-on-text.
3. **Sibling empty anchor only when the heading contains a link**: keeps wrap mode for plain headings but hides the permalink from mouse users in the link case (no visible affordance, hover-`#` glyph felt out of place).
4. **goldmark-anchor external extension**: external dependency; our custom transformer + renderer is minimal and keeps the dependency tree unchanged.
5. **JavaScript-only approach**: could update `window.location.hash` on click, but requires extra UI to expose a copy-pasteable link.
6. **`aria-label="Link to this section"` on every anchor**: overrides the anchor's accessible name and, by extension, the heading's accessible name — screen-reader heading navigation would announce every heading as "Link to this section" instead of the title. Rejected.

## Consequences

- Headings with auto-generated IDs are clickable: the whole heading text in the plain case, and every non-link span in the inline-link case. The inline link inside a heading retains its own destination on click.
- A heading whose entire content is a single link (e.g. `## [foo](/x)`) gets no permalink anchor — the inline link wins the click. Users still reach the section via the address bar / ToC sidebar / Tab key on the heading id. This is judged acceptable given the rarity of the pattern.
- The golden test suite gained `parser.WithAutoHeadingID()` in the test setup to match production; existing fixtures gained `id` attributes and the anchor markup.
- The extension requires `parser.WithAutoHeadingID()` to emit anchors. Without it, headings render plain — the transformer is a no-op when no id is present.
- No new external dependencies.
