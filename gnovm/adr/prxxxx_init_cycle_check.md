# PRxxxx: report initialization cycles at preprocess

## Context

Go rejects a package whose variable initializer depends on itself through
functions, methods or closures ("initialization cycle"). Gno orders package
initializers with the dependency edges `codaInitOrderDeps` records
(#5247) and panics with "circular dependency" only when `runFileDecls`
runs them; preprocess accepts the file. The errorcheck ledger (#6346)
lists 31 corpus files whose only gap is this rule.

## Decision

- `checkInitCycles` runs after the accumulate-mode walk
  (`PreprocessFilesCollect`): it follows `ATTR_DECL_DEPS` from each
  package variable, the way `resolveEffectiveDeps` does, and reports the
  first path back to a variable at that variable's declaration as
  `initialization cycle: x refers to f refers to x`. Recursion among
  functions alone is not a cycle. A variable on a reported path is not
  reported again.
- A method selected on a call result (`f().m`) now records its edge: the
  receiver's static type is a one-element tuple.
- Off-chain only for now: `Preprocess` (on chain) is unchanged, the same
  cycle still fails in `runFileDecls`, so deploy gas and messages do not
  move. Moving the check on chain is a separate decision once the
  type-checker series settles what preprocess rejects.
- Names accumulate mode dropped resolve to nothing and are skipped.

## Alternatives considered

- Reporting from `runFileDecls`' own detection: it runs at execution, so
  errorcheck files (never executed) would stay uncovered.
- A fresh dependency walk over the AST: duplicates `codaInitOrderDeps`,
  and the two could disagree.

## Consequences

- 31 errorcheck corpus files flip to `Clean`; `gno lint` reports cycles
  before deployment.
- Messages name the path, as gc's do, so a marker regex such as
  `initialization cycle|depends upon itself` matches.
