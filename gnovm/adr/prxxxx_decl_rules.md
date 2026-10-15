# PRxxxx: declaration rules and unused names at preprocess

## Context

The errorcheck ledger (#6346) lists marker lines gc rejects and Gno's
preprocess accepts. A group of them are declaration rules: `:=` shape,
range clause keys, `init`/`main` signatures, field/method and interface
method name clashes, and "declared and not used" for variables and
labels. Today the in-process go/types guard catches the unused names on
chain; preprocess itself does not, so errorcheck files (never executed,
never guarded) stay uncovered.

## Decision

- Declaration rules in preprocess: `:=` may not repeat a name and must
  declare at least one new variable in its Go scope (Gno folds a control
  statement's header into its body block, so header-declared names may be
  shadowed; sibling declarations and parameters may not); a range clause
  may not repeat its key as value; `init` and `main` may only be
  functions, and `init` takes no arguments or results (a `cur realm`
  parameter aside); a type may not have a field and a method of the same
  name; interface methods must be unique and non-blank.
- `checkUnused` walks a preprocessed function and reports every local
  variable without a use and every label no goto, break or continue
  names, at the declaration. A target of `=` or `:=` is not a use; unused
  constants and the blank label are fine; a name only a folded constant
  uses (`len(a)` of an array) counts as used. Keyed by the declaring
  statement, since faux blocks copy their header's name sources into
  each case block. In accumulate mode, names a dropped statement declared
  or mentioned are not reported.
- Native filetests that declared variables they never used now use them,
  so their intent and goldens stay; the five whose point was the unused
  name take the new error as their golden.

## Alternatives considered

- Leaving unused names to the go/types guard. Rejected: the guard is
  temporary, and errorcheck files never reach it.
- Reporting unused names as warnings. Rejected: Go rejects them, and the
  guard already does on chain, so nothing deployable changes.

## Consequences

- 10 errorcheck corpus files flip to `Clean`; the four `// compile` files
  with a blank label or a constant-only use stay `Clean compile`.
- The on-chain error for an unused variable moves from the guard's text
  to preprocess's, same line.
