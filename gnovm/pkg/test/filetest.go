package test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/debug"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	teststdlibs "github.com/gnolang/gno/gnovm/tests/stdlibs"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/pmezard/go-difflib/difflib"
	"go.uber.org/multierr"
)

// RunFiletest executes a gnovm internal filetest in test/files.
// If opts.Sync is enabled, and the filetest's golden output has changed,
// the first string is set to the new generated content of the file.
// Before the filetest is run it will be type-checked.
//
// A file declaring a top-level `// Unsupported: <reason>` directive is
// short-circuited before any execution: the returned error is a
// [*SkipError] whose Reason carries the directive's payload. The
// caller (e.g. the TestFiles walker) is expected to detect this with
// [errors.As] and convert it to a `t.Skip(reason)`.
func (opts *TestOptions) RunFiletest(fname string, source []byte, tgs gno.Store) (string, types.Gas, gno.StorageDiffs, error) {
	opts.outWriter.w = opts.Output
	opts.outWriter.errW = opts.Error
	tcheck := true // Go type-check filetests in test/files.
	return opts.runFiletest(fname, source, tgs, tcheck)
}

// SkipError is returned by RunFiletest when the source file ends in a
// `// Verdict: Unsupported <tag>: <reason>` line. The walker is
// expected to detect this with [errors.As] and call t.Skip(Reason).
//
// Replaces conformance's external skiplist.yaml + compat.go for the
// in-repo workflow: each file declares its own skip reason inline, so
// adding a corpus file that exercises an unsupported Gno feature is
// a single edit — no cross-file coordination.
type SkipError struct {
	Reason string
}

func (e *SkipError) Error() string { return "skipped: " + e.Reason }

// tcheck: only set to false pkg/test.Test(), since `gno test`
// (cmd/gno/test.go) already type-checked the whole package.
// Go type-checking in filetests is only available for gnovm internal filetests
// in test/files.
func (opts *TestOptions) runFiletest(fname string, source []byte, tgs gno.Store, tcheck bool) (newContent string, gas types.Gas, diffs gno.StorageDiffs, retErr error) {
	rawSource := source // as on disk, golden region included
	dirs, err := ParseDirectives(bytes.NewReader(source))
	if err != nil {
		return "", 0, nil, fmt.Errorf("error parsing directives: %w", err)
	}
	isGoFile := strings.HasSuffix(fname, ".go")

	// Verify: a second `// Verdict:` line must not hide behind an
	// Unsupported first one, so the count runs before the short-circuit.
	if isGoFile && !opts.Sync {
		if n := countDirectives(dirs, DirectiveVerdict); n > 1 {
			return "", 0, diffs, fmt.Errorf("expected exactly one `// Verdict:` line, found %d; re-run with --update-golden-tests", n)
		}
	}
	// Strip the trailing golden region (`// GnoError:` /
	// `// GoBuildError:` blocks, verdict lines) from the source
	// fed to Gno. Leaving it in would extend the file and shift the
	// line number of any EOF-positioned error (e.g. `expected ')',
	// found 'EOF'`), making the golden unstable across runs. Trailing
	// comments don't affect code line numbers, so this is safe. (Only
	// meaningful for errorcheck .go files; a no-op otherwise.)
	//
	// Normalize to exactly one trailing newline afterwards: a file
	// without a golden (first sync) and the same file with the golden
	// stripped (later verify) must be byte-identical, or the synthetic
	// `func main(){}` the rescue appends lands on a different line.
	// A `// GnoTrim:` directive (trim.go) is checked here, before any
	// skip, so a bad one fails verify and sync alike; the trimmed body
	// replaces source once originalSource has been captured below.
	var trim *gnoTrim
	var trimmed []byte
	if isGoFile {
		source = []byte(stripTrailingGoldenRegion(string(source)))
		source = append(bytes.TrimRight(source, "\n"), '\n')
		if d := dirs.First(DirectiveGnoTrim); d != nil {
			t, err := parseGnoTrim(d.Content)
			if err != nil {
				return "", 0, diffs, err
			}
			if trimmed, err = t.apply(source); err != nil {
				return "", 0, diffs, err
			}
			trim = &t
		}
	}
	// Unsupported short-circuit: declared via the file's
	// `// Verdict: Unsupported <tag>: <reason>` line. Surfaces as a SkipError
	// the walker turns into t.Skip. Checked before any dispatch /
	// rescue / machine construction so unsupported files cost nothing.
	// Corpus vocabulary only: in a .gno file the line stays a comment.
	// Sync re-evaluates a trimmed file: the trim may have removed the gap.
	if v, ok := verdictIn(dirs); isGoFile && ok && v.Kind == VerdictUnsupported && (trim == nil || !opts.Sync) {
		if opts.Sync {
			// Normalise: tag the reason, drop goldens from before the
			// file was marked (they are never read again).
			if nc := writeUnsupportedDirective(rawSource, trim, v.Mode+": "+v.Reason); nc != string(rawSource) {
				return nc, 0, diffs, nil
			}
		}
		return "", 0, diffs, &SkipError{Reason: v.Mode + ": " + v.Reason}
	}

	// .go filetests under tests/files/gocorpus/testdata/ are files lifted
	// from Go's test corpus; the mode comes from the source (gc's
	// directive word and inline markers). Run mode compares both
	// runtimes' output, see finalizeGoRunDivergence.
	var isGoRunMode bool
	var goStdout string
	// originalSource is the file as-authored, before any in-memory
	// PKGPATH/synthetic-main rescue. Golden writes serialize against
	// THIS, never the rescued source — the rescue transforms must
	// not be persisted (gocorpus files stay upstream-verbatim).
	originalSource := source
	if trimmed != nil {
		source = trimmed
	}
	// Capture the golden region now, before the compile-mode rescue
	// re-parses dirs from the rescued source.
	origDirs := dirs
	// prependedLines is how many lines the rescue added at the top (0 or
	// 1), used to map Gno's error lines back to original-file coordinates.
	prependedLines := 0
	prependRescue := func() error {
		source = PrependPkgPathIfNeeded(source)
		if bytes.HasPrefix(source, []byte("// PKGPATH:")) &&
			!bytes.HasPrefix(originalSource, []byte("// PKGPATH:")) {
			prependedLines = 1
		}
		// Re-parse so the prepended `// PKGPATH:` line reaches the
		// directive parser.
		dirs, err = ParseDirectives(bytes.NewReader(source))
		if err != nil {
			return fmt.Errorf("error re-parsing directives after pkgpath rescue: %w", err)
		}
		return nil
	}
	compileMode := false
	var errorcheckMarkers []InlineError
	// .gno files opt INTO Gno-vs-Go comparison by declaring
	// `// GoOutput:` (or a KnownDivergence verdict). Without it, .gno
	// files keep their pure-Gno behavior — the 1600+ existing files
	// are untouched.
	hasGoOptIn := dirs.First(DirectiveGoOutput) != nil ||
		verdictKind(dirs) == VerdictKnownDivergence
	if isGoFile {
		hasErrorDir := dirs.First(DirectiveError) != nil
		hasTypeCheckErrorDir := dirs.First(DirectiveTypeCheckError) != nil
		switch {
		case HasInlineErrorMarkers(source) && !hasErrorDir:
			// `// errorcheck`: gc's inline markers name the lines it
			// rejects; Gno preprocesses once in accumulate mode.
			errorcheckMarkers = ParseInlineErrors(source)
			if len(errorcheckMarkers) == 0 {
				// A marker on a full-comment line (a pragma, as in
				// fixedbugs/issue18882.go) is not checkable.
				reason := "errorcheck marker on a non-code line (e.g. pragma comment); not checkable by the harness"
				if opts.Sync {
					return writeUnsupportedDirective(originalSource, trim, reason), 0, diffs, nil
				}
				return "", 0, diffs, &SkipError{Reason: reason}
			}
			if err := prependRescue(); err != nil {
				return "", 0, diffs, err
			}
		case CorpusDirective(source) == "compile" && !hasErrorDir && !hasTypeCheckErrorDir:
			// `// compile`: gc compiles but never runs it. Gno's preprocess
			// (and package init) only, no main; go/types' verdict is not
			// part of the ledger. Takes precedence over run mode even when
			// the file is runnable.
			compileMode = true
			if err := prependRescue(); err != nil {
				return "", 0, diffs, err
			}
		case !IsRunnable(source) && !hasErrorDir && !hasTypeCheckErrorDir:
			// No main to run and no markers to check: compile-only,
			// whatever the header says.
			compileMode = true
			if err := prependRescue(); err != nil {
				return "", 0, diffs, err
			}
		default:
			// Runnable .go corpus file: symmetric Gno-vs-Go.
			isGoRunMode = true
		}
		if isGoRunMode && trim != nil {
			return "", 0, diffs, fmt.Errorf("`// GnoTrim:` is for errorcheck and compile files; a trimmed run file is not the upstream program")
		}
	} else if hasGoOptIn && IsRunnable(source) {
		// .gno run-mode opt-in: harness ALSO runs Go and compares.
		// The existing `// Output:` directive remains Gno's pinned
		// golden (handled by the main match loop); the new triple
		// directives drive the Gno-vs-Go finalize at the end.
		isGoRunMode = true
	}
	// One verdict line per Go-compared file (.go corpus files in run or
	// compile mode, opted-in .gno files); sync writes it, verify demands it.
	// origDirs: the compile-mode rescue re-parses dirs from the source
	// whose golden tail was stripped above, so they no longer see it.
	if (isGoFile || isGoRunMode || compileMode) && !opts.Sync {
		if n := countDirectives(origDirs, DirectiveVerdict); n != 1 {
			return "", 0, diffs, fmt.Errorf("expected exactly one `// Verdict:` line, found %d; re-run with --update-golden-tests", n)
		}
		v, ok := verdictIn(origDirs)
		if !ok {
			return "", 0, diffs, fmt.Errorf("malformed `// Verdict:` line: want `<Kind> <mode>[: <reason>]`")
		}
		if err := v.check(); err != nil {
			return "", 0, diffs, err
		}
		// The mode names the file's mode; an Unsupported verdict carries a tag there.
		mode := "run"
		switch {
		case len(errorcheckMarkers) > 0:
			mode = "errorcheck"
		case compileMode:
			mode = "compile"
		}
		if v.Kind != VerdictUnsupported && v.Mode != mode {
			return "", 0, diffs, fmt.Errorf("`// Verdict: %s` names mode %q, but this is a %s file", v.Kind, v.Mode, mode)
		}
	}
	// The Go side is a golden: `go run` only under sync (with the toolchain
	// pinned in gocorpus/VERSION); verify compares against `// GoOutput:`
	// offline, so the host's Go version cannot move the result.
	if isGoRunMode {
		if opts.Sync {
			out, runErr := runGoToolchain(source, false)
			if runErr != nil {
				return "", 0, diffs, fmt.Errorf("filetest %s: cannot run via Go toolchain "+"(install go, or mark file `// Unsupported:`): %w", fname, runErr)
			}
			goStdout = out
		} else {
			d := dirs.First(DirectiveGoOutput)
			if d == nil {
				return "", 0, diffs, fmt.Errorf("filetest %s: missing `// GoOutput:` golden; "+"re-run with --update-golden-tests (needs the Go toolchain)", fname)
			}
			goStdout = d.Content
		}
	}

	// Sanity check: type-check directives are not available
	// with `gno test` of user packages.
	if !tcheck && dirs.FirstDefault(DirectiveTypeCheckError, "") != "" {
		panic("type-check error directive is only available for gnovm internal test files")
	}

	// Initialize Machine.Context and Machine.Alloc according to the input directives.
	pkgPath := dirs.FirstDefault(DirectivePkgPath, "main")
	coins, err := std.ParseCoins(dirs.FirstDefault(DirectiveSend, ""))
	if err != nil {
		return "", 0, nil, err
	}
	ctx := Context("", pkgPath, coins)
	maxAllocRaw := dirs.FirstDefault(DirectiveMaxAlloc, "0")
	maxAlloc, err := strconv.ParseInt(maxAllocRaw, 10, 64)
	if err != nil {
		return "", 0, nil, fmt.Errorf("could not parse MAXALLOC directive: %w", err)
	}
	// Force a non-nil allocator so interrealm v2 Phase 2 PkgID stamping
	// fires under filetests the same way it does in production. MAXALLOC
	// directive remains the override for tests that want a specific cap.
	if maxAlloc == 0 {
		maxAlloc = math.MaxInt64
	}

	var opslog io.Writer
	if dirs.First(DirectiveRealm) != nil {
		opslog = new(bytes.Buffer)
	}
	gasMeter := store.NewInfiniteGasMeter()
	// Create machine for execution and run test
	tcw := opts.BaseStore.CacheWrap()
	m := gno.NewMachineWithOptions(gno.MachineOptions{
		Output:        &opts.outWriter,
		Store:         tgs.BeginTransaction(tcw, tcw, nil, gasMeter),
		Context:       ctx,
		MaxAllocBytes: maxAlloc,
		GasMeter:      gasMeter,
		Debug:         opts.Debug,
		ReviveEnabled: true,
	})
	defer m.Release()

	// RUN THE FILETEST /////////////////////////////////////
	// Errorcheck and compile files are not executed (gc doesn't run
	// them; executing can hang on infinite loops Gno fails to reject) —
	// preprocess only.
	var result runResult
	if len(errorcheckMarkers) > 0 {
		// The on-chain path (first error, Preprocess) runs on its own
		// store wrap so the two passes do not share a package.
		cw2 := opts.BaseStore.CacheWrap()
		m2 := gno.NewMachineWithOptions(gno.MachineOptions{
			Output:        &opts.outWriter,
			Store:         tgs.BeginTransaction(cw2, cw2, nil, gasMeter),
			Context:       ctx,
			MaxAllocBytes: maxAlloc,
			GasMeter:      gasMeter,
			ReviveEnabled: true,
		})
		defer m2.Release()
		result = opts.runErrorcheck(m, m2, pkgPath, fname, source)
	} else {
		result = opts.runTest(m, pkgPath, fname, source, opslog, tcheck, compileMode)
	}

	// Collect storage diffs after test execution.
	diffs = m.Store.RealmStorageDiffs()

	// Feature gap: a .go corpus file whose Gno error is a "feature not
	// implemented" message (channels, goroutines, generics, imaginary
	// literals, dot imports, …) or an unsupported stdlib import — Gno
	// can't process it at all, so route to `// Unsupported:` (skip)
	// rather than letting it fail or land in KnownIssue. Detected from
	// Gno's ACTUAL error, so prose like "go to" in a comment can't trip
	// it. Auto-written under sync; thereafter skipped pre-dispatch. Verify
	// never skips here: a file that reaches this point carries a Clean or
	// Known* verdict, so a gap message now is a regression, not a gap.
	if isGoFile {
		reason := fileGapReason(result, len(errorcheckMarkers) > 0)
		// Gno's preprocess may surface an unrelated error before the
		// "unknown import path" (alias2.go imports `reflect` and gets a
		// redeclaration error first), so also look at the import list:
		// a stdlib Gno lacks is a feature gap, not a divergence to pin.
		// A file pinning its behavior with a native `// Error:` /
		// `// TypeCheckError:` directive is left alone.
		nativePin := dirs.First(DirectiveError) != nil || dirs.First(DirectiveTypeCheckError) != nil
		if reason == "" && !nativePin {
			if imp := MissingStdlibImport(source, opts.RootDir); imp != "" {
				reason = "unknown import path " + imp
			}
		}
		// Generics are a feature gap, but Gno reports them inconsistently
		// (preprocess "type parameter", or a downstream runtime "name T
		// not declared"), so detect from the AST instead of the message —
		// otherwise a generic file that fails at runtime is misrouted to
		// // KnownIssue: instead of // Unsupported:.
		if reason == "" && result.Error != "" && UsesGenerics(source) {
			reason = "generics not supported in Gno"
		}
		// Run-mode programs whose output is non-reproducible (panic
		// addresses, goroutine dumps, or the harness temp path the
		// program prints when probing its own file/line) can't have a
		// stable golden either — route them to `// Unsupported:` too.
		if reason == "" && isGoRunMode {
			if r := NondeterministicOutput(goStdout); r != "" {
				reason = r
			} else if r := NondeterministicOutput(result.Output); r != "" {
				reason = r
			} else if r := NondeterministicOutput(result.Error); r != "" {
				reason = r
			}
		}
		if reason != "" {
			if opts.Sync {
				return writeUnsupportedDirective(originalSource, trim, reason), m.GasMeter.GasConsumed(), diffs, nil
			}
			return "", m.GasMeter.GasConsumed(), diffs, fmt.Errorf(
				"stale verdict: Gno now hits a feature gap (%s) — re-run with --update-golden-tests to mark the file `// Verdict: Unsupported`:\n%s",
				reason, result.Error)
		}
	}

	// Errorcheck mode: gc's inline `// ERROR "regex"` markers are the
	// reference. One accumulate-mode pass gives every Gno error with its
	// span (`GnoError:`). Clean when every checkable marker line has a
	// Gno error starting on it (see creditLines) and no Gno error starts
	// on an unmarked line; otherwise a KnownIssue that cites the
	// preprocess-conformance umbrella issue, since every such gap is in
	// its scope (a per-cause issue written by hand survives re-sync).
	if len(errorcheckMarkers) > 0 {
		gas := m.GasMeter.GasConsumed()
		if reason := gcOnlyErrorcheckReason(originalSource, errorcheckMarkers); reason != "" {
			if opts.Sync {
				return writeUnsupportedDirective(originalSource, trim, reason), gas, diffs, nil
			}
			return "", gas, diffs, &SkipError{Reason: reason}
		}
		bodyLines := strings.Count(string(originalSource), "\n")
		cov := attributeErrorcheck(errorcheckMarkers, result.gnoErrors, prependedLines, bodyLines)
		// The on-chain path must reject the file iff accumulate mode did,
		// with its first error on a reported line; a parse-blocked file
		// has no on-chain checking error to compare.
		onChain := ""
		if !result.parseBlocked {
			lo, _ := errorSpan(result.onChainError)
			lo = min(lo-prependedLines, bodyLines)
			_, reported := cov.gnoErr[lo]
			switch {
			case result.onChainError == "" && len(result.gnoErrors) > 0:
				onChain = "accepts the file"
			case result.onChainError != "" && len(result.gnoErrors) == 0:
				onChain = "rejects the file: " + CleanErrorMessage(result.onChainError)
			case result.onChainError != "" && lo > 0 && !reported:
				onChain = fmt.Sprintf("first error on line %d, not reported by accumulate mode", lo)
			}
		}
		uncaught, overStrict, internal := cov.uncaught, cov.overStrict, cov.internal
		v := verdict{Kind: VerdictClean, Mode: "errorcheck"}
		if len(uncaught) > 0 || len(overStrict) > 0 || len(internal) > 0 || onChain != "" {
			v = verdict{Kind: VerdictKnownIssue, Mode: "errorcheck", Reason: errorcheckGapReason(uncaught, overStrict, internal, onChain)}
			// A reason written by hand (a per-cause issue) survives re-sync;
			// the sync-written umbrella form is recomputed every time, and
			// verify accepts its legacy `#6344; …` spelling until the next sync.
			if prev, ok := verdictIn(origDirs); ok && prev.check() == nil {
				switch {
				case prev.Kind == VerdictKnownDivergence:
					// A divergence is a decision: it stands while the sides
					// still disagree, whatever the gap list says.
					v = prev
				case prev.Kind != VerdictKnownIssue:
				case !isSyncWrittenReason(prev.Reason):
					v.Reason = prev.Reason
				case !opts.Sync && legacySyncReason(prev.Reason) == v.Reason:
					v.Reason = prev.Reason
				}
			}
		}
		sections := append(trimSections(trim, originalSource),
			goldenSection{name: DirectiveGnoError, block: FormatGnoErrorBlock(cov.gnoErr)},
			goldenSection{name: DirectiveVerdict, block: v.String(), inline: true},
		)
		newContent, err := opts.resolveStaticGolden(rawSource, originalSource, origDirs, sections)
		return newContent, gas, diffs, err
	}

	// Compile mode: Gno's preprocess against gc. Two evidence blocks
	// (GnoError, and GoBuildError as line numbers) and one verdict:
	// `Clean compile` when Gno accepts as gc did upstream, a Known*
	// verdict whenever Gno rejects (compileVerdict). The verdict's reason
	// obeys the same rules as run mode (verdict.check: no TODO, a
	// KnownIssue cites its issue).
	if compileMode {
		gas := m.GasMeter.GasConsumed()
		gnoErr := PerLineErrors(result.Error, prependedLines)
		if len(gnoErr) == 0 {
			// Gno accepts, as gc does upstream: no `go build` needed.
			sections := append(trimSections(trim, originalSource),
				goldenSection{name: DirectiveGnoError}, goldenSection{name: DirectiveGoBuildError},
				goldenSection{name: DirectiveVerdict, block: verdict{Kind: VerdictClean, Mode: "compile"}.String(), inline: true},
			)
			newContent, err := opts.resolveStaticGolden(rawSource, originalSource, origDirs, sections)
			return newContent, gas, diffs, err
		}
		// gc is the reference: `// compile` upstream means gc accepts the
		// file. Whenever Gno rejects, ask `go build` (sync; verify reuses
		// the pinned lines, and an absent block means gc accepted).
		var goBuildErr map[int]string
		if opts.Sync {
			buildOut, err := runGoToolchain(source, true)
			if err != nil {
				return "", gas, diffs, fmt.Errorf("filetest %s: cannot build via Go toolchain "+"(install go, or mark file `// Unsupported:`): %w", fname, err)
			}
			goBuildErr = PerLineErrors(buildOut, prependedLines)
		} else if d := origDirs.First(DirectiveGoBuildError); d != nil {
			goBuildErr = ParseLineBlock(d.Content)
		}
		prev, hasPrev := verdictIn(origDirs)
		v := compileVerdict(goBuildErr, prev, hasPrev)
		sections := append(trimSections(trim, originalSource),
			goldenSection{name: DirectiveGnoError, block: FormatGnoErrorBlock(gnoErr)},
			goldenSection{name: DirectiveGoBuildError, block: FormatLineOnlyBlock(goBuildErr)},
			goldenSection{name: DirectiveVerdict, block: v.String(), inline: true},
		)
		newContent, err := opts.resolveStaticGolden(rawSource, originalSource, origDirs, sections)
		return newContent, gas, diffs, err
	}

	// updated tells whether the directives have been mutated and the
	// regenerated filetest should be returned (only true under
	// opts.Sync). Declared here so the errorcheck short-circuit below
	// can also set it for `// GnoError:` golden refresh.
	updated := false

	// returnErr is used as the return value, and may be a MultiError if
	// multiple mismatches occurred. `updated` is declared above the
	// errorcheck short-circuit so both code paths can flip it.
	var returnErr error
	// `match` verifies the content against dir.Content; if different,
	// either updates dir.Content (for opts.Sync) or appends a new returnErr.
	match := func(dir *Directive, actual string) {
		content := dir.Content
		actual = strings.TrimRight(actual, "\n")
		content = strings.TrimRight(content, "\n")
		if content != actual {
			if opts.Sync {
				dir.Content = actual
				updated = true
			} else {
				if dir.Name == DirectiveError {
					returnErr = multierr.Append(
						returnErr,
						fmt.Errorf("%s diff:\n%s\nstacktrace:\n%s\nstack:\n%v",
							dir.Name, unifiedDiff(content, actual),
							result.GnoStacktrace, string(result.GoPanicStack)),
					)
				} else {
					returnErr = multierr.Append(
						returnErr,
						fmt.Errorf("%s diff:\n%s", dir.Name, unifiedDiff(content, actual)),
					)
				}
			}
		}
	}

	// Ensure needed the directives are present.
	// Run-mode .go files skip the unexpected-output/unexpected-panic
	// shortcuts: their verdict is determined by the explicit symmetric
	// Gno-vs-Go check at the end (finalizeGoRunDivergence call below).
	if result.Error != "" && !isGoRunMode {
		// Ensure this error was supposed to happen.
		errDirective := dirs.First(DirectiveError)
		if errDirective == nil {
			if opts.Sync {
				dirs = append(dirs, Directive{
					Name:    DirectiveError,
					Content: "",
				})
			} else {
				return "", m.GasMeter.GasConsumed(), diffs, fmt.Errorf("unexpected panic: %s\noutput:\n%s\nstacktrace:\n%s\nstack:\n%v", result.Error, result.Output, result.GnoStacktrace, string(result.GoPanicStack))
			}
		}
	} else if result.Output != "" && !isGoRunMode {
		outputDirective := dirs.First(DirectiveOutput)
		if outputDirective == nil {
			if opts.Sync {
				dirs = append(dirs, Directive{
					Name:    DirectiveOutput,
					Content: "",
				})
			} else {
				return "", m.GasMeter.GasConsumed(), diffs, fmt.Errorf("unexpected output:\n%s", result.Output)
			}
		}
	} else if !isGoRunMode {
		err = m.CheckEmpty()
		if err != nil {
			return "", m.GasMeter.GasConsumed(), diffs, fmt.Errorf("machine not empty after main: %w", err)
		}
		if gno.HasDebugErrors() {
			return "", m.GasMeter.GasConsumed(), diffs, fmt.Errorf("got unexpected debug error(s): %v", gno.GetDebugErrors())
		}
	}

	// Set to true if there was a go-typecheck directive..
	var hasTypeCheckErrorDirective bool

	// Check through each directive and verify it against the values from the test.
	for idx := range dirs {
		dir := &dirs[idx]
		switch dir.Name {
		case DirectiveOutput:
			match(dir, trimTrailingSpaces(result.Output))
		case DirectiveError:
			match(dir, result.Error)
		case DirectiveRealm:
			res := opslog.(*bytes.Buffer).String()
			match(dir, res)
		case DirectiveEvents:
			events := m.Context.(*teststdlibs.TestExecContext).EventLogger.Events()
			evtjson, err := json.MarshalIndent(events, "", "  ")
			if err != nil {
				panic(err)
			}
			evtstr := string(evtjson)
			match(dir, evtstr)
		case DirectivePreprocessed:
			pn := m.Store.GetBlockNodeSafe(gno.PackageNodeLocation(pkgPath))
			if pn == nil {
				return "", m.GasMeter.GasConsumed(), diffs, fmt.Errorf("package %q not preprocessed: %s", pkgPath, result.Error)
			}
			pre := pn.(*gno.PackageNode).FileSet.Files[0].String()
			match(dir, pre)
		case DirectiveStacktrace:
			match(dir, result.GnoStacktrace)
		case DirectiveGas:
			match(dir, strconv.FormatInt(m.GasMeter.GasConsumed(), 10))
		case DirectiveStorage:
			rlmDiff := realmDiffsString(m.Store.RealmStorageDiffs())
			match(dir, rlmDiff)
		case DirectiveTypes:
			match(dir, packageTypesString(m, pkgPath))
		case DirectiveTypeCheckError:
			hasTypeCheckErrorDirective = true
			if isGoFile {
				// Corpus files compare Gno with gc only; the native
				// go/types block is not part of their ledger.
				continue
			}
			match(dir, result.TypeCheckError)
			// DirectiveGnoOutput / DirectiveGnoError / DirectiveGoOutput are
			// the run-mode golden region; finalizeGoRunDivergence is their
			// sole authority (presence + content + divergence blessing).
		}
	}

	if !hasTypeCheckErrorDirective && result.TypeCheckError != "" && !isGoFile {
		dir := Directive{
			Name:    DirectiveTypeCheckError,
			Content: "",
		}
		match(&dir, result.TypeCheckError)
		dirs = append(dirs, dir)
	}

	// Symmetric Gno-vs-Go finalize for run-mode .go files or opted-in
	// .gno files. For .gno, Gno's pinned golden is the existing
	// `// Output:` directive (already match-checked above); for .go,
	// it's `// GnoOutput:` (new). Auto-append uses the appropriate
	// directive name per extension.
	if isGoRunMode {
		var newDirs Directives
		newDirs, returnErr = finalizeGoRunDivergence(dirs, result.Output, result.Error, goStdout, returnErr, opts.Sync, isGoFile)
		if newDirs != nil {
			dirs = newDirs
			updated = true
		}
	}

	if updated { // only true if sync == true
		return dirs.FileTest(), m.GasMeter.GasConsumed(), diffs, returnErr
	}

	return "", m.GasMeter.GasConsumed(), diffs, returnErr
}

// compileVerdict decides the verdict of a compile-mode file Gno rejects.
// gc accepted the file upstream (`// compile`), so a gc rejection here is
// a toolchain or rescue artefact, not agreement: the file is never Clean.
// A cited human Known* verdict is kept across re-syncs.
func compileVerdict(goBuildErr map[int]string, prev verdict, hasPrev bool) verdict {
	v := verdict{Kind: VerdictKnownIssue, Mode: "compile", Reason: "TODO: #NNNN explain the Gno bug (Gno rejects code gc accepts)"}
	if len(goBuildErr) != 0 {
		v = verdict{Kind: VerdictKnownDivergence, Mode: "compile", Reason: "TODO: cite why gc rejects this `// compile` file here (toolchain flag or rescue artefact); Gno rejects it too"}
	}
	if hasPrev && prev.Kind != VerdictClean && prev.Kind != VerdictUnsupported && prev.check() == nil {
		v = prev
	}
	return v
}

// goldenSection is one named per-line golden block (`// GnoError:` or
// `// GoBuildError:`) to (re)write at the bottom of a compile-mode file.
type goldenSection struct {
	name   string // directive name, e.g. DirectiveGnoError
	block  string // FormatGnoErrorBlock output (may be empty → omitted)
	inline bool   // single-line `// Name: block` (verdicts)
}

// resolveStaticGolden compares the wanted golden sections with the ones
// on disk. Verify fails on a missing, stale or differing block and on a
// stray block from an older ledger format; sync rewrites the region.
func (opts *TestOptions) resolveStaticGolden(rawSource, originalSource []byte, dirs Directives, sections []goldenSection) (string, error) {
	allOK := true
	for _, s := range sections {
		d := dirs.First(s.name)
		ok := (s.block == "" && d == nil) || (d != nil && strings.TrimRight(d.Content, "\n") == s.block)
		allOK = allOK && ok
	}
	allowed := make([]string, 0, len(sections))
	for _, s := range sections {
		allowed = append(allowed, s.name)
	}
	stray := strayGoldenBlocks(string(rawSource), allowed)
	if allOK && len(stray) == 0 {
		return "", nil
	}
	if opts.Sync {
		return writeStaticGolden(originalSource, sections), nil
	}
	for _, s := range sections {
		d := dirs.First(s.name)
		switch {
		case s.block != "" && d == nil:
			return "", fmt.Errorf(
				"no `// %s:` block present; re-run with "+
					"--update-golden-tests to record it:\n%s", s.name, indent(s.block, "  "))
		case s.block == "" && d != nil:
			return "", fmt.Errorf("stale `// %s:` block; re-run with --update-golden-tests to remove it", s.name)
		case d != nil && strings.TrimRight(d.Content, "\n") != s.block:
			return "", fmt.Errorf("// %s: diff:\n%s", s.name, unifiedDiff(d.Content, s.block))
		}
	}
	return "", fmt.Errorf("stray `// %s:` block in the golden region; re-run with --update-golden-tests to drop it", stray[0])
}

var reGoldenHeader = regexp.MustCompile(`^// ([A-Z][A-Za-z]*):`)

// strayGoldenBlocks lists the `// Name:` headers in src's trailing golden
// region that are not in allowed (blocks an older ledger format wrote).
func strayGoldenBlocks(src string, allowed []string) []string {
	region := strings.TrimPrefix(src, stripTrailingGoldenRegion(src))
	var stray []string
	for l := range strings.SplitSeq(region, "\n") {
		m := reGoldenHeader.FindStringSubmatch(l)
		if m == nil || slices.Contains(allowed, m[1]) {
			continue
		}
		stray = append(stray, m[1])
	}
	return stray
}

// writeUnsupportedDirective returns originalSource (golden region
// stripped) with a trailing `// Verdict: Unsupported <reason>` line (and
// the file's `// GnoTrim:` line, if any), so a file that fails on a
// Gno-unsupported construct is skipped on future runs (the verdict is
// detected pre-dispatch) and self-documents why.
func writeUnsupportedDirective(originalSource []byte, trim *gnoTrim, reason string) string {
	return writeStaticGolden(originalSource, append(trimSections(trim, nil),
		goldenSection{name: DirectiveVerdict, block: VerdictUnsupported + " " + TagUnsupportedReason(reason), inline: true}))
}

// writeStaticGolden returns the file content for a compile-mode golden
// update: originalSource (upstream-verbatim — NOT the rescued source, so
// in-memory transforms aren't persisted) with a trailing golden region
// appended, each non-empty section blank-line separated. Any existing
// trailing golden region is replaced rather than duplicated.
func writeStaticGolden(originalSource []byte, sections []goldenSection) string {
	body := strings.TrimRight(stripTrailingGoldenRegion(string(originalSource)), "\n")
	var sb strings.Builder
	sb.WriteString(body)
	sb.WriteString("\n")
	for _, s := range sections {
		if s.block == "" {
			continue
		}
		if s.inline {
			sb.WriteString("\n// " + s.name + ": " + s.block + "\n")
			continue
		}
		sb.WriteString("\n// ")
		sb.WriteString(s.name)
		sb.WriteString(":\n")
		for line := range strings.SplitSeq(s.block, "\n") {
			if line == "" {
				sb.WriteString("//\n")
				continue
			}
			sb.WriteString("// ")
			sb.WriteString(line)
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

// goldenRegionStarts are the directive names that can begin the trailing
// golden region of an errorcheck file.
var goldenRegionStarts = []string{
	DirectiveGnoTrim, DirectiveGnoTrimmed, DirectiveGnoError, DirectiveGoBuildError,
	// .go corpus files only: the run-mode goldens and the verdict line,
	// so a rewrite leaves no stale block behind.
	DirectiveGnoOutput, DirectiveGoOutput, DirectiveTypeCheckError,
	DirectiveVerdict,
}

// legacyGoldenBlocks are block names older ledger formats wrote; a
// re-sync strips them with the region and verify reports them as stray.
var legacyGoldenBlocks = []string{
	"GoTypeCheckError", "GoTypesOnly", "GnoOverStrictError", "UncaughtError",
	"GnoPreprocessError", "GoError", "Divergence",
}

// stripTrailingGoldenRegion removes the trailing golden region — one or
// more blank-line-separated `//`-comment blocks each beginning with a
// golden directive (goldenRegionStarts) — so a refresh replaces rather
// than appends. Returns src unchanged if the trailing comment block isn't
// a golden block. Only those directives open the region: any other
// trailing note is body, so the upstream manifest guards it.
func stripTrailingGoldenRegion(src string) string {
	lines := strings.Split(src, "\n")
	end := len(lines)
	for end > 0 && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}
	// The trailing run of comment and blank lines.
	start := end
	for start > 0 {
		t := strings.TrimSpace(lines[start-1])
		if t != "" && !strings.HasPrefix(t, "//") {
			break
		}
		start--
	}
	// Cut at the topmost block that opens with a golden directive. Free
	// text below it is part of the region too: notes belong in the
	// verdict's reason, so the verdict stays the last line of the file.
	for i := start; i < end; i++ {
		t := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(t, "//") {
			continue
		}
		if i > start && strings.TrimSpace(lines[i-1]) != "" {
			continue // not a block top
		}
		for _, name := range slices.Concat(goldenRegionStarts, legacyGoldenBlocks) {
			if strings.HasPrefix(t, "// "+name+":") {
				return strings.Join(lines[:i], "\n")
			}
		}
	}
	return src
}

// finalizeGoRunDivergence drives the symmetric Gno-vs-Go verdict for a
// run-mode file (a .go corpus file, or a .gno file with the
// Go-comparison opt-in) once both outputs are known. It is the sole
// authority for the run-mode golden region — the main match loop skips
// // GnoOutput: / // GnoError: / // GoOutput:.
//
// EVERY run file pins both sides, so a reviewer can judge bug-vs-expected
// from the file alone (a Gno panic where Go exits 0 is visible as a
// // GnoError: against an empty // GoOutput:):
//
//   - // GnoOutput: Gno's stdout — always (even empty). .go files only;
//     .gno files reuse their existing // Output: golden.
//   - // GnoError:  Gno's panic/error — only when non-empty (.go only).
//   - // GoOutput:  Go's combined stdout+stderr — always (even empty).
//   - // Verdict:   exactly one, for .go and opted-in .gno alike: `Clean run`
//     when the sides match, `KnownIssue` / `KnownDivergence` when they
//     differ (a human-written reason is preserved on re-sync).
//
// Gno's panic lands in gnoError, not gnoStdout; it is folded into the
// comparison to mirror Go's combined stream — otherwise a Gno panic with
// no prior stdout would compare equal to a clean Go run (both "") and
// silently pass a real bug (a panicking main with no prior output is
// exactly that shape).
//
// Returns (newDirs, err). In run mode sync always rewrites (both sides
// are re-pinned), so newDirs is non-nil under opts.Sync.
func finalizeGoRunDivergence(dirs Directives, gnoStdout, gnoError, goStdout string, prior error, sync, isGoFile bool) (Directives, error) {
	gnoOut := strings.TrimRight(trimTrailingSpaces(gnoStdout), "\n")
	gnoErr := strings.TrimRight(gnoError, "\n")
	goExp := strings.TrimRight(goStdout, "\n")

	gnoCombined := gnoOut
	if gnoErr != "" {
		if gnoCombined != "" {
			gnoCombined += "\n"
		}
		gnoCombined += gnoErr
	}
	diverges := gnoCombined != goExp

	// The Gno- and Go-side golden directives this file should carry.
	want := Directives{}
	if isGoFile {
		want = append(want, Directive{Name: DirectiveGnoOutput, Content: gnoOut})
		if gnoErr != "" {
			want = append(want, Directive{Name: DirectiveGnoError, Content: gnoErr})
		}
	}
	want = append(want, Directive{Name: DirectiveGoOutput, Content: goExp})
	if !diverges {
		want = append(want, verdict{Kind: VerdictClean, Mode: "run"}.directive())
	}

	if sync {
		// Rebuild the region: drop the old golden directives, re-append
		// the freshly-computed ones, then the verdict if the sides
		// differ (preserving the human's tag + reason).
		stale := []string{
			DirectiveGnoOutput, DirectiveGnoError, DirectiveGoOutput, DirectiveVerdict,
		}
		if isGoFile { // a .gno file keeps its native TypeCheckError golden
			stale = append(stale, DirectiveTypeCheckError)
		}
		out := append(removeDirectives(dirs, stale...), want...)
		if diverges {
			out = append(out, runVerdict(dirs, gnoErr != "").directive())
		}
		return out, prior
	}

	// Verify: each wanted directive must be present and match; a stale
	// // GnoError: (Gno no longer errors) must be gone; a verdict
	// (// KnownIssue: bug, or // KnownDivergence: accepted) must be
	// present iff the sides differ.
	for _, w := range want {
		switch d := dirs.First(w.Name); {
		case d == nil:
			prior = multierr.Append(prior, fmt.Errorf(
				"missing `// %s:` golden — re-run with --update-golden-tests", w.Name))
		case strings.TrimRight(d.Content, "\n") != w.Content:
			prior = multierr.Append(prior, fmt.Errorf("// %s: diff:\n%s",
				w.Name, unifiedDiff(d.Content, w.Content)))
		}
	}
	if isGoFile && gnoErr == "" && dirs.First(DirectiveGnoError) != nil {
		prior = multierr.Append(prior, fmt.Errorf(
			"stale `// GnoError:` golden (Gno no longer errors) — re-run with --update-golden-tests"))
	}
	kind := verdictKind(dirs)
	if diverges && kind == VerdictClean {
		prior = multierr.Append(prior, fmt.Errorf(
			"stale `// Verdict: Clean run`: Gno now diverges from Go — re-run with --update-golden-tests"))
	}
	hasVerdict := kind == VerdictKnownIssue || kind == VerdictKnownDivergence
	switch {
	case diverges && !hasVerdict:
		prior = multierr.Append(prior, fmt.Errorf(
			"Gno-vs-Go divergence detected — classify as `// Verdict: KnownIssue run: ...` (a Gno bug) or "+
				"`// Verdict: KnownDivergence run: ...` (accepted), or re-run with --update-golden-tests:\n%s",
			unifiedDiff(goExp, gnoCombined)))
	case !diverges && hasVerdict:
		prior = multierr.Append(prior, fmt.Errorf(
			"stale verdict: Gno's output now matches Go's — remove the KnownIssue / "+
				"KnownDivergence verdict (or re-run with --update-golden-tests)"))
	}
	return nil, prior
}

// runVerdict picks the verdict directive + reason for a diverging
// run-mode file. A human's prior choice (either tag, with a real reason)
// is preserved verbatim. Otherwise it auto-defaults: Gno erroring where
// Go ran clean (Go crashes are routed to // Unsupported before here, so
// gnoErrored here means a one-sided Gno failure) is a bug → KnownIssue;
// any other difference is provisionally an accepted KnownDivergence the
// contributor reclassifies if it's actually a bug.
func runVerdict(dirs Directives, gnoErrored bool) verdict {
	if v, ok := verdictIn(dirs); ok && (v.Kind == VerdictKnownIssue || v.Kind == VerdictKnownDivergence) &&
		v.Reason != "" && !strings.HasPrefix(v.Reason, "TODO") {
		return v
	}
	if gnoErrored {
		return verdict{Kind: VerdictKnownIssue, Mode: "run", Reason: "TODO: explain the Gno bug (Gno errors where Go runs clean)"}
	}
	return verdict{Kind: VerdictKnownDivergence, Mode: "run", Reason: "TODO: <category>: explain why this divergence is acceptable"}
}

// countDirectives returns how many entries of dirs carry name.
func countDirectives(dirs Directives, name string) int {
	n := 0
	for _, d := range dirs {
		if d.Name == name {
			n++
		}
	}
	return n
}

// removeDirectives returns a copy of dirs with all entries whose Name
// is in names omitted. Used by the divergence finalize to strip the
// triple when a blessed divergence has become stale (sync mode).
func removeDirectives(dirs Directives, names ...string) Directives {
	skip := make(map[string]bool, len(names))
	for _, n := range names {
		skip[n] = true
	}
	out := make(Directives, 0, len(dirs))
	for _, d := range dirs {
		if skip[d.Name] {
			continue
		}
		out = append(out, d)
	}
	return out
}

// packageTypesString returns a deterministic listing of every type
// declaration in the given package's block, one entry per line group:
//
//	<DeclName>[<TypeID>]=
//	    <indented amino JSON of the persisted form>
//
// The persisted form is produced via gno.PersistedTypeFormForTypeValue,
// so DeclaredTypes appear as RefType{ID} (matching the on-the-wire shape
// that copyValueWithRefs's TypeValue case emits), and aliases share the
// referenced type's TypeID.
//
// Entries are emitted in declaration (block-index) order. Unlike "Realm:"
// this is NOT a diff — every declared type is printed on every run.
func packageTypesString(m *gno.Machine, pkgPath string) string {
	pv := m.Store.GetPackage(pkgPath, false)
	if pv == nil {
		return ""
	}
	pb := pv.GetBlock(m.Store)
	if pb == nil || pb.Source == nil {
		return ""
	}
	names := pb.Source.GetBlockNames()
	var sb strings.Builder
	for i, tv := range pb.Values {
		if tv.T == nil || tv.T.Kind() != gno.TypeKind {
			continue
		}
		var name gno.Name
		if i < len(names) {
			name = names[i]
		}
		t := tv.GetType()
		tid := t.TypeID()
		persisted := gno.PersistedTypeFormForTypeValue(t)
		bz := amino.MustMarshalJSON(persisted)
		fmt.Fprintf(&sb, "%s[%s]=\n", name, tid)
		pretty := prettyTypeJSON(bz)
		indented := "    " + strings.ReplaceAll(string(pretty), "\n", "\n    ")
		sb.WriteString(indented)
		sb.WriteString("\n")
	}
	return sb.String()
}

// prettyTypeJSON indents JSON for readability, matching the Realm
// directive's style.
func prettyTypeJSON(jstr []byte) []byte {
	var c any
	if err := json.Unmarshal(jstr, &c); err != nil {
		return jstr
	}
	out, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		return jstr
	}
	return out
}

// returns a sorted string representation of realm diffs map
func realmDiffsString(m gno.StorageDiffs) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(fmt.Sprintf("%s: %d\n", k, m[k]))
	}
	return sb.String()
}

func trimTrailingSpaces(in string) string {
	lines := strings.Split(in, "\n")
	for i, line := range lines {
		line = strings.TrimRight(line, " ")
		lines[i] = line
	}
	return strings.Join(lines, "\n")
}

func unifiedDiff(wanted, actual string) string {
	diff, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(wanted),
		B:        difflib.SplitLines(actual),
		FromFile: "Expected",
		FromDate: "",
		ToFile:   "Actual",
		ToDate:   "",
		Context:  1,
	})
	if err != nil {
		panic(fmt.Errorf("error generating unified diff: %w", err))
	}
	return diff
}

type runResult struct {
	Output string
	Error  string
	// gnoErrors holds every preprocess error of an errorcheck run
	// (accumulate mode), with location prefixes; Error joins them.
	gnoErrors []string
	// parseBlocked: the file has no Gno form at all (no package clause,
	// or Go2Gno failed), so the checker never ran and markers it did not
	// reach say nothing about it. A file with syntax errors is checked.
	parseBlocked bool
	// onChainError is the first error of the on-chain path (plain
	// Preprocess, which stops at one), "" when it accepts the file.
	onChainError string
	// Set if there was an issue with type-checking.
	TypeCheckError string
	// Set if there was a panic within gno code.
	GnoStacktrace string
	// Set if this was recovered from a panic.
	GoPanicStack []byte
}

// preprocessOnly, when true, runs the file through parse + preprocess +
// typecheck but does NOT execute main(). Used for errorcheck filetests:
// they're compile-error tests gc never runs, and executing a file Gno
// failed to reject at compile time can hang (e.g. an undefined-label
// `break L2` Gno is lenient about, leaving a `for {}` infinite loop).
// All compile/type errors are caught during preprocess, so the verdict
// is unaffected.
func (opts *TestOptions) runTest(m *gno.Machine, pkgPath, fname string, content []byte, opslog io.Writer, tcheck, preprocessOnly bool) (rr runResult) {
	pkgName, err := gno.PackageNameFromFileBody(fname, string(content))
	if err != nil {
		if !strings.HasSuffix(fname, ".go") {
			return runResult{Error: err.Error()}
		}
		// A corpus file mid-errorcheck pass may have its package clause
		// neutralised; the rescue's PKGPATH names the package then.
		pkgName = gno.Name(pkgPath[strings.LastIndexByte(pkgPath, '/')+1:])
	}
	tcError := ""
	fname = filepath.Base(fname)
	if opts.tcCache == nil {
		opts.tcCache = make(gno.TypeCheckCache)
	}

	// Eagerly load imports.
	// LoadImports is run using opts.Store, rather than the transaction store;
	// it allows us to only have to load the imports once (and re-use the cached
	// versions). Running the tests in separate "transactions" means that they
	// don't get the parent store dirty.
	abortOnError := true
	if err := LoadImports(opts.TestStore, &std.MemPackage{
		Type: gno.MPFiletests,
		Name: string(pkgName),
		Path: pkgPath,
		Files: []*std.MemFile{
			{Name: "gnomod.toml", Body: gno.GenGnoModLatest(pkgPath)},
			{Name: fname, Body: string(content)},
		},
	}, abortOnError); err != nil {
		// NOTE: we perform this here, so we can capture the runResult.
		if swe, ok := err.(*stackWrappedError); ok {
			return runResult{Error: err.Error(), GoPanicStack: swe.stack}
		}
		return runResult{Error: err.Error()}
	}

	// Reset and start capturing stdout.
	opts.filetestBuffer.Reset()
	revert := opts.outWriter.tee(&opts.filetestBuffer)
	defer revert()

	defer func() {
		if r := recover(); r != nil {
			rr.Output = opts.filetestBuffer.String()
			rr.GoPanicStack = debug.Stack()
			rr.TypeCheckError = tcError
			switch v := r.(type) {
			case *gno.TypedValue:
				rr.Error = v.Sprint(m)
			case *gno.PreprocessError:
				rr.Error = v.Unwrap().Error()
			case gno.UnhandledPanicError:
				rr.Error = v.Error()
				rr.GnoStacktrace = m.ExceptionStacktrace()
			default:
				rr.Error = fmt.Sprint(v)
				rr.GnoStacktrace = m.Stacktrace().String()
			}
		}
	}()

	// Validate that package name matches path last element.
	// See https://github.com/gnolang/gno/issues/1571
	if err := gno.ValidatePkgNameMatchesPath(pkgName, pkgPath); err != nil {
		panic(err)
	}

	// Remove filetest from name, as that can lead to the package not being
	// parsed correctly when using RunMemPackage.
	fname = strings.ReplaceAll(fname, "_filetest", "")

	// Use last element after / (works also if slash is missing).
	if !gno.IsRealmPath(pkgPath) { // Simple case - pure package.
		// Determine package type based on path
		mptype := gno.MPUserProd
		if strings.HasSuffix(pkgPath, "_test") {
			mptype = gno.MPUserIntegration
		}
		// Construct mem package for single filetest.
		mpkg := &std.MemPackage{
			Type: mptype,
			Name: string(pkgName),
			Path: pkgPath,
			Files: []*std.MemFile{
				{Name: "gnomod.toml", Body: gno.GenGnoModLatest(pkgPath)},
				{Name: fname, Body: string(content)},
			},
		}
		// Validate Gno syntax and type check.
		if tcheck {
			if _, err := gno.TypeCheckMemPackage(memPackageForTypeCheck(mpkg), gno.TypeCheckOptions{
				// Use Teststore to load imported packages,
				// mimicing the loading behavior with on-chain.
				// (if using m.Store, the realm package will
				// be preloaded during typecheck)
				Getter:     opts.TestStore,
				TestGetter: m.Store,
				Mode:       gno.TCLatestRelaxed,
				Cache:      opts.tcCache,
			}); err != nil {
				tcError = restoreGoExtInError(fname, fmt.Sprintf("%v", err.Error()))
			}
		}
		// Must parse before set pn&pv.
		fn := m.MustParseFile(fname, string(content))
		// Construct throwaway package and parse file.
		pn := gno.NewPackageNode(pkgName, pkgPath, &gno.FileSet{})
		pv := pn.NewPackage(m.Alloc)
		m.Store.SetBlockNode(pn)
		m.Store.SetCachePackage(pv)
		m.SetActivePackage(pv)
		m.Context.(*teststdlibs.TestExecContext).OriginCaller = DefaultCaller
		// Run (add) file, and then run main().
		m.RunFiles(fn)
		if !preprocessOnly {
			m.RunMain()
		}
	} else { // Realm case.
		gno.DisableDebug() // until main call.

		// Save package using realm crawl procedure.
		// Realms are always MPUserProd because they need to be stored
		mpkg := &std.MemPackage{
			Type: gno.MPUserProd,
			Name: string(pkgName),
			Path: pkgPath,
			Files: []*std.MemFile{
				{Name: "gnomod.toml", Body: gno.GenGnoModLatest(pkgPath)},
				{Name: fname, Body: string(content)},
			},
		}
		// Start transaction store.
		orig, txs := m.Store, m.Store.BeginTransaction(nil, nil, nil, nil)
		m.Store = txs
		// Validate Gno syntax and type check.
		if tcheck {
			if _, err := gno.TypeCheckMemPackage(memPackageForTypeCheck(mpkg), gno.TypeCheckOptions{
				Getter:     m.Store,
				TestGetter: m.Store,
				Mode:       gno.TCLatestRelaxed,
				Cache:      opts.tcCache,
			}); err != nil {
				tcError = restoreGoExtInError(fname, fmt.Sprintf("%v", err.Error()))
			}
		}
		// Set OriginCaller BEFORE running package init so package-level
		// var initializers that read runtime.OriginCaller() (e.g. `var c
		// = runtime.OriginCaller()` at package scope) see the proper
		// DefaultCaller, matching the package-main path's ordering above.
		m.Context.(*teststdlibs.TestExecContext).OriginCaller = DefaultCaller

		// Run decls and init functions.
		m.RunMemPackage(mpkg, true)

		// Clear store cache and reconstruct machine from committed info
		// (mimicking on-chain behaviour).
		// (jae) why is this needed?
		txs.Write()
		m.Store = orig
		pv2 := m.Store.GetPackage(pkgPath, false)
		m.SetActivePackage(pv2)
		gno.EnableDebug()

		// Clear store.opslog from init function(s).
		m.Store.SetLogStoreOps(opslog) // resets.
		if !preprocessOnly {
			m.RunMainMaybeCrossing()
		}
	}
	return runResult{
		Output:         opts.filetestBuffer.String(),
		GnoStacktrace:  m.Stacktrace().String(),
		TypeCheckError: tcError,
	}
}

// ---------------------------------------
// directives and directive parsing

const (
	// These directives are used to set input variables, which should change for
	// the specific filetest. They must be specified on a single line.
	DirectivePkgPath  = "PKGPATH"
	DirectiveMaxAlloc = "MAXALLOC"
	DirectiveSend     = "SEND"

	// These are used to match the result of the filetest against known golden
	// values.
	DirectiveOutput         = "Output"
	DirectiveError          = "Error"
	DirectiveRealm          = "Realm"
	DirectiveEvents         = "Events"
	DirectivePreprocessed   = "Preprocessed"
	DirectiveStacktrace     = "Stacktrace"
	DirectiveGas            = "Gas"
	DirectiveStorage        = "Storage"
	DirectiveTypes          = "Types"
	DirectiveTypeCheckError = "TypeCheckError"

	// DirectiveVerdict is the one verdict line every Go-compared file
	// (.go corpus file or opted-in .gno) ends with; see verdict.go.
	DirectiveVerdict = "Verdict"

	// Symmetric Gno-vs-Go golden directives for .go corpus files.
	// They mirror existing native directives but split the actual
	// outputs by source so a reader sees both sides structurally:
	//
	//   // GnoOutput: <Gno's actual stdout>
	//   // GoOutput:  <`go run`'s actual stdout>
	//   // Verdict:   Clean run, or KnownIssue / KnownDivergence with a reason
	//
	// Gno's panic goes to // GnoError:. All are written by
	// `--update-golden-tests`; a Known* verdict gets a TODO placeholder
	// the contributor replaces.
	DirectiveGnoOutput = "GnoOutput"
	DirectiveGoOutput  = "GoOutput"
	DirectiveGnoError  = "GnoError"
	// DirectiveGoBuildError pins the lines `go build` flagged on a
	// compile-mode file (numbers only, wording follows the Go release).
	// Gno's side of a compile file is the same `// GnoError:` block run
	// mode uses; the verdict's mode says which stage produced it.
	DirectiveGoBuildError = "GoBuildError"
)

// nativeDirectives are the .gno filetest directives; corpusDirectives
// are the .go corpus additions. Only corpus directives accept the
// inline `// Name: content` form, so `// Output: foo` in a .gno comment
// stays a comment.
var nativeDirectives = []string{
	DirectivePkgPath,
	DirectiveMaxAlloc,
	DirectiveSend,
	DirectiveOutput,
	DirectiveError,
	DirectiveRealm,
	DirectiveEvents,
	DirectivePreprocessed,
	DirectiveStacktrace,
	DirectiveGas,
	DirectiveStorage,
	DirectiveTypes,
	DirectiveTypeCheckError,
}

var corpusDirectives = []string{
	DirectiveVerdict,
	DirectiveGnoTrim,
	DirectiveGnoTrimmed,
	DirectiveGnoOutput,
	DirectiveGoOutput,
	DirectiveGnoError,
	DirectiveGoBuildError,
}

var allDirectives = slices.Concat(nativeDirectives, corpusDirectives)

// singleLinePascalDirectives holds PascalCase directives whose content
// is ALWAYS single-line — even when empty — and that must be parsed
// without the bare-PascalCase multi-line absorbing behavior. Members
// (the verdict line) carry one-line text.
//
// Other PascalCase directives (Output / Error / GnoOutput / GoOutput /
// GnoError / GoError / Realm / …) are multi-line markers by default,
// matching the .gno convention: directive on its own line, then
// `//`-prefixed content lines, terminated by a blank line or end of
// file. Corpus directives also accept the inline `// Name: foo` form
// — see ParseDirectives.
var singleLinePascalDirectives = map[string]bool{
	DirectiveVerdict:    true,
	DirectiveGnoTrim:    true,
	DirectiveGnoTrimmed: true,
}

// pinnedGoldenDirectives lists PascalCase directives whose empty
// content is meaningful and must still be serialized (rather than
// skipped). `// GoOutput:` with no lines means "Go produces no stdout"
// — a pinned-golden assertion we want visible in the file.
var pinnedGoldenDirectives = map[string]bool{
	DirectiveGnoOutput:    true,
	DirectiveGoOutput:     true,
	DirectiveGnoError:     true,
	DirectiveGoBuildError: true,
}

// Directives contains the directives of a file.
// It may also contains directives with empty names, to indicate parts of the
// original source file (used to re-construct the filetest at the end).
type Directives []Directive

// First returns the first directive with the corresponding name.
func (d Directives) First(name string) *Directive {
	if name == "" {
		return nil
	}
	for i := range d {
		if d[i].Name == name {
			return &d[i]
		}
	}
	return nil
}

// FirstDefault returns the [Directive.Content] of First(name); if First(name)
// returns nil, then defaultValue is returned.
func (d Directives) FirstDefault(name, defaultValue string) string {
	v := d.First(name)
	if v == nil {
		return defaultValue
	}
	return v.Content
}

// FileTest re-generates the filetest from the given directives; the inverse of ParseDirectives.
func (d Directives) FileTest() string {
	var bld strings.Builder
	for i, dir := range d {
		ll := ""
		if i < len(d)-1 {
			ll = "\n"
		}
		switch {
		case dir.Name == "":
			cnt := strings.TrimRight(dir.Content, "\n ")
			bld.WriteString(cnt + "\n" + ll)
		case strings.ToUpper(dir.Name) == dir.Name: // ALLCAPS:
			bld.WriteString("// " + dir.Name + ": " + dir.Content + ll)
		case singleLinePascalDirectives[dir.Name]:
			// Single-line PascalCase meta-directives. Always one line,
			// content (possibly empty) right after the colon.
			if dir.Content == "" {
				bld.WriteString("// " + dir.Name + ":" + ll)
			} else {
				bld.WriteString("// " + dir.Name + ": " + dir.Content + ll)
			}
			if ll == "" { // last directive (a verdict): the file still ends in a newline
				bld.WriteString("\n")
			}
		default:
			if dir.Content == "" || dir.Content == "\n" {
				// Pinned-golden directives (`// GoOutput:` etc.) carry
				// meaning even when empty — "Go produces no stdout"
				// is a positive assertion. Emit the bare marker plus a
				// blank-line separator so the parser doesn't absorb
				// subsequent directives into it.
				if pinnedGoldenDirectives[dir.Name] {
					bld.WriteString("// " + dir.Name + ":\n" + ll)
				}
				continue
			}
			bld.WriteString("// " + dir.Name + ":\n")
			cnt := strings.TrimRight(dir.Content, "\n ")
			lines := strings.Split(cnt, "\n")
			for _, line := range lines {
				if line == "" {
					bld.WriteString("//\n")
					continue
				}
				bld.WriteString("// ")
				bld.WriteString(line)
				bld.WriteString("\n")
			}
			bld.WriteString(ll)
		}
	}
	return bld.String()
}

// Directive represents a directive in a filetest.
// A [Directives] slice may also contain directives with empty Names:
// these compose the source file itself, and are used to re-construct the file
// when a directive is changed.
type Directive struct {
	Name     string
	Content  string
	Complete bool
	LastLine string
}

// Allows either a `ALLCAPS: content` on a single line, or a `PascalCase:`,
// with content on the following lines.
var reDirectiveLine = regexp.MustCompile("^(?:([A-Za-z]*):|([A-Z]+): ?(.*))$")

// Single-line PascalCase directives: `Name:` or `Name: content` on
// one line. Used by meta-directives like `Verdict:` /
// `GnoOutput:` / `GoOutput:` whose payload is a short
// single-line value (often empty for `GoOutput:` when Go produces no
// stdout). PascalCase distinguishes them from the ALLCAPS
// input-parameter family (PKGPATH/MAXALLOC/SEND); a bare `Name:` is
// single-line only for [singleLinePascalDirectives], an inline
// `Name: content` only for [corpusDirectives] — see ParseDirectives.
var reDirectiveSingleLinePascal = regexp.MustCompile(`^([A-Z][a-z][A-Za-z]*):(?: (.*))?$`)

// ParseDirectives parses all the directives in the filetest given at source.
func ParseDirectives(source io.Reader) (Directives, error) {
	sc := bufio.NewScanner(source)
	parsed := make(Directives, 0, 8)
	parsed = append(parsed, Directive{LastLine: "// FAUX: faux directive", Complete: true}) // faux directive.
	for sc.Scan() {
		last := &parsed[len(parsed)-1]
		txt := sc.Text()
		if !strings.HasPrefix(txt, "//") {
			// If we're already in an incomplete text directive, simply append there.
			if last.Name == "" && !last.Complete {
				last.Content += txt + "\n"
				last.LastLine = txt
				continue
			}
			// Otherwise make a new directive.
			parsed = append(parsed,
				Directive{
					Content:  txt + "\n",
					LastLine: txt,
				})
			continue
		}

		comment := txt[2:]                         // leading double slash
		comment = strings.TrimPrefix(comment, " ") // leading space (if any)

		// Special case if following an incomplete comment line,
		// always append to it even if it looks like `// TODO: ...`.
		if strings.HasPrefix(txt, "//") &&
			strings.HasPrefix(last.LastLine, "//") &&
			!last.Complete {
			if last.Name == "" {
				// Just append text to it.
				last.Content += txt + "\n"
				last.LastLine = txt
				continue
			} else {
				// Just append comment to it.
				last.Content += comment + "\n"
				last.LastLine = txt
				continue
			}
		}

		// PascalCase single-line directive. Two acceptance paths:
		//   1. Name is in [singleLinePascalDirectives] (Unsupported,
		//      Divergence) — always single-line, content may be empty.
		//   2. Inline content is non-empty and Name is a corpus
		//      directive ([corpusDirectives]). Bare `// Name:` falls
		//      through to the multi-line marker rule below.
		// Checked before reDirectiveLine so multi-line absorption
		// doesn't eat the next comment line for case 1.
		if subm2 := reDirectiveSingleLinePascal.FindStringSubmatch(comment); subm2 != nil {
			name := subm2[1]
			content := subm2[2]
			isSingle := singleLinePascalDirectives[name] ||
				(content != "" && slices.Contains(corpusDirectives, name))
			if isSingle {
				parsed = append(parsed,
					Directive{
						Name:     name,
						Content:  content,
						Complete: true,
					})
				continue
			}
		}
		// Find if there is a colon (indicating a possible directive).
		subm := reDirectiveLine.FindStringSubmatch(comment)
		if subm != nil && slices.Contains(allDirectives, subm[1]) {
			// CamelCase directive.
			parsed = append(parsed,
				Directive{
					Name:     subm[1],
					LastLine: txt,
				})
			continue
		}
		if subm != nil && slices.Contains(allDirectives, subm[2]) {
			// APPCAPS directive.
			parsed = append(parsed,
				Directive{
					Name:     subm[2],
					Content:  subm[3],
					Complete: true,
				})
			continue
		}

		// Not a directive, just a comment.
		// If we're already in an incomplete directive, simply append there.
		if !last.Complete {
			if last.Name == "" {
				last.Content += txt + "\n"
				last.LastLine = txt
				continue
			} else {
				last.Content += comment + "\n"
				last.LastLine = txt
				continue
			}
		}
		// Otherwise make a new directive.
		parsed = append(parsed,
			Directive{
				Content:  txt + "\n",
				LastLine: txt,
			})
	}

	// Remove trailing (newline|space)* and filter empty directives —
	// except pinned-golden ones (`// GnoOutput:` / `// GoOutput:` / …),
	// whose emptiness is a meaningful assertion ("produces no output")
	// and must round-trip, mirroring the write side in FileTest.
	result := make([]Directive, 0, len(parsed))
	parsed = parsed[1:] // remove faux directive
	for _, dir := range parsed {
		content := dir.Content
		content = strings.TrimRight(content, "\n ")
		if content == "" && !pinnedGoldenDirectives[dir.Name] {
			continue
		}
		dir.Content = content
		result = append(result, dir)
	}

	return result, sc.Err()
}

// goSubprocessTimeout caps how long `go run` is allowed to take when
// auto-deriving the expected output for a .go filetest. Most corpus
// files complete in well under a second; the cap is here so a
// pathological hang doesn't deadlock CI.
var goSubprocessTimeout = 30 * time.Second

// runGoToolchain compiles+runs source via the host Go toolchain and
// returns Go's user-visible output (stdout + stderr combined) and any
// build/compile errors. Dispatches on source content: runnable files
// (package main + func main) go through `go run`; non-runnable files
// go through `go build` so non-main packages also compile cleanly.
//
// Stdout and stderr are combined, in the order the program wrote them,
// because Gno's runtime exposes one output stream — comparing only Go's
// stdout would flag the builtin `println` (Go: stderr; Gno: stdout) as
// a divergence purely from the stream choice. Both streams share one
// file descriptor; two buffers would lose the interleaving.
//
// Non-zero exit (panic, compile error) is NOT a function-level error
// — it's the corpus file's expected behavior. The only error returned
// is a genuine exec failure (Go toolchain not on PATH, timeout).
// runGoToolchain runs source with the Go toolchain on PATH: `go run` for a
// runnable program, else `go build` (always, when buildOnly: compile-mode
// files are never executed).
func runGoToolchain(source []byte, buildOnly bool) (output string, err error) {
	runnable := IsRunnable(source) && !buildOnly
	dir, mkErr := os.MkdirTemp("", "gno-filetest-go-*")
	if mkErr != nil {
		return "", fmt.Errorf("mkdir temp: %w", mkErr)
	}
	defer os.RemoveAll(dir)
	srcPath := filepath.Join(dir, "main.go")
	if err := os.WriteFile(srcPath, source, 0o644); err != nil {
		return "", fmt.Errorf("write temp source: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), goSubprocessTimeout)
	defer cancel()

	// `go build foo.go` (single-file form) requires package main.
	// For non-main shapes (errorcheck/compile-only files declaring
	// `package p`), `go build .` on the temp directory works — BUT
	// requires a go.mod, otherwise Go complains "go.mod not found"
	// before it even parses the source. Write a minimal go.mod so we
	// get the real compile errors (e.g. "missing function body").
	// Its `go` line is the toolchain's own version: an older pin would
	// set -lang and make gc reject `any`, unions or `max` as "requires
	// go1.N or later", which is not the reference verdict we want.
	if !runnable {
		modPath := filepath.Join(dir, "go.mod")
		mod := "module gnofiletest\n\ngo " + goModLangVersion(runtime.Version()) + "\n"
		if err := os.WriteFile(modPath, []byte(mod), 0o644); err != nil {
			return "", fmt.Errorf("write temp go.mod: %w", err)
		}
	}
	var args []string
	if runnable {
		args = []string{"run", srcPath}
	} else {
		args = []string{"build", "-o", os.DevNull, "."}
	}
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = dir
	// Disable module network I/O — corpus files use stdlib only, and
	// `go build` in a fresh dir with go.mod will otherwise try to reach
	// GOPROXY (proxy.golang.org), which is slow / can hang. GOPROXY=off
	// prevents any download attempt.
	cmd.Env = append(os.Environ(), "GOPROXY=off", "GOFLAGS=-mod=mod", "GONOSUMCHECK=1")
	// WaitDelay caps how long cmd.Wait blocks AFTER the context expires
	// waiting for pipe-reader goroutines. Without it, a `go build` that
	// spawns child compile/link processes (which inherit stdout/stderr)
	// can hang for hours after SIGKILL — the children keep the pipes
	// open. E.g. fixedbugs/bug137.go's compile spun >19min on this.
	// 5s is well past any legitimate reap time.
	cmd.WaitDelay = 5 * time.Second
	outPath := filepath.Join(dir, "output")
	outFile, createErr := os.Create(outPath)
	if createErr != nil {
		return "", fmt.Errorf("create temp output: %w", createErr)
	}
	cmd.Stdout = outFile
	cmd.Stderr = outFile
	runErr := cmd.Run()
	outFile.Close()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return "", fmt.Errorf("`go %s` exceeded %s timeout — mark the file with `// Unsupported:`",
			args[0], goSubprocessTimeout)
	}
	combined, readErr := os.ReadFile(outPath)
	if readErr != nil {
		return "", fmt.Errorf("read temp output: %w", readErr)
	}
	output = strings.TrimRight(string(combined), "\n")
	if runErr != nil {
		var ee *exec.ExitError
		if !errors.As(runErr, &ee) {
			return "", fmt.Errorf("`go %s` failed (is `go` on PATH?): %w", args[0], runErr)
		}
		// Non-zero exit is expected for panic-class corpus files; the
		// diagnostics are part of the captured output.
	}
	return output, nil
}

// memPackageForTypeCheck wraps mpkg so its files are visible to
// gno.TypeCheckMemPackage. That call filters input by extension and
// silently skips anything not ending in .gno (the .gno suffix is the
// canonical user-source extension; .go in the gno repo conventionally
// means VM/tooling implementation, which never goes through the
// typecheck pipeline). For .go files dropped under tests/files/
// gocorpus/testdata/ as regression tests for Go's standard test corpus,
// the in-memory MemFile name gets aliased to .gno here so the typecheck
// actually runs. The outer fname (used for parser provenance, error
// attribution, and the runtime path) stays as .go everywhere else.
//
// Returns mpkg unchanged when no .go files are present.
func memPackageForTypeCheck(mpkg *std.MemPackage) *std.MemPackage {
	needsRename := false
	for _, f := range mpkg.Files {
		if filepath.Ext(f.Name) == ".go" {
			needsRename = true
			break
		}
	}
	if !needsRename {
		return mpkg
	}
	out := &std.MemPackage{
		Type:  mpkg.Type,
		Name:  mpkg.Name,
		Path:  mpkg.Path,
		Files: make([]*std.MemFile, len(mpkg.Files)),
	}
	for i, f := range mpkg.Files {
		if filepath.Ext(f.Name) == ".go" {
			out.Files[i] = &std.MemFile{
				Name: strings.TrimSuffix(f.Name, ".go") + ".gno",
				Body: f.Body,
			}
		} else {
			out.Files[i] = f
		}
	}
	return out
}

// restoreGoExtInError post-processes a TypeCheckMemPackage error
// message so it references the original .go filename rather than the
// .gno-suffix alias used internally by [memPackageForTypeCheck].
// Acts on the basename only to avoid touching unrelated paths in the
// message body. No-op when originalFname doesn't end in .go.
func restoreGoExtInError(originalFname, tcError string) string {
	if filepath.Ext(originalFname) != ".go" {
		return tcError
	}
	base := strings.TrimSuffix(filepath.Base(originalFname), ".go")
	return strings.ReplaceAll(tcError, base+".gno", base+".go")
}

var reGoVersion = regexp.MustCompile(`^go(\d+\.\d+)`)

// goModLangVersion turns runtime.Version() ("go1.25.9") into a go.mod
// `go` line value ("1.25"); a devel build falls back to the go.mod floor.
func goModLangVersion(v string) string {
	if m := reGoVersion.FindStringSubmatch(v); m != nil {
		return m[1]
	}
	return "1.21"
}

// runErrorcheck parses and preprocesses content in accumulate mode;
// nothing is type-checked by go/types or executed. It also runs the
// on-chain path (plain Preprocess, first error) on onChain, so the
// ledger can check that the two agree on whether the file is rejected.
func (opts *TestOptions) runErrorcheck(m, onChain *gno.Machine, pkgPath, fname string, content []byte) (rr runResult) {
	pkgName, err := gno.PackageNameFromFileBody(fname, string(content))
	if err != nil {
		return runResult{Error: err.Error()}
	}
	fname = filepath.Base(fname)
	rr.onChainError = firstPreprocessError(onChain, pkgName, pkgPath, fname, content)
	defer func() {
		if r := recover(); r != nil { // a VM bug: keep what was collected before it
			msg := fmt.Sprint(r)
			if pe, ok := r.(*gno.PreprocessError); ok {
				msg = pe.Unwrap().Error()
			}
			rr.gnoErrors = append(rr.gnoErrors, stripPreprocessStack(msg))
			rr.Error = strings.Join(rr.gnoErrors, "; ")
		}
	}()
	// Every syntax and declaration error go/parser reports is pinned,
	// cascades included: a resynchronisation error on an unmarked line is
	// an honest over-strict line. The checker then runs on the partial
	// tree, each unparsed node a dropped statement.
	fn, perrs := m.ParseFileCollect(fname, string(content))
	for _, e := range perrs {
		rr.gnoErrors = append(rr.gnoErrors, e.Error())
	}
	if len(perrs) > 0 {
		// The parse error rejects the file on chain too, as its first error.
		rr.onChainError = rr.gnoErrors[0]
	}
	if fn == nil {
		// No package clause, or the partial tree has no Gno form: the
		// checker never ran, so markers it did not reach say nothing.
		rr.parseBlocked = true
		rr.Error = strings.Join(rr.gnoErrors, "; ")
		return rr
	}
	fset := &gno.FileSet{Files: []*gno.FileNode{fn}}
	pn := gno.NewPackageNode(pkgName, pkgPath, fset)
	pv := pn.NewPackage(m.Alloc)
	m.Store.SetBlockNode(pn)
	m.Store.SetCachePackage(pv)
	m.SetActivePackage(pv)
	for _, e := range gno.PreprocessFilesCollect(m.Store, pn, fset) {
		rr.gnoErrors = append(rr.gnoErrors, stripPreprocessStack(e.Unwrap().Error()))
	}
	rr.Error = strings.Join(rr.gnoErrors, "; ")
	return rr
}

// firstPreprocessError runs the on-chain path on m: parse, predefine and
// preprocess the file (PreprocessFiles, the deploy path's checking half,
// which stops at the first error), without running any initializer or
// init function. It returns that error's text, or "" when the path
// accepts the file. A parse error or a runtime panic is not a checking
// error.
func firstPreprocessError(m *gno.Machine, pkgName gno.Name, pkgPath, fname string, content []byte) (msg string) {
	defer func() {
		if r := recover(); r != nil {
			if pe, ok := r.(*gno.PreprocessError); ok {
				msg = stripPreprocessStack(pe.Unwrap().Error())
			}
		}
	}()
	fn, err := m.ParseFile(fname, string(content))
	if err != nil {
		return ""
	}
	m.PreprocessFiles(string(pkgName), pkgPath, &gno.FileSet{Files: []*gno.FileNode{fn}}, false, false)
	return ""
}

// stripPreprocessStack drops the preprocess stack dump an internal error
// carries after its message; the ledger pins the message alone.
func stripPreprocessStack(msg string) string {
	if i := strings.Index(msg, "\n--- preprocess stack ---"); i >= 0 {
		return strings.TrimSuffix(msg[:i], ":")
	}
	return msg
}

// bodilessFuncGap is Gno's message for a function declared without a
// body (gc allows it: the body is assembly). In errorcheck mode it is a
// per-line gap, not a file gap: the error stays in the GnoError block
// and the function stays declared, so calls to it still type-check.
const bodilessFuncGap = "does not have a body but is not natively defined"

// fileGapReason is the feature-gap reason that routes a file to
// Unsupported, or "". An errorcheck file is judged per error, with the
// bodiless-function message exempt.
func fileGapReason(result runResult, errorcheck bool) string {
	if !errorcheck {
		return UnsupportedFeatureError(result.Error)
	}
	for _, e := range result.gnoErrors {
		if strings.Contains(e, bodilessFuncGap) {
			continue
		}
		if r := UnsupportedFeatureError(e); r != "" {
			return r
		}
	}
	return ""
}

// errorcheckCoverage is the attribution of accumulate-mode errors to the
// file's markers, by line.
type errorcheckCoverage struct {
	gnoErr     map[int]string // first message per start line, source coordinates
	uncaught   []int          // checkable marker lines no error credits
	overStrict []int          // start lines of errors that credit no marker
	internal   []int          // start lines of internal errors
}

// creditLines lists the lines an error credits: the line its span starts
// on, and the line it ends on for a closing-brace rule (a function that
// does not terminate), where gc puts the marker. One error credits
// nothing else, however many lines it spans.
func creditLines(lo, hi int, msg string) []int {
	if hi > lo && strings.Contains(msg, "does not terminate") {
		return []int{lo, hi}
	}
	return []int{lo}
}

// attributeErrorcheck maps gnoErrors (located, in machine coordinates)
// onto markers. An internal error is a VM bug, not a verdict: it credits
// no marker and is listed on its own. An EOF error is positioned one past
// the last line; gc's marker for it sits on the last line.
func attributeErrorcheck(markers []InlineError, gnoErrors []string, prependedLines, bodyLines int) errorcheckCoverage {
	cov := errorcheckCoverage{gnoErr: map[int]string{}}
	marked := make(map[int]bool, len(markers))
	for _, mk := range markers {
		marked[mk.Line] = true
	}
	credited := map[int]bool{}
	for _, e := range gnoErrors {
		lo, hi := errorSpan(e)
		if lo == 0 {
			continue
		}
		lo, hi = lo-prependedLines, hi-prependedLines
		if lo > bodyLines {
			lo, hi = bodyLines, bodyLines
		}
		msg := CleanErrorMessage(e)
		if _, ok := cov.gnoErr[lo]; !ok {
			cov.gnoErr[lo] = msg
		}
		if IsInternalError(e) {
			if !slices.Contains(cov.internal, lo) {
				cov.internal = append(cov.internal, lo)
			}
			continue
		}
		hit := false
		for _, line := range creditLines(lo, hi, msg) {
			credited[line] = true
			hit = hit || marked[line]
		}
		// A feature gap on one line of an otherwise checkable file (a
		// stdlib symbol Gno lacks, a newer language feature) is not
		// over-strictness; the GnoError block still shows it.
		if !hit && UnsupportedFeatureError(msg) == "" && !slices.Contains(cov.overStrict, lo) {
			cov.overStrict = append(cov.overStrict, lo)
		}
	}
	for _, mk := range markers {
		if !mk.GcOnly && !credited[mk.Line] {
			cov.uncaught = append(cov.uncaught, mk.Line)
		}
	}
	slices.Sort(cov.overStrict)
	slices.Sort(cov.internal)
	return cov
}

// A located message is "file:L:C: msg", "file:L:C-C2: msg" (same line),
// "file:L:C-L2:C2: msg", or "file:L: msg"; a span may end in primes
// (nodes synthesised at one position).
var reErrorSpan = regexp.MustCompile(`:(\d+)(?::\d+)?(?:-(?:(\d+):)?\d+)?'*:`)

// errorSpan returns the first and last line of a located error message
// ("pkg/file.go:12:3-14:2: msg"); 0, 0 when it has no location.
func errorSpan(msg string) (lo, hi int) {
	m := reErrorSpan.FindStringSubmatch(msg)
	if m == nil {
		return 0, 0
	}
	lo, _ = strconv.Atoi(m[1])
	hi = lo
	if m[2] != "" {
		hi, _ = strconv.Atoi(m[2])
	}
	return lo, hi
}

// errorcheckGapReason is the sync-written KnownIssue reason for an
// errorcheck file: the umbrella issue, then the marker lines Gno missed,
// the unmarked lines it rejected, the lines where it failed internally,
// and how the on-chain path disagrees with accumulate mode.
func errorcheckGapReason(uncaught, overStrict, internal []int, onChain string) string {
	lines := lineList
	var parts []string
	if len(uncaught) > 0 {
		parts = append(parts, "uncaught markers: lines "+lines(uncaught))
	}
	if len(overStrict) > 0 {
		slices.Sort(overStrict)
		parts = append(parts, "over-strict: lines "+lines(overStrict))
	}
	if len(internal) > 0 {
		slices.Sort(internal)
		parts = append(parts, "internal errors: lines "+lines(internal))
	}
	if onChain != "" {
		parts = append(parts, "on-chain path "+onChain)
	}
	if len(parts) == 0 {
		return preprocessUmbrella
	}
	return preprocessUmbrella + " " + strings.Join(parts, "; ")
}

// syncReasonParts are the heads of the gap parts errorcheckGapReason writes.
var syncReasonParts = []string{"uncaught markers:", "over-strict:", "internal errors:", "on-chain path "}

// isSyncWrittenReason reports whether r is a reason errorcheckGapReason
// wrote (the umbrella alone, or followed by its parts), in the current
// spelling or the legacy `#6344; …` one; anything else was written by hand.
func isSyncWrittenReason(r string) bool {
	rest, ok := strings.CutPrefix(r, preprocessUmbrella)
	if !ok {
		return false
	}
	rest = strings.TrimSpace(strings.TrimPrefix(rest, ";"))
	if rest == "" {
		return true
	}
	for _, p := range syncReasonParts {
		if strings.HasPrefix(rest, p) {
			return true
		}
	}
	return false
}

// legacySyncReason rewrites the legacy `#6344; …` spelling to the current
// `#6344 …` one; other reasons are returned unchanged.
func legacySyncReason(r string) string {
	if rest, ok := strings.CutPrefix(r, preprocessUmbrella+"; "); ok {
		return preprocessUmbrella + " " + rest
	}
	return r
}

// gcOnlyErrorcheckReason names why an errorcheck file has nothing for
// Gno to check: gc diagnostic or runtime-package flags, a -lang pin below
// go1.17, only GC_ERROR markers, or only pragma-placement markers.
func gcOnlyErrorcheckReason(source []byte, markers []InlineError) string {
	// Positions the harness cannot attribute: `//line` and `/*line*/`
	// directives remap them (the error lines then follow the directive,
	// not the file, and the rescue offset no longer applies), and a byte
	// order mark inside the file is a parser test.
	if bytes.Contains(source, []byte("\n//line ")) {
		return "gc-specific: //line directives remap the positions the markers refer to"
	}
	if bytes.Contains(source, []byte("/*line ")) {
		return "gc-specific: /*line*/ directives remap the positions the markers refer to"
	}
	if bytes.Contains(source[1:], []byte("\ufeff")) {
		return "gc-specific: byte order mark test; the markers sit on the BOM lines"
	}
	if firstLine := strings.SplitN(string(source), "\n", 2)[0]; strings.HasPrefix(firstLine, "//") {
		flags := strings.Fields(strings.TrimPrefix(firstLine, "//"))
		if len(flags) > 0 && flags[0] == "errorcheckwithauto" {
			return "gc optimization-diagnostic errorcheck (-0/-m); markers are compiler diagnostics, not errors"
		}
		for _, fl := range flags[1:] {
			switch {
			case fl == "-0" || fl == "-m":
				return "gc optimization-diagnostic errorcheck (-0/-m); markers are compiler diagnostics, not errors"
			case fl == "-+" || strings.HasPrefix(fl, "-p="):
				return "gc runtime-package errorcheck (-+/-p=); markers assert gc-internal contracts"
			case langBelowPin(fl):
				return "gc -lang version-gating test below Gno's go1.17 pin (" + strings.TrimPrefix(fl, "-lang=") + ")"
			}
		}
	}
	checkable, pragmaOnly := 0, true
	for _, mk := range markers {
		if mk.GcOnly {
			continue
		}
		checkable++
		hit := false
		for _, pat := range mk.Patterns {
			if strings.Contains(pat, "compiler directive") {
				hit = true
			}
		}
		if !hit {
			pragmaOnly = false
		}
	}
	if checkable == 0 {
		return "only gc-specific (GC_ERROR) markers; not part of Gno's contract"
	}
	if pragmaOnly {
		return "gc pragma-placement enforcement; //go: directives are inert comments in Gno"
	}
	return ""
}
