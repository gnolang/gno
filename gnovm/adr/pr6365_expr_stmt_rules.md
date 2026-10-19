# PR6365: expression statement and builtin rules at preprocess

## Context

The errorcheck ledger (#6346) lists marker lines gc rejects and Gno's
preprocess accepts: expression statements that are not calls, misuse of
`...` and of the value builtins, taking a function's address, and
non-constant literal indices. Three native filetests relied on go/types
refusing what Gno then ran.

## Decision

Checks in `check_exprs.go`, called from preprocess on the preprocessed
tree:

- An expression statement must be a call; conversions and the value
  builtins (`append cap complex imag len make new real`) are not used
  there.
- No `...` argument to a builtin other than `append`; `append`'s first
  argument is a typed slice.
- A declared function has no address and is not assignable.
- Array and slice literal indices are constants; `len` of a constant
  string folds to a constant.
- Unused locals are also checked in package-level function literals.
- A type assertion statement `x.(T)` stays accepted, which go/types
  refuses ("comma, ok expression ... is not used"): Gno runs it for its
  panic and programs rely on that. This divergence from Go is deliberate
  and survives the on-chain switch.

`map42`, `var27` and `composite16` are error filetests now; `defer0`
keeps its run coverage without the spread into `println`. Goldens print
the assignment target as written.

## Alternatives considered

- Leaving these to the go/types guard: temporary, and errorcheck files
  never reach it.

## Consequences

- 6 errorcheck corpus files flip to `Clean`.
- A program whose statement was a bare conversion or value builtin is
  rejected at deploy, as gc rejects it.
