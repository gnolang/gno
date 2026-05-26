# tests

This directory contains integration tests for the GnoVM. This file aims to provide a brief overview.

GnoVM tests and filetests run in a special context relating to its imports.
You can see the additional Gonative functions in [gnovm/pkg/test/imports.go](../pkg/test/imports.go).
You can see additional standard libraries and standard library functions
available in testing in [gnovm/tests/stdlibs](./stdlibs).

## `files`: GnoVM filetests

The most important directory is `files`, which contains filetests for the Gno
project. These are executed by the `TestFiles` test in the `gnovm/pkg/gnolang`
directory.

The `files/extern` directory contains several packages used to test the import
system. The packages here are imported with the prefix
`filetests/extern/`, exclusively within these filetests.

Tests with the `_long` suffix are skipped when the `-short` flag is passed.

These tests are largely derived from Yaegi, licensed under Apache 2.0.

### `files/gocorpus/testdata`: filetests with the `.go` extension

`TestFiles` accepts `.go` files dropped under `files/gocorpus/testdata/`
verbatim — no rename, no conversion. Primary use: regression tests
for files lifted from Go's standard test corpus (`/usr/local/go/test/`).
The `testdata/` segment hides `.go` files from `go list` / `go build`
/ `go test`; the walker rejects `.go` files outside such a segment.

The harness picks a mode from file content; an explicit native
directive bypasses each mode.

| Mode | Trigger | Pass criterion | Bypass with |
|---|---|---|---|
| errorcheck | inline `// ERROR "regex"` (or `// GC_ERROR`) markers | at least one marker matches Gno's preprocess/typecheck/runtime error (loose: Gno bails on first error, per-line matching is too strict) | `// Error:` |
| compile-only | not runnable (non-`main` package or no `func main()`) | Gno preprocess **and** go/types both produce no error | `// TypeCheckError:` or `// Error:` |
| run | runnable, no `// Output:` | Gno's stdout matches `go run`'s stdout (auto-derived) | `// Output:` |

For non-`main` files (errorcheck, compile-only), a PKGPATH +
synthetic-`main` rescue is applied so they reach Gno preprocess
instead of bouncing on the realm-naming check.

Escape hatches — two single-line meta-directives:

- `// Unsupported: <reason>` — `t.Skip(reason)` before any execution.
  Use for feature gaps (channels, goroutines, generics, dot-imports,
  `complex`, …). Replaces the cross-file skiplist YAML the external
  [`gno-go-conformance`](https://github.com/gnolang/gno-go-conformance)
  tool uses. Example: `run/unsupported_canary.go`.
- `// Divergence: <category>: <reason>` — blesses a real Gno-vs-Go
  difference; the match verdict is **inverted** (passes iff Gno
  actually diverges). When Gno is later fixed, the directive becomes
  stale and the test FAILS so the blessing doesn't rot. Categories:
  `error-wording`, `error-early-bail`, `stdlib-formatting`,
  `stdlib-symbol-missing`, `stdlib-behavior`, `resource-budget`,
  `determinism`. Example: `run/divergence_canary.go`.

Canaries: `gocorpus/testdata/{run,errorcheck,compile}/canary.go`.

Notes:
- Go-corpus directives (`// run`, `// errorcheck`, `// compile`, …)
  on the first line are treated as plain comments.
- `go run` subprocess has a 30s timeout; `// GCCGO_ERROR` is ignored
  (mirroring gc); LINE/LINE+N substitution in marker patterns is NOT
  performed.
- Multi-file tests (`// *dir`) and gc-internal tests don't fit any
  mode — convert manually to a Gno filetest instead.

## `stdlibs`: testing standard libraries

These contain standard libraries which are only available in testing, and
extensions of them, like `std.TestSkipHeights`.

## other directories

- `backup` has been here since forever; and somebody should come around and delete it at some point.
- `challenges` contains code that supposedly doesn't work, but should.
- `integ` contains some files for integration tests which likely should have
  been in some `testdata` directory to begin with. You guessed it,
  they're here until someone bothers to move them out.
