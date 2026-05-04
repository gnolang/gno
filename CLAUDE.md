# Project Instructions

## Verification rules

- After changing gas constants or allocation/GC logic, always run these before declaring done:
  - `go test ./gno.land/pkg/sdk/vm/ -run Gas`
  - `go test ./gno.land/pkg/integration/ -run txtar`
  - `go test ./gnovm/pkg/gnolang/ -run Files -test.short`
- Always run `/simplify` before presenting completed work on non-trivial changes.

## Before/after comparisons

- When comparing gas numbers or performance metrics before vs after, always verify the test logic hasn't changed (e.g. loop counts, input sizes). Show reasoning, not just the numbers.
- Never claim a percentage improvement without confirming the test is doing the same work in both cases.

## PR descriptions

- When writing PR descriptions, grep for all new/modified files in the diff (`git diff --stat`) and categorize them. Don't omit major new files like benchmarks, tooling, or calibration scripts.
- List all categories of work (features, bug fixes, tooling, tests) — not just the headline feature.

## Gno interrealm semantics

- Before writing or reviewing any caller-authentication, access-control, or cross-realm code in Gno (`/r/`, `/p/`, `/e/` packages), read `docs/resources/gno-interrealm.md`. Do not pattern-match from Solidity `msg.sender` or other-language intuition.
- `runtime.PreviousRealm()` only shifts on explicit cross-calls (`fn(cross, ...)`) into crossing functions (`func fn(cur realm, ...){...}`). A `PreviousRealm().PkgPath() == "..."` check inside a non-crossing function does NOT identify the immediate caller and is a security bug.
