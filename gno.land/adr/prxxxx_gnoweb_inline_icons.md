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
- **One look, one source.** The chrome already has an icon sprite,
  `components/ui/icons.html`, inlined in every page and referenced as
  `<use href="#ico-…">`. New icons have to look like those, the chrome icons
  have to be callable as well, and no icon may be defined twice.
- **License.** The icons ship in the gno repo and in every page, with no
  credit and no notice to keep. That rules out MIT, ISC, Apache and CC-BY
  sets (Lucide, Tabler, Phosphor, Heroicons, Material, Font Awesome): each
  requires keeping a notice. Only public-domain material, or icons we draw.
- **Light and scalable.** A page pays only for the icons it shows; adding an
  icon costs a line, not page weight; render-time work is a map read and a
  few writes.

## Decision

### Syntax

One self-closing tag, in the style of `<gno-input />` and the other `gno-*`
tags:

```markdown
<gno-icon name="star" />
<gno-icon name="check-circle" label="Verified" />
```

- `name` picks the icon, matched exactly against the table: lowercase, as
  listed, read raw (`name="st&#97;r"` is an unknown name). There is no path,
  prefix or fragment handling.
- `label` is optional (see Accessibility). Its character references and
  backslash escapes are resolved, as goldmark resolves text: `&amp;`, and
  `\|` or `&#124;` for a `|` inside a table cell.
- No other attribute is read. There is no class, style or size: an icon
  takes the size (1.15em, as `.c-icon`) and color (`currentColor`) of the
  text around it, so it works in headings, links and both themes untouched.
- Tag and attribute names are case-insensitive, like every `gno-*` tag. The
  tag is written on one line and is at most 512 bytes.

No shortcode (`:icon-star:`). A second syntax would double the surface to
document, test and escape, collide with emoji shortcodes, and give nothing the
tag does not; the tag also degrades to nothing (safe mode strips unknown HTML)
on a renderer that does not know it.

### Parsing

- An inline parser on `<`, priority 250: before goldmark's autolink (300) and
  raw-HTML (400) parsers. It checks the `<gno-icon` prefix with a byte compare,
  then reads the tag with a small hand-written scanner: attribute names,
  quoted or unquoted values (`label="a > b"` is one value), up to `/>` or `>`.
  No regex, no tokenizer, no allocation: name and label are slices of the
  source, and only a recognized tag allocates its AST node.
- **The scan is bounded** to 512 bytes and to the current line. An earlier
  version handed the rest of the line to `x/net/html`'s tokenizer; with no
  closing `>`, every repeated `<gno-icon ` read to the end of the line, so
  200 KB of them on one line took 15 s to parse. Bounded, the same input
  takes milliseconds (`TestIconParseLinear`, `BenchmarkIconParse/unterminated`).
- A tag that does not end within the bound, or spans lines, is not claimed:
  goldmark shows it as text or strips it as raw HTML. A tag that ends but is
  not self-closing (`<gno-icon name="x">`) is claimed and renders
  `<!-- gno-icon: write it self-closing, <gno-icon name="…" /> -->`; a
  missing or unknown name renders `<!-- gno-icon: unknown name "…" -->`
  (escaped), the same silent-but-visible-in-source style as `<gno-form>`.
- Code spans, fenced and indented code keep the tag literal: goldmark never
  runs inline parsers inside them.
- A line holding only `<gno-icon … />` would be a CommonMark type-7 HTML
  block, which safe mode strips. A block parser at priority 899 (ahead of the
  HTML block parser, 900) opens a paragraph on a line that starts with an
  icon tag, delegating to goldmark's own `NewParagraphParser()`.
- **Heading IDs.** goldmark derives an auto ID from the heading's raw source
  line, which would give `## <gno-icon name="rocket" /> Launch` the ID
  `gno-icon-namerocket-launch`. An AST transformer rebuilds the IDs of a
  document that has an icon in a heading: from the raw line minus the byte
  ranges of the `Icon` nodes goldmark actually parsed, with a fresh ID
  generator walking headings in document order. The ID is `launch`, a later
  `Launch` gets `launch-1`, and a tag shown as text (code span, backslash
  escape) stays in the ID as any text does. Other inline syntax is left as
  goldmark leaves it. The parser flags the context when an icon lands in a
  heading, so documents without one skip the transformer. The TOC reads
  heading text nodes, so it shows `Launch` with no markup.
- Icons are allowed inside `<gno-foreign>`. They are static allowlisted
  glyphs with no link or script surface, so foreign content gains nothing it
  could abuse. Brand marks are excluded everywhere (below).

### Icon set

**System UIcons**: https://github.com/CoreyGinnivan/system-uicons, site
https://www.systemuicons.com, pinned at commit
`4a17b006f9f3a6f549631d408daae322737be366`. Its `LICENSE` is the Unlicense
("This is free and unencumbered software released into the public domain"),
and the README states "Use the icons how you want, for free, and without any
attribution". No notice ships in the repo or the pages.

It was the only general UI set found under a public-domain grant (CC0, 0BSD,
Unlicense): 430 outline icons on a 21px grid, `fill="none"`,
`stroke="currentColor"`, round caps and joins, the same construction as the
chrome icons. **All of it is imported (428 icons)** except two third-party
brand marks, `airplay` (Apple) and `bluetooth` (Bluetooth SIG). The 1px
upstream stroke on a 21px grid is set to 1.3, which matches the chrome's 1.5
on 24px at the same rendered size. Where an upstream name is already a chrome
icon's (`search`, `check`, `link`, `arrow-down`… 14 of them), the chrome icon
keeps the name and the outline one is `NAME-outline`, so no name changes
meaning.

Six names the set lacks (`shield`, `shield-check`, `key`, `layers`, `music`,
`map`) are drawn for gnoweb in the same construction, in
`markdown/icons/drawn.svg`.

The chrome icons from `components/ui/icons.html` are callable too, except
four brand marks (`github`, `twitter`, `discord`, `telegram`): realm content
must not wear another organization's logo. Its two dead duplicate
definitions (a second `ddl` and `warning`, which browsers never used) are
removed. Total: **495 icons**, listed with a picture in `markdown/ICONS.md`.

### Registry: one generated Go table

`markdown/icons_gen.go` is a generated Go map literal, one line per icon,
sorted:

```go
"star":   {iconHeadStroke, `<path d="m7.5 11.5-5 3…" transform="translate(3 3)"/>`},
"search": {`viewBox="0 0 14 14"`, `<path d="…" stroke="currentColor" …/>`},
```

- Nothing is parsed or embedded at run time; a lookup is a map read.
- Each entry stores only what is its own. The 434 outline icons share
  `iconHeadStroke` (`viewBox="0 0 21 21" fill="none" stroke="currentColor"
  stroke-width="1.3" stroke-linecap="round" stroke-linejoin="round"`), which
  the renderer writes; the generator strips those values from each body
  wherever an element would inherit them anyway (it tracks inheritance, so a
  `fill="none"` under a filled group stays). That removes 51 KB of repeats.
  Chrome icons keep their own `viewBox`/`fill` head.
- **The generator is the safety boundary.** It re-serializes every glyph
  through an element and attribute allowlist (`g`, `path`, `circle`,
  `ellipse`, `line`, `polyline`, `polygon`, `rect`; geometry, `transform`,
  fill and stroke attributes). `<title>`, `<defs>`, `id`s, event handlers and
  any `url(#…)` reference are dropped, so the output is plain shape markup
  whatever the source files hold, and repeated inline copies never duplicate
  an `id`. Values are HTML-escaped. A name defined twice, in one source or
  across sources, is an error. `TestIconRegistry` re-checks the shipped table
  against the allowlist.
- Single source of truth: the table is derived from `components/ui/icons.html`
  (the chrome sprite, unchanged in role), `icons/drawn.svg` and
  `icons/vendored.svg`; `TestIconTable` fails CI when it is stale.

### Rendering: inline `<svg>` per use

Each `<gno-icon />` renders its own `<svg class="gno-icon" …>` with the
glyph's shapes inline. A page pays only for the icons it uses, the page sprite
does not grow, there is no asset to fetch and nothing to add to the CSP.

### Accessibility

- Decorative (no label): `aria-hidden="true" focusable="false"`.
- With a label: `role="img" aria-label="…"`. No `<title>`, so no `id` to
  duplicate when the icon repeats.
- An unlabeled icon that is all a link or a heading holds leaves it with no
  accessible name; the renderer adds
  `<!-- gno-icon: alone in a link or heading, add label="…" to name it -->`
  next to it, and the docs say a label is required there. A heading holding
  only an icon has no text, so it gets no TOC entry, label or not.
- No `tabindex`; nothing an icon renders is focusable.
- Forced colors / high contrast: glyphs use `currentColor` only (fill or
  stroke), so they follow the system text color.
- Size in `em` and `vertical-align: -0.2em`, so the icon follows the font
  size and sits on the text in `h1`…`h6`, `li` and `td` (screenshots in the
  PR). Goldens `a11y_*` pin each case.

### Sanitizer (`chain/markdown`, `sanitize.Block` / `BlockRich`)

`sanitize.Block` and `BlockRich` escape any line starting with `<gno-…>` so
user content cannot open a structural block. That guard only looks at line
starts, so before this change an icon rendered mid-line but showed as
literal text at the start of a line. Icons are now rendered at every
position: `isExtDelimiter` exempts the `<gno-icon` opener. An icon is an
inline allowlisted glyph with no structural effect (gnoweb opens a plain
paragraph on such a line), and `Block` already lets through
`data:image/svg+xml` images, so user content gains no new way to look like
realm chrome. `</gno-icon>` and look-alikes (`<gno-iconic>`) stay escaped.
`InlineText` escapes `<`, so icons never render through it. Locked in by
native unit tests and `golden/sanitize/{block,blockrich}-gno-icon-*` and
`inline-gno-icon` fixtures.

The alternative, escaping `<gno-` mid-line too, needs the line-based guard to
learn code spans (a backslash inside a code span is visible) and would turn
every mid-line `<gno-…>` that safe mode used to strip into visible text.

### Tooling

`markdown/icons/icons.txt` lists the imported upstream icons, one per line,
with an optional upstream file name to rename (`unlock lock_open`). `make
icons`:

1. `tools/cmd/iconset` downloads the pinned upstream archive and writes
   `icons/vendored.svg` (the listed icons as `<symbol>`s; a source file, not
   shipped);
2. `go test -run TestIconTable -update-golden-tests` regenerates
   `icons_gen.go`;
3. `go test -run 'TestIconRegistry|TestIconCatalog' -update-golden-tests`
   regenerates `ICONS.md` and `ICONS.svg` (GitHub strips `<gno-icon>` from
   markdown, so the catalog embeds a plain SVG picture of every glyph over its
   name, from the same table).

Adding or removing an icon is one line plus `make icons`; the output is
sorted and gofmt'd, so the diff is one line per icon.

## Measurements

Apple M4 Max, `go test ./gno.land/pkg/gnoweb/markdown -bench BenchmarkIcon`.

| What | Value |
|---|---|
| `gnoweb` binary, before → after (495 icons) | 45,894,994 → 46,109,090 bytes: **+214 KB (+0.47%)** |
| Table data (bodies + non-shared heads) | 162 KB; the shared head saves 51 KB |
| Rendered icon, all 495 | mean **498 B**, median 447 B, min 136 B, max 2,300 B (chrome) |
| Render one icon (`star`, decorative) | 40 ns/op, 275 B, **0 allocs** |
| Render one icon with a label | 58 ns/op, 296 B, **0 allocs** |
| Render a chrome icon (`search`) | 48 ns/op, 611 B, 0 allocs |
| Scan one tag (`parseIconTag`) | 0 allocs (`TestParseIconTagAllocs`) |
| Parse a line of 400 icons | 133 µs, ~2 allocs per icon (the node and goldmark's text segment) |
| Parse a line of 1,000 non-icon tags | unchanged from the base pipeline (415 KB, 5,020 allocs) |
| 20,000 unterminated `<gno-icon ` on one line | 15.5 s before the bound; the whole `TestIconParseLinear` now runs in < 0.2 s |

## Alternatives considered

- **A sprite file in `/public` (`<use href="/public/icons.svg?v=…#star">`).**
  Same origin, so CSP and `<use>` would work, but the markdown renderer knows
  neither the assets path nor the assets version; both would have to be
  threaded through `RenderConfig`, the extension and the `<gno-foreign>` inner
  instance. Every page using one icon would also fetch the whole set.
- **Growing `components/ui/icons.html`.** Zero renderer work, since the page
  already carries the sprite, but every HTML page would carry the whole set
  (≈200 KB of markup) whether it shows an icon or not.
- **Emitting only the used `<symbol>`s into the page sprite.** Smallest
  output, but needs renderer→layout plumbing to collect names per render.
  Inline per use is within a few hundred bytes of it for realistic pages.
- **Embedding the SVG sources and parsing them at init.** Simpler to wire,
  but repeats every shared attribute 434 times in the binary and parses HTML
  at every start; the generated table does neither.
- **A selection of the upstream set.** Every icon costs only binary size
  (≈330 bytes on average) and none costs a page anything, so the whole set is
  imported; the cost of the full import is the measurement above.
- **Lucide / Tabler / Phosphor / Heroicons.** Better coverage, but MIT/ISC
  require keeping a notice, which the project does not want to ship.
- **Shortcode `:icon-star:`.** See Syntax.

## Consequences

- Authors get 495 icons with one obvious syntax; mistakes fail visibly in the
  page source and invisibly on the page.
- `gnoweb` grows by 214 KB; pages grow by ≈500 bytes per icon used and not at
  all otherwise; the served page sprite is unchanged.
- Parse cost: a `<` that is not an icon pays a 9-byte compare and allocates
  nothing; an icon allocates its node only; an unterminated tag costs at most
  a 512-byte scan. Render cost: a map read and a few writes, no allocation.
- `chain/markdown`'s `EscapeBlockHazards` changes behavior for one input: a
  line starting with `<gno-icon` no longer gets a leading backslash. Realm
  output built with `sanitize.Block`/`BlockRich` before and after this change
  differs only for such lines.
- `.gno-alert summary svg` sizing now skips `.gno-icon`, so an icon in an
  alert title keeps its inline size.
- Chrome icon names become a public API for realm content: renaming or
  removing one in `icons.html` breaks the realms that use it. A chrome
  symbol added later (including a brand mark, which would need adding to the
  exclusion list) becomes callable once `make icons` runs; it shows up in the
  `icons_gen.go` and `ICONS.md` diffs, which CI forces to be regenerated.
- The renaming of upstream icons and the drawn icons are ours to maintain.
  Upstream has not changed since 2023; bumping the pin is a one-line change
  followed by `make icons`.
