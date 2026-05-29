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
| errorcheck | inline `// ERROR "regex"` (or `// GC_ERROR`) markers | golden snapshot: per-line errors pinned in `// GnoError:` (Gno's own) + `// GoTypeCheckError:` (go/types guard); markers are gc provenance, not a gate |
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

  *Not yet implemented for compile-only mode.* That mode still uses
  the legacy single-line `// Divergence: <reason>` verdict-inversion
  directive: presence flips the verdict, stale blessings fail loudly,
  no pinned goldens.

- **Errorcheck golden snapshot.** For files with inline `// ERROR
  "regex"` markers, the harness walks the per-line errors and pins them
  in two golden blocks at the bottom of the file:
    - `// GnoError:` — **Gno's own** static (preprocess) / runtime errors.
    - `// GoTypeCheckError:` — the **go/types guard's** errors. go/types
      is the Go type checker that gno.land's deploy gate runs *ahead* of
      GnoVM preprocess; it's not Gno's own behavior, so it gets its own
      block. Crucially it still rejects even when GnoVM preprocess is
      permissive, so it's a real guard worth pinning separately. It also
      reports every error in one pass (no first-error bail), so it often
      covers markers GnoVM preprocess bails before reaching.

  The inline `// ERROR` markers are upstream (gc) **provenance**, NOT a
  pass/fail gate — wording may differ (the whole point of the migrated
  "known divergences"). What the test verifies is that the pinned
  behavior hasn't *changed*: both blocks must match. Mechanics:
    - go/types is captured once from the initial run (all its errors).
      GnoVM preprocess bails on the first error, so the harness
      neutralizes that line and re-runs to surface the next.
    - Gno joins multiple errors with `; `; the per-line message is the
      matching `; `-segment, marker-aligned when several apply (internal
      "should not happen" assertions are skipped for the real one).
    - A `package <x>` line is neutralized to `package main` (not
      commented out) so iteration continues past an invalid-package-name
      test — the package clause is the file's one global dependency.
    - On pass 1 a GnoVM error on an *unmarked* line is Gno's genuine
      first error (recorded); on a later pass it's a neutralization
      artifact (skipped).

  Verdict:
  - Both blocks present and matching → PASS. Missing/stale block → FAIL
    "run `--update-golden-tests`" / diff. The blocks are the contract;
    any change in either checker's per-line behavior shifts them → FAIL.
  - Neither Gno nor the go/types guard rejects the file (gc does) → FAIL:
    a real leniency divergence. Mark `// Unsupported: <reason>` if
    intentional (e.g. Gno runs neither gc's liveness nor stack-frame
    analysis, and neither does go/types), since there's no error to pin.

  **Partial coverage (`// GnoIncomplete:`).** When Gno bails in the
  declaration/preprocess phase before reaching every marker, the golden
  covers only the markers it reached. Such files carry an auto-written
  `// GnoIncomplete: covered N of M markers; …` tag (required whenever
  coverage is partial — the harness fails without it and flags a stale
  one when coverage later becomes complete). The file still passes on
  its pinned golden; the tag makes it greppable as a candidate for a
  future runnable variant (valid `package main` + declarations) that
  would exercise the unreached markers.

Canaries: `gocorpus/testdata/{run,errorcheck,compile}/canary.go`. The
208 migrated "known divergence" errorcheck files live at their upstream
corpus paths under `gocorpus/testdata/` (`fixedbugs/`, `syntax/`, …).

Notes:
- Go-corpus directives (`// run`, `// errorcheck`, `// compile`, …)
  on the first line are treated as plain comments.
- `go run` subprocess has a 30s timeout; `// GCCGO_ERROR` is ignored
  (mirroring gc); LINE/LINE+N substitution in marker patterns is NOT
  performed.
- A trailing `// GnoError:` block is stripped from the source before
  it reaches Gno (and trailing newlines normalized), so the golden
  doesn't perturb EOF-positioned error line numbers.
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
