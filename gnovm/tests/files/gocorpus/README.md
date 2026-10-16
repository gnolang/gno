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
| `Output:` / `GoOutput:` | the same pair for a native `.gno` filetest that opts in: adding `GoOutput:` makes the harness run Go too and demand the `Verdict:` line; `.gno` files without it are untouched |
| `GnoError:` | Gno's own errors, per line, wording kept (a runtime panic in run mode, preprocess errors in compile mode) |
| `GoBuildError:` | lines gc (`go build`) flagged, compile files; lines only, their wording follows the Go release |
| `ParseBlocked:` | checkable markers in a file Gno could not parse: the checker never ran, so they count as neither caught nor missed |
| `GnoWordingDiffers:` | attributed lines whose Gno message matches none of gc's patterns: the line agreed, the wording did not; read these when a line match may be a coincidence |

| verdict | meaning |
|---|---|
| `Verdict: Clean <mode>` | both checkers reach the same verdict (run \| compile \| errorcheck); wording differences stay in the evidence blocks |
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

Files with inline `// ERROR "regex"` markers (gc's errorcheck tests) are
preprocessed once in accumulate mode (`PreprocessFilesCollect`): every Gno
error with its span, no execution, no Go side. The markers are gc's verdict.
`Clean errorcheck` when a Gno error covers every checkable marker line and
none falls on an unmarked line; otherwise sync writes
`KnownIssue errorcheck: #6344 uncaught markers: lines …; over-strict: lines …`
directly, since every such gap is a preprocess conformance gap in that
umbrella issue's scope (a per-cause `#NNNN` written by hand survives re-sync).
An over-strict line whose Gno message is a feature gap (the `unsupportedTags`
needles: a stdlib symbol Gno lacks, a newer language feature) does not count;
the `GnoError:` block still shows it. An EOF error maps to the last source
line, where gc's marker sits. gc-only files (`-0`/`-m` diagnostics, `-+`/`-p=`,
`-lang` below go1.17, `GC_ERROR`-only, pragma placement, `//line` directives,
a byte order mark) are `Unsupported`.
