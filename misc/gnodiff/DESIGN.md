# gnodiff design

Differential tester comparing GnoVM against the standard Go toolchain (`gc`)
on the Gno subset of Go. This tool lives under `misc/gnodiff` and does not
modify the VM.

## Architecture

Each program is evaluated three ways:

1. **`RunGo` (`go run .`)**: compiles and executes the program in a throwaway
   `go 1.22` module with `CGO_ENABLED=0`, `GOWORK=off`, `GOPROXY=off`,
   `GOSUMDB=off`, `GOTOOLCHAIN=local`, and a hard timeout + process-group kill.
   Pinning `go 1.22` matches Gno's per-iteration loop variables without letting
   newer Go 1.23+ syntax slip past `gc`. `println` output is read from stderr
   and separated from a panic dump only when the process exits non-zero and a
   top-level `panic: ` header precedes a `goroutine N [...]` stack trace.
   Runtime fatal errors (`fatal error: stack overflow`, out of memory) map to
   `KindNonTerm`.
2. **`RunGno` (in-process GnoVM)**: fresh `test.ProdStore`, `ParseFile`,
   `RunFiles`, `RunMain`, with a gas meter and `MaxAllocBytes` ceiling.
   Gas is recorded for inspection and is not a pass/fail signal. Recovery
   tracks execution stage (`stageSetup`, `stageLoad`, `stageRun`):
   `*gnolang.PreprocessError` is unwrapped cleanly (keeping out-of-gas as
   `KindNonTerm` and `runtime.Error` / `"should not happen"` as `KindInternal`),
   and unexpected panics during `stageRun` always map to `KindInternal`.
   Stdout printed before a panic or non-termination is preserved and compared.
3. **`RunLint` (`gnolang.TypeCheckMemPackage`)**: the `go/types` pass used by
   `gno lint`, run with `ProdOnly: true` and a per-root `TypeCheckCache`
   guarded by a mutex. Any panic inside the type-checker maps to `KindInternal`.

## Classification invariants

`gno run` does not type-check. A program that both `gc` and `gno lint` reject
while the VM still executes is classified as `lint-run` rather than a missed
type-check. Representative seeds are `corpus/control/unused_local`
([gnolang/gno#3426](https://github.com/gnolang/gno/issues/3426)) and
`corpus/const/unsigned_sub`
([gnolang/gno#6392](https://github.com/gnolang/gno/issues/6392)).

Key soundness rules in `Classify`:

- **Dedicated `internal` class.** Runner failures (`KindInternal` from gc,
  GnoVM, or lint) classify as `ClassInternal`, distinct from `ClassReal`
  language divergences, and fail the run without writing misleading repros.
- **Anchored runtime panic folding.** `panic-wording`
  ([gnolang/gno#6538](https://github.com/gnolang/gno/issues/6538)) only folds
  exact `runtime error: ...` messages for the same event, keeping numeric
  operands (`index [3]` vs `[4]`, negative shift `-1` vs `-2` when both sides
  carry an amount, and nil-slice index `0` with length `0` only). User panics
  without `runtime error: ` never fold into runtime panics, and
  `slice bounds out of range [lo:hi] with capacity N` never folds into
  `lo > hi`.
- **All reject diagnostics compared.** `normalizeDiag` strips file paths (with
  multi-line spans) and Gno's `--- preprocess stack ---` trailer while keeping
  every diagnostic line. `reject-wording` requires every diagnostic line on
  both sides to be a constant overflow/underflow (`bigint overflows` /
  `bigint underflows` vs gc's `overflows <type>`); mixed errors or different
  named types stay `real`.
- **Non-termination.** Two non-terminations compare stdout printed prior to
  timeout/out-of-gas; both-empty non-terminations agree because the limit that
  fired is not language semantics.
- **Drift and staleness detection.** Seeds marked `// gnodiff:known` validate
  their `expect-go` / `expect-gno` directives (drift becomes `real`) and flip
  to `known-agrees` if both toolchains start agreeing.
- **Output normalization.** Normalizes line endings, trailing whitespace, and
  6- to 16-digit hex addresses (`0xPTR`) while preserving shorter hex values
  such as `0x0`, `0x1`, and `0xff`.

## Corpus and generator

- **Hand-written seeds (`corpus/`):** 16 `.src` files (named `.src` so `go test`
  does not vet intentionally rejected or lint-run cases), including the three
  annotated seeds (`runcap`, `control/unused_local`, `const/unsigned_sub`).
- **In-memory catalog (`internal/gen.RawCatalog`):** integer width arithmetic
  and conversion grid plus language feature programs, combined with `corpus/`
  at runtime by `LoadCorpus` (309 programs total).
- **Deterministic generator (`internal/gen.Generate`):** driven by a seeded
  PCG PRNG. Most seeds are type-correct runtime programs composing randomized
  feature snippets (closures, defer/recover, interfaces/type switches, struct
  embedding and method values, labeled break/continue, goto, fallthrough,
  arrays/slices, overlapping `copy`/`append`, `string`/`[]byte`, struct-keyed
  maps, shadowing, float truncation/NaN, and per-width wraparound). One seed
  in 23 emits a typed constant overflow (`// gnodiff:const-overflow`), and one
  in 23 exercises `strconv.FormatFloat` across verbs `f`/`e`/`g`, bit sizes
  `32`/`64`, and `Inf`/`NaN`.
- **Token-accurate subset guard (`Unsupported`):** uses `go/scanner` and
  `go/parser` (`ImportsOnly`) to reject out-of-scope constructs (goroutines,
  channels/`<-`, `select`, `unsafe`, complex types/literals/builtins, generics,
  integer `range`, Go 1.21+ `min`/`max`/`clear` builtins, and imports outside
  `math`, `strings`, `strconv`, and `unicode/utf8`).

## Trade-offs

- **Sequential execution:** `RunCorpus` runs programs one at a time rather than
  in parallel goroutines. Each `RunGo` invokes `go run .`; sequential execution
  keeps memory/CPU predictable on 2-core CI runners and avoids spurious timeouts
  from `go build` cache contention.
- **Fresh `ProdStore` per `RunGno`:** prevents any program from leaking state
  into the next, while `RunLint` shares a per-root `ProdStore` and
  `TypeCheckCache` behind a mutex since `TypeCheckMemPackage` is read-only
  against the store.
- **Kind-preserving line minimizer:** `minimizeDivergence` is deterministic,
  budget-bounded, and keeps only reductions that preserve `(go.Kind, gno.Kind)`,
  parseability (when the original parsed), and `ClassReal`.
