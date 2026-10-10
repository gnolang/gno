# gocorpus: Go's test corpus as Gno filetests

`testdata/` holds files lifted verbatim from Go's `$GOROOT/test` at the
release in `UPSTREAM`, each followed by a comment tail the harness writes
(`gnovm/pkg/test/filetest.go`). Nothing above the tail is edited. `VERSION`
is the Go toolchain the Go-side evidence was recorded with.

## Per-file contract

Evidence blocks (raw, auto-written), then exactly one `// Verdict:` line, last:

| block | meaning |
|---|---|
| `GnoOutput:` / `GoOutput:` | stdout of Gno and of `go run` (run files) |
| `GnoError:` | Gno's own errors, per line, wording kept |
| `GoTypeCheckError:` / `GoBuildError:` | lines the go/types guard / `go build` flagged; lines only, their wording follows the Go release |

| verdict | meaning |
|---|---|
| `Verdict: Clean <mode>` | both checkers reach the same verdict (`run`); wording differences stay in the evidence blocks |
| `Verdict: Unsupported <tag>: <reason>` | skipped, feature gap or gc-only test; tags in `corpus.go` (`unsupportedTags`) |
| `Verdict: KnownIssue <mode>: <reason>` | Gno is wrong here, tracked |
| `Verdict: KnownDivergence <mode>: <reason>` | Gno differs on purpose (a run output or a rule decided for Gno) |

Verify (`go test ./gnovm/pkg/gnolang -run TestFiles/gocorpus`) runs Gno only
and compares with the pinned Go side; it never shells out to `go`. One
dependency on the Go release remains: the go/types guard runs in-process, so
the *set of lines* in `GoTypeCheckError:` follows the Go that
built the test binary. Run with `GOTOOLCHAIN=$(cat VERSION)` to match the
goldens; `make -C gnovm _test.filetest` does not pin the toolchain unless
`go.mod` carries a `toolchain` line, so set it explicitly.

## Re-sync

    GOTOOLCHAIN=$(cat gnovm/tests/files/gocorpus/VERSION) \
      go test ./gnovm/pkg/gnolang -run TestFiles/gocorpus -short -update-golden-tests

Sync runs `go run` / `go build` for the Go side; the pinned toolchain keeps
`GoOutput` stable. Bump `VERSION` and re-sync together.

Files with inline `// ERROR` markers (gc's errorcheck tests) are refused for
now; that mode lands with the error-accumulating preprocessor.
