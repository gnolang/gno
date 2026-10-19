# PR6328: the preprocessor's checking half — the series design

Status: proposed; tracked in #6344. The harness and ledgers (#6328, #6329,
#6340, #6346, #6367) measure; the accumulate mode (#6345) and the rule PRs
(#6347 to #6371) build; #6254 lands the type-declaration cycle prescan.

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
   first error panics exactly as before: no behaviour, gas or wording
   change; a rule that now names its node moves a position (composite16:
   the non-constant literal index is reported at the index).
3. **Error codes** (`ErrCode`): every rule raises `fail(code, msg, at...)`
   and the code travels on `PreprocessError.Code()`. Wording may change;
   the code is the stable identity, the future hash input when error text
   leaves the ABCI result (not done here). Sites not yet converted carry
   `CodeNone`. Audit at this step: 237 `panic(` calls in `preprocess.go`
   and 76 in `type_check.go` (checking and internal) against 50 `fail`
   sites; `TestErrCodesUnique` keeps the codes distinct.
4. **Positions**: a rule names the node to point at; a folded constant
   borrows its source's span; a plain panic raised while predefining a
   declaration is positioned at that declaration.
5. **Accumulate mode** stays at statement/declaration granularity
   (#6345): a failed statement is dropped and the names it declares are
   poisoned; a dropped dependency poisons its names on the way up; a rule
   that finds several faults in one statement (duplicate cases) reports
   them all through `report`.

## Verification

The errorcheck ledger (`gnovm/tests/files/gocorpus`, 681 gc files and 334
go/types files): markers are gc's verdict, a file is `Clean` only when
preprocess alone reaches it, and verify fails a file whose behaviour
moves without its verdict. `KnownIssue errorcheck` goes 178 → 48 over
#6347 to #6370; each PR's description carries its flips. Every native
filetest where only go/types refused the program (except addressability,
#6083) now fails in Gno. Unit tests: `preprocess_collect_test.go`.

## Consequences

- New checks are protocol changes only on paper: go/types already refused
  every program they refuse, so nothing deployable changes until the
  on-chain switch PR removes go/types. Every walk a rule adds is metered
  in the PR that adds it (`OpCPUSlopeCheckNode`).
- Left: addressability (#6083), type-declaration cycle prescan (#6254,
  which replaces the cycle positioning here), the on-chain switch, and
  the gnomod `gno = "x.y"` versioning of rule sets.
