# Stop the gnoweb alert parser from pinning the markdown reader

## Context

`>>>>>` followed by a TAB — six bytes — made gnoweb's markdown renderer
allocate until the process was OOM-killed. Upstream goldmark renders the same
input in microseconds; loading `ExtAlerts` alone was enough to trigger it. Any
realm could return those bytes from `Render(path string)`, so a six-byte
`Render` output could take down a gnoweb instance.

`alertParser.process` ended with:

```go
if line[pos+advanceBy-1] == '\t' {
    reader.SetPadding(2)
}
```

Two things are wrong with it.

First, `process` is called from `Open` *before* the `[!KIND]` regex decides
whether the line is an alert at all. When the regex does not match, `Open`
returns `nil, parser.NoChildren` — but the padding has already been set and
stays set for whichever parser handles the line next. A block parser that
declines to open must leave the reader exactly as it found it.

Second, the padding value is wrong in kind, not just in magnitude. goldmark's
`blockquoteParser.process` *consumes* the tab and sets padding to the
remainder of the tab stop (`util.TabWidth(reader.LineOffset()) - 1`). Setting a
fixed padding without consuming the tab leaves the reader on the same byte,
and `reader.Advance` spends padding units before it touches `pos.Start`. The
blockquote parser's own advance is therefore absorbed, `blockquoteParser.Open`
succeeds on the byte it already handled, `HasChildren` sends
`parser.openBlocks` back to its `retry` label, and the loop appends one
`ast.Blockquote` per pass forever. Traced, the reader sits at `segStart=4`,
`peek=">\t"` indefinitely.

The bug was not only a DoS. Because the alert extension is always loaded in
production, a plain `>` + TAB blockquote rendered as a literal code block
containing the marker:

```
>	body   ->   <blockquote><pre><code>&gt;\tbody</code></pre></blockquote>
```

where goldmark alone gives `<blockquote><p>body</p></blockquote>`. Tab-indented
alert bodies were mangled the same way. No golden fixture covered a tab after
`>`, which is presumably why the breakage went unnoticed.

A second, unrelated tab bug sat one parser over, in
`alertHeaderParser.Open`, and review of the fix above surfaced it:

```go
w, _ := util.IndentWidth(line, reader.LineOffset())
reader.Advance(w)
```

`IndentWidth` returns a width in COLUMNS and a position in BYTES. A tab
between `]` and the title is one byte but three or four columns, so advancing
by the width eats title bytes: `> [!NOTE]\ttitle` rendered as `tle`, and
`> [!WARNING]\t` — where the tab is the last byte on the line — advanced past
the newline and pulled the following line up into the summary. It predates the
reader pin, but it is the same class of mistake in the same tab arithmetic, so
it is fixed here rather than filed separately.

Reviewing that fix surfaced two more offset bugs of the same shape, both
predating it and both fixed here for the same reason.

`process` computed its advance relative to `pos`, the indent position, but
returned it as though it were relative to the start of the line, and `Open`
sliced `line[advanceBy:]` with it. Any leading indent therefore stayed in the
slice, the `[!KIND]` regex missed, and the alert silently degraded to a plain
blockquote — `  > [!NOTE] x` was not an alert. CommonMark allows a block start
to be indented up to three columns, and `consumeMarker` accepts exactly that
on every continuation line, so the two halves of the same parser disagreed
about what a marker looks like. The same slice bug hid alerts behind a
tab-padded outer blockquote (`>\t> [!NOTE] x`), where the padding arrives as
leading spaces from `PeekLine`.

`alertHeaderParser.Open` trimmed the line terminator with

```go
if len(line) > 0 && line[len(line)-1] == '\n' {
    segment.Stop = segment.Stop - 1
}
```

which leaves a lone `\r` behind on CRLF input. A `\r` is not empty, so it
counted as a title and suppressed the default kind label: `> [!NOTE]\r\n`
rendered `<summary>` with no text where `> [!NOTE]\n` renders `Note`. Nothing
between a realm's `Render` output and goldmark normalizes line endings, so this
is reachable from chain data. Plain goldmark strips the `\r` itself, which is
why only the alert summary showed it.

## Decision

Make `process` pure — detection only, no reader mutation — and move marker
consumption into a new `consumeMarker`, called from `Continue`.

`consumeMarker` mirrors `blockquoteParser.process` byte for byte, deliberately.
The two parsers share the `>` trigger and alternate on nested `>` lines, so
they have to agree on how far a marker advances the reader; the tab arithmetic
is copied rather than reinvented.

In `alertHeaderParser.Open`, advance by `IndentWidth`'s byte position instead
of its column width.

Have `process` return an offset measured from the start of the line, so `Open`
slices past the indent as well as the marker. That offset is also the correct
argument for `reader.Advance`, which spends reader padding before it touches
real bytes, so an indented and a padding-prefixed marker both work out. `Open`
now locates the `]` inside the subline the regex actually matched rather than
searching the whole line, which drops a `string(line)` conversion too.

Trim the line terminator with `bytes.TrimRight(line, "\r\n")`, subtracting the
bytes removed from `segment.Stop` rather than recomputing it, so any reader
padding stays accounted for.

## Alternatives considered

- Drop the `SetPadding` call and nothing else. Kills the DoS, and was the
  experiment that confirmed the cause, but leaves tab handling wrong in the
  other direction: the tab is fully consumed, so indentation inside an alert
  body stops being significant.
- Keep the padding in `process` and have `Open` restore it on refusal. Needs
  `process` to report what it changed and every caller to unwind it correctly —
  more moving parts than not mutating in the first place.
- Cap nesting depth for `>`. Bounds the blowup without fixing it; the reader is
  still pinned, and the wrong tab rendering remains.
- Fix it upstream in goldmark. Nothing here is goldmark's bug — `openBlocks`
  behaves correctly given a parser that reports success without consuming
  input.

## Consequences

`>` + TAB now renders exactly as upstream goldmark renders it, with or without
the alert extension loaded, and tab-indented alert bodies keep CommonMark tab-
stop semantics (`>\t\tcode` is an indented code block with the residual two
columns). Tab-separated alert titles survive intact. An alert marker may now be
indented up to three columns, or sit behind a tab-padded outer blockquote, and
still be an alert; four columns remains an indented code block. Three new
golden fixtures — `ext_alerts/valid_tab_body.md.txtar`,
`ext_alerts/valid_tab_title.md.txtar` and
`ext_alerts/valid_indented_marker.md.txtar` — pin that output; no existing
golden changed.

Widening what opens an alert widens the surface the pin lived on, so the fix
was re-checked against the same sweeps used to confirm it: ~598k inputs over
`> \t [ ! ] - \n N` differentially against plain goldmark (non-alert inputs
must render identically), and ~814k token-level inputs through the full
production extension stack checked for non-termination and output blowup. Clean
on both.

Five regression tests. Four are pure unit tests on the invariants that broke —
`process` must not move the reader, `process`'s offset must land on the `[`
whatever the indent, `consumeMarker` must advance the reader, and the title
terminator must be trimmed in full — and cost nothing. The first two are
table-driven over a source plus the offset to start at: the padding only got
set when the byte AFTER a `>` was a tab, so on `>>>>>\t` it fired at the
innermost marker, and a case that starts at offset 0 never reaches that branch.
The CRLF case is a unit test rather than a golden because txtar round-tripping
and git checkout normalization both eat lone `\r` bytes. The fifth asserts the
user-visible output matches
plain goldmark for nested markers; a regression makes it diverge at n=4 and
fail immediately, before reaching the depth that pins the reader, and it is
additionally fenced with a soft memory limit and a deadline so a regression
cannot exhaust a CI runner.
