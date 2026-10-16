# PR6364: import path, constant width and branch label rules

## Context

The errorcheck ledger (#6346) lists marker lines gc rejects and Gno's
preprocess accepts, among them misspelled import paths, untyped
constants of unbounded width, and labeled branches to a non-loop. Import
errors also surfaced once per file, at the package lookup, so accumulate
mode could report only the first.

## Decision

Checks in `check_consts.go` and the import step, on the preprocessed
tree:

- An import path is checked for spelling before the package lookup, per
  declaration, so accumulate mode reports each one.
- An untyped integer constant wider than 512 bits is a constant
  overflow, as in go/types. This also bounds constant folding, so the
  chained-shift gas test now expects the cap.
- A labeled `break` or `continue` must name an enclosing loop (`break`
  also a switch or select).

## Alternatives considered

- Leaving these to the go/types guard: temporary, and errorcheck files
  never reach it.
- A larger constant width cap: go/types' 512 bits is what Go programs
  are written against, and it bounds folding cost.

## Consequences

- 3 errorcheck corpus files flip to `Clean`; the other import and
  constant files in the group keep their verdict with updated evidence.
- Programs that folded constants past 512 bits are rejected at deploy,
  as gc rejects them.
