# gnodiff — Gno vs Go differential fuzzer

`gnodiff` runs the same small program on the standard Go toolchain (`go run`)
and on GnoVM, then classifies what a caller can observe.

This is the "Gno vs Go differential fuzzer" project from the gnolang grants
list. It does not change the VM.

## Classes

| Class | Meaning | Fails the run |
| --- | --- | --- |
| exact | Same outcome and the same text | no |
| known | `// gnodiff:known`, and any `expect-go` / `expect-gno` lines still match | no |
| panic-wording | Both sides panic at the same operation (same indexes, same output before the panic) but the message text differs | no |
| real | Anything else, including a known seed whose expect lines moved, and a runner crash | yes |

`gnodiff-out/panic-wording.txt` lists every wording pair (gc text, Gno text,
the programs that hit it, and one repro). Real disagreements are line-reduced
into `gnodiff-out/repro/`.

## Layout

```
misc/gnodiff/
  main.go                         CLI
  corpus/<category>/*.src         enumerated seeds (checked in)
  corpus/<name>/prog.src          a few hand-written seeds
  internal/gen/                   enumerated catalog and the AST generator
  internal/diff/                  runners, classifier, minimizer, fuzz targets
```

Regenerate the enumerated files with `go run . -write-corpus corpus` after
editing `internal/gen`. A test fails if the files drift.

Generics are not generated. Gno does not implement them. `fmt` is not
imported: Gno's `fmt` is test-only (it needs `reflect`). The stdlib seeds
use `math`, `strings`, `strconv`, and `utf8`.

## Running

```bash
make gnodiff                         # seed corpus
cd misc/gnodiff && go test ./...     # fast tests only
go run . -generate 2000 -seed 1 -out gnodiff-out/generated
go test -fuzz=FuzzGenerated -fuzztime=60s ./internal/diff/
```

`-generate` builds type-correct programs from a deterministic AST generator
(same seed, same source). `-fuzz=FuzzGenerated` is that generator under
`testing.F`. `-fuzz=FuzzPrograms` still mutates seed bytes; most mutants do
not parse, so it is the secondary entry.

Ordinary `go test` does not run the corpus or the fuzzer. The nightly workflow
is `.github/workflows/ci-gnodiff.yml` (also `workflow_dispatch`, and pull
requests that touch this directory).

## What the seed corpus found

319 programs on Go 1.25.9: 290 exact, 3 known, 26 panic-wording, 0 real.

Known, annotated on the seeds:

- `corpus/runcap`: `[]rune` of a non-constant string. gc printed cap 32, Gno
  printed cap 5 (`cap == len`, the determinism rule). A constant conversion
  and short ASCII conversions agreed.
- `corpus/control/unused_local`: gc rejects `x := 1` when `x` is unread
  (`declared and not used`). Gno runs the program. This looks like a
  typechecker gap, not a rule Gno documents.
- `corpus/const/unsigned_sub`: `uint64(8) - uint64(40)` is a typed constant
  that does not fit. gc rejects it (`overflows uint64`). Gno wraps and prints
  `18446744073709551584` (`2^64 - 32`). The runtime wrap of a non-constant subtraction
  agrees; this is the constant rule. It looks like a constant-evaluator gap.

Panic wording (same panic, different text), from `panic-wording.txt`:

- division by zero: gc `integer divide by zero`, Gno `division by zero`
- negative shift: gc `negative shift amount`, Gno `negative shift amount: (-1 <type>)`
- index: gc `index out of range [3] with length 2`, Gno `slice index out of bounds: 3 (len=2)`
- nil pointer: gc `invalid memory address or nil pointer dereference`, Gno `nil pointer dereference`
- nil slice index: gc `index out of range [0] with length 0`, Gno `nil slice index (out of bounds)`
- low > high slice: gc `slice bounds out of range [2:1]`, Gno `invalid slice index 2 > 1`

`math`, `strings`, `strconv`, and `utf8` seeds agreed, including `Float64bits`
of negative zero, `NaN` inequality, and `IsInf`. `new(struct{})` pointer
identity agreed on this toolchain (`false`, then `true` for a pointer compared
with itself).

## Subset that is not generated

Goroutines, channels, `select`, `unsafe`, complex numbers, type parameters,
integer `range` (`for i := range n`), and imports other than `math`,
`strings`, `strconv`, and `utf8`. `int` is 64-bit on both sides here; a
32-bit gc run is not an oracle.
