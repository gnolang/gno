# gocorpus: Go's test corpus as Gno filetests

`testdata/` holds files lifted verbatim from Go's `$GOROOT/test` at the
release in `UPSTREAM`, each followed by a comment tail the harness writes
(`gnovm/pkg/test/filetest.go`). Nothing above the tail is edited. `VERSION`
is the Go toolchain the Go-side evidence was recorded with.

`testdata/gotypes/` is the type checker's own suite, lifted the same way
from `$GOROOT/src/internal/types/testdata` (`check`, `fixedbugs`, `spec`,
`examples`): one rule per file, markers in go/types' block form
(`/* ERROR "text" */`, `/* ERRORx "regex" */`, several per line), checked
by the same errorcheck mode; a file without markers must preprocess
cleanly (compile mode). `// -lang=` and `// -gotypesalias=` headers
configure go/types; those below Gno's go1.17 pin or disabling alias
nodes are `Unsupported`.

## Per-file contract

Evidence blocks (raw, auto-written), then exactly one `// Verdict:` line, last:

| block | meaning |
|---|---|
| `GnoOutput:` / `GoOutput:` | stdout of Gno and of `go run` (run files) |
| `GnoError:` | Gno's own errors, per line, wording kept |
| `GoTypeCheckError:` / `GoBuildError:` | lines the go/types guard / `go build` flagged; lines only, their wording follows the Go release |
| `GoTypesOnly:` | errorcheck markers only the guard caught: what Gno's checker still lacks |
| `GnoOverStrictError:` / `UncaughtError:` | Gno rejects a line gc did not mark / a checkable marker neither checker caught |
| `ParseBlocked:` | checkable markers in a file Gno could not parse: the checker never ran, so they count as neither caught nor missed |
| `GnoWordingDiffers:` | attributed lines whose Gno message matches none of gc's patterns: the line agreed, the wording did not; read these when a line match may be a coincidence |
| `GnoPreprocessError:` | Gno's errors on compile-only files |

| verdict | meaning |
|---|---|
| `Verdict: Clean <mode>` | both checkers reach the same verdict (run \| compile \| errorcheck); wording differences stay in the evidence blocks |
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

Files with inline `// ERROR "regex"` markers (gc's errorcheck tests) are
preprocessed once in accumulate mode (`PreprocessFilesCollect`): every Gno
error with its span, no execution. A marker is attributed to Gno when a Gno
error span covers its line, to the guard otherwise (`GoTypesOnly`). The
verdict is `Clean: errorcheck` unless Gno is over-strict or a checkable
marker is uncaught, in which case it is `KnownIssue`.
