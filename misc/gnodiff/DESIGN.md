# gnodiff

Differential tester for the Gno subset of Go. It does not change the VM.

## What is compared

Each program is run three ways:

- `go run` in a throwaway `go 1.22` module (gc).
- An in-process GnoVM machine, the same shape as `gno run`: `test.ProdStore`, `ParseFile`, `RunFiles`, `RunMain`, with a gas meter and an allocation ceiling. Gas is recorded and is not a pass/fail signal.
- `gnolang.TypeCheckMemPackage`, the go/types pass `gno lint` uses.

`gno run` does not type-check. A program gc and lint both reject, and the VM still executes, is class `lint-run` ("run accepts what lint rejects"). That is not reported as a typechecker miss. `gno lint` already rejects an unread local and an overflowing typed constant such as `uint64(8) - uint64(40)`. Those two are filed as [gnolang/gno#3426](https://github.com/gnolang/gno/issues/3426) and [gnolang/gno#6392](https://github.com/gnolang/gno/issues/6392). Panic text differences are [gnolang/gno#6538](https://github.com/gnolang/gno/issues/6538).

## Classes

| Class | Meaning | Fails the run |
| --- | --- | --- |
| exact | Same outcome and the same text, or the same path-stripped reject diagnostic, or the same stdout before a non-termination | no |
| known | `// gnodiff:known`, and any `expect-go` / `expect-gno` lines still match | no |
| known-agrees | A known seed now matches. The annotation is stale | no |
| panic-wording | Both sides panic on the same runtime event. Indexes and shift operands are part of the event | no |
| lint-run | gc and gno lint reject; the VM runs | no |
| reject-wording | Both sides reject the same constant that does not fit. The VM says `bigint overflows`, or `bigint underflows` for a negative constant, and does not name the type; gc does. Two different named types, or an overflow against a different error, are real | no |
| real | Anything else, including a known or lint-run seed whose expect lines moved | yes |
| internal | One of the runners (gc, GnoVM, or lint) failed internally | yes |

Two rejections whose diagnostics still differ after stripping `file:line` are real. Two non-terminations with different stdout are real. Both non-terminations with empty stdout are exact: the limit that fired is not the language behavior.

Panic folding applies only to `runtime error:` panics and keeps numbers. `index out of range [3]` does not match `[4]`. A negative shift of `-1` does not match `-2` when both messages carry the operand. gc's negative-shift text has no operand, so that one pair cannot confirm the number Gno prints. The nil-slice wording fold applies only to gc's `[0] with length 0` against Gno's `nil slice index (out of bounds)` text. Any other index is real.

6- to 16-digit hex pointer addresses in stdout are rewritten to `0xPTR`. Shorter hex values (including `0x0`) are kept so nil and non-nil pointers do not collapse together. A program that needs to compare identity should print a boolean.

A `panic: ` line is a panic only when `go run` exits non-zero and a `goroutine N [` stack follows it. `println("panic: ...")` stays program output. Runtime fatal errors (`fatal error: stack overflow`, out of memory) are classified as non-termination.

## Corpus

Hand-written seeds live in `corpus/` as `*.src` (not `.go`, so `go test` does not vet them). The enumerated grid — integer widths, conversions, and one program per language area — is `internal/gen.RawCatalog` and is built in memory by `LoadCorpus`. It is not checked in.

`[]rune` of a non-constant string is a known difference: Gno sets `cap == len` (determinism); gc may not. The seed prints `cap == len`, not the raw cap.

## Generator

`Generate(seed)` is deterministic.

- Most seeds are type-correct programs: closures (including per-iteration loop variables), defer/recover, interfaces and type switches, struct embedding, method values, labeled break/continue, goto, fallthrough, arrays versus slices, overlapping `copy` and `append` aliasing, `string`/`[]byte` conversions, maps with struct keys, shadowing, float truncation and NaN (printed as integers and booleans), and per-width wrap using variables.
- One seed in 23 is a typed constant that does not fit (`// gnodiff:const-overflow`). gc and lint should reject it. The VM may wrap it, which is `lint-run`, not a new bug.
- One seed in 23 imports `strconv` and prints `FormatFloat` with verb `f`.

Not generated, and rejected by `Unsupported` if a byte mutant introduces them: goroutines, channels, `select`, `unsafe`, complex numbers, type parameters, integer `range`, Go 1.21+ builtins (`min`, `max`, `clear`), and imports other than `math`, `strings`, `strconv`, and `unicode/utf8`. `fmt` is test-only in Gno. `int` is 64-bit on both sides here.

Map iteration in generated programs is reduced to an order-independent aggregate. `FuzzGenerated` is the fuzz entry. `FuzzPrograms` does not load the seed corpus.

## Running

`make gnodiff` runs the hand seeds plus the in-memory catalog. `go test` in this module does not. `.github/workflows/ci-gnodiff.yml` runs the fast tests, that corpus, and 600 generated programs. The per-program gc timeout on the generated batch is 30s so a cold compile is not recorded as non-termination.
