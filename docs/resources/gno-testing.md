# Running & Testing Gno code

The `gno` binary runs and tests Gno code locally: `gno test` for unit tests,
`gno run` for evaluating one-off expressions, and filetests for golden tests of
realms.

All of them run against a mocked GnoVM, so there is no real chain and any state
changes stay in memory for that one command. Imports resolve from your local
gno installation rather than a chain, so tests see the same standard library and
examples packages you have on disk.

## Prerequisites

`gno` installed. See [Installation](../builders/install.md). The examples below
build on the counter realm from
[Getting started](../builders/getting-started.md): the `myrealm` package, with
`myrealm.gno` and `myrealm_test.gno` in the package directory.

## `gno test`

`gno test` runs a package's `_test.gno` files, much like `go test`. From inside
the package directory:

```
$ gno test .
ok      .       0.81s
```

Add `-v` for verbose output:

```
$ gno test . -v
=== RUN   TestIncrement
--- PASS: TestIncrement (0.00s)
ok      .       0.81s
```

Other flags cover test timeouts and performance checks. See `gno test --help`.

## Gas profiling

`gno test -gasprofile=<file>` writes a source-level gas profile of every
executed test (unit tests and filetests) in standard [pprof](https://github.com/google/pprof)
format, so the whole `go tool pprof` ecosystem works on gno gas. Use it to find
which functions in your realm actually spend the gas.

```
$ gno test -gasprofile=gas.pprof .
ok      .       0.81s
gas profile written to gas.pprof (view with: go tool pprof gas.pprof)
```

Explore it interactively — top functions, per-line annotation, or a flame graph
in the browser:

```
$ go tool pprof -top gas.pprof            # top functions by gas
$ go tool pprof -http=:8080 gas.pprof     # flame graph + call graph (needs Graphviz)
```

The profile records gas per gno function across several dimensions. Select one
with `-sample_index`:

| Sample index          | What it measures                              |
|-----------------------|-----------------------------------------------|
| `total_gas` (default) | cpu + alloc + store + other (gross, pre-refund) |
| `cpu_gas`             | CPU-cycle gas (execution)                     |
| `alloc_gas`           | allocation gas                                |
| `store_gas`           | storage read/write + amino (serialization)    |
| `other_gas`           | gas not classified into the dimensions above  |
| `refund_gas`          | gas refunded during execution (tracked separately, not netted into `total_gas`) |

`total_gas` is the gross billable gas; the net cost is `total_gas - refund_gas`.
Refunds are kept as their own dimension rather than subtracted so a child frame
never appears to cost more than its parent.

```
$ go tool pprof -sample_index=store_gas -http=:8080 gas.pprof   # storage flame graph
```

Notes:

- `-gasprofile` requires `-p 1` (the profiler runs single-threaded); the flag
  sets it, and errors if you pass `-p` greater than 1.
- A plain `gno test` run charges no store gas and does not meter allocation
  against the test meter; `-gasprofile` additionally wires both, so a profiled
  run reports `store_gas` and `alloc_gas` dimensions a normal run does not — and
  its `--- GAS:` total is correspondingly higher. This is expected: the extra
  metering is what makes those dimensions observable. On-chain profiling
  (`.app/profiletx`) is observation-only and does not change gas.
- `go tool pprof -top`/`-tree`/`-peek` work without Graphviz; the graph views
  `-http`/`-svg` need it installed.
- `-list <func>` needs `-source_path=<pkgdir>` because frames record only the
  file basename (full paths would be non-deterministic). Even then it resolves
  only for source under that one root, so stdlib frames and multi-package
  profiles (`gno test -gasprofile ./...`) may not fully resolve; `-top`/`-tree`
  do not need it.

To profile a **transaction** running on a local dev node instead of tests, see
[Profiling a transaction](#profiling-a-transaction) below.

## Profiling a transaction

A gas-profiler-enabled node exposes an `.app/profiletx` ABCI query that runs a
transaction through simulation (no state is committed) and returns a pprof
profile of its gas usage — the on-chain analogue of `gno test -gasprofile`, and
a way to answer "where did this transaction's gas go?" across the cpu, alloc,
and **store** dimensions.

[`gnodev`](./gnodev.md) enables this query automatically. The profiler is
**off by default on all other nodes** (never on validators) and is enabled per
node via `AppOptions.EnableGasProfiler` — it is a local-development feature.

The easiest way is the `-profile` flag on `gnokey maketx`: it signs the tx (like
`-simulate only`, without broadcasting) and writes the pprof to a file. Point
`-remote` at a profiler-enabled node such as `gnodev`:

```sh
gnokey maketx call \
  -pkgpath "gno.land/r/dev/counter" -func "Increment" \
  -gas-wanted 20000000 -gas-fee 1000000ugnot \
  -profile gas.pprof \
  -remote 127.0.0.1:26657 -chainid dev \
  devtest
# gas profile written to gas.pprof (ok)
# view with: go tool pprof gas.pprof
```

The flag works on every `maketx` subcommand (`call`, `run`, `addpkg`, `send`,
and the `session` subcommands).
Set a generous `-gas-wanted`: the tx runs under it, so too low a limit yields a
profile truncated at the out-of-gas point (the confirmation line reports `ok` vs
a "partial" note). Against a node without the profiler enabled, the command
fails with a clear "gas profiling is not enabled on this node" error.

Or query it from Go with the gno client:

```go
profile, log, err := client.ProfileTx(tx) // tx is a *std.Tx
if err != nil {
    // the node has profiling disabled, or the tx could not be decoded
}
_ = log // "ok", or a note that the profile is partial (tx ran out of gas / failed)
os.WriteFile("gas.pprof", profile, 0o644)
// then: go tool pprof gas.pprof
```

`profile` is the same gzipped pprof produced by `gno test -gasprofile`, so all
the `go tool pprof` viewing and `-sample_index` dimension-switching shown above
apply.

## `gno run`

`gno run` evaluates an expression against your package code, a quick way to
check a function during development without deploying. It works with pure
packages and plain, non-crossing functions. To exercise realm functions that
take a `realm` argument, use `gno test` or a filetest instead.

It's a program runner, not a REPL, so return values aren't printed
automatically. Wrap the expression in `println()`:

```
$ gno run -expr "println(Add(2, 3))" .
5
```

Pass `-debug` to start the GnoVM debugger. See this
[blog post](https://gno.land/r/gnoland/blog:p/gno-debugger).

## Example tests

`gno test` also supports example tests, [similar to Go](https://go.dev/blog/examples). An
example test function takes no arguments and begins with the word `Example`. Like the test shown
above, it must be in a file ending in `_test.gno`.
The function prints output which is compared to the expected output in the `// Output:` comment.

To try it, create a file `example_test.gno` which checks the expected value of the `Render` function:
```
touch example_test.gno
```

`example_test.gno`:
```go
package myrealm

import (
	"fmt"
)

func ExampleRender() {
	count = 10
	fmt.Println(Render(""))
	// Output:
	// Count: 10
}
```

:::warning Reserved function name
Your test file can have local helper functions, but `init()` is reserved for other types of tests.
Use something like `initialize()` instead.
:::

## Filetests

Filetests are golden tests typically used to test realms. They execute a `main`
function and compare actual output against expected output written as comment
directives at the bottom of the file.

Filetests use the `*_filetest.gno` suffix and are placed in a `filetests/`
subdirectory of the realm package.

:::warning Stability notice
Filetests are primarily intended as an internal tool. Their API and behavior
are not guaranteed to be as stable as standard `gno test` testing.
:::

### Example

```go
// PKGPATH: gno.land/r/demo/counter_test
// SEND: 1000000ugnot
package counter_test

import "gno.land/r/demo/counter"

func main() {
	counter.Increment(cross)
	println(counter.Render(""))
}

// Output:
// 1
```

### Running filetests

```bash
# Only run the filetest for a package (from the package directory)
gno test -run "_filetest.gno" .
# Update expected values when output intentionally changes
gno test --update-golden-tests .
```

### Directives

**Input directives** are single-line comments at the top of the file:

| Directive  | Description                          | Default |
|------------|--------------------------------------|---------|
| `PKGPATH`  | Package path. Use `r/` for realms.   | `main`  |
| `MAXALLOC` | Max memory allocation in bytes.      | `0`     |
| `SEND`     | Coins sent with the transaction.     | (none)  |

**Output directives** are multi-line comments at the bottom:

| Directive        | Matches                                       |
|------------------|-----------------------------------------------|
| `Output`         | Standard output.                              |
| `Error`          | Panic or error message.                       |
| `Realm`          | Realm state change operations.                |
| `Events`         | Emitted events (JSON).                        |
| `Preprocessed`   | Preprocessed AST.                             |
| `Stacktrace`     | Gno stacktrace on panic.                      |
| `Gas`            | Gas consumed.                                 |
| `Storage`        | Realm storage size diff.                      |
| `TypeCheckError` | Go type-checker error.                        |

:::info Pure package imports
Imports of pure packages are processed separately. If a pure package contains a
line like `println(1)`, its output cannot be checked by an `// Output:` directive.
:::
