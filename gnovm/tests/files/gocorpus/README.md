# gocorpus: Go's test corpus as Gno filetests

`testdata/` holds files lifted verbatim from Go's `$GOROOT/test` (the
release in `VERSION`), each followed by a comment tail the harness writes
(`gnovm/pkg/test/filetest.go`). Nothing above the tail is edited.

## Per-file contract

Evidence blocks (raw, auto-written), then exactly one `// Verdict:` line, last:

| block | meaning |
|---|---|
| `GnoOutput:` / `GoOutput:` | stdout of Gno and of `go run` (run files) |
| `Output:` / `GoOutput:` | the same pair for a native `.gno` filetest that opts in: adding `GoOutput:` makes the harness run Go too and demand the `Verdict:` line; `.gno` files without it are untouched |
| `GnoError:` | Gno's own errors, per line, wording kept |
| `GoBuildError:` | lines gc (`go build`) flagged, compile files; lines only, their wording follows the Go release |

| verdict | meaning |
|---|---|
| `Verdict: Clean <mode>` | both checkers reach the same verdict (`run`); wording differences stay in the evidence blocks |
| `Verdict: Unsupported <tag>: <reason>` | skipped, feature gap or gc-only test; tags in `corpus.go` (`unsupportedTags`) |
| `Verdict: KnownIssue <mode>: <reason>` | Gno is wrong here; the reason must cite the tracking issue (`#NNNN` or an issue URL) |
| `Verdict: KnownDivergence <mode>: <reason>` | Gno differs on purpose; the reason must cite the Gno decision it rests on (a `docs/` page, `#NNNN` or a URL); the Go spec merely permitting a difference is not one. A divergence nobody has decided is a `KnownIssue` |

Sync proposes a `Known*` verdict with a `TODO:` reason; verify refuses it until
a human replaces the reason, so a divergence is never blessed by accident. File
one issue per cause and list the corpus files it covers: the same `#NNNN` goes
into each of them, and the fix PR flips those files to `Clean` (verify fails
them as "stale verdict" until it does). A file whose cause has no issue yet
stays out of the tree.

Verify (`go test ./gnovm/pkg/gnolang -run TestFiles/gocorpus`) runs Gno only
and compares with the pinned Go side; it never shells out to `go`. Sync runs
`go run` / `go build` from the toolchain on PATH, so run it with
`GOTOOLCHAIN=$(cat VERSION)` to keep the goldens on one Go release;
`make -C gnovm _test.filetest` does not pin the toolchain unless `go.mod`
carries a `toolchain` line, so set it explicitly.

`UPSTREAM` names the Go release the files were lifted from; `UPSTREAM.sha256`
holds the sha256 of each file's upstream body, and `TestCorpusBodiesMatchUpstream`
fails any file whose body (golden region stripped) drifts from it. After adding
files, regenerate the manifest with `./upstream-manifest.sh <go source tree>`.

## Re-sync

    GOTOOLCHAIN=$(cat gnovm/tests/files/gocorpus/VERSION) \
      go test ./gnovm/pkg/gnolang -run TestFiles/gocorpus -short -update-golden-tests

Sync runs `go run` / `go build` for the Go side; the pinned toolchain keeps
`GoOutput` stable. `VERSION` is kept equal to the `go` line in the root
`go.mod`, so CI needs one toolchain; bump both together and re-sync.

Files with inline `// ERROR` markers (gc's errorcheck tests) are refused for
now; that mode lands with the error-accumulating preprocessor.
