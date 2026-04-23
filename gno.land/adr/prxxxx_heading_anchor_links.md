# ADR: Heading Anchor Links in gnoweb

## Status

Accepted

## Context

Issue [#5579](https://github.com/gnolang/gno/issues/5579): In rendered realm/readme markdown, headings get auto-generated IDs via `parser.WithAutoHeadingID()` but the heading text itself is not a link. Clicking a `<h1>`/`<h2>`/etc. does nothing — the URL hash doesn't update, and there's no way to copy a stable link to a section without going through the ToC sidebar.

## Decision

Add a custom `headingRenderer` in the GnoExtension that overrides goldmark's default heading renderer. On the closing pass (after child content is rendered), append an empty `<a class="heading-anchor" href="#id" aria-hidden="true"></a>` element. CSS shows a `§` symbol on hover via `::after`, providing a clickable self-link.

This approach avoids nesting `<a>` tags (invalid HTML), which would occur if we wrapped the heading content in an anchor — headings can contain links from the GnoLink extension.

The `aria-hidden="true"` attribute prevents screen readers from announcing the anchor, while the link remains clickable for sighted users.

## Alternatives Considered

1. **Wrap heading content in `<a href="#id">`**: Would nest `<a>` tags when headings contain links — invalid HTML per spec.
2. **goldmark-anchor extension**: External dependency; our custom renderer is minimal (same pattern as existing GnoExtension renderers) and keeps the dependency tree unchanged.
3. **JavaScript-only approach**: Could update `window.location.hash` on click, but doesn't provide a shareable link for copy-paste without additional URL manipulation.
4. **Always-visible anchor icon**: Would add visual noise; hover-to-show is the established pattern (GitHub, MDN, Rust docs).

## Consequences

- Headings with auto-generated IDs now have a clickable anchor link that updates the URL hash.
- The golden test suite needed `parser.WithAutoHeadingID()` added to match production config, updating existing golden outputs to include `id` attributes.
- No new external dependencies.
