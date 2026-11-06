# benchci

A benchmark regression gate for pull requests, run by
[`ci-bench.yml`](../../.github/workflows/ci-bench.yml).

It runs a small fixed set of Go benchmarks ([`suite.txt`](./suite.txt)) on the
pull request and on its merge base, then fails when
[benchstat](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat) reports a large,
statistically significant slowdown.

## How it handles noise

CI runners are shared machines, so an absolute threshold ("this benchmark must
stay under 2 ms") would be red or green depending on the runner, not on the
code. Instead:

1. **Same job, same runner.** Base and head are measured back to back on one
   VM, so they share its CPU and its neighbours.
2. **Compiled once, interleaved.** Each side's test binaries are built once,
   then run in `COUNT` rounds (default 10). Odd rounds run base first, even
   rounds head first, so slow drift during the job hits both sides equally.
3. **A significance test.** benchstat compares the two samples of 10 with a
   Mann-Whitney U test and only reports a delta when p < 0.05.
4. **A wide effect floor.** A significant delta still has to be at least
   `MAX_TIME_REGRESSION` percent (default 20) on sec/op to fail. allocs/op does
   not depend on the machine, so its floor is tighter: `MAX_ALLOC_REGRESSION`
   percent (default 10).

The gate is meant to catch the 2x slowdown nobody noticed, not a 3% drift.

## Running it locally

```sh
git worktree add ../gno-base master
go install golang.org/x/perf/cmd/benchstat@latest
misc/benchci/run.sh ../gno-base . /tmp/benchci
misc/benchci/check.sh /tmp/benchci/benchstat.csv
```

`COUNT=4` shortens a local run.

## Adding a benchmark

Add a line to `suite.txt`. Keep the whole suite within its time budget: each
line runs `COUNT` times on each side. Prefer CPU-bound benchmarks without disk
I/O, and use a fixed iteration count (`20x`) for a benchmark with expensive
setup, since the `b.N` ramp repeats the setup.
