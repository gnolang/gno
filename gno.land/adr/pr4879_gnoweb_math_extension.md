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

`markdown/ext_math.go` adds an inline parser, a block parser and a renderer,
registered by `NewGnoExtension` (`ExtMath`). The renderer converts each
expression to MathML on the server and writes it in place, inside a
`<semantics>` element that keeps the TeX source as an
`<annotation encoding="application/x-tex">`.

Delimiters, as written in the markdown source:

- `$…$` inline. Pandoc's rules decide what is math: the opening `$` must be
  followed by a non-space, the closing `$` must not follow a space or an
  escaping backslash nor be followed by a digit, so `$5 and $10` and `$ 5`
  stay text, and `\$` is a literal dollar.
- `$$…$$` display.
- `\\(…\\)` inline and `\\[…\\]` display. The backslash is doubled because
  `\(` is a CommonMark backslash escape: a single `\(x\)` renders as `(x)`.

A display delimiter written inside a line of text is inline-level: the
expression is rendered as `<math display="block">` inside the paragraph. A
display delimiter that opens a line, with its closing delimiter on a later
line, opens a math block.

### Parsing rules

- A math block is read like a paragraph. It ends at a blank line (TeX
  forbids paragraph breaks in math mode) or at any line that would interrupt
  a paragraph: thematic break, ATX or setext heading, code fence, list item,
  HTML block, blockquote. The check offers the line to goldmark's own
  CommonMark block parsers in a scratch context (`startsBlock`), so their
  interruption rules apply as written. An unclosed `$$` cannot swallow the
  lists, quotes and headings after it when a later line holds `$$`.
- A block opens only when a closing line follows within `MaxMathInputLen`
  bytes and before such a line; otherwise the opener is text.
- A closing line ends with the delimiter, which appears on it only once:
  the delimiter alone, or after the last line of math. Like a closing code
  fence, it may be followed only by spaces, so `$$y$$ trailing` does not
  close a block and is not split into a paragraph with a stray `$$`.
- A block whose container (blockquote, list item) ends before its closing
  line is rendered as its source text.
- An empty expression (`$$$$`, `\\(\\)`, a blank block) is text: it holds no
  math and would cost a whole `<math>` element for a few input bytes.
- Code spans, fenced and indented code keep their `$` literally.

### The converter: a port of TreeBlood

`markdown/mathml` is a port of TreeBlood (https://github.com/Wyatt915/treeblood),
a pure-Go TeX to MathML converter by Wyatt Sheffield, under the MIT licence.
The licence and the attribution are kept in `markdown/mathml/LICENCE.MD`.
The port is vendored in the tree, and gnoweb's changes (escaping, bounds,
output) are applied to it, so they are reviewed with gnoweb. The converter is created per expression
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
  taken in em) of at most 20pt, 2em
  (`maxRaisePt`); anything else is dropped and the content rendered
  unshifted or unspanned. An unbounded span would set the minimum size of a
  stretched arrow thousands of em tall; an unbounded shift would move math
  over the page around it.
- `\class{name}{x}` renders `x` and drops the class: page authors must not
  apply the site's CSS classes (to overlay the page, for instance). The only
  classes in the output are the converter's own (`math-displaystyle`,
  `math-textstyle`, `mathcal`, `mathscr`).
- `\newcommand`, `\renewcommand` and `\def` are accepted but their
  definitions are not expanded, so there is no macro expansion to bound.
- `FuzzMathRender` checks that no input produces `<script>`, `<style>`,
  `<iframe>`, `<object>`, `<embed>` or `<svg>`, an `on*` attribute or a
  `javascript:` value, and that the output stays within the size bounds
  below.

### DoS model

- Per expression, input: an expression longer than `MaxMathInputLen` (8 KiB)
  is not converted and is rendered as escaped text.
- Per expression, depth: `ParseTex` panics past `MaxParseDepth` (64) nested
  levels, and the converter recovers it as a conversion error; the
  pretty-print indentation is capped at 16 levels (`maxIndent`) so deep
  nesting cannot make the output quadratic.
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
  unclosed `$$` or `\\[` lines is scanned once overall.

### Styling

Table cell alignment is written as a `columnalign` attribute on each `<mtd>`,
decided once per logical column (`columnAlign`), and the stylesheet
(`05-composition.css`) maps it to `text-align` and to the side padding that
pairs up the columns of `aligned` environments. Chrome's MathML Core does not
implement `columnalign`, so the stylesheet is what aligns cells there. Math
uses the site's monospace family and a font size token, and display math
gets the page's block spacing.

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
- Font: math inherits the site's monospace stack, whose first family is the
  bundled Roboto Mono face. It has no OpenType `MATH` table, so stretchy
  delimiters, radicals and large operators are drawn with the browser's
  fallback and look plainer than with a math font. Shipping a math font
  (Latin Modern Math, STIX Two Math) would fix it at the cost of a large
  download on every page with math.
- Every limit falls back to the escaped TeX source, never to an error page:
  an over-long, too deep, too expanding or over-budget expression is shown
  as text, and the rest of the page renders.
- Each `<math>` carries its TeX source as an annotation, which adds the
  source's size to every formula; it is counted in the output bounds.
- The converter is a fork. Fixes made upstream in TreeBlood do not reach
  gnoweb on their own, and the hardening here must be kept when porting
  them.
