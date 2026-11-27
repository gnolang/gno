# gnodiff — Gno vs Go differential fuzzer

`gnodiff` runs the same small program on the standard Go toolchain (`go run`)
and on GnoVM, then compares what a caller can observe: printed output, and
whether the program returned, panicked, or was rejected. Disagreements are
written to a report, and unexpected ones are reduced to a shorter repro.

This is the scaffold for the "Gno vs Go differential fuzzer" project listed
in the gnolang grants README. It does not change the VM.

## Layout

```
misc/gnodiff/
  main.go                 CLI
  corpus/<name>/prog.src  seed programs, valid Go and valid Gno
  internal/diff/          runners, comparator, minimizer, fuzz target
```

The module is separate from the root `go.mod`, same as the other tools under
`misc/` that link the VM (`replace github.com/gnolang/gno => ../..`).

## How a run works

1. Each seed is one `package main` with no imports. `println` is the only
   output channel, so the programs stay inside the language subset both
   sides implement.
2. **gc.** The source is copied into a throwaway module and executed with
   `go run`. `println` writes to stderr on gc; the runner splits that stream
   into program output and a panic value, and drops the goroutine dump.
   A timeout, or a killed process, is non-termination.
3. **GnoVM.** An in-process machine is built the way `gno run` and
   `misc/multiarch-determinism` build one: `test.ProdStore`, `ParseFile`,
   `RunFiles`, `RunMain`. A gas meter and an allocation ceiling stop
   non-terminating mutants. Gas is recorded and not compared.
4. **Compare.** Successes must print the same text. Panics must have the same
   message after a short list of wording synonyms (nil-pointer text). If both
   sides reject the program, or both fail to terminate, that is agreement —
   compiler wording is not the behavior under test. A failure inside the
   runner itself fails the run.
5. **Minimize.** For an unexpected divergence, lines are deleted while the
   disagreement remains. The reduced file is written to `gnodiff-out/repro/`.
   A seed marked `// gnodiff:known <reason>` is reported and does not fail
   the process.

Pointer addresses in output are masked (`0xPTR`). Seeds that care about
pointer identity print a boolean instead.

## Running

From the repo root:

```bash
make gnodiff
```

Or from this directory:

```bash
go test ./...          # comparator, loader, one gc/GnoVM smoke program
go run . -out gnodiff-out
make corpus            # same as go run
```

`go test` does not execute the corpus. The corpus shells out to `gc` once
per seed and is the slow path; keep it out of CI's default test job.

Flags:

- `-corpus` (default `./corpus`): directory of `<name>/prog.src` seeds.
- `-out` (default `gnodiff-out`): `report.txt`, `report.json`, and `repro/`.
- `-timeout` (default `20s`): per-program `go run` deadline.
- `-gas` (default `50000000`): GnoVM gas limit. Exhaustion counts as non-termination.
- `-minimize` (default `true`): shrink unexpected divergences.
- `-budget` (default `30`): minimization attempts per divergence.

`GNOROOT` is optional. When it is unset the tool uses `gnoenv.GuessRootDir`
and then walks up from the working directory looking for `gnovm/stdlibs`.

Exit status is 0 when every seed agrees or is annotated `gnodiff:known`.
Unexpected divergences and runner failures exit 1.

## Fuzzing

`testing.F` mutates the seed sources. Ordinary `go test` skips it.

```bash
cd misc/gnodiff
go test -fuzz=FuzzPrograms -fuzztime=30s ./internal/diff/
```

Byte-level mutation of Go source mostly fails to parse. Inputs that do not
parse, or that use a construct from the unsupported list below, are discarded.
A parsed program that disagrees is a failing input; a reduced repro is written
under `gnodiff-out/fuzz/` (or `$GNODIFF_OUT/fuzz`).

## Known-unsupported Go features

Seeds and the fuzzer stay off this list. Gno's own compatibility notes are in
`docs/resources/go-gno-compatibility.md`. The target language is Go 1.17, with
the exceptions below.

- goroutines (`go`) and channels (`chan`, `select`)
- `unsafe`, `uintptr`, `complex64`, `complex128`
- type parameters (generics) and interface type-set unions (`~T`, `|`)
- `min`, `max`, and `clear` builtins (added after Go 1.17)
- integer `range` (`for i := range n`, Go 1.22)
- imports of any kind, including the standard library — there is no shared
  subset wired up yet, and map iteration order would not be comparable
- dot imports, and Gno-only types (`realm`, and the chain `address` type)

`int` is 64-bit on the gc side of this tool (it is run on a 64-bit host) and
in GnoVM. A 32-bit gc run is not a supported oracle.

Panic text that names the same event is rewritten to a canonical form, and the
numbers stay in the form. A wrong index still fails. The first corpus run
panicked on both sides for each of these, with the same line printed before
the panic:

- nil pointer: gc `invalid memory address or nil pointer dereference`, Gno
  `nil pointer dereference`
- division by zero: gc `integer divide by zero`, Gno `division by zero`
- negative shift: gc `negative shift amount`, Gno `negative shift amount: (-1 int)`
- index: gc `index out of range [3] with length 2`, Gno
  `slice index out of bounds: 3 (len=2)`
- slice bounds: gc `slice bounds out of range [2:1]`, Gno
  `invalid slice index 2 > 1`

`corpus/runcap` is annotated `gnodiff:known`. Converting a non-constant string
to `[]rune` sets `cap == len` in Gno (5 for `"héllo"`). gc returned cap 32 for
that value. A constant conversion `[]rune("héllo")` and the short ASCII cases
in `corpus/strcap` agreed, including `cap == len`. Every other seed agreed.

## Next steps

- **Gas.** The Gno result already carries `gas`. A follow-up can snapshot it
  per seed and fail when a VM change moves the cost, after checking the seed
  itself did not change (same loop bounds, same inputs).
- **Generation.** Replace byte mutation with an AST mutator that stays inside
  the supported subset: swap operators, nudge integer literals, reorder
  independent statements. That is what makes `go test -fuzz` spend its time
  on real programs.
- **Stdlib.** Allow a fixed import list of pure packages (`strconv`, `utf8`,
  `bytes`) once those packages are known to match, and keep the comparator
  away from map iteration and from anything that reads the clock.
- **CI.** A scheduled job, not the pull-request `go test` matrix, that runs
  `make gnodiff` and uploads `gnodiff-out/` on failure. The fast `go test` for
  this module is already in `ci / misc`.
