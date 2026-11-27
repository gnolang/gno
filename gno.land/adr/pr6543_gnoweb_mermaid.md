# ADR: gnoweb draws a subset of Mermaid as server-side SVG

## Context

Realms explain themselves in markdown: a DAO's proposal flow, the calls
between realms, how votes or a treasury split, a grant's milestones. Mermaid
is the text format authors and assistants already write for diagrams, and
GitHub draws it, but in gnoweb a ```` ```mermaid ```` fence shows as code.

The usual way to draw Mermaid is mermaid.js in the browser, and none of it
fits gnoweb:

- **Weight.** `mermaid.min.js` 12.3 is 5.5 MB minified; all of gnoweb's
  JavaScript is 148 KB.
- **Isolation.** The diagram source is realm markdown, so attacker input.
  mermaid.js has had stored-XSS bugs, which is why GitHub, GitLab and Gitea
  run it in a sandboxed iframe. gnoweb sends `X-Frame-Options: DENY` and a
  CSP whose `default-src 'self'` covers frames, so it cannot frame a
  sandbox of its own, and running mermaid.js in the page would put it next
  to the wallet connection.
- **Styles.** mermaid.js writes a `<style>` element into its SVG, which the
  CSP's `style-src` drops.
- **No JavaScript.** gnoweb pages are rendered on the server and work
  without scripts (ADR-003); math made the same choice (pr4879: TeX to
  MathML on the server).

## Decision

### A package of our own: `markdown/mermaid`

`mermaid.Render(src) (svg []byte, width float64, err error)` parses a
source and draws it as an `<svg>` element `width` px wide. It depends on
the standard library and on `golang.org/x/text`, which gnoweb already
uses, copies no code from another renderer, and supports a subset of
Mermaid, chosen for what realms show:

| Type | Statements |
|---|---|
| `sequenceDiagram` | `participant`, `actor` (with `as`), the message arrows (`->`, `-->`, `->>`, `-->>`, `-x`, `--x`, `-)`, `--)`, `<<->>`, `<<-->>`), activation markers and `activate`/`deactivate` (accepted, not drawn), notes (`left of`, `right of`, `over` one or two), `loop`, `alt`/`else`, `opt`, `par`/`and`, `critical`/`option`, `break`, `autonumber` (with a start, a step and `off`), `title`; `;` separates statements |
| `pie` | `showData`, `title`, `"label" : value` (a repeated label keeps its first value, as in Mermaid) |
| `xychart-beta`, `xychart` | `title`, `x-axis` with an optional title and categories, `y-axis` with an optional title and range, `bar`, `line`; vertical only |
| `timeline` | `title`, `section`, periods with events on their line or on continuation lines; a colon followed by a space separates events |
| `kanban` | columns and cards by indentation, `id[label]` or a bare label, the `label`, `ticket`, `assigned` and `priority` metadata |

Every type also takes `accTitle`, `accDescr`, `%%` comment lines, a YAML
frontmatter (its `title` is drawn) and entity codes (`#59;`, `#quot;`, any
HTML entity name). Keywords ignore ASCII case, as in Mermaid. `%%{init}%%` directives and the frontmatter's
configuration are ignored, since gnoweb's theme decides the colors, and so
are kanban metadata keys outside the four above.

Anything else (another diagram type, a statement outside the table, a
malformed line, a node shape or icon) makes `Render` return an `*Error`, and
the fence shows as the code block it is, as an unconverted ```` ```math ````
fence does. The rule is never to draw a diagram that means something else
than in Mermaid: what is not understood is refused, not guessed. A few
deliberate differences remain: unquoted multi-word axis titles keep their
spaces, and an xychart without a range starts its y-axis at zero. `#`
starts a comment only at the start of a line, as in Mermaid; elsewhere it
stays text (Mermaid would end the line there), and a sequence line where a
`;` follows such a `#` is refused, since Mermaid would drop the statements
after it.

### Two forms: the fence and `<gno-mermaid>`

A ```` ```mermaid ```` fence is the form Mermaid, GitHub and every editor
know; it draws the diagram at its natural size, centered. A `<gno-mermaid>`
block takes the same source as its body and options as attributes, the
way gnoweb's other blocks do (`<gno-details summary="…">`,
`<gno-columns>`), with no fence inside:

    <gno-mermaid width="full">
    pie title Votes
        "Yes" : 42
    </gno-mermaid>

Its one attribute, `width`, is `fit` (shrunk to the container when wider,
never enlarged, as Mermaid draws by default) or `full` (the container's
width); without it the diagram keeps its natural size and scrolls when
wider, so its text stays the size of the page's. The body is read verbatim
up to a `</gno-mermaid>` line, like a fence. Another attribute or value, a
block its container or the document ends before the close tag, or a page
with diagrams off shows the body as code, so a mistake shows the source
rather than a guess. A tag line that opens no block (a self-closing
`<gno-mermaid/>`, a close tag with no block to end) renders as a comment,
as a stray `<gno-details>` tag does, rather than as an HTML block hiding
the lines up to the next blank one. In a `<gno-columns>` column, a diagram
without a width shrinks to the column, down to three quarters of its size
(its 12 px labels stay 9 px), then scrolls; an empty `<svg>` three
quarters as wide holds that floor, since a style attribute is not
allowed. Inside `<gno-foreign>`, which does not load this extension, the
tag is omitted like any other unknown one.

### The SVG: inline, CSP-clean, themed

The SVG is inlined in the page, so its text uses the page font and the
theme. Assistive technology reads it as one image (`role="img"`) labelled by
`accTitle`, the title or the type, with `<desc>` from `accDescr`. The
source follows the drawing in a visually hidden `<pre class="u-sr-only">`,
so a screen reader still reaches every label, as it did when the fence was
code; it is also what the Copy button of a rendered file copies (the
drawing carries `data-copy-exclude`). The omnibar's in-page find skips
diagrams: SVG text cannot hold its `<mark>`.

`Render` draws the title of every type above the diagram the type
returns, which it moves below the title, and centers under it when the
title is wider, with a `translate` of two numbers.

The SVG carries geometry as attributes and nothing else, like the chain map
(chainmap/templates/map.html): no `style` attribute and no `<style>`
element. Shapes and text get `gno-mermaid__*` classes, and the stylesheet's
MERMAID COMPONENT colors them from the theme tokens, so both themes apply.
Series take twelve color classes in turn, all light enough to carry dark
text at 4.5:1 in both themes.

`svgWriter` is the only code that writes the SVG, and its methods are its
allowlist: the elements `svg`, `desc`, `g`, `rect`, `line`, `circle`,
`polygon`, `polyline`, `path`, `text` and `tspan`; attribute values that are
plain decimals it formats (clamped, never an exponent or an infinity), a
`translate` of two of them, classes of a `class` type whose values are Go
constants, anchors, or text escaped by `encoding/xml` (which also replaces
the characters XML forbids).
No method writes a style, an event handler, a link or a reference to another
resource. A test parses every golden SVG and every fuzzed one against that
list.

### Text measured in the served font

Boxes are sized from the advance widths of Inter, the font the page sets the
text in. `gen_widths.py` (`make mermaid-widths`) reads them from the
`Intervar.woff2` gnoweb serves, at the weights 400 and 600 and the 14 px
optical size, into `widths_gen.go` (230 runes). A rune the font lacks is
estimated: a full em for East Asian scripts and emoji, 0.6 em otherwise.
Text wraps at spaces, and a word too long for its box (an address, a path,
East Asian text) breaks after a `/ . - _ ,` or between characters.

### Bounded work

- `Render` refuses a source over 16 KiB (`MaxSourceLen`).
- Layouts take time linear in the source, except the spacing of sequence
  participants, which takes messages times participants: a sequence
  diagram has at most 50 participants. A diagram takes a few milliseconds
  at worst.
- `Render` refuses a drawing wider or taller than 20,000 px or larger than
  512 KiB (`MaxOutputLen`), which also bounds how many items a type draws.
  `svgWriter` stops writing past 512 KiB, so such a drawing is not written
  in full first.
- A document tries at most 20 diagrams, fences and `<gno-mermaid>` blocks
  together, counted as they are rendered (a fence in a footnote never
  referenced takes none), and draws at most 2 MiB of SVG, which the first
  diagram that does not fit spends; the diagrams past a bound stay code,
  and are not drawn.
- `Render` turns a panic into an `*Error`, so a bug on some input shows the
  source rather than failing the page.
- The page still goes through gnoweb's output cap (`maxRenderOutputBytes`).

### Off by default

`RenderConfig.Mermaid`, set by `-mermaid` on gnoweb and `-web-mermaid` on
gnodev, enables the transformer through the parser context
(`GnoContext.Mermaid`), as the IPFS gateway does. Off, a fence is the code
block it always was. Content inside `<gno-foreign>` is not drawn: the
sandbox loads its own extensions, this one is not among them, and each
foreign block would otherwise get bounds of its own.

## Alternatives considered

- **mermaid.js in the browser.** Rejected for the reasons in the context:
  weight, an attacker-controlled input next to the wallet, a sandbox gnoweb
  cannot frame, and styles the CSP drops.
- **Reusing a Go Mermaid renderer.** Four pure-Go renderers were evaluated on
  the same corpus (valid diagrams, injection attempts, inputs built to blow
  up the layout, a minute of fuzzing each). None is proven: the most
  complete are weeks-old, single-maintainer ports that need Go 1.26 or later
  (gno builds with 1.25.9) or embed JavaScript engines (+37 MB); the light
  one sets no limit on its layout work; and all but one write inline styles
  the CSP drops. Vendoring one would mean owning tens of thousands of
  unproven lines.
- **Rendering as an image** (`<img src="data:image/svg+xml…">`). It would
  contain any SVG a third-party renderer writes, but loses the theme, text
  selection and the page font. Writing the SVG ourselves makes the
  containment unnecessary.
- **Full Mermaid.** Graph types (flowchart, state, class, ER) need a graph
  layout, the hard part of Mermaid. They come in follow-ups: gantt, then a
  layered layout with flowchart, then state. For the layout, porting dagre
  (MIT, the layout mermaid.js uses) with its license and a link to it, as
  mathml credits its origin, will be weighed against writing one.

## Consequences

- Realms can show diagrams without a script, in gnoweb's look, in both
  themes, on deployments that enable them.
- Compatibility is partial: a diagram that uses syntax outside the subset
  shows as code. The table above is the reference, and the supported subset
  grows type by type.
- gnoweb maintains the parser and layout of each type it draws.
- `widths_gen.go` must be regenerated when the served font changes.
