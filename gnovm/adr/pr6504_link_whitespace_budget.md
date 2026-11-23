# Charge a link's whitespace to the bracket scan budget

## Context

`findBracketSpans` in `gnovm/stdlibs/chain/markdown/markdown.go` pays for its
walks from one budget of `8 * len(s)` steps, so finding every link stays
linear in the input. `scanLinkText` and the destination and title loops of
`scanLinkURL` and `scanLRDTail` take one step per byte. Their whitespace
skips took none. A failed link scan resumes one byte past its `[`, so in
`"[a](" * k + " " * m + "!"` every opener walks the same spaces again, for
free. The budget lets `k` reach about `2·sqrt(m)`, so the free walk of `k·m`
bytes grows as the input to the power 1.5. The natives' gas grows only with
the input. On a 997,961-byte input with `k = 1990`, the two skips around a
link's title walked 1,969,110,000 bytes, about 1,973 per input byte, where
the budget allows 8.

## Decision

All six whitespace skips in `scanLinkURL` and `scanLRDTail` go through
`skipSpace`, which takes one budget step per byte. Out of budget, a link and
a definition's `:` fail the scan, as the destination loops do. Later in a
definition's tail the title is dropped, or the span ends where the skip
stopped, as the title scan there already does.

## Alternatives considered

- **Charge only the two skips a later opener can walk again**: the skip
  before a title or `)`, and the skip before `)` after a title. Each of the
  other four starts right after a byte only one opener owns. That bound rests
  on how the walker resumes, which no test pins, so all six are charged.
- **Remember where each whitespace run ends**, so no run is walked twice and
  no output changes, which would avoid a hardfork. A one-entry memo fails on
  two interleaved runs, which evict the entry at every opener, and a memo of
  any fixed size fails on more runs. Only a map from run start to run end
  holds, adding machinery whose only gain is the upgrade class.

## Consequences

- An input whose scans fit the budget keeps its output byte for byte.
- An input that runs out of budget sooner gets its later links and
  definitions escaped as plain text. Behind 632 openers and 100,000 spaces, a
  trailing `[t](u)` comes out as `\[t\](u)`. That output can be stored, so the
  fix is a hardfork.
- The natives' gas charge does not change: it is
  `Base + Slope·len(s)/1024`, taken before they run.
- The line ending a reference definition's title is still read outside the
  budget, in `isBlockInterrupt`. A separate change reads that line once, and
  `TestScanLRDTailChargesWhitespace` passes its `interruptAt` to
  `scanLRDTail`.
