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
- Keep a link destination whole only where goldmark reads one. Outside
  `<…>`, a destination ends at its first space, which only a title or the
  `)` may follow; otherwise the brackets are text and the search goes on:
  `\\( [a,b](x \\) and [l](/r/y)` closes at the first `\\)` instead of
  hiding it until the link's `)`. A destination that may continue on the
  next line, or holds a title or `<…>`, is still skipped to its balanced
  `)` or the line end. A `(` left open in a destination read as no link
  opens no link either (its destination ends at the same space), which
  keeps the scan linear when such destinations nest.
- A search stopped at the `](` of the label the expression opened in does
  not go on to the next line: `[Send $10](…)` then `Balance: 100$` stays a
  link and text.
- Carry a code span the opener's line leaves open into the next line's
  search, and look for a same-line closer of a display block with the
  inline parser's finder, so a code span hides a closer from the block
  parser too (``$$ a `$$` b`` opens a block).
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
  destination still skipped to the line end (one without a space, with a
  title or `<…>`, that never closes on its line: `$$[0,1](x$$y$$`), leaving
  that math as text. It never makes a formula cross a link. A destination
  with a space and no title (`$$[0,1](x … $$y$$`) is no longer skipped.
- Known limit: a code span is seen across the line break from the opener's
  line to the next only. One opened on the opener's line and closed on the
  next does not hide a closer that follows it on the opener's line
  (``\\(a `x\\)`` then ``b` c\\)``): telling would need the next line
  for every opener, outside the search cache.
- Known limit: `EscapeInline` turns a trailing `\` into `\\`, so a realm
  that writes `(` or `[` right after a sanitized value
  (`sanitize.InlineText(name) + "("`) can form a `\\(` or `\\[` opener; a
  second such value followed by `)` or `]` closes it. Those bytes are the
  ones a realm writes for AMS math, so the parser cannot tell them apart,
  and escaping the backslash as `&#92;` instead would break heading ids
  and the table of contents. The reach is the text styled as math: every
  backslash of the user's value is doubled, so no TeX command gets through.
  Realms should not put `(` or `[` right after a sanitized value.
- Known limit, unchanged: `sanitize.Block`, `BlockRich`, `Blockquote*` and
  `FootnoteDefinition` keep user math (their escaping targets block
  structure, not `$`). Math in a user post stays inside its blank-line
  envelope; whether posts may hold math is left to a follow-up.
