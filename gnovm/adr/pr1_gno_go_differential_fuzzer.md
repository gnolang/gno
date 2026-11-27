# Gno vs Go differential fuzzer

## Status

Proposed. Scaffold only: no VM behavior change.

## Context

Gno targets the Go 1.17 language spec, with deliberate exceptions (no
goroutines or channels, and a few later features such as per-iteration loop
variables that the VM has since picked up). Regressions show up as a program
that `go run` and GnoVM do not agree on: different printed results, a panic on
one side only, or a panic whose message describes a different event. Those
bugs have been found by hand — integer-to-string conversion of an out-of-range
code point is one — and the gnolang grants list asks for a fuzzer that looks
for them continuously.

`gno run` and `misc/multiarch-determinism` already build an in-process
`gnolang.Machine` (`test.ProdStore`, `ParseFile`, `RunFiles`, `RunMain`). The
fuzzer should call that, not grow a second executor inside the VM.

## Decision

Add `misc/gnodiff`, a separate module in the same style as the other `misc/`
tools that link the VM.

- Seeds are `package main` programs with no imports, one directory each, using
  `println` as the only output. The corpus covers arithmetic, conversions,
  strings, slices, maps, closures, defer/panic/recover, structs, interfaces,
  switches, goto, and shifts. Map iteration is reduced to an order-independent
  aggregate. `int` is compared as 64-bit, which is what GnoVM uses and what gc
  uses on the hosts this tool runs on.
- The gc side is `go run` in a throwaway module. The Gno side is the in-process
  machine above, with a gas meter and an allocation ceiling so a non-terminating
  mutant cannot hang the process. Gas is stored on the result and is not a
  pass/fail signal.
- The comparator matches on outcome class. Successful runs must print the same
  text. Panics must carry the same message after one wording synonym for the
  nil-pointer runtime error. Both sides rejecting the source, or both failing
  to terminate, counts as agreement. A crash of either runner does not.
- Unexpected disagreements are line-reduced and written under `gnodiff-out/`.
  A seed can be annotated `// gnodiff:known` when the difference is an accepted
  language rule (for example zero-sized pointer identity, if a run still shows
  it). The annotation is reported and does not fail the process.
- `testing.F` mutates seed bytes. Normal `go test` skips it. Inputs outside
  the supported subset, and inputs that do not parse, are dropped. `make
  gnodiff` runs the corpus; root `make test` only runs the fast tests.

Core packages under `gnovm/pkg/gnolang` are unchanged.

## Alternatives considered

- **Shelling out to the `gno` CLI.** Same semantics, but a process per seed,
  no gas figure, and a harder timeout story. The in-process machine is the
  path `gno run` already uses.
- **One package of test functions, executed once.** A single panic aborts the
  rest of the corpus, and a divergence is not a minimized file.
- **A grammar fuzzer as the first cut.** That is the right second step. A
  byte-level `testing.F` plus a fixed corpus is enough to wire the runners,
  the report, and CI, and it already executes real programs.

## Consequences

- `make gnodiff` is a manual and, later, a scheduled check. It is not on the
  pull-request test path, because each seed starts `go run`.
- The fast tests of the module are on the `ci / misc` matrix.
- Panic text that differs only by wording is canonicalized when the numbers
  agree (division by zero, negative shift, index, slice bounds, nil pointer).
  The first corpus run produced those five wording pairs and one behavioral
  difference: `[]rune` of a non-ASCII string has spare cap on gc and `cap == len`
  in Gno. That seed is annotated `gnodiff:known`. A new wording pair the table
  does not know is still a divergence.
- Gas, stdlib imports, and an AST mutator are follow-ups. They are described
  in `misc/gnodiff/README.md`.
