# ADR: the preprocessor's checking half — rule files, run context, error codes

Status: proposed (PRs 4–12 of the type-checker ladder that follows the
errorcheck corpus PR).

## Context

The preprocessor is Gno's `types2 + walk`: it resolves names, folds
constants and lowers the tree the interpreter executes, and it checks.
Its checking half was incomplete, so go/types ran before it at deploy.
go/types checks Go, not Gno; its verdict follows the Go release that
built the node; it is unmetered; and two checkers drift both ways. The
ladder finishes the preprocessor's own checks so go/types can leave the
consensus path (kept off-chain as referee).

## Decisions

1. **One rule file per Go spec chapter**, pure functions over `Type` and
   `Node`: `check_types.go`, `check_exprs.go`, `check_assign.go`,
   `check_decls.go`, `check_consts.go`, `check_unused.go`. A rule takes
   what it needs (types, a node), decides, and `fail`s. `preprocess.go`
   keeps the walk and calls the rules from the TRANS_LEAVE hooks; it does
   not grow new inline checks.
2. **A run context** (`checkRun`, `check.go`): the store, the block stack
   the walk owns (shared by pointer) and, in accumulate mode, the error
   sink. The accumulate-mode handlers (`onPanic`, `predefineGuarded`,
   `report`, `reportUnused`) are its methods. On chain (`acc == nil`) the
   first error panics exactly as before: no behaviour, gas or text change.
3. **Error codes** (`ErrCode`): every rule raises `fail(code, msg, at...)`
   and the code travels on `PreprocessError.Code()`. Wording may change;
   the code is the stable identity, the future hash input when error text
   leaves the ABCI result (not done here). Sites not yet converted carry
   `CodeNone`.
4. **Positions**: a rule names the node to point at; a folded constant
   borrows its source's span; a plain panic raised while predefining a
   declaration is positioned at that declaration.
5. **Accumulate mode** stays at statement/declaration granularity
   (PR 2): a failed statement is dropped and the names it declares are
   poisoned; a dropped dependency poisons its names on the way up; a rule
   that finds several faults in one statement (duplicate cases) reports
   them all through `report`.

## Verification

`gnovm/tests/files/gocorpus` errorcheck corpus (681 files): the
`GoTypesOnly` metric (marker lines only go/types catches) went 359 → 70
over PRs 4–11 with zero over-strict lines added; each PR's commit message
carries its row. Every native filetest where only go/types refused the
program (except addressability, #6083) now fails in Gno. Unit tests: `preprocess_collect_test.go`.

## Consequences

- New checks are protocol changes only on paper: go/types already refused
  every program they refuse, so nothing deployable changes until the
  on-chain switch PR removes go/types.
- Left: addressability (#6083), type-declaration cycle prescan (#6254,
  which replaces the cycle positioning here), the on-chain switch, and
  the gnomod `gno = "x.y"` versioning of rule sets.
