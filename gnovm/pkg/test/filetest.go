package test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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
func (opts *TestOptions) RunFiletest(fname string, source []byte, tgs gno.Store) (string, types.Gas, error) {
	opts.outWriter.w = opts.Output
	opts.outWriter.errW = opts.Error
	tcheck := true // Go type-check filetests in test/files.
	return opts.runFiletest(fname, source, tgs, tcheck)
}

// SkipError is returned by RunFiletest when the source file declares
// a top-level `// Unsupported: <reason>` directive. The walker is
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
func (opts *TestOptions) runFiletest(fname string, source []byte, tgs gno.Store, tcheck bool) (newContent string, gas types.Gas, retErr error) {
	dirs, err := ParseDirectives(bytes.NewReader(source))
	if err != nil {
		return "", 0, fmt.Errorf("error parsing directives: %w", err)
	}

	// Unsupported short-circuit: declared via a top-level
	// `// Unsupported: <reason>` directive. Surfaces as a SkipError
	// the walker turns into t.Skip. Checked before any dispatch /
	// rescue / machine construction so unsupported files cost nothing.
	if u := dirs.First(DirectiveUnsupported); u != nil {
		return "", 0, &SkipError{Reason: u.Content}
	}

	// Capture the `// GnoError:` golden block and `// GnoStaticIncomplete:`
	// golden region now, before any strip or re-parse below loses them
	// (prependRescue re-parses dirs from the stripped source). Used by
	// the errorcheck verdict.
	origDirs := dirs

	// Strip the trailing golden region (`// GnoStaticIncomplete:` /
	// `// GnoError:` / `// GoTypeCheckError:` blocks) from the source
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
	source = []byte(stripTrailingGoldenRegion(string(source)))
	source = append(bytes.TrimRight(source, "\n"), '\n')

	// .go filetests under tests/files/gocorpus/testdata/ are regression tests for
	// files lifted from Go's standard test corpus. Three dispatch modes,
	// detected from source content:
	//
	//   - errorcheck (inline `// ERROR "regex"` markers): legacy loose
	//     marker matching. PKGPATH+synthetic-main rescue applies. The
	//     verdict-inversion `// Divergence:` flow still drives this mode.
	//
	//   - compile-only (not runnable): same legacy flow.
	//
	//   - run (runnable): symmetric Gno-vs-Go comparison. Both runtimes
	//     are run, their stdouts compared. A divergence requires the
	//     blessing triple (// GnoOutput: / // GoOutput: / // Divergence:);
	//     stale or missing-blessing cases FAIL. See finalizeGoRunDivergence
	//     below.
	var errorcheckMarkers []InlineError
	var divergenceReason string // legacy errorcheck/compile inversion path only
	var isGoRunMode bool
	var goStdout string
	// originalSource is the file as-authored, before any in-memory
	// PKGPATH/synthetic-main rescue. Golden writes serialize against
	// THIS, never the rescued source — the rescue transforms must
	// not be persisted (gocorpus files stay upstream-verbatim).
	originalSource := source
	// prependedLines is how many lines the rescue added at the top
	// (0 or 1), used to translate Gno's error line numbers back into
	// original-file coordinates for marker lookup.
	prependedLines := 0
	isGoFile := strings.HasSuffix(fname, ".go")
	// .gno files opt INTO Gno-vs-Go comparison by declaring at least
	// one of `// GoOutput:` / `// GoError:` / `// Divergence:`. Without
	// such a directive, .gno files keep their pure-Gno behavior — the
	// 1600+ existing files are untouched.
	hasGoOptIn := dirs.First(DirectiveGoOutput) != nil ||
		dirs.First(DirectiveGoError) != nil ||
		dirs.First(DirectiveKnownDivergence) != nil
	prependRescue := func() error {
		source = PrependPkgPathIfNeeded(source)
		// We prepended a PKGPATH line iff source now starts with one
		// and the original didn't — detect by-construction rather than
		// sniffing (a file that legitimately ships its own PKGPATH
		// must not be miscounted).
		if bytes.HasPrefix(source, []byte("// PKGPATH:")) &&
			!bytes.HasPrefix(originalSource, []byte("// PKGPATH:")) {
			prependedLines = 1
		}
		// Re-parse: the prepended `// PKGPATH:` line must be visible to
		// the directive parser.
		dirs, err = ParseDirectives(bytes.NewReader(source))
		if err != nil {
			return fmt.Errorf("error re-parsing directives after pkgpath rescue: %w", err)
		}
		if d := dirs.First(DirectiveKnownDivergence); d != nil {
			divergenceReason = d.Content
		}
		return nil
	}
	compileMode := false
	if isGoFile {
		hasErrorDir := dirs.First(DirectiveError) != nil
		hasTypeCheckErrorDir := dirs.First(DirectiveTypeCheckError) != nil
		switch {
		case HasInlineErrorMarkers(source) && !hasErrorDir:
			errorcheckMarkers = ParseInlineErrors(source)
			// A marker tag present but no parseable marker (e.g. the
			// `// ERROR` sits on a full-comment pragma line, as in
			// fixedbugs/issue18882.go) must NOT silently fall through
			// to a pure-Gno run with no marker checking — route to
			// `// Unsupported:` so the gap is pinned and greppable.
			if len(errorcheckMarkers) == 0 {
				reason := "errorcheck marker on a non-code line (e.g. pragma comment); not checkable by the harness"
				if opts.Sync {
					return writeUnsupportedDirective(originalSource, reason), 0, nil
				}
				return "", 0, &SkipError{Reason: reason}
			}
			if err := prependRescue(); err != nil {
				return "", 0, err
			}
		case CorpusDirective(source) == "compile" && !hasErrorDir && !hasTypeCheckErrorDir:
			// `// compile`: gc compiles but never runs it. Preprocess
			// only (regardless of whether it's runnable) and check that
			// Gno + go/types accept it. Takes precedence over run-mode.
			compileMode = true
			if err := prependRescue(); err != nil {
				return "", 0, err
			}
		case !IsRunnable(source) && !hasErrorDir && !hasTypeCheckErrorDir:
			if err := prependRescue(); err != nil {
				return "", 0, err
			}
		default:
			// Runnable .go corpus file: symmetric Gno-vs-Go.
			isGoRunMode = true
		}
	} else if hasGoOptIn && IsRunnable(source) {
		// .gno run-mode opt-in: harness ALSO runs Go and compares.
		// The existing `// Output:` directive remains Gno's pinned
		// golden (handled by the main match loop); the new triple
		// directives drive the Gno-vs-Go finalize at the end.
		isGoRunMode = true
	}
	if isGoRunMode {
		out, _, runErr := runGoToolchain(source)
		if runErr != nil {
			return "", 0, fmt.Errorf("filetest %s: cannot run via Go toolchain "+
				"(install go, or mark file `// Unsupported:`): %w", fname, runErr)
		}
		goStdout = out
	}

	// Legacy verdict-inversion for the errorcheck/compile-only modes
	// (run mode uses the explicit triple via finalizeGoRunDivergence
	// at the end). When `// Divergence:` was set above for those two
	// modes, suppress a match-path error (real divergence) or surface
	// a "stale directive" failure if Gno turned out to match.
	defer func() {
		if isGoRunMode || divergenceReason == "" {
			return
		}
		if retErr == nil {
			retErr = fmt.Errorf("stale `// Divergence: %s` directive: Gno's behavior now matches Go; remove the directive",
				divergenceReason)
			return
		}
		retErr = nil
	}()

	// Sanity check: type-check directives are not available
	// with `gno test` of user packages.
	if !tcheck && dirs.FirstDefault(DirectiveTypeCheckError, "") != "" {
		panic("type-check error directive is only available for gnovm internal test files")
	}

	// Initialize Machine.Context and Machine.Alloc according to the input directives.
	pkgPath := dirs.FirstDefault(DirectivePkgPath, "main")
	coins, err := std.ParseCoins(dirs.FirstDefault(DirectiveSend, ""))
	if err != nil {
		return "", 0, err
	}
	ctx := Context("", pkgPath, coins)
	maxAllocRaw := dirs.FirstDefault(DirectiveMaxAlloc, "0")
	maxAlloc, err := strconv.ParseInt(maxAllocRaw, 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("could not parse MAXALLOC directive: %w", err)
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
	result := opts.runTest(m, pkgPath, fname, source, opslog, tcheck, len(errorcheckMarkers) > 0 || compileMode)

	// Feature gap: a .go corpus file whose Gno error is a "feature not
	// implemented" message (channels, goroutines, generics, imaginary
	// literals, dot imports, …) or an unsupported stdlib import — Gno
	// can't process it at all, so route to `// Unsupported:` (skip)
	// rather than letting it fail or land in KnownIssue. Detected from
	// Gno's ACTUAL error, so prose like "go to" in a comment can't trip
	// it. Auto-written under sync; thereafter skipped pre-dispatch.
	if isGoFile {
		reason := UnsupportedFeatureError(result.Error)
		// Gno's own preprocess can stay lenient about an import it
		// lacks — it may surface an unrelated error first (or none
		// line-mappable) instead of "unknown import path". But if the
		// go/types guard couldn't import a package, the file can't be
		// type-checked in Gno's universe at all, so it's a feature gap,
		// not a divergence to pin. (e.g. alias2.go imports `reflect`,
		// which Gno has no stdlib for; Gno preprocess reports a
		// package-level redeclaration instead, so the gap only shows in
		// the guard's TypeCheckError.)
		//
		// Two guards keep this narrow: (1) only a normal package path
		// counts — an invalid-import-path syntax test like `import
		// "/foo"` also yields "unknown import path" but is a real
		// errorcheck case to pin, not a missing stdlib; (2) a file that
		// already pins its behavior with a native `// Error:` /
		// `// TypeCheckError:` directive is left alone.
		nativePin := dirs.First(DirectiveError) != nil || dirs.First(DirectiveTypeCheckError) != nil
		if reason == "" && !nativePin {
			if imp := UnsupportedImport(result.TypeCheckError); imp != "" && looksLikePackagePath(imp) {
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
				return writeUnsupportedDirective(originalSource, reason),
					m.GasMeter.GasConsumed(), nil
			}
			return "", m.GasMeter.GasConsumed(), &SkipError{Reason: reason}
		}
	}

	// Compile-mode short-circuit. `// compile` files are compiled but
	// never run. The old model dumped Gno's preprocess output straight
	// into `// KnownIssue:`, conflating evidence (what Gno actually
	// said) with the verdict (is Gno buggy or not).
	//
	// New model — three EVIDENCE blocks + one VERDICT:
	//
	//   Evidence (raw per-line output from each checker):
	//     // GnoPreprocessError:  Gno's own preprocess errors
	//     // GoBuildError:        `go build`'s errors (full Go pipeline)
	//     // GoTypeCheckError:    go/types guard's errors
	//
	//   Verdict (one of):
	//     // KnownIssue:          Gno-only rejection — real bug (gc + guard
	//                             accept, Gno rejects)
	//     // KnownDivergence:     both Gno and Go pipelines reject —
	//                             wording/stage differ; not a bug
	//     (none)                  all four accept — file is Clean, no
	//                             golden section written
	//
	// The verdict is auto-classified from evidence; a human can refine
	// KnownDivergence's category note or reclassify to Unsupported when
	// the file's construct is fundamentally not user-relevant on-chain.
	if compileMode {
		gas := m.GasMeter.GasConsumed()
		gnoErr := PerLineErrors(result.Error, prependedLines)
		goTC := PerLineErrors(result.TypeCheckError, prependedLines)
		// Fast path: if Gno accepts AND go/types accepts, the file is
		// Clean — no verdict needed, so no need to run the expensive
		// `go build` (which ~doubles the per-file test time across 467
		// files). We only need go-build evidence to distinguish "real
		// Gno bug" from "both reject" when Gno rejects.
		if len(gnoErr) == 0 && len(goTC) == 0 {
			return "", gas, nil // Clean — everyone accepts
		}
		var goBuildErr map[int]string
		if len(gnoErr) > 0 {
			if _, buildStderr, err := runGoToolchain(source); err == nil {
				goBuildErr = PerLineErrors(buildStderr, prependedLines)
			}
		}
		// Verdict: KnownIssue only when Gno-only rejects; otherwise
		// KnownDivergence (both reject, wording differs).
		verdict := ""
		verdictName := ""
		if len(gnoErr) > 0 {
			if len(goBuildErr) == 0 && len(goTC) == 0 {
				verdictName = DirectiveKnownIssue
				verdict = "TODO: explain the Gno bug (Gno rejects code gc + go/types accept)"
			} else {
				verdictName = DirectiveKnownDivergence
				verdict = "compile-error-wording: both Gno and Go reject; wording/stage differ"
			}
		}
		// Preserve a human-refined verdict across re-sync. Skip preservation
		// when the existing content looks like the LEGACY per-line dump
		// (`line N: ...` form — that was the old harness's evidence-in-
		// verdict format, not human-written). Skip TODO placeholders.
		if verdictName != "" {
			if d := origDirs.First(verdictName); d != nil {
				c := strings.TrimSpace(d.Content)
				if c != "" &&
					!strings.HasPrefix(c, "TODO") &&
					!regexp.MustCompile(`^line \d+:`).MatchString(c) {
					verdict = strings.TrimRight(d.Content, "\n")
				}
			}
		}
		sections := []goldenSection{
			{name: DirectiveGnoPreprocessError, block: FormatGnoErrorBlock(gnoErr)},
			{name: DirectiveGoBuildError, block: FormatGnoErrorBlock(goBuildErr)},
			{name: DirectiveGoTypeCheckError, block: FormatGnoErrorBlock(goTC)},
		}
		if verdictName != "" {
			sections = append(sections, goldenSection{
				name:  verdictName,
				block: verdict,
			})
		}
		newContent, err := opts.resolveErrorcheckGolden(originalSource, origDirs, "", sections)
		return newContent, gas, err
	}

	// updated tells whether the directives have been mutated and the
	// regenerated filetest should be returned (only true under
	// opts.Sync). Declared here so the errorcheck short-circuit below
	// can also set it for `// GnoError:` golden refresh.
	updated := false

	// Errorcheck-mode short-circuit: golden-snapshot multi-pass.
	// The inline `// ERROR` markers are upstream (gc) provenance, NOT a
	// pass/fail gate. The harness walks Gno's per-line errors and pins
	// them in a `// GnoError:` golden block; the test passes iff that
	// block matches Gno's current behavior. So Gno's wording may differ
	// from gc's marker (the known-divergence case) — what's verified is
	// that Gno's behavior hasn't *changed*. See [runErrorcheckMultiPass].
	//
	// A file Gno doesn't reject at all (no error anywhere) is a real
	// leniency divergence — fail, so it's noticed (mark `// Unsupported:`
	// if intentional). Files Gno can't run (feature gaps) are skipped
	// via `// Unsupported:` before dispatch.
	if len(errorcheckMarkers) > 0 {
		// divergenceReason is a compile-only concept; clear it so the
		// verdict-inversion defer doesn't process this branch's return.
		divergenceReason = ""

		gnoErrLines, goTCLines := opts.runErrorcheckMultiPass(
			result, source, fname, pkgPath, errorcheckMarkers,
			prependedLines, tgs, tcheck)
		gas := m.GasMeter.GasConsumed()

		// checkable = the markers that are part of Gno's contract.
		// GC_ERROR markers are gc-implementation diagnostics (backend
		// limits, pragma checks) — catching them is fine, missing them
		// is NOT a leak. A file whose markers are all gc-only has
		// nothing for Gno to verify → Unsupported (skip), same class
		// as the pragma-comment markers.
		var checkable []InlineError
		for _, mk := range errorcheckMarkers {
			if !mk.GcOnly {
				checkable = append(checkable, mk)
			}
		}
		if len(checkable) == 0 {
			reason := "only gc-specific (GC_ERROR) markers; not part of Gno's contract"
			if opts.Sync {
				return writeUnsupportedDirective(originalSource, reason), gas, nil
			}
			return "", gas, &SkipError{Reason: reason}
		}

		// Evidence vs verdict, set-relation model:
		//
		//   Evidence (raw per-line output from each checker):
		//     // GnoError:          Gno's own catches, unfiltered
		//     // GoTypeCheckError:  go/types guard's catches, unfiltered
		//
		//   Verdicts (DERIVED from the sets; no prose block):
		//     // GnoOverStrictError:   Gno lines OUTSIDE markers ∪ guard —
		//                              Gno rejects code both gc and go/types
		//                              accept (over-strict; the deferred-
		//                              KnownIssue equivalent for errorcheck)
		//     // GnoStaticIncomplete:  markers not fully covered by
		//                              Gno ∪ guard (coverage-gap note)
		//
		// Production-safe ⟺ markers ⊆ Gno ∪ guard (both gate deploy).
		marked := make(map[int]bool, len(errorcheckMarkers))
		for _, mk := range errorcheckMarkers {
			marked[mk.Line] = true
		}
		gnoErr := gnoErrLines
		overStrict := make(map[int]string)
		for ln, msg := range gnoErrLines {
			_, inGuard := goTCLines[ln]
			if !marked[ln] && !inGuard {
				overStrict[ln] = msg
			}
		}

		// Leak accounting: a checkable marker caught by NEITHER Gno nor
		// the guard is an UNCAUGHT error — gc-invalid code the whole Gno
		// stack would deploy. Pinned per-line in `// UncaughtError:`
		// (under-rejection; strictly worse than over-strictness).
		uncaught := make(map[int]string)
		for _, mk := range checkable {
			_, a := gnoErr[mk.Line]
			_, b := goTCLines[mk.Line]
			if !a && !b {
				uncaught[mk.Line] = "uncaught; gc expects: " + strings.Join(mk.Patterns, " | ")
			}
		}

		// Evidence blocks are raw; GnoOverStrictError is the derived
		// verdict (its lines repeat the offending GnoError entries so
		// the deferred bucket is greppable). Triage prose lives in plain
		// comments + state markers, exactly like run mode — errorcheck
		// files carry no `// KnownIssue:` block.
		sections := []goldenSection{
			{name: DirectiveGnoError, block: FormatGnoErrorBlock(gnoErr)},
			{name: DirectiveGoTypeCheckError, block: FormatGnoErrorBlock(goTCLines)},
			{name: DirectiveGnoOverStrictError, block: FormatGnoErrorBlock(overStrict)},
			{name: DirectiveUncaughtError, block: FormatGnoErrorBlock(uncaught)},
			// Empty block ⇒ "must be absent": scrubs the legacy per-file
			// KnownIssue verdicts errorcheck files used to carry.
			{name: DirectiveKnownIssue, block: ""},
		}
		// incompleteNote is retired: the leak information lives in the
		// per-line UncaughtError verdict; passing "" scrubs legacy
		// GnoStaticIncomplete tags on sync.
		newContent, err := opts.resolveErrorcheckGolden(originalSource, origDirs, "", sections)
		return newContent, gas, err
	}

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
				return "", m.GasMeter.GasConsumed(), fmt.Errorf("unexpected panic: %s\noutput:\n%s\nstacktrace:\n%s\nstack:\n%v",
					result.Error, result.Output, result.GnoStacktrace, string(result.GoPanicStack))
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
				return "", m.GasMeter.GasConsumed(), fmt.Errorf("unexpected output:\n%s", result.Output)
			}
		}
	} else if !isGoRunMode {
		err = m.CheckEmpty()
		if err != nil {
			return "", m.GasMeter.GasConsumed(), fmt.Errorf("machine not empty after main: %w", err)
		}
		if gno.HasDebugErrors() {
			return "", m.GasMeter.GasConsumed(), fmt.Errorf("got unexpected debug error(s): %v", gno.GetDebugErrors())
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
				return "", m.GasMeter.GasConsumed(), fmt.Errorf("package %q not preprocessed: %s", pkgPath, result.Error)
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
			match(dir, result.TypeCheckError)
		// DirectiveGnoOutput / DirectiveGnoError / DirectiveGoOutput are
		// the run-mode golden region; finalizeGoRunDivergence is their
		// sole authority (presence + content + divergence blessing).
		}
	}

	if !hasTypeCheckErrorDirective && result.TypeCheckError != "" {
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
		return dirs.FileTest(), m.GasMeter.GasConsumed(), returnErr
	}

	return "", m.GasMeter.GasConsumed(), returnErr
}

// goldenSection is one named per-line golden block (`// GnoError:` or
// `// GoTypeCheckError:`) to (re)write at the bottom of an errorcheck file.
type goldenSection struct {
	name  string // directive name, e.g. DirectiveGnoError
	block string // FormatGnoErrorBlock output (may be empty → omitted)
}

// resolveErrorcheckGolden reconciles the trailing golden region — an
// optional `// GnoStaticIncomplete:` tag plus the named per-line blocks —
// against the freshly-computed values. `dirs` is the file's parsed
// directives (to look up the on-disk blocks/tag).
//
// Returns (newContent, err): newContent is non-empty only when sync
// rewrote the file. Non-sync failures report the first diff / missing
// block.
func (opts *TestOptions) resolveErrorcheckGolden(originalSource []byte, dirs Directives, incompleteNote string, sections []goldenSection) (string, error) {
	inc := dirs.First(DirectiveGnoStaticIncomplete)
	allOK := (incompleteNote == "") == (inc == nil) &&
		(inc == nil || strings.TrimRight(inc.Content, "\n") == incompleteNote)
	for _, s := range sections {
		d := dirs.First(s.name)
		ok := (s.block == "" && d == nil) || (d != nil && strings.TrimRight(d.Content, "\n") == s.block)
		allOK = allOK && ok
	}
	if allOK {
		return "", nil
	}
	if opts.Sync {
		return writeErrorcheckGolden(originalSource, incompleteNote, sections), nil
	}
	for _, s := range sections {
		d := dirs.First(s.name)
		switch {
		case s.block != "" && d == nil:
			return "", fmt.Errorf(
				"errorcheck: no `// %s:` block present; re-run with "+
					"--update-golden-tests to record it:\n%s", s.name, indent(s.block, "  "))
		case s.block == "" && d != nil:
			return "", fmt.Errorf("errorcheck: stale `// %s:` block; re-run with --update-golden-tests to remove it", s.name)
		case d != nil && strings.TrimRight(d.Content, "\n") != s.block:
			return "", fmt.Errorf("// %s: diff:\n%s", s.name, unifiedDiff(d.Content, s.block))
		}
	}
	// Only the tag differs.
	return "", fmt.Errorf("errorcheck: `// GnoStaticIncomplete:` tag inconsistent with coverage; re-run with --update-golden-tests")
}

// writeUnsupportedDirective returns originalSource (golden region
// stripped) with a trailing `// Unsupported: <reason>` directive, so a
// file that fails on a Gno-unsupported import is skipped on future
// runs (the directive is detected pre-dispatch) and self-documents why.
func writeUnsupportedDirective(originalSource []byte, reason string) string {
	body := strings.TrimRight(stripTrailingGoldenRegion(string(originalSource)), "\n")
	return body + "\n\n// " + DirectiveUnsupported + ": " + reason + "\n"
}

// writeErrorcheckGolden returns the file content for an errorcheck
// golden update: originalSource (upstream-verbatim — NOT the
// rescued source, so in-memory transforms aren't persisted) with a
// trailing golden region appended — an optional `// GnoStaticIncomplete:`
// tag followed by each non-empty section, blank-line separated. Any
// existing trailing golden region is replaced rather than duplicated.
func writeErrorcheckGolden(originalSource []byte, incompleteNote string, sections []goldenSection) string {
	// Capture triage state marker lines before stripping so a golden
	// refresh never drops manual triage status.
	var stateLines []string
	for _, ln := range strings.Split(string(originalSource), "\n") {
		t := strings.TrimSpace(ln)
		for _, name := range triageStateMarkers {
			if strings.HasPrefix(t, "// "+name+":") {
				stateLines = append(stateLines, t)
			}
		}
	}
	body := strings.TrimRight(stripTrailingGoldenRegion(string(originalSource)), "\n")
	var sb strings.Builder
	sb.WriteString(body)
	sb.WriteString("\n")
	// Canonical marker position: between the code body and the pinned
	// golden region (same place the run-mode writer keeps them).
	for _, ln := range stateLines {
		sb.WriteString("\n")
		sb.WriteString(ln)
		sb.WriteString("\n")
	}
	if incompleteNote != "" {
		sb.WriteString("\n// ")
		sb.WriteString(DirectiveGnoStaticIncomplete)
		sb.WriteString(": ")
		sb.WriteString(incompleteNote)
		sb.WriteString("\n")
	}
	for _, s := range sections {
		if s.block == "" {
			continue
		}
		sb.WriteString("\n// ")
		sb.WriteString(s.name)
		sb.WriteString(":\n")
		for _, line := range strings.Split(s.block, "\n") {
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
	DirectiveGnoStaticIncomplete, DirectiveGnoError, DirectiveGoTypeCheckError, DirectiveKnownIssue,
	DirectiveGnoPreprocessError, DirectiveGoBuildError, DirectiveKnownDivergence,
	DirectiveGnoOverStrictError, DirectiveUncaughtError,
}

// triageStateMarkers are the standalone triage status lines (`// Fixed:`,
// `// Fixing:`, `// Tracked:`) written by manual triage after the golden
// region. They are NOT harness directives (never add them to
// allDirectives) but the golden writer must carry them across re-syncs:
// stripTrailingGoldenRegion treats a trailing marker block as strippable
// and writeErrorcheckGolden re-appends the captured lines at the end.
var triageStateMarkers = []string{"Fixed", "Fixing", "Tracked"}

// stripTrailingGoldenRegion removes the trailing golden region — one or
// more blank-line-separated `//`-comment blocks each beginning with a
// golden directive (GnoStaticIncomplete / GnoError / GoTypeCheckError) — so a
// refresh replaces rather than appends. Returns src unchanged if the
// trailing comment block isn't a golden block.
func stripTrailingGoldenRegion(src string) string {
	lines := strings.Split(src, "\n")
	for {
		end := len(lines)
		for end > 0 && strings.TrimSpace(lines[end-1]) == "" {
			end--
		}
		start := end
		for start > 0 && strings.HasPrefix(strings.TrimSpace(lines[start-1]), "//") {
			start--
		}
		if start >= end {
			break
		}
		top := strings.TrimSpace(lines[start])
		isGolden := false
		for _, name := range goldenRegionStarts {
			if strings.HasPrefix(top, "// "+name+":") {
				isGolden = true
				break
			}
		}
		for _, name := range triageStateMarkers {
			if strings.HasPrefix(top, "// "+name+":") {
				isGolden = true
				break
			}
		}
		if !isGolden {
			break
		}
		lines = lines[:start]
	}
	return strings.Join(lines, "\n")
}

// runErrorcheckMultiPass walks Gno's per-line errors and records them
// into a golden snapshot. The inline `// ERROR` markers are NOT used
// as a pass/fail gate — they're upstream (gc) provenance. The contract
// is "Gno errors at this line, with this wording", captured per line
// so the snapshot detects any change in Gno's behavior.
//
// Each pass:
//  1. Reads Gno's error (preprocess + go/types) from the prior run.
//  2. Extracts the source line, records Gno's clean per-line message.
//  3. Neutralizes that line (package decls → `package main`, else
//     commented out) so the next pass surfaces the next error.
//
// Iteration stops when Gno accepts the (progressively-neutralized)
// file, when it errors on a line WITHOUT a marker (see below), or when
// neutralizing fails to clear a line (cycle guard).
//
// Unmarked-line handling distinguishes signal from artifact:
//   - On pass 1 (original file, no neutralization yet) an unmarked
//     error is Gno's genuine first error — recorded, then stop. This
//     captures files where Gno bails before the marked lines (e.g. a
//     too-large constant the markers don't cover).
//   - On a later pass an unmarked error is almost always a neutralize
//     artifact (a commented-out func signature orphaning its body) —
//     stop WITHOUT recording, to keep the golden clean.
//
// Pass 1 reuses `initial`; later passes spin up a fresh machine each
// (Gno's preprocess state is not idempotent across runs).
//
// prependedLines is how many lines the PKGPATH rescue added at the top
// (0 or 1) — used to translate Gno's line numbers to source coords.
func (opts *TestOptions) runErrorcheckMultiPass(
	initial runResult, source []byte, fname, pkgPath string,
	markers []InlineError, prependedLines int,
	tgs gno.Store, tcheck bool,
) (gnoErrLines, goTCLines map[int]string) {
	gnoErrLines = make(map[int]string) // GnoVM's own preprocess/runtime errors
	goTCLines = make(map[int]string)   // the go/types guard's errors
	seen := make(map[int]bool)
	markerByLine := make(map[int]InlineError, len(markers))
	for _, mk := range markers {
		markerByLine[mk.Line] = mk
	}

	// recordGoTypes folds the go/types guard's per-line catches for any
	// not-yet-covered marker into goTCLines. go/types reports ALL its
	// errors in one pass (no first-error bail) and is run every pass —
	// some catches only surface AFTER neutralization (e.g. a method on a
	// non-local type, reachable only once an invalid `package _` is
	// rewritten to `package main`). First pass to catch a line wins
	// (the initial, un-neutralized run is most authoritative for the
	// lines it does reach).
	recordGoTypes := func(r runResult) {
		tcSegs := gnoErrSegments(r.TypeCheckError)
		for _, mk := range markers {
			if _, done := goTCLines[mk.Line]; done {
				continue
			}
			gnoLn := mk.Line + prependedLines
			if segHasLine(tcSegs, gnoLn) {
				m := mk
				goTCLines[mk.Line] = errorForLine(nil, tcSegs, gnoLn, &m)
			}
		}
	}

	// GnoVM preprocess iteration — Gno's OWN static errors. Preprocess
	// bails on the first error, so neutralize that line and re-run to
	// surface the next (package decls → `package main`, else commented
	// out). Each pass also sweeps the go/types guard (recordGoTypes).
	// GnoVM errors go to gnoErrLines (the `// GnoError:` block); an
	// internal "should not happen" assertion is not a real diagnostic,
	// so it's skipped there (the real error, if any, is go/types').
	// The walk continues PAST unmarked (over-strict) errors so Gno's
	// behavior on every marker line is still measured — a single
	// over-strict reject must not mask the file's whole native-coverage
	// picture. Neutralizing a line removes any declarations it carried,
	// so later "name X not declared"-style errors that merely reference
	// a neutralized declaration are CASCADE ARTIFACTS: they are
	// neutralized and skipped, never pinned (the affected marker line
	// simply stays uncovered by Gno and shows up in the coverage note).
	artifactIDs := map[string]bool{}
	isArtifact := func(msg string) bool {
		m := reUndeclaredIdent.FindStringSubmatch(msg)
		if m == nil {
			return false
		}
		id := m[1]
		if id == "" {
			id = m[2]
		}
		return artifactIDs[id]
	}
	currentSource := source
	currentPkgPath := pkgPath
	result := initial
	for pass := 1; pass <= 2*len(markers)+8; pass++ {
		recordGoTypes(result)

		gnoLine := ExtractErrorLine(result.Error)
		if gnoLine == 0 {
			return gnoErrLines, goTCLines // no preprocess error left
		}
		sourceLine := gnoLine - prependedLines
		if seen[sourceLine] {
			return gnoErrLines, goTCLines // neutralizing didn't clear it; cycle guard
		}
		seen[sourceLine] = true

		errSegs := gnoErrSegments(result.Error)
		marker, marked := markerByLine[sourceLine]
		var mkPtr *InlineError
		if marked {
			mkPtr = &marker
		}
		if _, done := gnoErrLines[sourceLine]; !done {
			if msg := errorForLine(errSegs, nil, gnoLine, mkPtr); msg != "" &&
				!internalNoise(msg) && !isArtifact(msg) {
				gnoErrLines[sourceLine] = msg
			}
		}

		// Neutralize and continue. gnoLine is post-prepend coords,
		// which is what indexes into currentSource. Remember the
		// identifiers the neutralized line declared, so their cascade
		// errors are recognized as artifacts on later passes.
		for _, id := range declaredIdents(lineAt(currentSource, gnoLine)) {
			artifactIDs[id] = true
		}
		var wasPkg bool
		currentSource, wasPkg = NeutralizeLine(currentSource, gnoLine)
		if wasPkg {
			currentPkgPath = "main"
		}
		result = opts.runErrorcheckPass(currentSource, fname, currentPkgPath, tgs, tcheck)
	}
	return gnoErrLines, goTCLines
}

// reUndeclaredIdent matches Gno's undefined-reference wordings and
// captures the identifier — used to recognize cascade artifacts after a
// declaration-carrying line was neutralized.
var reUndeclaredIdent = regexp.MustCompile(`(?:name (\w+) not (?:declared|defined)|undefined: (\w+))`)

// lineAt returns the 1-indexed line of src, or "" when out of range.
func lineAt(src []byte, line int) string {
	lines := strings.Split(string(src), "\n")
	if line < 1 || line > len(lines) {
		return ""
	}
	return lines[line-1]
}

// declaredIdents extracts the identifiers a source line declares:
// `const/var/type/func NAME ...`, a keywordless const/var-block entry
// (`NAME = ...`, `NAME TYPE = ...`, or a bare iota-continuation `NAME`),
// including comma-separated name lists. Heuristic by design — it feeds
// artifact suppression, where over-matching a name only risks skipping a
// cascade error that was garbage anyway.
func declaredIdents(line string) []string {
	s := strings.TrimSpace(line)
	if i := strings.Index(s, "//"); i >= 0 {
		s = strings.TrimSpace(s[:i])
	}
	for _, kw := range []string{"const ", "var ", "type ", "func "} {
		if rest, ok := strings.CutPrefix(s, kw); ok {
			s = rest
			break
		}
	}
	m := regexp.MustCompile(`^([A-Za-z_]\w*(?:\s*,\s*[A-Za-z_]\w*)*)\s*(?:=|$|[([{ \t])`).FindStringSubmatch(s)
	if m == nil {
		return nil
	}
	var ids []string
	for _, id := range strings.Split(m[1], ",") {
		id = strings.TrimSpace(id)
		if id != "" && id != "_" {
			ids = append(ids, id)
		}
	}
	return ids
}

// runErrorcheckPass executes one fresh-machine pass for the
// errorcheck multi-pass driver. Each pass needs its own machine and
// transaction store because Gno's preprocess records package state
// that would otherwise carry over between passes (and re-surface as
// a different error than the source line the caller expects).
func (opts *TestOptions) runErrorcheckPass(source []byte, fname, pkgPath string, tgs gno.Store, tcheck bool) runResult {
	tcw := opts.BaseStore.CacheWrap()
	gasMeter := store.NewInfiniteGasMeter()
	m := gno.NewMachineWithOptions(gno.MachineOptions{
		Output:        &opts.outWriter,
		Store:         tgs.BeginTransaction(tcw, tcw, nil, gasMeter),
		Context:       Context("", pkgPath, nil),
		GasMeter:      gasMeter,
		Debug:         opts.Debug,
		ReviveEnabled: true,
	})
	defer m.Release()
	// Errorcheck passes are always preprocess-only (no main execution).
	return opts.runTest(m, pkgPath, fname, source, nil, tcheck, true)
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
//   - // Divergence: present iff the two sides differ; its reason text
//     blesses the diff (a human-written reason is preserved on re-sync).
//
// Gno's panic lands in gnoError, not gnoStdout; it is folded into the
// comparison to mirror Go's combined stream — otherwise a Gno panic with
// no prior stdout would compare equal to a clean Go run (both "") and
// silently pass a real bug. (bug446.go: Gno mis-orders package init and
// panics where Go exits 0.)
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

	if sync {
		// Rebuild the region: drop the old golden directives, re-append
		// the freshly-computed ones, then the verdict if the sides
		// differ (preserving the human's tag + reason).
		out := removeDirectives(dirs,
			DirectiveGnoOutput, DirectiveGnoError, DirectiveGoOutput,
			DirectiveKnownIssue, DirectiveKnownDivergence)
		out = append(out, want...)
		if diverges {
			name, reason := runVerdict(dirs, gnoErr != "")
			out = append(out, Directive{Name: name, Content: reason, Complete: true})
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
	hasVerdict := dirs.First(DirectiveKnownIssue) != nil || dirs.First(DirectiveKnownDivergence) != nil
	switch {
	case diverges && !hasVerdict:
		prior = multierr.Append(prior, fmt.Errorf(
			"Gno-vs-Go divergence detected — classify as `// KnownIssue:` (a Gno bug) or "+
				"`// KnownDivergence:` (accepted), or re-run with --update-golden-tests:\n%s",
			unifiedDiff(goExp, gnoCombined)))
	case !diverges && hasVerdict:
		prior = multierr.Append(prior, fmt.Errorf(
			"stale verdict: Gno's output now matches Go's — remove the `// KnownIssue:` / "+
				"`// KnownDivergence:` directive (or re-run with --update-golden-tests)"))
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
func runVerdict(dirs Directives, gnoErrored bool) (name, reason string) {
	for _, n := range []string{DirectiveKnownIssue, DirectiveKnownDivergence} {
		if d := dirs.First(n); d != nil {
			if r := strings.TrimSpace(d.Content); r != "" && !strings.HasPrefix(r, "TODO") {
				return n, d.Content
			}
		}
	}
	if gnoErrored {
		return DirectiveKnownIssue, "TODO: explain the Gno bug (Gno errors where Go runs clean)"
	}
	return DirectiveKnownDivergence, "TODO: <category>: explain why this divergence is acceptable"
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
func realmDiffsString(m map[string]int64) string {
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
	pkgName := gno.Name(pkgPath[strings.LastIndexByte(pkgPath, '/')+1:])
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
		// Run decls and init functions.
		m.RunMemPackage(mpkg, true)

		// Clear store cache and reconstruct machine from committed info
		// (mimicking on-chain behaviour).
		// (jae) why is this needed?
		txs.Write()
		m.Store = orig
		pv2 := m.Store.GetPackage(pkgPath, false)
		m.SetActivePackage(pv2)
		m.Context.(*teststdlibs.TestExecContext).OriginCaller = DefaultCaller
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

	// Single-line PascalCase meta-directives that short-circuit the
	// match logic. Reason is the single-line text after the colon.
	DirectiveUnsupported = "Unsupported"
	// DirectiveKnownDivergence blesses an ACCEPTED run-mode difference
	// (Gno's output legitimately differs from Go's — formatting, map
	// order, error wording, …). It is NOT a bug: a real Gno bug is a
	// `// KnownIssue:` instead. Single-line `<category>: <reason>`.
	DirectiveKnownDivergence = "KnownDivergence"
	// DirectiveGnoStaticIncomplete tags an errorcheck file whose golden
	// covers only SOME of its inline markers: Gno bailed (or stayed
	// lenient) before the rest were reached. STATIC-only — errorcheck
	// files are preprocess-only, never run, so there's no runtime
	// dimension. The file still passes (its golden is pinned); the tag
	// flags it as a candidate for a future runnable variant that would
	// exercise the unreached markers. Auto-written under
	// --update-golden-tests.
	DirectiveGnoStaticIncomplete = "GnoStaticIncomplete"

	// Symmetric Gno-vs-Go golden directives for .go corpus files.
	// They mirror existing native directives but split the actual
	// outputs by source so a reader sees both sides structurally:
	//
	//   // GnoOutput: <Gno's actual stdout>
	//   // GoOutput:  <`go run`'s actual stdout>
	//   // Divergence: <category>: <reason>
	//
	// Same for errors (// GnoError: / // GoError:). Both are
	// auto-writable via `--update-golden-tests`; `// Divergence:`
	// gets an auto-filled TODO placeholder the contributor refines.
	DirectiveGnoOutput = "GnoOutput"
	DirectiveGoOutput  = "GoOutput"
	DirectiveGnoError  = "GnoError"
	DirectiveGoError   = "GoError"
	// DirectiveGoTypeCheckError pins the per-line errors from the
	// go/types guard (the Go type checker gno.land's deploy gate runs
	// ahead of GnoVM preprocess). Kept separate from `// GnoError:`
	// (which is Gno's OWN static/runtime behavior) because go/types is
	// not Gno — even when GnoVM preprocess is permissive, this guard
	// still rejects, and that's worth pinning on its own.
	DirectiveGoTypeCheckError = "GoTypeCheckError"
	// DirectiveGnoOverStrictError pins, on errorcheck files, Gno's
	// per-line preprocess errors on lines that neither the gc markers
	// nor the go/types guard flag — the evidence of Gno-only
	// (over-strict) rejection backing a `// KnownIssue:` verdict.
	DirectiveGnoOverStrictError = "GnoOverStrictError"
	// DirectiveUncaughtError pins, on errorcheck files, the checkable
	// (non-GC_ERROR) marker lines caught by NEITHER Gno's preprocess
	// nor the go/types guard — gc-invalid code the Gno stack would
	// deploy (a leak / under-rejection; worse than over-strictness).
	DirectiveUncaughtError = "UncaughtError"
	// DirectiveGnoPreprocessError pins Gno's per-line preprocess errors
	// in compile-mode files. Evidence-only — the fact that Gno rejects
	// these lines. The VERDICT (bug vs benign divergence) is separate:
	// see `// KnownIssue:` / `// KnownDivergence:`. Previously the
	// harness dumped Gno's preprocess output into `// KnownIssue:`
	// itself, conflating evidence with verdict.
	DirectiveGnoPreprocessError = "GnoPreprocessError"
	// DirectiveGoBuildError pins the Go toolchain's `go build` output
	// for compile-mode files — evidence of whether the file's rejection
	// is Gno-only or shared with the Go build/link pipeline. Populated
	// automatically alongside GoTypeCheckError. Distinguishes real
	// Gno-only bugs from cases where both Gno and Go reject the file
	// with different wording (a benign KnownDivergence, not a bug).
	DirectiveGoBuildError = "GoBuildError"
	// DirectiveKnownIssue marks a Gno BUG — Gno disagrees with reality
	// and it matters. Two shapes share the tag:
	//   - errorcheck/compile (static): a per-line block of Gno errors at
	//     lines carrying NO gc marker — Gno rejects code gc accepts
	//     (over-strict).
	//   - run (runtime): a free-text reason — Gno's run result diverges
	//     from Go's in a way that's a bug, not an accepted difference
	//     (contrast `// KnownDivergence:`). e.g. bug446: Gno panics where
	//     Go exits 0.
	// The file still passes (the behavior is pinned); the tag makes the
	// bug greppable. Urgency (fix-now vs defer) is DERIVED by the ledger
	// from the facts, not stored here.
	DirectiveKnownIssue = "KnownIssue"
)

var allDirectives = []string{
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
	DirectiveUnsupported,
	DirectiveKnownDivergence,
	DirectiveGnoStaticIncomplete,
	DirectiveGnoOutput,
	DirectiveGoOutput,
	DirectiveGnoError,
	DirectiveGoError,
	DirectiveGoTypeCheckError,
	DirectiveGnoOverStrictError,
	DirectiveUncaughtError,
	DirectiveGnoPreprocessError,
	DirectiveGoBuildError,
	DirectiveKnownIssue,
}

// singleLinePascalDirectives holds PascalCase directives whose content
// is ALWAYS single-line — even when empty — and that must be parsed
// without the bare-PascalCase multi-line absorbing behavior. Members
// (currently `Unsupported`, `Divergence`) carry one-line reason text.
//
// Other PascalCase directives (Output / Error / GnoOutput / GoOutput /
// GnoError / GoError / Realm / …) are multi-line markers by default,
// matching the .gno convention: directive on its own line, then
// `//`-prefixed content lines, terminated by a blank line or end of
// file. Inline-content single-line form (`// Output: foo`) is also
// accepted via the same parser path — see ParseDirectives.
var singleLinePascalDirectives = map[string]bool{
	DirectiveUnsupported:         true,
	DirectiveGnoStaticIncomplete: true,
	// KnownDivergence is deliberately NOT here: like KnownIssue it is a
	// verdict that carries multi-line human prose, so a bare
	// `// KnownDivergence:` must absorb its following comment lines.
	// (When it was single-line, a multi-line human verdict parsed as an
	// EMPTY directive + orphan comment, and golden sync clobbered the
	// human's classification back to an auto KnownIssue.)
}

// pinnedGoldenDirectives lists PascalCase directives whose empty
// content is meaningful and must still be serialized (rather than
// skipped). `// GoOutput:` with no lines means "Go produces no stdout"
// — a pinned-golden assertion we want visible in the file.
var pinnedGoldenDirectives = map[string]bool{
	DirectiveGnoOutput:          true,
	DirectiveGoOutput:           true,
	DirectiveGnoError:           true,
	DirectiveGoError:            true,
	DirectiveGnoPreprocessError: true,
	DirectiveGoBuildError:       true,
	DirectiveGoTypeCheckError:   true,
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
// one line. Used by meta-directives like `Unsupported:` /
// `Divergence:` / `GnoOutput:` / `GoOutput:` whose payload is a short
// single-line value (often empty for `GoOutput:` when Go produces no
// stdout). PascalCase distinguishes them from the ALLCAPS
// input-parameter family (PKGPATH/MAXALLOC/SEND); single-line-vs-
// multi-line discrimination is by the directive name's membership in
// [singleLinePascalDirectives] — without that set, the parser's
// PascalCase-bare rule would absorb the next comment line as the
// directive's content.
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
		//   2. Inline content is non-empty — `// Name: foo` for any
		//      known PascalCase directive collapses to single-line
		//      form. Bare `// Name:` falls through to the multi-line
		//      marker rule below.
		// Checked before reDirectiveLine so multi-line absorption
		// doesn't eat the next comment line for case 1.
		if subm2 := reDirectiveSingleLinePascal.FindStringSubmatch(comment); subm2 != nil {
			name := subm2[1]
			content := subm2[2]
			isSingle := singleLinePascalDirectives[name] ||
				(content != "" && slices.Contains(allDirectives, name))
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
// Stdout and stderr are combined because Gno's runtime exposes one
// output stream — comparing only Go's stdout would flag the builtin
// `println` (Go: stderr; Gno: stdout) as a divergence purely from
// the stream choice, not a real semantic difference. Combining puts
// the comparison on the same footing as `go test`'s default.
//
// Non-zero exit (panic, compile error) is NOT a function-level error
// — it's the corpus file's expected behavior. The only error returned
// is a genuine exec failure (Go toolchain not on PATH, timeout).
func runGoToolchain(source []byte) (output, stderr string, err error) {
	dir, mkErr := os.MkdirTemp("", "gno-filetest-go-*")
	if mkErr != nil {
		return "", "", fmt.Errorf("mkdir temp: %w", mkErr)
	}
	defer os.RemoveAll(dir)
	srcPath := filepath.Join(dir, "main.go")
	if err := os.WriteFile(srcPath, source, 0o644); err != nil {
		return "", "", fmt.Errorf("write temp source: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), goSubprocessTimeout)
	defer cancel()

	// `go build foo.go` (single-file form) requires package main.
	// For non-main shapes (errorcheck/compile-only files declaring
	// `package p`), `go build .` on the temp directory works — BUT
	// requires a go.mod, otherwise Go complains "go.mod not found"
	// before it even parses the source. Write a minimal go.mod so we
	// get the real compile errors (e.g. "missing function body").
	if !IsRunnable(source) {
		modPath := filepath.Join(dir, "go.mod")
		if err := os.WriteFile(modPath, []byte("module gnofiletest\n\ngo 1.17\n"), 0o644); err != nil {
			return "", "", fmt.Errorf("write temp go.mod: %w", err)
		}
	}
	var args []string
	if IsRunnable(source) {
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
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	runErr := cmd.Run()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return "", "", fmt.Errorf("`go %s` exceeded %s timeout — mark the file with `// Unsupported:`",
			args[0], goSubprocessTimeout)
	}
	// Combine stdout and stderr into one user-visible stream — Gno's
	// runtime has a single output channel, so this puts comparison on
	// the same footing. Stderr is also returned separately for callers
	// that want to distinguish (e.g. errorcheck/compile modes that
	// look at compiler diagnostics specifically).
	combined := outBuf.String() + errBuf.String()
	output = strings.TrimRight(combined, "\n")
	stderr = strings.TrimRight(errBuf.String(), "\n")
	if runErr != nil {
		var ee *exec.ExitError
		if !errors.As(runErr, &ee) {
			return "", "", fmt.Errorf("`go %s` failed (is `go` on PATH?): %w", args[0], runErr)
		}
		// Non-zero exit is expected for errorcheck / panic-class
		// corpus files; stderr carries the diagnostics.
	}
	return output, stderr, nil
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
