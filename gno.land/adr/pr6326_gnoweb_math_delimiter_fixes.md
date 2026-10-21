# ADR: math delimiter fixes after the gnoweb math extension

## Context

The math extension (`pr4879_gnoweb_math_extension.md`) made `$`, `$$`,
`\\(` and `\\[` open a formula. A review after the merge found two defects.

1. `sanitize.InlineText`, the helper AGENTS.md prescribes for attacker
   input in `Render`, escapes through the `chain/markdown` native
   `EscapeInline`, whose set did not hold `$`. A user name such as
   `$x^{2} + ADMIN_{verified}$` went through the sanitizer intact and
   rendered as a formula (italics, superscripts, subscripts) in the realm's
   own text. Commands such as `\color` stay out of reach, since the
   sanitizer already escapes `\`, but the text is no longer literal.
2. The closer of `$$`, `\\(` and `\\[` was found with `bytes.Index`, so it
   could sit inside a code span or a link destination:
   ``\\(a `x\\)y` b\\)`` cut the code span in two, and
   `$$a [l](/r/x$$help) b$$` broke the link. Only the single `$` closer
   skipped code spans, link destinations and autolinks.

## Decision

- Add `$` to the inline escape set of `chain/markdown` (`EscapeInline`).
  gnoweb renders `\$` as a literal dollar, so escaped text reads the same;
  it can no longer open math. `EscapeTitle` (`LinkTitle`) keeps `$` as is:
  a link title is an HTML attribute, never parsed as math, and gnoweb shows
  its backslashes, so "Price $5" would have read "Price \$5".
- Split the scan of `findDollarClose` into `findClose`: the code span, link
  and autolink skipping is shared, and only the test for a closer differs.
  Single `$` keeps its pandoc rules and its stop at a `](` that closes a
  label opened before it (`[Send $10](...)`). `$$`, `\\)` and `\\]` do not
  stop there: `$$ (0,1](x) $$` is a half-open interval, and stopping would
  turn formulas that render today into text.
- Keep the escape set's explanation in `markdown.go` only: stdlib `.gno`
  files are part of genesis, and even a comment change there moves the app
  hash pinned by `TestAppHashCrossrealm38`.

## Alternatives considered

- Escape `$` in `p/nt/markdown/sanitize` only. The escaping lives in the
  native so that every caller of `EscapeInline` shares one set; a second
  set in the package would drift from it.
- A separate scanner for the other closers. It would duplicate the
  skipping rules, which is how the single `$` closer and the others came to
  differ in the first place.

## Consequences

- `chain/markdown` is a VM native: realms that call `sanitize.InlineText` or
  `TableCell` now get `\$` where they got `$`. The rendered
  page is unchanged except that the dollar no longer opens math. The Gno
  examples suite passes.
- A `$$`, `\\(` or `\\[` expression now keeps a code span or a link
  destination whole, as `$` already did; golden cases are added to
  `golden/ext_math/dollar_link_test.txtar`, and the linear-scan test now
  runs the finders the parser uses for `\\(` and `\\[`.
- A formula opened inside a link label (`[Send $$10](/r/x$$help)`) still
  closes there, as on master: ending `$$`, `\\(` and `\\[` at the label's
  `](` would turn `$$ (0,1](x) $$` into text. In ``$$a` b$$ c``, the code
  span now wins over the closer, as CommonMark has it.
- Known limit, unchanged for `$`, now shared by the other closers: the
  search cache can answer "no closer" for an opener placed after a link
  destination that never closes on its line (`$$[0,1](x … $$y$$`), leaving
  that math as text. It never makes a formula cross a link.
- Known limit, unchanged: `sanitize.Block`, `BlockRich`, `Blockquote*` and
  `FootnoteDefinition` keep user math (their escaping targets block
  structure, not `$`). Math in a user post stays inside its blank-line
  envelope; whether posts may hold math is left to a follow-up.
