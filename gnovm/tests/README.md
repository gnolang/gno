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

### `files/testdata`: filetests with the `.go` extension

The `TestFiles` walker accepts both `.gno` (the native extension) and
`.go` files. `.go` filetests live under `files/testdata/`; Go's own
tooling (`go list`, `go build`, `go test`) ignores any directory named
`testdata/`, so `.go` files dropped there don't conflict with
package-discovery elsewhere in the repo.

Primary use case: regression tests for files lifted from Go's standard
test corpus (`/usr/local/go/test/`). After fixing a Gno bug surfaced on
a corpus file, drop the original `.go` file verbatim into
`files/testdata/` to lock the fix in CI — no rename, no conversion.

**Comparison behavior.** When a `.go` filetest has no explicit
`// Output:` directive, the harness derives the expected output by
running the file through the Go toolchain (`go run`) and compares
Gno's output to Go's. Test passes only when Gno matches Go.

To bless an intentional Gno-vs-Go divergence (e.g. an error-wording
difference Gno will keep), add an explicit `// Output:` directive with
Gno's actual output. The directive bypasses the auto-derive step and
serves as documentation of the accepted divergence.

See `files/testdata/run/canary.go` for the minimal auto-derive
pattern.

Notes:
- Native filetest directives (`// PKGPATH:`, `// Output:`, `// Error:`,
  `// MAXALLOC:`, etc.) are interpreted normally.
- Go-corpus directives (`// run`, `// errorcheck`, `// compile`, etc.)
  are treated as plain comments — only the file's actual behavior under
  Go and Gno is compared.
- The `go run` subprocess has a 30s timeout; tests are expected to be
  short-running. For files Go can't `go run` (errorcheck-style with
  `// ERROR` markers, `// *dir` multi-file tests, gc-internal): don't
  add them to `testdata/` — convert manually to a Gno filetest with
  the appropriate native directive instead.

## `stdlibs`: testing standard libraries

These contain standard libraries which are only available in testing, and
extensions of them, like `std.TestSkipHeights`.

## other directories

- `backup` has been here since forever; and somebody should come around and delete it at some point.
- `challenges` contains code that supposedly doesn't work, but should.
- `integ` contains some files for integration tests which likely should have
  been in some `testdata` directory to begin with. You guessed it,
  they're here until someone bothers to move them out.
