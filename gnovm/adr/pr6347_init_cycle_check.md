# PR6347: report initialization cycles at preprocess

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
- On chain too: `CheckInitCycles` runs after a package's files are
  preprocessed (`RunFiles`, `runMemPackage`) and panics with the first
  cycle as a positioned `*PreprocessError`, where `runFileDecls` used to
  fail later with "circular dependency". Same programs rejected, earlier
  and with the cycle named; the walk bills `OpCPUSlopeCheckNode` per
  declaration visited. The ledger's on-chain check (#6346) requires the
  two paths to agree, which an accumulate-only rule would fail.
- Names accumulate mode dropped resolve to nothing and are skipped.

## Alternatives considered

- Reporting from `runFileDecls`' own detection: it runs at execution, so
  errorcheck files (never executed) would stay uncovered.
- A fresh dependency walk over the AST: duplicates `codaInitOrderDeps`,
  and the two could disagree.

## Consequences

- 31 errorcheck corpus files flip to `Clean`; `gno lint` and deploys
  report cycles at preprocess. Six native filetests take the positioned
  message as their golden.
- Messages name the path, as gc's do, so a marker regex such as
  `initialization cycle|depends upon itself` matches.
