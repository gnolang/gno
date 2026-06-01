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

| Mode | Trigger | What's pinned |
|---|---|---|
| run | runnable (`package main` + `func main()`) | both sides always: `// GnoOutput:` / `// GnoError:` (Gno's run) + `// GoOutput:` (`go run`). Pass iff they match; a difference needs a verdict (below) |
| errorcheck | inline `// ERROR "regex"` (or `// GC_ERROR`) markers | golden snapshot: per-line errors in `// GnoError:` (Gno's own) + `// GoTypeCheckError:` (go/types guard); markers are gc provenance, not a gate |
| compile-only | not runnable (non-`main` package or no `func main()`) | Gno preprocess **and** go/types both produce no error |

For non-`main` files (errorcheck, compile-only), a PKGPATH +
synthetic-`main` rescue is applied so they reach Gno preprocess
instead of bouncing on the realm-naming check.

### Tag taxonomy

Two kinds of trailing `//` tags. **Facts** are raw observations, one per
origin: `// GnoOutput:` / `// GnoError:` (Gno runtime / preprocess),
`// GoOutput:` (`go run`, stdout+stderr combined), `// GoTypeCheckError:`
(the go/types guard gno.land runs ahead of preprocess). **Verdicts** are
the judgment derived from the facts — each file gets one:

| Verdict | Meaning |
|---|---|
| `// KnownIssue:` | a Gno **bug** — Gno disagrees with gc + go/types + Go, and it matters. The thing to fix. |
| `// KnownDivergence:` | a run-mode difference that's **accepted** / benign (formatting, map order, error wording) — not a bug. |
| `// Unsupported:` | Gno **can't process** it (feature gap) → `t.Skip`. |
| none = Clean | Gno matches everyone. |

Plus one caveat (not a verdict): `// GnoStaticIncomplete:` — errorcheck
marker coverage is partial (static-only; these files never run).

The goal is the `// KnownIssue:` set; the facts exist to *yield* it.
Urgency is **derived** by `gocorpus/gen_ledger.sh`, never stored: a
run-mode KnownIssue is a runtime divergence (ships past deploy, breaks in
production → 🔥 urgent); a static KnownIssue is preprocess over-strictness
(caught at deploy, can't ship → 🟠 deferred).

Escape hatches:

- `// Unsupported: <reason>` — `t.Skip(reason)` before any execution.
  Use for feature gaps (channels, goroutines, generics, dot-imports,
  `complex`, …). Each file declares its own skip reason inline;
  there's no cross-file skiplist. Example:
  `gocorpus/testdata/run/unsupported_canary.go`.

- **Run mode pins BOTH sides, always.** Every run file carries its
  Gno-side and Go-side facts at the bottom (blank-line separated,
  matching the `.gno` `// Output:` convention), so a reviewer can judge
  bug-vs-expected from the file alone:

  ```
  // GnoOutput:          # Gno's stdout (always, even empty)
  // <Gno's stdout>

  // GnoError:           # Gno's panic/error — ONLY when it errors
  // <Gno's panic>

  // GoOutput:           # `go run`'s stdout+stderr (always, even empty)
  // <Go's output>
  ```

  Gno's panic lands in `// GnoError:`, not `// GnoOutput:`; it's folded
  into the comparison to mirror Go's combined stream — otherwise a Gno
  panic with no stdout would compare equal to a clean Go run (both empty)
  and silently pass a real bug. Example where it diverges:
  `gocorpus/testdata/fixedbugs/bug446.go` (Gno panics where Go exits 0).

  **A difference needs a verdict.** If the Gno and Go sides differ, the
  file must carry exactly one:

  - `// KnownDivergence: <category>: <reason>` — the difference is
    **accepted** (benign). Its presence blesses the diff; the reason is
    for the reader.
  - `// KnownIssue: <reason>` — the difference is a **Gno bug** (a wrong
    result, or a panic where Go succeeds).

  The harness verifies the facts match current actuals and that a verdict
  is present iff the sides differ (a stale verdict on a now-matching file
  FAILS). `--update-golden-tests` auto-writes a `TODO:` default — a
  one-sided Gno error defaults to `// KnownIssue:` (a bug signal), other
  diffs to `// KnownDivergence:` — which the contributor then refines or
  reclassifies. A human's tag choice + reason is preserved across re-sync.

  **Recommended `KnownDivergence` category** (advisory):
  `// KnownDivergence: <category>: <explanation>`.

  | Category | Meaning |
  |---|---|
  | `error-wording` | Same kind of error/panic, different message text. |
  | `error-early-bail` | Multi-error file: Gno's preprocessor bails on a different error first. |
  | `stdlib-formatting` | Same logic, formatted output differs (e.g. `%v` for floats). |
  | `stdlib-symbol-missing` | Gno doesn't expose this stdlib symbol yet. |
  | `stdlib-behavior` | Same symbol, different observable behavior. |
  | `resource-budget` | Exceeds Gno's default alloc/gas budget. |
  | `determinism` | Output depends on map order, GC timing, or scheduling. |

  Use `unclassified` when none fit; add a category if a pattern recurs.

  **Output comparison.** Go's stdout and stderr are combined (Gno has a
  single output stream), so artifacts like Go's builtin `println`
  (stderr) aren't flagged when both runtimes emit the same text.

- **`.gno` opt-in.** Runnable `.gno` filetests (anywhere under
  `tests/files/`) opt into the Go comparison by carrying at least one of
  `// GoOutput:` / `// GoError:` / `// KnownDivergence:`; the harness then
  pins `// GoOutput:` too. Without these they keep pure-Gno behavior — the
  existing 1600+ files are untouched. Gno's golden stays the existing
  `// Output:` (not `// GnoOutput:`). Example:
  `gocorpus/testdata/gno/optin_canary.gno`.

  *Not implemented for compile-only mode*, which uses the legacy
  single-line `// KnownDivergence: <reason>` verdict-inversion directive
  (presence flips the verdict; no pinned goldens).

- **Errorcheck golden snapshot.** For files with inline `// ERROR
  "regex"` markers, the harness walks the per-line errors and pins them
  in two golden blocks at the bottom of the file:
    - `// GnoError:` — **Gno's own** static (preprocess) / runtime errors.
    - `// GoTypeCheckError:` — the **go/types guard's** full per-line
      errors. go/types is the Go type checker that gno.land's deploy
      gate runs *ahead* of GnoVM preprocess; it's not Gno's own
      behavior, so it gets its own block. It still rejects even when
      GnoVM preprocess is permissive, and reports every error in one
      pass (no first-error bail), so it often covers markers GnoVM
      preprocess bails before reaching. Listed in full (not deduped
      against `// GnoError:`) so the `GnoError ⊆ GoTypeCheckError`
      relation below is visible in-file.

  The model: a Gno error is **legitimate** if gc marks that line or the
  go/types guard also caught it — `// GnoError:` lines are a subset of
  (markers ∪ go/types). Any Gno error left over — Gno rejecting code
  *both* gc and the guard accept — is **over-strict** and goes to the
  `// KnownIssue:` block (below) instead.

  A third block, `// KnownIssue:`, pins the over-strict Gno errors —
  lines Gno rejects that **neither** a gc marker **nor** the go/types
  guard backs (Gno rejects code both accept; a Gno bug to fix). This is
  the same `// KnownIssue:` verdict as run mode, in its static shape (a
  per-line block instead of a free-text reason); the ledger buckets it as
  🟠 deferred (over-rejection is caught at deploy — no inconsistent
  state). Kept out of `// GnoError:` so legitimate behavior isn't
  conflated with bugs. The file still passes (the go/types guard's
  coverage is the contract); when Gno is fixed and stops erroring there,
  the block goes stale → re-sync removes it. Example: `const2.go` (Gno
  wrongly rejects the literal `1e+500000000` while go/types correctly
  flags only the overflow on use).

  The inline `// ERROR` markers are upstream (gc) **provenance**, NOT a
  pass/fail gate — wording may differ (the whole point of the migrated
  "known divergences"). What the test verifies is that the pinned
  behavior hasn't *changed*: the blocks must match. Mechanics:
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
  - Gno's error is a **feature-gap** message — an unsupported import
    (`unknown import path unsafe`/`syscall`/`net/http`/…) or an
    unimplemented Go feature (`channels`/`goroutines are not permitted`,
    generics, imaginary literals, dot imports, builtin shadowing) →
    auto-routed to `// Unsupported:` (skip): Gno can't process the file
    at all. Detected from Gno's *actual* error message (not the source
    text, so prose like "go to" in a comment can't trip it).
    `--update-golden-tests` writes the directive; thereafter the file is
    skipped pre-dispatch. (Mirrors gno-go-conformance's compat/classify
    feature-gap triage.) The go/types guard's own import failure
    (`could not import reflect (unknown import path "reflect")`) is
    treated the same way: Gno's preprocess can stay *lenient* about a
    missing stdlib (it may surface an unrelated error first, or none
    line-mappable), so a guard import-failure for a normal package path
    also routes to `// Unsupported:`. Invalid-import-path syntax tests
    (`import "/foo"`, control chars) are excluded — those keep their
    errorcheck golden (e.g. `import6.go`).

  **Partial coverage (`// GnoStaticIncomplete:`).** Static-only —
  errorcheck files are preprocess-only, never run, so there's no runtime
  dimension. When fewer than all markers are covered (a marker counts if
  Gno's own preprocess **or** the go/types guard caught it), the golden
  covers only the reached markers. Such files carry an auto-written tag of
  the form `// GnoStaticIncomplete: covered N of M markers (Gno
  preprocess: X, go/types guard: Y); …` (required whenever coverage is
  partial — the harness
  fails without it, and flags a stale one when coverage later becomes
  complete or the per-checker split changes). The two per-checker counts
  make Gno's **leniency** explicit: `Gno preprocess: 0` means Gno itself
  flags none of the markers and the coverage is carried entirely by the
  guard (the note reads "lenient"); `Gno preprocess: X` with X<M means
  Gno bailed before reaching the rest. The file still passes on its
  pinned golden; the tag makes it greppable as a candidate for a future
  runnable variant (valid `package main` + declarations) that would
  exercise the unreached markers.

  **Manual `// Unsupported:` override.** A hand-added `// Unsupported:`
  directive takes precedence over all of the above — it's checked
  pre-dispatch, so the file is skipped before any golden/KnownIssue
  logic runs, and it survives `--update-golden-tests` (the early skip
  never rewrites the file). Use it when a file's markers are
  fundamentally untestable even if Gno's own errors were fixed — e.g.
  gc liveness/codegen (`-live`) tests, whose markers aren't type errors
  so neither Gno nor the go/types guard can check them. (Replace any
  auto-written blocks with the directive when overriding.)

Canaries: `gocorpus/testdata/{run,errorcheck,compile}/canary.go`. The
~2000 migrated corpus files live at their upstream paths under
`gocorpus/testdata/` (`fixedbugs/`, `syntax/`, `typeparam/`, …). A
per-file ledger, bucketed by verdict with urgency derived (🔥 urgent
runtime KnownIssues first), is in `gocorpus/MIGRATION.md` (regenerate
with `gocorpus/gen_ledger.sh`).

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
