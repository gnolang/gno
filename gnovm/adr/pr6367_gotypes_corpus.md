# PR6367: go/types' test suite as a second corpus

## Context

gc's errorcheck tests (#6346) exercise the compiler front end broadly,
one marker per rule occurrence. go/types' own suite
(`$GOROOT/src/internal/types/testdata`) is written one rule per file
with the checker's exact message, so it measures a type checker more
finely and names the rule a gap belongs to.

## Decision

- `testdata/gotypes/` holds the suite's `check`, `fixedbugs`, `spec` and
  `examples` files from the same upstream release, verbatim, checked by
  the same errorcheck mode. The harness reads go/types' block markers
  (`/* ERROR "text" */`, `/* ERRORx "regex" */`, several per line) and
  its `// -lang=` and `// -gotypesalias=` headers; a file without markers
  must preprocess cleanly (compile mode).
- Two more Gno-side evidence blocks: `ParseBlocked:` (checkable markers
  in a file Gno could not parse; the checker never ran, so they stay
  uncaught and the file is a `KnownIssue`, the block saying why) and
  `GnoWordingDiffers:` (covered lines whose Gno message matches none of
  gc's patterns, so a line match may be a coincidence; `ERROR` is a
  substring in go/types' files, `ERRORx` a regexp).
- Type set unions, `~` and `comparable` route to `Unsupported generics`;
  `-gotypesalias=0` files are `Unsupported`.

## Alternatives considered

- Only gc's corpus: it has no file for many go/types rules and does not
  state the message, so wording drift goes unseen.
- Counting a wording mismatch as uncaught: too strict while Gno's
  messages are its own; the block makes the mismatch visible instead.

## Consequences

- 334 more files: 57 Clean, 13 KnownIssue errorcheck, 8 marker-less
  files Gno rejects (five language-version features, three gaps), the
  rest Unsupported, 163 of them generics.
- Two rules the suite caught over-reporting are fixed in this PR.
