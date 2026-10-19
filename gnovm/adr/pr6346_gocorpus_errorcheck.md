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
- `Clean errorcheck` when every checkable marker line has a Gno error
  starting on it and no Gno error starts on an unmarked line.
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
- Coverage is by line: a marker is covered when a Gno checking error
  starts on its line (or ends on it, for a function that does not
  terminate), wording kept as evidence; one error credits at most those
  lines, so a span over several marked lines does not cover them all.
  An internal error (`should not happen` and kin) covers nothing: it is
  a VM bug, listed on its own without its stack dump. Matching by error
  code, once every rule carries one, is the follow-up.
- The on-chain path (`PreprocessFiles`, first error, nothing executed)
  also runs on every file and must agree with accumulate mode on whether
  the file is rejected, with its first error on a reported line; the
  ledger vouches for the path deploys take, not only for the measuring one.

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

## Syntax errors

A file go/parser rejects is still checked: `ParseFileCollect` pins every
parser error and the checker runs on the partial tree, so a marker on a
line the parser could not read is credited by the parser's error and the
rest of the file is measured. Only a file with no Gno form at all (no
package clause, or a partial tree Go2Gno refuses) is parse-blocked. A
gccgo-only marker is optional: never uncaught, never over-strict.
