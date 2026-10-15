# PR6346: errorcheck mode for the Go corpus

## Context

Go's `// errorcheck` tests carry inline `// ERROR "regex"` markers on the
lines gc must reject. They are the largest conformance corpus for a type
checker, and the Gno type checker that will replace the in-process go/types
guard is measured by the set of lines it reports. The run and compile modes
(#6329, #6340) compare Gno with gc's output; errorcheck files need no Go
side, the markers are gc's verdict. Gno's preprocess stops at its first
error, so one pass could cover at most one marker; the accumulate-mode
preprocessor (`PreprocessFilesCollect`) gives every error with its span.

## Decision

- Evidence is one block, `GnoError:` (every accumulate-mode error, per
  line, wording kept), and one `Verdict:` line.
- `Clean errorcheck` when a Gno error span covers every checkable marker
  line and no Gno error falls on an unmarked line.
- Otherwise `KnownIssue errorcheck: #6344 uncaught markers: lines …;
  over-strict: lines …`. Sync writes the citation itself: a marker Gno
  misses, or an unmarked line it rejects, is by definition a preprocess
  conformance gap, which #6344 tracks for the type-checker PR series; a
  per-cause `#NNNN` written by hand survives re-sync.
- gc-only files are `Unsupported`: `-0`/`-m` optimisation diagnostics,
  `-+`/`-p=` runtime-package contracts, `-lang` pins below go1.17,
  `GC_ERROR`-only and pragma-placement markers, markers on non-code lines.
- Parse errors are pinned up to the first syntax error; the parser's
  resynchronisation cascades are not gc's.

## Alternatives considered

- Counting a marker as covered when the go/types guard flags it (the
  first draft's `GoTypesOnly:` block). Rejected: the guard is temporary,
  and a guard-only catch is exactly a gap the series must close.
- The `TODO:` placeholder of run and compile mode. Rejected for this mode:
  hundreds of files would need the same hand-written `#6344`.
- Neutralising one line per error and re-running preprocess. Rejected:
  it cannot attribute errors exactly and needs no accumulate mode; it is
  what #6344's fixes make unnecessary.

## Consequences

- The `Clean` count is Gno's own: a file is Clean only when preprocess
  alone reaches gc's verdict.
- Each type-checker PR flips the files it fixes to `Clean`; verify fails a
  file whose Gno errors changed without a verdict update.
- Sync for errorcheck files is Gno-only and fast; `VERSION` does not
  affect them.
