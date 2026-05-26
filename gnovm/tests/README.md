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

The `TestFiles` walker accepts both `.gno` (the native extension) and
`.go` files. `.go` filetests live under `files/gocorpus/testdata/` —
the `gocorpus/` parent names the purpose ("regression tests for files
from Go's standard test corpus"); the `testdata/` segment is the
Go-tooling shield (`go list` / `go build` / `go test` ignore any
directory named `testdata/`, so `.go` files there don't conflict with
package-discovery elsewhere in the repo). The walker enforces this:
`.go` files outside a `testdata/` segment are skipped.

Primary use case: regression tests for files lifted from Go's standard
test corpus (`/usr/local/go/test/`). After fixing a Gno bug surfaced on
a corpus file, drop the original `.go` file verbatim into
`files/gocorpus/testdata/` to lock the fix in CI — no rename, no
conversion.

**Comparison behavior.** The harness picks one of three modes based
on file content; each is bypassed by the corresponding explicit
directive (the blessed-divergence escape hatch).

*Errorcheck mode.* When the source carries inline `// ERROR "regex"`
markers (Go's standard test corpus convention), the harness applies a
PKGPATH+synthetic-main rescue (so files declaring `package p` reach
preprocess+typecheck instead of bouncing on the realm-naming
requirement), runs the file through Gno, and verifies at least one
marker's regex matches Gno's preprocess / typecheck / runtime error
output. Pass criterion is intentionally loose: Gno's preprocessor
stops at the first error, so requiring per-line marker matches would
fail most corpus errorcheck files. The signal we want is "Gno
catches the kind of error gc does", not "Gno enumerates every
individual error". Bypass with an explicit `// Error:` directive
carrying Gno's actual error wording.

*Compile-only mode.* When the source is not runnable (non-main
package, or no `func main()`), the harness applies the same
PKGPATH+synthetic-main rescue and PASSes iff Gno's preprocess and
the bundled go/types both produce no error. Mirrors gc's `// compile`
semantics: gc accepts the file and never runs it. Bypass with an
explicit `// TypeCheckError:` or `// Error:` directive carrying Gno's
actual error wording — useful when Gno legitimately rejects code gc
accepts and you want to lock that in (note: gno.land's deploy gate
runs go/types ahead of Gno preprocess, so divergences here are
production-shielded).

*Run mode* (default). When a `.go` filetest is runnable and has no
explicit `// Output:` directive, the harness derives the expected
output by running the file through the Go toolchain (`go run`) and
compares Gno's output to Go's. Test passes only when Gno matches Go.
Bypass with an explicit `// Output:` directive carrying Gno's actual
output.

See `files/gocorpus/testdata/run/canary.go` (run mode),
`files/gocorpus/testdata/errorcheck/canary.go` (errorcheck), and
`files/gocorpus/testdata/compile/canary.go` (compile-only) for the
minimal patterns.

Notes:
- Native filetest directives (`// PKGPATH:`, `// Output:`, `// Error:`,
  `// MAXALLOC:`, etc.) are interpreted normally.
- Go-corpus directives (`// run`, `// errorcheck`, `// compile`, etc.)
  on the file's first line are treated as plain comments — only the
  file's actual behavior is what gets compared.
- The `go run` subprocess (run mode only) has a 30s timeout; tests
  are expected to be short-running.
- `// GC_ERROR` markers are accepted as equivalent to `// ERROR`
  (corpus files mix them). `// GCCGO_ERROR` is intentionally ignored
  — this harness mirrors gc semantics.
- LINE/LINE+N substitution in marker patterns is NOT performed;
  literal text stays in the regex (matching is best-effort).
- For files outside both modes (`// *dir` multi-file tests,
  gc-internal): don't add them to `gocorpus/testdata/` — convert
  manually to a Gno filetest with the appropriate native directive
  instead.

## `stdlibs`: testing standard libraries

These contain standard libraries which are only available in testing, and
extensions of them, like `std.TestSkipHeights`.

## other directories

- `backup` has been here since forever; and somebody should come around and delete it at some point.
- `challenges` contains code that supposedly doesn't work, but should.
- `integ` contains some files for integration tests which likely should have
  been in some `testdata` directory to begin with. You guessed it,
  they're here until someone bothers to move them out.
