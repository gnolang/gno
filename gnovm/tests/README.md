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

`TestFiles` also accepts `.go` files under `files/gocorpus/testdata/`,
verbatim copies of Go's own test corpus (`$GOROOT/test`). The harness picks
a mode from the file's content (runnable → run; compile-only files and files
with inline `// ERROR` markers are refused until their modes land),
pins both Gno's and Go's behaviour as
trailing `//` blocks, and ends every file with exactly one `// Verdict:` line
(`Clean`, `Unsupported`, `KnownIssue` or `KnownDivergence` as its first word). The contract,
the block vocabulary and the re-sync command are in
[`files/gocorpus/README.md`](files/gocorpus/README.md). The corpus files are
copyright The Go Authors, BSD-3-Clause (`files/gocorpus/LICENSE`); their
headers are kept verbatim.

## `stdlibs`: testing standard libraries

These contain standard libraries which are only available in testing, and
extensions of them, like `std.TestSkipHeights`.

## other directories

- `backup` has been here since forever; and somebody should come around and delete it at some point.
- `challenges` contains code that supposedly doesn't work, but should.
- `integ` contains some files for integration tests which likely should have
  been in some `testdata` directory to begin with. You guessed it,
  they're here until someone bothers to move them out.
