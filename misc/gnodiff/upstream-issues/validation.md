# Validation against upstream gnolang/gno master

Nothing in this folder was filed. No upstream issue or pull request was opened.

## Environment

Fetched `https://github.com/gnolang/gno` (default branch `master`) into a clean clone. This is not the fork's working tree.

| | |
| --- | --- |
| upstream commit | `cf8ff126706cc9ba2d226bc256b3060dfa7731e2` |
| upstream subject | `feat(tm2,gno.land)!: let governance freeze accounts through auth:p:frozen_addrs (#6520)` |
| commit date | 2026-11-26 19:49:34 +0100 |
| Go | `go1.25.9 linux/amd64` |
| gno binary | `go build -o /tmp/bin/gno ./cmd/gno` from `gnovm/` of that commit |
| `gno version` | `develop` (the binary does not embed the git hash) |
| `GNOROOT` | the clone root |

That commit is also the current `master` of this fork. The commands below were still run from the fresh upstream clone, not from the fuzzer branch.

## How the paths differ

Gno does not use one type checker for every command.

- `gno lint` and `gno test` call `TypeCheckMemPackage`, which is go/types. Both bugs are rejected there, for a package (`gno.land/p/...`) and for a realm (`gno.land/r/...`).
- `gno run` does not fail on that go/types result. It preprocesses and executes. Both bugs run.
- A `gnovm/tests/files` filetest type-checks with go/types and records `TypeCheckError`, then executes anyway. Both bugs produce output and a typecheck error. That split is the open issue [gnolang/gno#5601](https://github.com/gnolang/gno/issues/5601).

So "Gno accepts the program" is true of `gno run` and of filetest execution. It is false of `gno lint` and `gno test`. Deployment-style checking, which uses the same go/types path as lint, would reject both.

Runtime integer wrap is not the bug. The same subtraction through variables is accepted by gc and by `gno run`, and both print `18446744073709551584`.

## Bug 1: unused local

Validated on upstream master: **yes, on `gno run` and on filetest execution. No, on `gno lint` and `gno test`.**

Go 1.25.9 rejects the program. Gno's go/types path rejects it with the go/types wording. `gno run` prints `2` and exits 0.

Not described as intentional in `docs/resources/go-gno-compatibility.md`. Open issue [gnolang/gno#3426](https://github.com/gnolang/gno/issues/3426) already says unused values should fail at compile time and that `gno run` executes them. It is not an exact duplicate: that report uses a bare `1` expression, and today go/types does catch `x := 1`. The hole that remains is `gno run` ignoring a check lint already makes, plus filetests still executing ([#5601](https://github.com/gnolang/gno/issues/5601)). Closed [gnolang/gno#3965](https://github.com/gnolang/gno/issues/3965) is a different program (an inner `pid := 42` that is unread, outer `pid` prints 0) and was closed `not_planned`.

Draft: [unused-local.md](unused-local.md). Recommendation: do not open a third issue. Comment on #3426 with this repro, and point at #5601 for the filetest half.

## Bug 2: typed constant uint64 subtraction

Validated on upstream master: **yes, on `gno run` and on filetest execution. No, on `gno lint` and `gno test`.**

gc rejects `uint64(8) - uint64(40)`. `gno run` prints `18446744073709551584` (`2^64 - 32`). `gno lint` and `gno test` reject with the same overflow text go/types uses.

This is the open bug [gnolang/gno#6392](https://github.com/gnolang/gno/issues/6392), which already lists typed constant `+`, `-`, `*`, and negation wrapping, including `const d uint8 = 0; d - 1`. The `uint64(8) - uint64(40)` row is the same bug, not a new one. Not described as intentional. Runtime wrap of non-constant unsigned subtraction matches gc and should stay.

Related open PRs, not merged on this commit: [gnolang/gno#6409](https://github.com/gnolang/gno/pull/6409) (fold constants with `go/constant`), [gnolang/gno#6390](https://github.com/gnolang/gno/pull/6390) (untyped rune constants only; #6392 says that fix leaves typed constants alone).

Draft: [unsigned-constant-overflow.md](unsigned-constant-overflow.md). Recommendation: do not open a new issue. Add the row to #6392.

## Panic wording

The six pairs from the fuzzer do not change whether the program panics or the indexes it reports. They are not worth six issues. One combined low-priority issue is worth filing so panic-string diffs stop being re-reported as behavioral bugs. Draft: [panic-wording.md](panic-wording.md).

## `fmt` today

There is no production stdlib `fmt`.

- `gnovm/stdlibs/fmt` does not exist on this commit.
- A fmt implementation exists only as test overrides: `gnovm/tests/stdlibs/fmt/` (`print.gno`, `format.gno`, `fmt_test.gno`, and the rest). Filetests and `gno test` can import `fmt` because the test store overlays that directory. `gno run` and production package load do not.
- The compatibility doc says so directly. `docs/resources/go-gno-compatibility.md` marks `fmt` as `` `test` `` and footnote 9: it depends on `reflect`, and `gno.land/p/nt/ufmt/v0` is the suggested stand-in. "In test functions, `fmt` works."
- `ufmt` is that stand-in: `examples/gno.land/p/nt/ufmt/v0/ufmt.gno`, import `gno.land/p/nt/ufmt/v0`, with `Sprintf`, `Errorf`, `Fprintf`, and the print family. The package README says v0 is unaudited.

`gno fmt` is the formatter (gofmt for `.gno` files), not the fmt package.
