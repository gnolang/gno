# A benchmark regression gate on pull requests

## Status

Accepted.

## Context

The repository has close to a thousand `func Benchmark` functions that
compile, and no CI job runs any of them: a pull request that makes the VM or
the store twice as slow passes every check. Two of the benchmarks also could
not fail when they should:

- `gnovm/cmd/benchstore`: every benchmark called `b.Skip` when `-db` was not
  given. A skip is silent without `-v`, so `go test -bench . ./gnovm/cmd/benchstore/`
  printed `PASS` after measuring nothing.
- `tm2/pkg/bft/wal`: five `BenchmarkWalRead*` sizes (100KB to 1GB) were behind
  `b.Skip("TODO: benchmark failing")`. The cause was the helper: it wrote the
  message with a limit of `n+64` bytes but read it back with
  `maxTestMsgSize` (64KiB), so every message above 64KiB was rejected with
  `DataCorruptionError[length 102445 exceeded maximum possible value of 65536 bytes]`.

## Decision

1. `benchstore`: a missing `-db` is now a `b.Fatal`, so the run is red and the
   message says which flag to pass.
2. `wal`: the helper uses one limit for the writer and the reader, and checks
   the error of the write it used to ignore. 100KB and 1MB are revived. 10MB,
   100MB and 1GB are deleted: consensus opens its WAL with
   `maxMsgSize = 1MB` (`consensus/reactor.go`), so those sizes cannot reach a
   production WAL, and the 1GB case needs several GB of memory per iteration.
3. A new workflow, `ci-bench.yml`, runs `misc/benchci/suite.txt` on the pull
   request and on its merge base, in the same job, and fails on a large,
   significant regression. Mechanics and the noise handling are in
   `misc/benchci/README.md`: binaries compiled once per side, ten interleaved
   rounds with alternating order, benchstat's Mann-Whitney U test at
   p < 0.05, and an effect floor of +20% on sec/op and +10% on allocs/op.

## Alternatives considered

- **An absolute threshold per benchmark.** Rejected: the result would depend
  on which runner the job landed on.
- **Storing a baseline from master and comparing against it.** Rejected for
  now: it compares two different machines, which is exactly the noise this
  avoids. Measuring both sides in one job costs twice the run time, and buys a
  comparison that means something.
- **Running every benchmark.** Rejected: hundreds of comparisons at p < 0.05
  produce false positives on every run, and the wall clock is far beyond what
  maintainers will keep. The suite is about 40 sec/op rows.
- **Defaulting `benchstore` to pebbledb.** Rejected: its key sizes go to 1e9
  keys on disk; an implicit default would turn a stray `-bench .` into an
  hours-long disk fill.

## Consequences

- A pull request touching Go code under `gnovm/`, `tm2/` or `gno.land/` gets
  one more job, about 3 to 6 minutes of runner time.
- A slowdown under 20% passes. The gate is for the regression nobody noticed,
  not for drift; a reviewer can still read the benchstat table in the job
  summary.
- The suite is a text file. Adding a benchmark is one line, and the README
  says how to keep it inside its budget.
