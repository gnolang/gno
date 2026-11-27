# gnodiff — Gno vs Go differential fuzzer

`gnodiff` runs small Go programs on the standard Go toolchain (`go run`), on
GnoVM (`gno run`'s in-process machine), and through `gno lint`'s type-checker
(`gnolang.TypeCheckMemPackage`), reporting observable disagreements. It does
not modify the VM.

See [DESIGN.md](DESIGN.md) for architecture, invariants, and trade-offs.

## Classes

| Class | Meaning | Fails the run |
| --- | --- | --- |
| exact | Same outcome and text (path-stripped diagnostics for rejections; stdout before non-termination) | no |
| known | `// gnodiff:known`, and any `expect-go` / `expect-gno` lines still match | no |
| known-agrees | A `// gnodiff:known` seed now matches; the annotation is stale | no |
| panic-wording | Both sides panic on the same `runtime error:` event with different wording (indexes and operands preserved) | no |
| reject-wording | Both sides reject a constant overflow/underflow (`bigint overflows`/`underflows` vs gc's `overflows <type>`) | no |
| lint-run | `gc` and `gno lint` both reject, and `gno run` still executes | no |
| real | Anything else, including drifted `known`/`lint-run` expect lines | yes |
| internal | One of the runners (`gc`, GnoVM, or `lint`) failed internally | yes |

`gno run` does not type-check. `lint-run` separates that known gap
([gnolang/gno#3426](https://github.com/gnolang/gno/issues/3426),
[gnolang/gno#6392](https://github.com/gnolang/gno/issues/6392)) from real VM
or type-checker bugs. Panic wording differences are tracked in
[gnolang/gno#6538](https://github.com/gnolang/gno/issues/6538).

## Layout

```
misc/gnodiff/
  main.go                  CLI
  DESIGN.md                architecture and invariants
  corpus/<name>/prog.src   hand-written seeds (including annotated known/lint-run cases)
  internal/gen/            in-memory catalog and deterministic AST generator
  internal/diff/           runners (gc, GnoVM, lint), classifier, minimizer, fuzz targets
```

The integer-width grid and language catalog are built in memory from
`internal/gen.RawCatalog` by `LoadCorpus`. Pass `-write-corpus <dir>` to dump
them to disk for inspection.

## Running

```bash
make gnodiff                         # hand seeds plus the in-memory catalog
cd misc/gnodiff && go test ./...     # fast unit + smoke tests
go run . -generate 2500 -seed 1 -timeout 30s -out gnodiff-out/generated
go test -fuzz=FuzzGenerated -fuzztime=60s ./internal/diff/
```

Out of scope (not generated, and filtered by `Unsupported` for byte-mutation
fuzzing): goroutines, channels, `select`, `unsafe`, complex types/literals/builtins,
generics, integer `range`, Go 1.21+ `min`/`max`/`clear` builtins, and imports
other than `math`, `strings`, `strconv`, and `unicode/utf8` (`fmt` is test-only
in Gno). `int` is 64-bit on both sides.

Ordinary `go test` runs only the fast unit and smoke tests.
`.github/workflows/ci-gnodiff.yml` runs on a nightly schedule,
`workflow_dispatch`, and PRs touching `misc/gnodiff`, executing the fast tests,
the full seed corpus, and a 600-program generated batch with a 30s gc timeout.

## Corpus run

309 programs on Go 1.25.9: **282 exact, 1 known, 22 panic-wording,
2 reject-wording, 2 lint-run, 0 real, 0 runner failures.**

- `corpus/runcap` (`known`): `[]rune` of a non-constant string. Prints
  `cap == len` rather than raw cap: gc `5 false 2 true 5 true`, Gno
  `5 true 2 true 5 true`.
- `corpus/control/unused_local` (`lint-run`): gc and `gno lint` report
  `declared and not used: x`; the VM prints `2`.
- `corpus/const/unsigned_sub` (`lint-run`): gc and `gno lint` report
  `overflows uint64`; the VM prints `18446744073709551584`.
- `const/overflow_int8` and `const/overflow_shift` (`reject-wording`): both
  sides reject (`overflows` vs `bigint overflows target kind`).
- `panic-wording`: six families recorded in `panic-wording.txt` (division by
  zero, negative shift, slice index, nil pointer, nil slice index 0, low > high
  slice bounds).

## Generated run

2500 programs from seed 1 on Go 1.25.9: **1944 exact, 443 panic-wording,
100 reject-wording, 13 lint-run, 0 real, 0 runner failures.**

Panic wording covered division by zero and slice index out of bounds
(`[2..6]` with length `2`). The 13 `lint-run` programs are typed unsigned
constant underflows (`uint32(1) - uint32(5)` and `uint64(8) - uint64(40)`).
The 100 `reject-wording` programs are constant overflows and underflows
rejected by both sides.
