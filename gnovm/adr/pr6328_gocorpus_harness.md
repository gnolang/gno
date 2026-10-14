# Running Go's test corpus as Gno filetests

## Status

Proposed in #6328; the run-mode files (942) follow in a stacked PR.

## Context

Go ships a regression corpus under `$GOROOT/test`: small `package main`
programs with a directive on line one (`// run`, `// errorcheck`, ...). Gno
aims at Go semantics, so the corpus is a ready-made conformance suite, but
nothing in the repo could execute it: the filetest walker took only `.gno`,
goldens were Gno-only, and a Gno-vs-Go difference had nowhere to be recorded.

## Decision

Corpus files are copied verbatim under `gnovm/tests/files/gocorpus/testdata/`
(a `testdata/` segment keeps `go build ./...` away from them) and run by the
existing `TestFiles` walker, which now accepts `.go` there. Nothing above the
harness-written tail is edited; `VERSION` pins the Go toolchain that records
the Go side and equals the `go` line in `go.mod`, so CI needs one toolchain.

Each file ends with evidence blocks and exactly one `// Verdict:` line:

- `GnoOutput:` / `GoOutput:` pin both sides; `GnoError:` pins Gno's error.
- `Verdict: Clean run`: outputs match.
- `Verdict: Unsupported <tag>: <reason>`: feature gap or gc-only test; the file
  is skipped. Reasons are auto-tagged (`generics`, `channel`, `reflect`,
  `nondeterministic`, `long`, ...).
- `Verdict: KnownIssue run: #NNNN <reason>`: Gno is wrong; the reason must cite
  the tracking issue.
- `Verdict: KnownDivergence run: <reason>`: Gno differs on purpose; the reason
  must cite the Gno decision it rests on (a `docs/` page, `#NNNN` or a URL).

Sync (`--update-golden-tests`) runs `go run` under the pinned toolchain and
rewrites the tail; on a divergence it proposes a `Known*` verdict with a
`TODO:` reason. Verify runs Gno only, never shells out to `go`, and refuses a
`TODO:` reason, a `KnownIssue` without a ticket, a `KnownDivergence` without a
citation, a `Clean` verdict that now diverges, and a `Known*` verdict whose
sides now agree. A `Known*` file is therefore a test that keeps running: a fix
flips it to `Clean` in the fix PR, and a changed failure shape fails as a
`GnoError:` diff.

A native `.gno` filetest opts into the same comparison by adding `GoOutput:`;
its Gno side stays the usual `Output:` golden. Files without it are untouched.

Only run mode lands here. Files with inline `// ERROR` markers (errorcheck)
and compile-only files are refused with a clear message; those modes need the
error-accumulating preprocessor and arrive with it.

## Alternatives considered

- **Reference an external `$GOROOT/test` instead of copying.** Avoids the
  license file and re-syncs, and can report whole-tree coverage. Rejected:
  verify would need a Go toolchain on every run or a golden store beside the
  reference anyway, a reviewer could not judge a verdict from one diff, and
  Gno runs a small slice of the corpus, so a whole-tree runner would mostly
  emit `Unsupported`.
- **Import only `Clean` files; track everything else as issues.** Rejected:
  an issue is never executed, so a fix goes unnoticed (the harness itself once
  cited `fixedbugs/bug446.go` as a live init-order bug; it runs `Clean` today).
- **Let sync's verdict pass verify.** Rejected: a divergence would be blessed
  without a human decision.

## Consequences

- `gnovm/tests/files/gocorpus/LICENSE` carries Go's BSD notice, as the files'
  headers require.
- Bumping `VERSION` means bumping `go.mod` and re-syncing; `GoTypeCheckError:`
  line sets follow the Go that built the test binary.
- Verify cost per corpus file equals a native filetest's; sync is the slow,
  opt-in path.
- One canary per verdict kind lives in `testdata/run/` (plus a `.gno` opt-in
  canary); the `KnownIssue` canary rests on a live bug (#5787) and will need
  its line flipped when that bug is fixed, which is the intended lifecycle.
