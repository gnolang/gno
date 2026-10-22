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
| `Output:` / `GoOutput:` | the same pair for a native `.gno` filetest that opts in: adding `GoOutput:` makes the harness run Go too and demand the `Verdict:` line; `.gno` files without it are untouched |
| `GnoError:` | Gno's own errors, per line, wording kept (a runtime panic in run mode, preprocess errors in compile mode). Compile mode runs no `main`, but it does run package initialisers (`var` initialisers, `init`), which gc never runs for a `// compile` file: Gno has no init-free load, and an initialiser that panics or loops is a finding, not a mode bug |
| `GoBuildError:` | lines gc (`go build`) flagged, compile files; lines only, their wording follows the Go release |
| `GnoTrim:` / `GnoTrimmed:` | hand-written: body lines blanked before checking (see "Trimming an incidental construct"); auto-written: the trimmed lines that carried a marker, dropped from the expected set |
| `GnoTypeDiffers:` | the declared names (variables, constants, functions, parameters, results) whose static type Gno's preprocess records differs from go/types' on the same source, one `line N: x gno int, go int32` per name, every mode (see "The types check") |

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
(`GnoOutput:`, `GnoError:`, `GoOutput:`, `GnoTypeDiffers:`, `TypeCheckError:`, `Verdict:`); any
other trailing comment is body and the manifest guards it, so triage notes
go in the verdict's reason. Directive parsing is line-based: a column-0
`// Verdict:` or `// GnoOutput: x` line inside an upstream body (a raw string,
a comment) is read as a directive, verify fails the file for carrying two
verdicts, and a sync would drop the line, which the manifest then catches.

## The types check

Acceptance, error lines and output never show the static type Gno gives
a declaration, so every mode also compares it with go/types' (`typediff.go`):
both are spelt one way (rune/byte by kind, `untyped int|rune|float|string|bool`,
a declared type by its bare name, functions by signature, structs and
interfaces by their literal) and each mismatch is a `GnoTypeDiffers:` line.
A file with mismatches is `KnownIssue <mode>: #6344 types differ: lines …`;
a reason written by hand keeps its own `[#NNNN ]types differ: lines …`
group, reconciled like the errorcheck gap groups below (a stale block or
list fails verify). An errorcheck file is checked for the declarations
Gno accepted; a name go/types could not type is skipped.

## Trimming an incidental construct

A file skipped for one construct Gno lacks, while the rest of it is
checkable, can carry a hand-written tail line

    // GnoTrim: lines 12, 92-93

The harness blanks the listed body lines (1-based, ranges inclusive)
before parsing and classifying, so the Gno side never sees them and
the rest of the file gets a real verdict. Line numbers do not move:
markers keep their lines. A trimmed line that carried a marker is
dropped from the expected set and listed as `// GnoTrimmed: lines …`.
Sync keeps the directive as written and rewrites the evidence; verify
refuses a directive that names a line outside the body or a line
without a trimmable construct, a trim that blanks every marker line,
and a trim on a run file (the trimmed program is not the upstream one).

Use it for an incidental `import "unsafe"` plus its few (≤3) uses, a
stray complex, imaginary or `uintptr` line; those four constructs are
the only ones it may remove. Never trim a line to hide a checkable one,
and never trim a declaration another line uses: a `var z complex128`
whose `z` a marker line reads would turn that marker into a
`name not defined` credit, and an unmarked use into an over-strict
artefact; such a file stays `Unsupported` with no directive. If the
trimmed file still hits a gap, sync keeps the directive and writes the
new `Unsupported` tag, which is the honest bucket for it.

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
umbrella issue's scope. A reason written by hand is `;`-separated parts; a
gap group is `[#NNNN ]<kind>: lines a, b, c-d[ (note)]` with kind `uncaught
markers`, `over-strict`, `internal errors` or `types differ`, owned by the
last `#NNNN` in force; other parts are prose, kept as written. Every run reconciles the
groups with the gaps: a group keeps the lines still gaps of its kind (its
note with it, an emptied group is dropped), unowned gap lines go to `#6344`,
and verify fails on a listed line no longer a gap or a gap listed nowhere.
An over-strict line whose Gno message is a feature gap (the `unsupportedTags`
needles: a stdlib symbol Gno lacks, a newer language feature) does not count;
the `GnoError:` block still shows it. A bodiless function declaration (gc
accepts it, the body being assembly) is such a line-level gap in errorcheck
mode only: Gno reports it, keeps the function declared, and the calls on
marker lines are still checked; a `// compile` file with one stays
`Unsupported bodiless-func`. An internal error (`should not happen`
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
`-lang` below go1.17, `GC_ERROR`-only, pragma placement, `//line` and `/*line*/` directives,
a byte order mark) are `Unsupported`.
