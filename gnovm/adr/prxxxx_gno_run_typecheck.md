# Type-check `gno run` before execution

## Status

Implemented. Local interpreter only. On-chain admission already rejected these programs.

## Context

[gnolang/gno#3426](https://github.com/gnolang/gno/issues/3426) shows a program Go rejects and `gno run` executes:

```go
package main

func main() {
    const c1 = 1 < 8
    main()
    1
}
```

gc reports `1 (untyped int constant) is not used` and stops. `gno run` parsed the file and called `RunFiles`, so the unused expression never halted the recursive `main()` and the process ran until it was out of memory. An unread local (`x := 1`) is the same hole: gc says `declared and not used: x`, and `gno run` printed and exited 0.

The `go/types` pass behind `gno lint`, `gno test`, and `AddPackage` already applies Go's rules, including:

- a local variable must be read; an assignment that does not read it does not count
- a blank identifier is a use
- a variable used only inside a closure is used
- parameters, results, and package-level variables are exempt
- an expression statement that is not a call or receive is `... is not used`

On 2025-01-26 the program was refused on chain with that unused-expression diagnostic. [thehowl](https://github.com/gnolang/gno/issues/3426#issuecomment-2619235725) confirmed the on-chain checker is the gate. The 2026-11-27 repro on master narrowed the remaining hole to `gno run`: lint and test already exit 1, and a filetest records the `TypeCheckError` and still runs. That last part is [gnolang/gno#5601](https://github.com/gnolang/gno/issues/5601), not this bug. `misc/gnodiff` classifies the same split as `lint-run` (`corpus/control/unused_local`): gc and `TypeCheckMemPackage` reject, the VM run path executes.

## Decision

`gno run` calls `TypeCheckMemPackage` on the files it is about to run, after the Gno parser has accepted them and before `RunFiles`. Any error is printed with the existing `gnoIssue` formatter (`file:line:col: <go/types message> (code=gnoTypeCheckError)`) and the command exits 1. The program does not start.

Options:

- `ProdOnly`: `gno run` does not load test files.
- `TCLatestRelaxed`: a missing `gnomod.toml` is synthesized. The on-disk module is not consulted, so this pass checks the source rather than the module's version or draft bit.
- `AlreadyOnChain`: skips admission-only rules. Reserved declaration names stay on `addpkg`, `MsgRun`, and `enable`, as [pr6208](pr6208_reserved_name_decl_sites.md) requires for `gno run`. `go/types` diagnostics, including unused variables and unused expressions, still run.

## Alternatives considered

- **Reject in the preprocessor.** `PreprocessAllFilesAndSaveBlockNodes` re-runs every stored package at node boot. A new rejection there can refuse a package the chain already accepted and stop the node. `checkDeclNames` and `checkCurRebind` live in the type checker for that reason. A second implementation of Go's use rules would also drift from `go/types`, which is the wording gc ships.
- **Stop filetests on every `TypeCheckError`.** That is #5601. Filetests today record the diagnostic and still execute, including programs whose only type error is an unread local and whose `// Output:` is the point of the test. Folding that into this change rewrites those goldens without closing the `gno run` hole any further.
- **Reimplement the check inside GnoVM.** The nuisance program is already refused on chain. The interpreter path that still ran it is `gno run`'s decision to skip the checker, not a missing analysis in `go/types`.

## Consequences

- `gno run` no longer executes a program `gno lint` rejects. The recursive unused-expression program stops at the diagnostic instead of allocating until the process dies.
- A program the VM used to run and then fail inside preprocessing (a conversion `go/types` already rejects) now fails at type-check, with `go/types` wording, and does not start.
- On-chain admission, gas, and stored packages are unchanged. Node boot does not re-check this rule.
- Examples and filetests are unchanged: `gno test` already type-checks them, and filetests still run after recording `TypeCheckError`.
