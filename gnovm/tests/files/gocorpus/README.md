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
nodes are `Unsupported`, a `-lang=go1.N` with N ≥ 17 is ignored (the
harness emulates no language version: a "requires go1.N" marker counts
as caught only when Gno rejects the line for its own reason, which
`GnoWordingDiffers:` then shows). A marker that is the first token on
its line is refused for these files: go/types attributes it to the
previous line's token, which line-based attribution cannot follow.

## Per-file contract

Evidence blocks (raw, auto-written), then exactly one `// Verdict:` line, last:

| block | meaning |
|---|---|
| `GnoOutput:` / `GoOutput:` | stdout of Gno and of `go run` (run files) |
| `Output:` / `GoOutput:` | the same pair for a native `.gno` filetest that opts in: adding `GoOutput:` makes the harness run Go too and demand the `Verdict:` line; `.gno` files without it are untouched |
| `GnoError:` | Gno's own errors, per line, wording kept (a runtime panic in run mode, preprocess errors in compile mode). Compile mode runs no `main`, but it does run package initialisers (`var` initialisers, `init`), which gc never runs for a `// compile` file: Gno has no init-free load, and an initialiser that panics or loops is a finding, not a mode bug |
| `GoBuildError:` | lines gc (`go build`) flagged, compile files; lines only, their wording follows the Go release |
| `ParseBlocked:` | checkable markers in a file Gno could not parse: the checker never ran, so they stay uncaught (the verdict is `KnownIssue`) and this block says why |
| `GnoWordingDiffers:` | attributed lines whose Gno message matches none of gc's patterns (go/types' `ERROR` is a substring, `ERRORx` and gc's `ERROR` regexps): the line agreed, the wording did not; read these when a line match may be a coincidence |

| verdict | meaning |
|---|---|
| `Verdict: Clean <mode>` | both checkers reach the same verdict (run \| compile \| errorcheck); wording differences stay in the evidence blocks. A compile file is Clean only when Gno accepts it: gc accepted it upstream, so a `GoBuildError:` here is a toolchain or rescue artefact and sync proposes a `Known*` verdict |
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

Verify also refuses a Clean or `Known*` file whose Gno side now fails with a
feature-gap message (`stale verdict`): only sync routes a file to
`Unsupported`, so a Gno regression never turns into a skip.

`UPSTREAM` names the Go release the files were lifted from; `UPSTREAM.sha256`
holds the sha256 of each file's upstream body, and `TestCorpusBodiesMatchUpstream`
fails any file whose body (golden region stripped) drifts from it. After adding
files, regenerate the manifest with `./upstream-manifest.sh <go source tree>`.
The region is the trailing comment blocks that open with a golden directive
(`GnoOutput:`, `GnoError:`, `GoOutput:`, `TypeCheckError:`, `Verdict:`); any
other trailing comment is body and the manifest guards it, so triage notes
go in the verdict's reason. Directive parsing is line-based: a column-0
`// Verdict:` or `// GnoOutput: x` line inside an upstream body (a raw string,
a comment) is read as a directive, verify fails the file for carrying two
verdicts, and a sync would drop the line, which the manifest then catches.

## Re-sync

    GOTOOLCHAIN=$(cat gnovm/tests/files/gocorpus/VERSION) \
      go test ./gnovm/pkg/gnolang -run TestFiles/gocorpus -short -update-golden-tests

Sync runs `go run` / `go build` for the Go side; the pinned toolchain keeps
`GoOutput` stable. `VERSION` is kept equal to the `go` line in the root
`go.mod`, so CI needs one toolchain; bump both together and re-sync.

Files with inline `// ERROR "regex"` markers (gc's errorcheck tests) are
preprocessed once in accumulate mode (`PreprocessFilesCollect`): every Gno
error with its span, no execution, no Go side. The markers are gc's verdict.
`Clean errorcheck` when every checkable marker line has a Gno error starting
on it and no Gno error starts on an unmarked line; otherwise sync writes
`KnownIssue errorcheck: #6344 uncaught markers: lines …; over-strict: lines …`
directly, since every such gap is a preprocess conformance gap in that
umbrella issue's scope (a per-cause `#NNNN` written by hand survives re-sync).
An over-strict line whose Gno message is a feature gap (the `unsupportedTags`
needles: a stdlib symbol Gno lacks, a newer language feature) does not count;
the `GnoError:` block still shows it. An internal error (`should not happen`
and kin) covers no marker: it is a VM bug, listed as `internal errors: lines …`.
Every errorcheck file is also run through the on-chain path (plain
`Preprocess`, first error): it must reject the file iff accumulate mode did,
with its first error on a reported line, else the verdict says how the
`on-chain path` disagrees. Coverage is by line, not wording: a marker counts
as covered when a Gno checking error starts on its line (or ends on it, for
a function that does not terminate: gc's marker sits on the closing brace);
one error credits no other line, however many it spans. The wording stays
in the evidence. An EOF error maps to the last source line, where gc's
marker sits. Internal errors are pinned without their preprocess stack dump. gc-only files (`-0`/`-m` diagnostics, `-+`/`-p=`,
`-lang` below go1.17, `GC_ERROR`-only, pragma placement, `//line` directives,
a byte order mark) are `Unsupported`.
