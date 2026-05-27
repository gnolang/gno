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
	isGoFile := strings.HasSuffix(fname, ".go")
	// .gno files opt INTO Gno-vs-Go comparison by declaring at least
	// one of `// GoOutput:` / `// GoError:` / `// Divergence:`. Without
	// such a directive, .gno files keep their pure-Gno behavior — the
	// 1600+ existing files are untouched.
	hasGoOptIn := dirs.First(DirectiveGoOutput) != nil ||
		dirs.First(DirectiveGoError) != nil ||
		dirs.First(DirectiveDivergence) != nil
	if isGoFile {
		hasErrorDir := dirs.First(DirectiveError) != nil
		hasTypeCheckErrorDir := dirs.First(DirectiveTypeCheckError) != nil
		switch {
		case HasInlineErrorMarkers(source) && !hasErrorDir:
			errorcheckMarkers = ParseInlineErrors(source)
			source = PrependPkgPathIfNeeded(source)
			// Re-parse: PrependPkgPathIfNeeded may have prepended a
			// `// PKGPATH:` line, which the directive parser needs to see.
			dirs, err = ParseDirectives(bytes.NewReader(source))
			if err != nil {
				return "", 0, fmt.Errorf("error re-parsing directives after pkgpath rescue: %w", err)
			}
			if d := dirs.First(DirectiveDivergence); d != nil {
				divergenceReason = d.Content
			}
		case !IsRunnable(source) && !hasErrorDir && !hasTypeCheckErrorDir:
			source = PrependPkgPathIfNeeded(source)
			dirs, err = ParseDirectives(bytes.NewReader(source))
			if err != nil {
				return "", 0, fmt.Errorf("error re-parsing directives after pkgpath rescue: %w", err)
			}
			if d := dirs.First(DirectiveDivergence); d != nil {
				divergenceReason = d.Content
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
	result := opts.runTest(m, pkgPath, fname, source, opslog, tcheck)

	// updated tells whether the directives have been mutated and the
	// regenerated filetest should be returned (only true under
	// opts.Sync). Declared here so the errorcheck short-circuit below
	// can also set it for `// GnoError:` golden refresh.
	updated := false

	// Errorcheck-mode short-circuit: drive the multi-pass marker
	// verifier. Pass 0 reuses the initial result above; subsequent
	// passes create their own machines (Gno's preprocess state is
	// not idempotent across runs). See [runErrorcheckMultiPass].
	//
	// A file is considered opted-in to strict mode if it carries a
	// `// GnoError:` golden block. Opt-in files propagate multi-pass
	// failures as test errors and refresh the block under opts.Sync.
	// Files without the block fall back to the legacy single-pass
	// loose marker check (preserves backward compatibility for the
	// upstream-verbatim corpus files that haven't been blessed yet).
	if len(errorcheckMarkers) > 0 {
		prependedLines := 0
		if bytes.HasPrefix(source, []byte("// PKGPATH:")) {
			prependedLines = 1
		}
		perLine, mpFailures := opts.runErrorcheckMultiPass(
			result, source, fname, pkgPath, errorcheckMarkers,
			prependedLines, tgs, tcheck)

		errDir := dirs.First(DirectiveGnoError)
		strict := errDir != nil
		expectedBlock := FormatGnoErrorBlock(perLine)

		switch {
		case strict && len(mpFailures) > 0:
			return "", m.GasMeter.GasConsumed(), errors.Join(mpFailures...)
		case strict:
			// Strict mode: verify the pinned block matches multi-pass output.
			if strings.TrimRight(errDir.Content, "\n") != expectedBlock {
				if opts.Sync {
					errDir.Content = expectedBlock
					updated = true
				} else {
					return "", m.GasMeter.GasConsumed(),
						fmt.Errorf("// GnoError: diff:\n%s",
							unifiedDiff(errDir.Content, expectedBlock))
				}
			}
		case opts.Sync && expectedBlock != "":
			// No block today, but a clean multi-pass succeeded under
			// --update-golden-tests: opt the file in.
			dirs = append(dirs, Directive{Name: DirectiveGnoError, Content: expectedBlock})
			updated = true
		default:
			// Loose-mode fallback: at least one marker must match
			// Gno's error. Preserves the upstream-verbatim corpus
			// files that don't yet have a // GnoError: block.
			if err := VerifyErrorcheckMarkers(errorcheckMarkers, result.Error, result.TypeCheckError); err != nil {
				return "", m.GasMeter.GasConsumed(), err
			}
		}
		if updated {
			return dirs.FileTest(), m.GasMeter.GasConsumed(), nil
		}
		return "", m.GasMeter.GasConsumed(), nil
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
		case DirectiveGnoOutput:
			match(dir, trimTrailingSpaces(result.Output))
		case DirectiveGoOutput:
			match(dir, goStdout)
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
		newDirs, returnErr = finalizeGoRunDivergence(dirs, result.Output, goStdout, returnErr, opts.Sync, isGoFile)
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

// runErrorcheckMultiPass iteratively verifies every inline `// ERROR`
// marker in source. Each pass:
//
//  1. Reads Gno's error from the prior run (or runs fresh).
//  2. Extracts the source line Gno errored on.
//  3. Locates the inline marker at that line and verifies its regex
//     matches Gno's combined error+typecheck output (loose match —
//     same lenience as single-pass, just applied per-line).
//  4. Comments out the matched line in the source so the next pass
//     surfaces the next error.
//
// Pass 0 reuses the `result` already produced by runFiletest's
// single setup-and-run; subsequent passes spin up a fresh machine
// each (Gno's preprocess state is not idempotent across runs).
//
// Returns a per-marker-line map of cleaned Gno error text (used to
// populate the `// GnoError:` golden block) and any per-pass
// failures: marker pattern mismatches, unexpected lines, stuck
// cases (commenting-out didn't clear the error), or markers left
// unverified after all passes.
//
// prependedLines is the count of lines [PrependPkgPathIfNeeded]
// added at the top of source (currently 0 or 1) — used to translate
// Gno's error line numbers back into original-marker coordinates.
func (opts *TestOptions) runErrorcheckMultiPass(
	initial runResult, source []byte, fname, pkgPath string,
	markers []InlineError, prependedLines int,
	tgs gno.Store, tcheck bool,
) (map[int]string, []error) {
	perLine := make(map[int]string)
	seen := make(map[int]bool)
	var failures []error

	currentSource := source
	result := initial
	for pass := 1; pass <= len(markers)+1; pass++ {
		combined := strings.TrimSpace(result.Error + "\n" + result.TypeCheckError)
		if combined == "" {
			break // Gno accepted the (progressively-neutralized) file.
		}
		// Pick the line from Gno's preprocess/runtime error first;
		// fall back to the typecheck error if Gno didn't report one.
		gnoLine := ExtractErrorLine(result.Error)
		if gnoLine == 0 {
			gnoLine = ExtractErrorLine(result.TypeCheckError)
		}
		if gnoLine == 0 {
			failures = append(failures, fmt.Errorf(
				"errorcheck pass %d: Gno errored but no source line found in:\n%s",
				pass, indent(combined, "  ")))
			break
		}
		sourceLine := gnoLine - prependedLines
		if seen[sourceLine] {
			failures = append(failures, fmt.Errorf(
				"errorcheck pass %d: stuck at line %d (commenting-out didn't clear it):\n%s",
				pass, sourceLine, indent(combined, "  ")))
			break
		}
		seen[sourceLine] = true

		var marker *InlineError
		for i := range markers {
			if markers[i].Line == sourceLine {
				marker = &markers[i]
				break
			}
		}
		if marker == nil {
			failures = append(failures, fmt.Errorf(
				"errorcheck pass %d: Gno errored at line %d but no // ERROR marker is attached there:\n%s",
				pass, sourceLine, indent(combined, "  ")))
			break
		}
		if !MarkerMatches(*marker, combined) {
			failures = append(failures, fmt.Errorf(
				"errorcheck pass %d: marker at line %d patterns %s don't match Gno error:\n%s",
				pass, sourceLine, strings.Join(marker.Patterns, " | "), indent(combined, "  ")))
			break
		}
		// Store just Gno's preprocess error for the // GnoError:
		// golden — go/types' output (which lists ALL gc errors at
		// once) would pollute the per-line entry with messages
		// about other lines.
		perLine[sourceLine] = CleanErrorMessage(result.Error)

		// Comment out the matched line so the next pass surfaces the
		// next error. Use gnoLine (post-prepend coordinates) — that's
		// what indexes into currentSource.
		currentSource = CommentOutLine(currentSource, gnoLine)

		// Run the next pass, unless we've verified everything.
		if len(perLine) >= len(markers) {
			break
		}
		result = opts.runErrorcheckPass(currentSource, fname, pkgPath, tgs, tcheck)
	}

	// Surface markers that never matched.
	for _, mk := range markers {
		if _, ok := perLine[mk.Line]; !ok {
			failures = append(failures, fmt.Errorf(
				"errorcheck: marker at line %d (%s) never matched any Gno error",
				mk.Line, strings.Join(mk.Patterns, " | ")))
		}
	}
	return perLine, failures
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
	return opts.runTest(m, pkgPath, fname, source, nil, tcheck)
}

// finalizeGoRunDivergence drives the symmetric Gno-vs-Go verdict for
// a run-mode file (either a .go corpus file or a .gno file with the
// Go-comparison opt-in) once both outputs are known.
//
//   - With // Divergence: present, the contributor has blessed a real
//     divergence. The pinned Gno-side and // GoOutput: directives are
//     match-checked by the main loop above; here we additionally
//     verify gnoStdout != goStdout — equal outputs mean Gno now
//     matches Go and the blessing is stale.
//   - With no // Divergence:, the outputs must match. A diff is a real
//     regression: under opts.Sync the harness auto-appends the triple
//     so the contributor only edits the category + reason. Otherwise
//     the diff is returned as an error pointing at the same workflow.
//
// Gno-side pinned directive: .go files use new // GnoOutput:; .gno
// files reuse existing // Output: (per the long-standing convention).
// isGoFile selects between them for auto-append.
//
// Returns (newDirs, err). newDirs is non-nil only when sync mode wrote
// or removed directives — the caller uses it to regenerate the file.
func finalizeGoRunDivergence(dirs Directives, gnoOutput, goStdout string, prior error, sync, isGoFile bool) (Directives, error) {
	gnoStdout := strings.TrimRight(trimTrailingSpaces(gnoOutput), "\n")
	goExp := strings.TrimRight(goStdout, "\n")
	diverges := gnoStdout != goExp

	gnoDir := DirectiveOutput
	if isGoFile {
		gnoDir = DirectiveGnoOutput
	}

	dDir := dirs.First(DirectiveDivergence)
	if dDir != nil {
		// Blessing path. Pinned directives already checked in the main
		// match loop; surface their mismatches via `prior`.
		if !diverges {
			if sync {
				// Stale blessing: remove the divergence triple. For
				// .gno files keep the long-standing // Output:
				// (drops only the new triple, not Gno's golden).
				toRemove := []string{DirectiveGoOutput, DirectiveDivergence}
				if isGoFile {
					toRemove = append(toRemove, DirectiveGnoOutput)
				}
				return removeDirectives(dirs, toRemove...), prior
			}
			stale := fmt.Errorf("stale `// Divergence: %s` directive: Gno's output now matches Go's; remove the divergence triple",
				dDir.Content)
			return nil, multierr.Append(prior, stale)
		}
		return nil, prior
	}

	// No blessing. Pure pass/fail on actual outputs.
	if !diverges {
		return nil, prior
	}
	if sync {
		new := append(Directives{}, dirs...)
		if isGoFile {
			// Pin Gno's side via new // GnoOutput:. .gno files reuse
			// the existing // Output: directive (added by the main
			// match loop's sync path), so nothing to do here.
			new = append(new, Directive{Name: gnoDir, Content: gnoStdout})
		}
		new = append(new,
			Directive{Name: DirectiveGoOutput, Content: goExp},
			// Empty reason is technically valid (the directive is a
			// boolean blessing); the TODO placeholder nudges the
			// contributor to pick a category from the advisory
			// lexicon in tests/README.md and explain why the
			// divergence is acceptable. The diff itself is already
			// visible in the two output directives above, so the
			// reason text is for the *why*, not the *what*.
			Directive{
				Name:     DirectiveDivergence,
				Content:  "TODO: <category>: explain why this divergence is acceptable",
				Complete: true,
			},
		)
		return new, prior
	}
	diff := unifiedDiff(goExp, gnoStdout)
	return nil, multierr.Append(prior, fmt.Errorf(
		"Gno-vs-Go divergence detected — bless with the divergence triple, or re-run with --update-golden-tests to auto-append:\n%s",
		diff))
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

func (opts *TestOptions) runTest(m *gno.Machine, pkgPath, fname string, content []byte, opslog io.Writer, tcheck bool) (rr runResult) {
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
		m.RunMain()
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
		m.RunMainMaybeCrossing()
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
	DirectiveDivergence  = "Divergence"

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
	DirectiveDivergence,
	DirectiveGnoOutput,
	DirectiveGoOutput,
	DirectiveGnoError,
	DirectiveGoError,
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
	DirectiveUnsupported: true,
	DirectiveDivergence:  true,
}

// pinnedGoldenDirectives lists PascalCase directives whose empty
// content is meaningful and must still be serialized (rather than
// skipped). `// GoOutput:` with no lines means "Go produces no stdout"
// — a pinned-golden assertion we want visible in the file.
var pinnedGoldenDirectives = map[string]bool{
	DirectiveGnoOutput: true,
	DirectiveGoOutput:  true,
	DirectiveGnoError:  true,
	DirectiveGoError:   true,
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

	// Remove trailing (newline|space)* and filter empty directives.
	result := make([]Directive, 0, len(parsed))
	parsed = parsed[1:] // remove faux directive
	for _, dir := range parsed {
		content := dir.Content
		content = strings.TrimRight(content, "\n ")
		if content == "" {
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
	// `package p`), `go build .` on the temp directory works.
	var args []string
	if IsRunnable(source) {
		args = []string{"run", srcPath}
	} else {
		args = []string{"build", "-o", os.DevNull, "."}
	}
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = dir
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
