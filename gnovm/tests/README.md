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

| Mode | Trigger | Pass criterion |
|---|---|---|
| run | runnable (`package main` + `func main()`) | Gno's stdout == `go run`'s stdout |
| errorcheck | inline `// ERROR "regex"` (or `// GC_ERROR`) markers | at least one marker matches Gno's preprocess/typecheck/runtime error (loose) |
| compile-only | not runnable (non-`main` package or no `func main()`) | Gno preprocess **and** go/types both produce no error |

For non-`main` files (errorcheck, compile-only), a PKGPATH +
synthetic-`main` rescue is applied so they reach Gno preprocess
instead of bouncing on the realm-naming check.

Escape hatches:

- `// Unsupported: <reason>` — `t.Skip(reason)` before any execution.
  Use for feature gaps (channels, goroutines, generics, dot-imports,
  `complex`, …). Each file declares its own skip reason inline;
  there's no cross-file skiplist. Example:
  `gocorpus/testdata/run/unsupported_canary.go`.

- For **blessed Gno-vs-Go divergences in run mode**, a triple of
  pinned-golden directives records both sides + the blessing.
  Placed at the bottom of the file (matching the `.gno` `// Output:`
  convention), with blank-line separators between each entry:

  ```
  // GnoOutput:        # .gno files reuse the existing // Output:
  // <Gno's output>    # instead of // GnoOutput:.

  // GoOutput:
  // <`go run`'s output>

  // Divergence: <free-text reason>
  ```

  The harness verifies all three: the pinned outputs must match
  current actuals, the outputs must actually differ (otherwise the
  divergence is stale and the test FAILS with "remove the divergence
  triple"). The directive itself is a boolean — its presence blesses
  the diff visible in `// GnoOutput:` / `// GoOutput:`. The reason
  text isn't parsed; it exists for the future reader. Example:
  `gocorpus/testdata/run/divergence_panic.go`.

  **Recommended reason shape** (advisory, not enforced by the
  harness — `// Divergence: <free text>` works too):

  ```
  // Divergence: <category>: <one-line explanation>
  ```

  The seven categories below cover the divergences that have shown
  up so far. They name the *kind* of difference, which helps
  maintainers triage:

  | Category | Meaning |
  |---|---|
  | `error-wording` | Same kind of error or panic, different message text. |
  | `error-early-bail` | Multi-error file: Gno's preprocessor bails on a different error first, so the set of errors differs. |
  | `stdlib-formatting` | Same logic, formatted output differs (e.g. `%v` for floats). |
  | `stdlib-symbol-missing` | Gno doesn't expose this stdlib symbol yet. |
  | `stdlib-behavior` | Same symbol, different observable behavior. |
  | `resource-budget` | Exceeds Gno's default alloc/gas budget. |
  | `determinism` | Output depends on map order, GC timing, or scheduling — non-comparable. |

  Use `unclassified` when none fit. The list is gno's own — feel
  free to add a new category if a recurring pattern appears.

  **Output comparison.** Go's stdout and stderr are combined for the
  comparison (Gno has a single output stream). This keeps the
  comparison on the same footing as `go test`'s default and avoids
  flagging artifacts like Go's builtin `println` (writes to stderr)
  as divergences when both runtimes emit the same text.

  **Workflow.** Copy a corpus file verbatim → run it → if Gno
  matches Go, done. If diverges, re-run with `--update-golden-tests`
  and the harness auto-appends the triple (with a `TODO:`
  placeholder reason the contributor refines).

- **`.gno` opt-in.** The same triple works for runnable `.gno`
  filetests (anywhere under `tests/files/`): the harness invokes the
  Go toolchain and compares **only when** at least one of
  `// GoOutput:` / `// GoError:` / `// Divergence:` is present.
  Without these, `.gno` files keep their pure-Gno behavior — the
  existing 1600+ files are untouched. Example:
  `gocorpus/testdata/gno/optin_canary.gno`.

  *Not yet implemented for errorcheck/compile modes.* Those modes
  still use the legacy `// Divergence: <reason>` single-line
  verdict-inversion directive: presence flips the verdict, stale
  blessings fail loudly, no pinned goldens. Symmetric
  `// GnoError:` / `// GoError:` for these modes is planned as a
  follow-up.

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
