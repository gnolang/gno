# gnodiff — Gno vs Go differential fuzzer

`gnodiff` runs the same small program on the standard Go toolchain (`go run`),
on GnoVM (`gno run`'s in-process machine), and through the `gno lint`
type-check (`TypeCheckMemPackage`). It does not change the VM.

The design note is [DESIGN.md](DESIGN.md).

## Classes

| Class | Meaning | Fails the run |
| --- | --- | --- |
| exact | Same outcome and the same text. Two rejections compare a path-stripped diagnostic. Two non-terminations compare stdout, and match when both outputs are empty | no |
| known | `// gnodiff:known`, and any `expect-go` / `expect-gno` lines still match | no |
| known-agrees | A known seed now matches. The annotation is stale and is printed | no |
| panic-wording | Both sides panic on the same event. Indexes and, when both messages carry one, shift operands are part of the event | no |
| reject-wording | Both sides reject a constant that does not fit. The VM says `bigint overflows` or `bigint underflows` and does not name the type; gc does. Different named types, or an overflow against a different error, are real | no |
| lint-run | gc and gno lint both reject, and the VM still runs the program | no |
| real | Anything else, including a known or lint-run seed whose expect lines moved, and a runner crash | yes |

`gno run` does not type-check. `lint-run` is that gap, not a missed type-check:
`gno lint` already rejects an unread local and `uint64(8) - uint64(40)`.
Those are [gnolang/gno#3426](https://github.com/gnolang/gno/issues/3426) and
[gnolang/gno#6392](https://github.com/gnolang/gno/issues/6392). Panic text is
[gnolang/gno#6538](https://github.com/gnolang/gno/issues/6538).

Pointer addresses in stdout are rewritten to `0xPTR`. Compare identity with a
boolean. A line that starts with `panic: ` is a panic only when a goroutine
stack follows it.

## Layout

```
misc/gnodiff/
  main.go                  CLI
  DESIGN.md                design note
  corpus/<name>/prog.src   hand-written seeds
  corpus/control, const    the two lint-run seeds
  internal/gen/            in-memory catalog and the AST generator
  internal/diff/           runners, classifier, minimizer, fuzz targets
```

The integer-width grid and the language catalog are `internal/gen.RawCatalog`.
`LoadCorpus` builds them in memory. They are not checked in. `-write-corpus`
still writes them, for inspection.

## Running

```bash
make gnodiff                         # hand seeds plus the in-memory catalog
cd misc/gnodiff && go test ./...     # fast tests, including two generated programs
go run . -generate 2500 -seed 1 -timeout 30s -out gnodiff-out/generated
go test -fuzz=FuzzGenerated -fuzztime=60s ./internal/diff/
```

`-generate` is deterministic. One seed in 23 is a typed constant gc must
reject (`// gnodiff:const-overflow`). One seed in 23 prints `strconv.FormatFloat`
with verb `f`. The rest are type-correct and draw from: closures (including
per-iteration loop variables), defer/recover, interfaces and type switches,
struct embedding, method values, labeled break/continue, goto, fallthrough,
arrays versus slices, overlapping `copy` and `append` aliasing,
`string`/`[]byte` conversions, maps with struct keys, shadowing, float
truncation and NaN (printed as integers and booleans), and per-width wrap
using variables so the constant evaluator is not involved.

Not generated: goroutines, channels, `select`, `unsafe`, complex numbers,
type parameters, integer `range`, and imports other than `math`, `strings`,
`strconv`, and `unicode/utf8`. `fmt` is test-only in Gno. `int` is 64-bit
on both sides here.

`FuzzGenerated` is the fuzz entry. `FuzzPrograms` does not load the corpus.
Ordinary `go test` does not run the corpus. The nightly workflow is
`.github/workflows/ci-gnodiff.yml` (also `workflow_dispatch`, and pull
requests that touch this directory). Its generated batch uses a 30s gc
timeout.

## Corpus run

314 programs on Go 1.25.9: 283 exact, 1 known, 26 panic-wording,
2 reject-wording, 2 lint-run, 0 real.

- `corpus/runcap`: `[]rune` of a non-constant string. The seed prints
  `cap == len`, not the raw cap. gc printed `5 false 2 true 5 true`. Gno
  printed `5 true 2 true 5 true`. Constant conversions agreed.
- `corpus/control/unused_local`: lint-run. gc and gno lint report
  `declared and not used: x`. The VM prints `2`.
- `corpus/const/unsigned_sub`: lint-run. gc and gno lint report
  `overflows uint64`. The VM prints `18446744073709551584`.
- `const/overflow_int8` and `const/overflow_shift`: reject-wording. Both
  sides reject. gc says `overflows`; the VM says `bigint overflows target kind`.
- Panic wording is the six families in `panic-wording.txt` (division by zero,
  negative shift, index, nil pointer, nil slice index 0, low > high slice).
  A negative shift of a different amount, or a nil slice at an index other
  than 0, is real.

`unicode/utf8`, `math`, `strings`, and `strconv` agreed. An earlier seed
imported `"utf8"`, which both toolchains reject; the import is `unicode/utf8`.

## Generated run

2500 programs from seed 1, Go 1.25.9: 1929 exact, 458 panic-wording,
100 reject-wording, 13 lint-run, 0 real.

Panic wording was only `integer divide by zero` versus `division by zero`,
and `index out of range [5] with length 2` versus
`slice index out of bounds: 5 (len=2)`. The 13 lint-run programs are
`uint64(8) - uint64(40)`, the same constant as `unsigned_sub`. The 100
reject-wording programs are constants that do not fit, including negative
constants the VM reports as `bigint underflows` and gc reports as
`overflows <type>`. No new behavioral disagreement.
