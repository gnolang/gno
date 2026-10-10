# ADR: gnoweb math extension (TeX to native MathML, server-side)

## Context

Realms render markdown, and gnoweb turns it into HTML with goldmark and the
extensions in `gno.land/pkg/gnoweb/markdown`. Realms that publish anything
mathematical (DeFi formulas, governance quorum rules, documentation) had no
way to show a formula: TeX source stays as text, or goes in a code block.

Two constraints from gnoweb shape the answer. Pages are rendered on the
server and must work without JavaScript (the State Explorer ADR,
`adr-003-state-explorer.md`, states the same goal: shareable URLs, pages that
print and that browsers without JS still render). And the markdown is
written by whoever deployed the realm, so every byte of it is attacker
input: what it turns into must be safe HTML, and the work it causes must be
bounded by its size.

Browsers now render MathML natively: Chrome and Edge ship MathML Core since
version 109, and Firefox and Safari have long supported MathML. A formula
converted to MathML on the server therefore needs no script, no image and
no font download on the client.

## Decision

### A goldmark extension that emits MathML

`markdown/ext_math.go` adds an inline parser, a block parser, an AST
transformer and a renderer, registered by `NewGnoExtension` through
`NewExtMath`. The renderer converts each expression to MathML on the server
and writes it in place, on one line, inside a `<semantics>` element that
keeps the TeX source as an `<annotation encoding="application/x-tex">`.

Delimiters, as written in the markdown source:

- `$…$` inline. Pandoc's rules decide what is math: the opening `$` must be
  followed by a non-space, the closing `$` must not follow a space or an
  escaping backslash nor be followed by a digit, so `$5 and $10` and `$ 5`
  stay text, and `\$` is a literal dollar. Links are kept whole: a `$`
  inside a link destination or an autolink does not close (gnoweb's own
  realm function links read `/r/x$help&func=F`), and an expression that
  opens inside a link label ends with it.
- `$$…$$` display.
- `\\(…\\)` inline and `\\[…\\]` display. The backslash is doubled because
  `\(` is a CommonMark backslash escape: a single `\(x\)` renders as `(x)`.
- A fenced code block whose info string is exactly `math`, as on GitHub, is
  display math: the transformer replaces the fence with a math block held to
  the same limits and budget as `$$`, and keeps the fence as its child, so a
  fence that is not converted renders as the code block it would be without
  math.

A display delimiter written inside a line of text is inline-level: the
expression is rendered as `<math display="block">` inside the paragraph. A
display delimiter that opens a line, with its closing delimiter on a later
line, opens a math block. Like a code fence, the opener may be indented by
up to three spaces.

### Parsing rules

- A math block is read like a paragraph. It ends at a blank line (TeX
  forbids paragraph breaks in math mode) or at any line that would interrupt
  a paragraph: thematic break, ATX or setext heading, code fence, list item,
  HTML block, blockquote, and gnoweb's own columns, forms and alerts. The
  check offers the line to the block parsers themselves in a scratch context
  (`startsBlock`), so their interruption rules apply as written: goldmark's
  CommonMark parsers plus the block parsers the other gnoweb extensions
  register, which `NewGnoExtension` records and hands to `NewExtMath`.
  goldmark cannot list the parsers an extension loaded before this one
  registered, so those are declared with `WithPeerBlockParsers`: gnoweb's
  render config passes `extension.Footnote`'s, so a footnote definition ends
  display math too. The parsers are indexed once by trigger byte, so a line
  is offered only to the parsers its first byte can open, and the scratch
  context drops every value a probed parser sets. An unclosed `$$` cannot
  swallow the lists, quotes, headings or gnoweb blocks after it when a later
  line holds `$$`.
- A block opens only when a closing line follows within `MaxMathInputLen`
  bytes, before such a line, and inside the containers (blockquote, list
  item) the opener is in: the lookahead offers each line to those
  containers' own `Continue`, and a math block has no lazy continuation
  lines. Otherwise the opener stays paragraph text, read exactly as without
  math.
- A closing line ends with the delimiter, which appears on it only once:
  the delimiter alone, or after the last line of math. Like a closing code
  fence, it may be followed only by spaces, so `$$y$$ trailing` does not
  close a block and is not split into a paragraph with a stray `$$`.
- Dollars around a reference to a footnote defined in the document are
  prices, not math: an inline expression or a `$$` block that holds one is
  left as the text (or paragraph) it would be without math, so the reference
  and its footnote stay on the page.
- An inline math node holds its source, delimiters included, as an
  `ast.String` child that the renderer skips: text extractors (the table of
  contents, an image's alt text) read the expression as written instead of
  dropping it.
- An inline expression is read from the one or two line segments the parser
  already holds, not from `block.Value`, which walks the paragraph's lines
  and made a paragraph of one expression per line quadratic.
- An empty expression (`$$$$`, `\\(\\)`, a blank block) is text: it holds no
  math and would cost a whole `<math>` element for a few input bytes.
- Code spans, indented code and fences other than ` ```math ` keep their `$`
  literally. A code span on the same line as the closing `$` of an inline
  `$…$` keeps its `$` too.

### The converter: a port of TreeBlood

`markdown/mathml` is a port of TreeBlood (https://github.com/Wyatt915/treeblood),
a pure-Go TeX to MathML converter by Wyatt Sheffield, under the MIT licence.
The licence and the attribution are kept in `markdown/mathml/LICENCE.MD`.
The port is vendored in the tree, and gnoweb's changes (escaping, bounds,
output) are applied to it, so they are reviewed with gnoweb. Among them:
a `{group}` that crosses an environment or a `\left...\right` pair is
rejected when tokenizing, as TeX rejects it ("Extra }"), since reading it
as one unit ran past the end of the group around it (a hang, or work
doubling per group); operator names (`\sin`, `\lim`) are `<mo>` elements
spaced from their neighbours as TeX spaces them, because MathML Core
honours `lspace`, `rspace` and `movablelimits` on `<mo>` only; and nested
font commands replace one another, as in TeX. Other TeX rules the port
follows: an empty group or argument (`{}`, `\frac{}{b}`) is an empty
`<mrow>`, so a script or a fixed-arity element keeps all its parts;
`\over`, `\atop` and `\choose` take the whole group (or table cell) on each
side, and a second one in the same group is an error, as in TeX; a `\\`
outside an environment splits the expression into the rows of a
one-column table; spaces between `\left` (or `\big`) and its delimiter
are skipped; a `%` comment changes nothing after its line. An unknown
command is shown as written, backslash included, in an `<merror>`.
Supported beyond TreeBlood: `\operatorname` (and `*`), `\phantom`,
`\hphantom`, `\vphantom`, `\kern`, `\mkern`, `\hspace`, and `\tag`, shown
after the formula as written since there is no equation numbering;
`\label` is read and dropped, and `\ref` and `\eqref` are not supported.
The converter is created per expression
(`NewMathMLConverter` in `renderMath`), as it keeps per-expression state and
must not be shared between concurrent renders.

### Security model

- Everything is escaped. Element text and attribute values go through
  `writeEscaped`, which escapes `<`, `>`, `"` and every `&`, so a typed entity
  (`\text{&lt;b&gt;}`) displays as typed and cannot decode into markup. The
  fallback for an expression that is not converted is the TeX source,
  escaped with `html.EscapeString`, in `<span class="math-inline">` or
  `<span|div class="math-display">` (a `<span>` when the expression sits
  inside a paragraph, where a `<div>` would be invalid; the stylesheet shows
  `.math-display` as a block).
- Attribute names are never taken from the input: the converter sets them
  from its own code, and `isAttrName` drops any name that is not a plain XML
  name, as defence in depth. Values derived from input are escaped, and the
  ones that size or move content are parsed and bounded: `\multirow` and
  `\multicolumn` spans must be integers in 1..64 (`maxCellSpan`), and a
  `\raisebox` shift must be a length in a TeX unit or px (a bare number is
  taken in em) that keeps the sum of the enclosing shifts within 2em of
  the text around the formula (`maxRaisePt`), counting the size switches
  in between (`\Huge\raisebox{1em}` moves 2.49em); anything else is
  dropped and the content rendered unshifted or unspanned. A `\kern`,
  `\mkern` or `\hspace` must lie within [-2em, 20em] (`minKernPt`,
  `maxKernPt`), so a negative space cannot pull the math over the text
  before it. Size switches (`\tiny` … `\Huge`) multiply when
  nested, so their cumulative scale is clamped to [0.5, 2.488] (`minSize`,
  `maxSize`). An unbounded span would set the minimum size of a stretched
  arrow thousands of em tall; an unbounded shift or size would draw math
  over the page around it.
- `\class{name}{x}` renders `x` and drops the class: page authors must not
  apply the site's CSS classes (to overlay the page, for instance). The only
  classes in the output are the converter's own (`math-displaystyle`,
  `math-textstyle`, `mathcal`, `mathscr` and the `math-*` styling classes
  below).
- No inline style. gnoweb's production CSP is `style-src 'self'`
  (`SecureHeadersMiddleware` in `gno.land/cmd/gnoweb/main.go`), which drops
  every `style` attribute, so what the converter used to write as styles is
  a class the stylesheet styles: `math-dtls-on` (dotless letters under
  accents), `math-latex-*` and `math-tex-*` (logo kerning), `math-liminf`
  and `math-limsup`, and `math-aligned` (the column pairs of `aligned`).
- `\color` and `\textcolor` accept only the theme colours red, orange,
  green, blue, purple and gray (or grey), as a `math-color-<name>` class; any
  other name or value, or a value in a colour model (`\color[RGB]{…}`),
  leaves the content in the text colour. An arbitrary
  `mathcolor` would allow text in the background colour and ignore the dark
  theme.
- `\newcommand`, `\renewcommand` and `\def` are accepted (with their `[n]`
  argument count and default) but their definitions are not expanded, so
  there is no macro expansion to bound.
- `FuzzMathRender` checks that no input produces `<script>`, `<style>`,
  `<iframe>`, `<object>`, `<embed>`, `<use>` or `<svg>` (except gnoweb's own
  alert icon, an `<svg>` holding only `<use href="#ico-…">`), an `on*`
  attribute or a `javascript:` value, and that the output stays within the
  size bounds below. Each render runs under a timeout, and
  `FuzzConversionTerminates` checks the converter alone the same way: a
  `\sideset` loop had gone unnoticed because fuzzing just sat on the input.

### DoS model

- Per expression, input: an expression longer than `MaxMathInputLen` (8 KiB)
  is not converted and is rendered as escaped text.
- Per expression, depth: `ParseTex` panics past `MaxParseDepth` (64) nested
  levels, and the converter recovers it as a conversion error
  (`errMaxDepth`). A style switch (`\bf`, `\color`, `\large`) applies to the
  rest of its group, so a run of switches nests one level per switch: the
  run counts once against `MaxParseDepth`, and `maxSwitchDepth` (256)
  bounds it, and with it the number of times switches scan the rest of
  their cell. The MathML is written on one line, so nesting
  adds no indentation to the output.
- Per expression, output: MathML longer than `64·n + 4096` bytes for `n`
  bytes of TeX (`maxMathOutputLen`) is discarded for the escaped source, so
  a table of thousands of tiny cells cannot amplify.
- Per page: one budget per render, `min(64·len(page) + 4096, 2 MiB)`
  (`mathBudgetFrom`, `MaxMathPageOutput`). Every conversion is charged
  against it, including output that is then discarded, so it bounds the
  conversion work and not only the bytes written. Scaling with the page is
  what stops a page made only of tiny expressions (`$a$$b$…`), each costing
  a few hundred bytes of fixed `<math>` markup, from growing past the ratio
  one expression is held to. Once it is spent, the remaining expressions
  are rendered as escaped text.
- Delimiter search is linear. An unclosed inline opener would rescan the
  rest of its line (quadratic on a line of `$a `); `findCloseCached` reuses
  the previous search of the same line. The block lookahead
  (`hasClosingLine`) caches the range it has scanned, so a page full of
  unclosed `$$` or `\\[` lines is scanned once overall; the cache is kept
  per container.

### Styling

All math styling is in the stylesheet (`05-composition.css`, scoped to the
realm and readme views), with the theme's tokens:

- Table cell alignment is written as a `columnalign` attribute on each
  `<mtd>`, decided once per logical column (`columnAlign`), and the
  stylesheet maps it to `text-align`. Cells get back the browser's default
  padding (`0.5ex 0.4em`), which the global reset removes, and the tables
  of `align` and `aligned` (class `math-aligned`) get the side padding that
  pairs up their columns. Chrome's MathML Core does not
  implement `columnalign`, so the stylesheet is what aligns cells there.
- Math uses `font-family: math`, the system's math font, at `1.1em`, so it
  follows the text around it (headings, tables); `\text` uses the body font.
- The `math-color-*` classes map to the text tokens that have light and dark
  values (caution, success, info, tip, tertiary); orange, which had none,
  gets `--g-color-orange-400/600` and `--s-color-math-orange`.
- Display math scrolls sideways when wider than the column, as code blocks
  and tables do, with block padding so the scroll box does not clip limits
  and descenders.
- `merror` (whose text the converter puts in an `<mtext>`, as browsers draw
  no text placed directly in it) uses the caution colours, and the
  unconverted source fallbacks look like inline code.

The `r/docs/markdown` realm documents the syntax
(`examples/quarantined/gno.land/r/docs/markdown/markdown.gno`).

## Alternatives considered

- **Client-side KaTeX or MathJax.** The common choice, and the best
  typesetting. It needs a script and its fonts on every page that holds math,
  which gnoweb's server-rendered, no-JS pages avoid, and it runs a large
  third-party parser in the visitor's browser on attacker-written input.
  KaTeX's own `trust` and `maxExpand` options exist because its HTML
  extensions and macro expansion are an attack surface to configure.
- **Server-side KaTeX through a JavaScript runtime** (goja, or V8 through
  cgo). Keeps pages script-free and KaTeX can emit MathML only, but it puts
  a JS engine and a JS bundle in the gnoweb binary, makes per-render CPU and
  memory harder to bound than a Go converter with explicit caps, and adds a
  toolchain to audit.
- **SVG or image rendering.** Pixel-identical everywhere, but heavy, not
  selectable or searchable, poorer for screen readers than MathML, and
  gnoweb omits raw HTML from realm content, inline `<svg>` included; images
  would need generation, storage and caching.
- **No math.** Realm authors keep TeX in code blocks. No risk and no code,
  but formulas stay unreadable.

## Consequences

- A line of math that reads as a block start (`- x`, `+ 2y`, `> 0`, `1. a`)
  ends the math block, which then renders as text, just as it would end a
  paragraph. Writing it as `{}- x`, moving the operator to the end of the
  previous line, or indenting it by four spaces keeps it in the math. This is
  the price of not letting an unclosed `$$` swallow the blocks after it.
- Rendering depends on the browser's MathML. Chrome implements MathML Core,
  a subset of MathML 3: `menclose` (`\cancel`, `\bcancel`, `\xcancel`),
  `columnlines` (the `|` of an `array` column spec) and `rowspacing` are not
  drawn there. Cell alignment is covered by the stylesheet, and mathematical
  variants (`\mathbb`, `\mathbf`, `\mathcal`, ...) are written as Unicode
  mathematical alphanumeric characters rather than `mathvariant`, which
  MathML Core only honours as `normal`.
- Font: `font-family: math` uses whatever math font the browser picks for
  the system (for instance STIX Two Math on macOS, Cambria Math on
  Windows); a system without one
  falls back to a text font, where stretchy delimiters, radicals and large
  operators look plainer. Shipping a math font (Latin Modern Math, STIX Two
  Math) would make it uniform at the cost of a large download on every page
  with math.
- Inline math cannot wrap or scroll: a very long inline formula widens the
  page on a phone. Long formulas belong in display math, which scrolls.
- Every limit falls back to the escaped TeX source, never to an error page:
  an over-long, too deep, too expanding or over-budget expression is shown
  as text, and the rest of the page renders.
- Each `<math>` carries its TeX source as an annotation, which adds the
  source's size to every formula; it is counted in the output bounds.
- The converter is a fork. Fixes made upstream in TreeBlood do not reach
  gnoweb on their own, and the hardening here must be kept when porting
  them. The parts of its API gnoweb does not use (`TexToMML`, document
  numbering, macro arguments, the indenting writer) are removed.
- Pandoc's `$` rule has false positives: in `$GNOT/$ATOM` the second `$`
  follows a non-space and is not followed by a digit, so `GNOT/` renders as
  math. Authors write `\$` for a literal dollar.
- Realms that splice user text into their markdown now let it write TeX:
  `sanitize.InlineText` (`gno.land/p/nt/markdown/sanitize/v0`) and
  `chain/markdown.EscapeInline` (`gnovm/stdlibs/chain/markdown/markdown.go`)
  do not escape `$`, so a user's `$x$` renders as math in such a realm.
  The output is still escaped MathML within the bounds above, but escaping
  `$` there is left to a follow-up PR.
- Math is not rendered inside `<gno-foreign>`: the sandboxed renderer
  (`buildInnerForeignMarkdown`) does not load the math extension, so TeX
  there stays text.
