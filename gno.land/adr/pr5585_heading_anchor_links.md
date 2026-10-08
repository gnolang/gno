# ADR: Heading Anchor Links in gnoweb

## Context

Issue [#5579](https://github.com/gnolang/gno/issues/5579): rendered realm and readme headings get auto-generated IDs from `parser.WithAutoHeadingID()`, but nothing on the page links to them. There is no way to copy a link to a section without going through the ToC sidebar.

## Decision

Replace goldmark's heading renderer in `GnoExtension` with one that closes every heading that has an id and some content with a permalink anchor, using the icon and label of the `ui/pkg_anchor` template on package overview cards:

```html
<h2 id="title">Title<a href="#title" class="heading-anchor" aria-label="Permalink"><svg class="c-icon" aria-hidden="true"><use href="#ico-link"></use></svg></a></h2>
```

The icon is hidden at rest and shown when the heading is hovered or the anchor is focused. On devices with no hover (`@media (hover: none)`) it is always shown.

## Alternatives Considered

1. **Wrap the heading text in `<a href="#id">`**: the first version of this PR. A heading containing a link needs one anchor per text run around it, which repeats the full `href` per run and grows the output quadratically with the number of links. It makes the heading text unselectable, splits one heading into several tab stops named after fragments, and leaves link-only headings with no permalink. Footnote refs, task checkboxes and raw HTML can still end up inside the anchor.
2. **Anchor as a sibling of the heading, in a wrapper element** (GitHub's markup): keeps the anchor's `aria-label` out of the heading's accessible name, at the cost of a wrapper around every heading, which breaks the adjacent-heading rules (`h1 + h2`) in the realm view CSS.
3. **goldmark-anchor external extension**: an external dependency for what is one renderer.
4. **JavaScript-only approach**: no link to copy without extra UI.

## Consequences

- Link-only headings get a permalink too. Each heading grows by one anchor repeating its id, so the output stays linear in the heading's length.
- The heading text is left untouched: it stays selectable, and inline links, footnote refs and raw HTML inside it render as before.
- The heading's accessible name ends with "Permalink", since the anchor sits inside the heading.
- The doc view does not load this extension, so godoc headings get no anchor.
- The golden test setup gained `parser.WithAutoHeadingID()` to match production; existing fixtures gained `id` attributes and the anchor markup.
- No new external dependencies.
