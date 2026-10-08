# ADR: inline icons in gnoweb markdown with `<gno-icon />`

## Status

Proposed.

## Context

Realm authors and static pages want a small icon next to text: a status
mark in a table, a glyph in front of a heading, a hint inside a link label.
Gnoweb markdown had no way to do it. Raw `<svg>` is stripped by safe mode,
an `![](data:image/svg+xml…)` image does not take the text color and needs
alt text, and emoji render differently on every platform.

Constraints that shaped the design:

- **Safety.** Realm markdown is attacker-controlled, so an icon cannot carry
  author markup. Whatever renders has to come from an allowlist we own.
- **CSP.** Gnoweb serves `default-src 'self'; img-src 'self' data: …` and no
  inline styles, so nothing may be fetched from a CDN at render time.
- **One look.** The chrome already has an icon sprite,
  `components/ui/icons.html` (67 `<symbol>`s, inlined in every page,
  referenced as `<use href="#ico-…">`). New icons have to look like those
  and the chrome icons have to be callable as well.
- **License.** The icons ship in the gno repo and in every page, with no
  credit and no notice to keep. That rules out MIT, ISC, Apache and CC-BY
  sets (Lucide, Tabler, Phosphor, Heroicons, Material, Font Awesome): each
  requires keeping a notice. Only public-domain material, or icons we draw.
- **Page weight.** A page should pay only for the icons it shows.

## Decision

### Syntax

One self-closing tag, in the style of `<gno-input />` and the other `gno-*`
tags:

```markdown
<gno-icon name="star" />
<gno-icon name="check-circle" label="Verified" />
```

- `name` picks the icon. It is matched exactly (after trimming) against the
  registry; there is no path, prefix or fragment handling.
- `label` is optional. With it the `<svg>` gets `role="img"` and an escaped
  `aria-label`; without it, `aria-hidden="true"`.
- No other attribute is read. There is no class, style or size: an icon
  takes the size (1.15em, as `.c-icon`) and color (`currentColor`) of the
  text around it, so it works in headings, links and both themes untouched.
- Tag and attribute names are case-insensitive, like every `gno-*` tag
  (`x/net/html` lowercases them).

No shortcode (`:icon-star:`). A second syntax would double the surface to
document, test and escape, collide with emoji shortcodes, and give nothing the
tag does not; the tag also degrades to nothing (safe mode strips unknown HTML)
on a renderer that does not know it.

### Parsing

- An inline parser on `<`, priority 250: before goldmark's autolink (300) and
  raw-HTML (400) parsers. It checks the `<gno-icon` prefix with a byte compare
  and only then tokenizes the tag with `x/net/html`, which handles quoting
  (`label="a > b"`). No regex.
- Only a **well-formed self-closing** tag is claimed. `<gno-icon name="x">`
  or `</gno-icon>` fall through to the raw-HTML parser and are stripped by
  safe mode, as any unknown HTML is. A self-closing tag with a missing or
  unknown name is claimed and renders `<!-- gno-icon: unknown name "…" -->`
  (escaped), the same silent-but-visible-in-source error style as
  `<gno-form>`.
- Code spans, fenced and indented code keep the tag literal: goldmark never
  runs inline parsers inside them.
- A line holding only `<gno-icon … />` would be a CommonMark type-7 HTML
  block, which safe mode strips. A block parser at priority 899 (ahead of the
  HTML block parser, 900) opens a paragraph on a line that starts with a
  valid icon tag, delegating to goldmark's own `NewParagraphParser()`.
- Auto heading IDs are derived by goldmark from the raw source line, which
  would give `## <gno-icon name="rocket" /> Launch` the ID
  `gno-icon-namerocket-launch`. `NewGnoParserContext` installs an `IDs`
  wrapper that strips icon tags before goldmark's generator runs, so the ID is
  `launch`, numbered in document order with any other `Launch`. The TOC reads
  heading text nodes, so it already shows `Launch` with no markup.
- Icons are allowed inside `<gno-foreign>`. They are static allowlisted
  glyphs with no link or script surface, so foreign content gains nothing it
  could abuse. Brand marks are excluded from the registry everywhere (below).

### Icon set

**System UIcons** (github.com/CoreyGinnivan/system-uicons), pinned at commit
`4a17b006f9f3a6f549631d408daae322737be366`. Its `LICENSE` is the Unlicense
("This is free and unencumbered software released into the public domain"),
and the README states "Use the icons how you want, for free, and without any
attribution". No notice ships in the repo or the pages.

It was the only general UI set found under a public-domain grant (CC0, 0BSD,
Unlicense): 432 outline icons on a 21px grid, `fill="none"`,
`stroke="currentColor"`, round caps and joins — the same construction as the
chrome icons. 178 are imported. Their 1px stroke on a 21px grid is set to 1.3,
which matches the chrome's 1.5 on 24px at the same rendered size.

Six names the set lacks (`shield`, `shield-check`, `key`, `layers`, `music`,
`map`) are drawn for gnoweb in the same construction, in
`markdown/icons/drawn.svg`.

The chrome icons from `components/ui/icons.html` are callable too, except the
four third-party brand marks (`github`, `twitter`, `discord`, `telegram`):
realm content must not wear another organization's logo. Total: 245 icons,
listed in `markdown/ICONS.md`.

### Rendering: inline `<svg>` per use

Each `<gno-icon />` renders its own `<svg class="gno-icon" viewBox=…>` with
the glyph's shapes inline. A page pays only for the icons it uses (about
450 bytes each before compression, much less after since repeats compress).

The glyphs come from one registry, built once at init from three symbol
files read in order: the chrome sprite `components/ui/icons.html`,
`icons/drawn.svg`, `icons/vendored.svg`. So `icons.html` stays the single
source for the chrome icons; nothing is copied. Every glyph is
**re-serialized through an element and attribute allowlist** (`g`, `path`,
`circle`, `ellipse`, `line`, `polyline`, `polygon`, `rect`; geometry,
`transform`, fill and stroke attributes). `<title>`, `<defs>`, `id`s, event
handlers and any `url(#…)` reference are dropped, so the output is plain
shape markup whatever the source files hold, and repeated inline copies never
duplicate an `id`. Values are HTML-escaped. A name defined in an earlier file
wins; a test fails if a later file shadows one.

### Tooling

`markdown/icons/icons.txt` lists the imported names, one per line, with an
optional upstream file name to rename (`unlock lock_open`). `make icons` runs
`tools/cmd/iconset`, which downloads the pinned upstream archive and writes
`icons/vendored.svg`, then runs the registry tests: one fails on a name an
earlier file already defines (`TestIconRegistryNoShadowing`), another
regenerates `ICONS.md` (`TestIconCatalog`, which also fails CI when the
catalog is stale). Adding an icon is one line plus `make icons`.

## Alternatives considered

- **A sprite file in `/public` (`<use href="/public/icons.svg?v=…#star">`).**
  Same origin, so CSP and `<use>` would work, but the markdown renderer knows
  neither the assets path nor the assets version; both would have to be
  threaded through `RenderConfig`, the extension and the `<gno-foreign>` inner
  instance. Every page using one icon would also fetch the whole set.
- **Growing `components/ui/icons.html`.** Zero renderer work, since the page
  already carries the sprite, but every HTML page would carry ~75 KB more
  markup (≈15 KB gzipped) whether it shows an icon or not.
- **Emitting only the used `<symbol>`s into the page sprite.** Smallest
  output, but needs renderer→layout plumbing to collect names per render.
  Inline per use is within a few hundred bytes of it for realistic pages.
- **Lucide / Tabler / Phosphor / Heroicons.** Better coverage, but MIT/ISC
  require keeping a notice, which the project does not want to ship.
- **A Go file generated from the SVGs** (`go generate`). Another artifact
  that can drift from its source; building the registry at init from the
  embedded files costs about a millisecond once and cannot drift.
- **Shortcode `:icon-star:`.** See Syntax.

## Consequences

- Authors get 245 icons with one obvious syntax; unknown names fail visibly
  in the page source and invisibly on the page.
- Binary size grows by the vendored file (≈88 KB of SVG source, embedded).
  The served page sprite is unchanged.
- Rendered size grows by ≈450 bytes per icon used, nothing otherwise.
- Parse cost: lines without `<gno-icon` pay a 9-byte compare per `<`. Each
  icon costs one `x/net/html` tokenizer (≈4 KB, a few allocations;
  `BenchmarkIconParse`).
- `NewGnoParserContext` now installs a heading-ID generator. Contexts built
  elsewhere (plain `parser.NewContext()`) keep goldmark's raw-line IDs.
- `.gno-alert summary svg` sizing now skips `.gno-icon`, so an icon in an
  alert title keeps its inline size.
- Chrome icon names become a public API for realm content: renaming or
  removing one in `icons.html` breaks the realms that use it. A chrome
  symbol added later (including a brand mark, which would need adding to the
  exclusion list) becomes callable; it shows up in the `ICONS.md` diff,
  which CI forces to be regenerated, so the review sees it.
- The renaming of upstream icons and the drawn icons are ours to maintain.
  Upstream has not changed since 2023; bumping the pin is a one-line change
  followed by `make icons`.
