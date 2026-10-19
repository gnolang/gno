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
  (`PreprocessFilesCollect`): one Tarjan pass over the `ATTR_DECL_DEPS`
  graph (linear in declarations and edges, each visit metered) finds the
  strongly connected components; a component holding a variable is a
  cycle, reported once, at the variable the sorted-name walk from its
  earliest variable closes on, as `initialization cycle: x refers to f
  refers to x` (Go's position and path). Recursion among functions
  alone is not a cycle.
- A method selected on a call result (`f().m`) now records its edge: the
  receiver's static type is a one-element tuple.
- On chain too: `CheckInitCycles` runs after a package's files are
  preprocessed (`RunFiles`, `runMemPackage`) and panics with the first
  cycle as a positioned `*PreprocessError`, where `runFileDecls` used to
  fail later with "circular dependency". Rejected earlier and with the
  cycle named, and stricter: the old runtime walk's memo missed `var A =
  d(); var V = d(); func d() int { return V }`, which go/types already
  refused. The walk bills `OpCPUSlopeInitCycleVisit` per declaration and
  per edge visited. The ledger's on-chain check (#6346) requires the two
  paths to agree, which an accumulate-only rule would fail.
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
