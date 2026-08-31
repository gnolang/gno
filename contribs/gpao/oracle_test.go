package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// A valid BIP39 mnemonic (the standard gno integration test seed). The oracle
// only needs it to derive the approver address; no network access happens here.
const testMnemonic = "source bonus chronic canvas draft south burst lottery vacant surface solve popular case indicate oppose farm nothing bullet exhibit title speed wink action roast"

// TestMain lets this test binary stand in for the real one when it is spawned
// as a verify-one child.
//
// o.verify() re-invokes os.Executable(), which under `go test` is the test
// binary. Without this dispatch the budget tests below would exercise a child
// that immediately fails to parse its arguments, which looks like a rejection
// and would pass for the wrong reason. Dispatching on os.Args[1] is safe: a
// test binary's own arguments are all -test.* flags, so the only way to see
// "verify-one" there is to have been spawned as one.
// spinEnv puts a spawned child into an endless loop that appends a byte to the
// named file every 20ms, instead of verifying. Only TestOracleVerifyBudgetKills
// sets it; production never does. It is the only way to observe whether the
// budget actually STOPPED the work, as opposed to returning while it continued.
const spinEnv = "GPAO_TEST_SPIN_HEARTBEAT"

func TestMain(m *testing.M) {
	if len(os.Args) > 1 && os.Args[1] == verifyOneCmdName {
		if beat := os.Getenv(spinEnv); beat != "" {
			spinForever(beat)
		}
		// Drive the REAL command tree, not newVerifyOneCmd directly. Routing
		// straight to the subcommand left AddSubCommands and NoParentFlags
		// uncovered — and dropping NoParentFlags makes AddSubCommands
		// re-register the parent's -remote onto a flagset that already declares
		// it, which panics at construction. main() would die on startup while
		// the suite stayed green.
		if err := newRootCmd(commands.NewDefaultIO()).
			ParseAndRun(context.Background(), os.Args[1:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func spinForever(path string) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		os.Exit(2)
	}
	for {
		if _, err := f.WriteString("x"); err != nil {
			os.Exit(2)
		}
		_ = f.Sync()
		time.Sleep(20 * time.Millisecond)
	}
}

// newTestVerifier builds the verification half directly, without a signer or a
// chain. The pure typecheck/preprocess tests use this; only the budget tests
// need a real oracle and a real child process.
func newTestVerifier(t *testing.T) *verifier {
	t.Helper()
	v, err := newVerifier(gnoenv.RootDir(), "", os.Stderr)
	require.NoError(t, err)
	return v
}

func newTestOracle(t *testing.T) *oracle {
	t.Helper()
	cfg := config{
		remote:    "http://127.0.0.1:26657", // not contacted in these tests
		chainID:   "test",
		mnemonic:  testMnemonic,
		gnoRoot:   gnoenv.RootDir(),
		gasFee:    defaultGasFee,
		gasWanted: defaultGasWanted,
	}
	// Real writers, not NewTestIO's nil ones: the parent tees a child's stderr
	// through io.Err(), and a nil there is a crash the tests should surface
	// rather than dodge.
	tio := commands.NewTestIO()
	tio.SetOut(commands.WriteNopCloser(io.Discard))
	tio.SetErr(commands.WriteNopCloser(io.Discard))
	o, err := newOracle(cfg, tio)
	require.NoError(t, err)
	require.False(t, o.approver.IsZero(), "approver address must be derived")
	return o
}

// TestOracleTypecheckAcceptsValidPackage: a well-typed package importing a
// stdlib passes the off-chain typecheck — the oracle would approve it.
func TestOracleTypecheckAcceptsValidPackage(t *testing.T) {
	v := newTestVerifier(t)

	require.NoError(t, v.verifyPackage(validTestPackage()),
		"valid package must pass verification")
}

// TestOracleTypecheckRejectsInvalidPackage: an ill-typed package fails the
// off-chain typecheck — the oracle would NOT approve it (and the chain would
// reject it anyway).
func TestOracleTypecheckRejectsInvalidPackage(t *testing.T) {
	v := newTestVerifier(t)

	const path = "gno.land/r/test/bad"
	mpkg := &std.MemPackage{
		Name: "bad",
		Path: path,
		Type: gno.MPUserAll,
		Files: []*std.MemFile{
			{Name: "bad.gno", Body: `package bad

func Boom(cur realm) string {
	return undefinedSymbol
}`},
			{Name: "gnomod.toml", Body: gno.GenGnoModLatest(path)},
		},
	}

	assert.Error(t, v.verifyPackage(mpkg), "ill-typed package must fail typecheck")
}

// An overrun must be reported as a budget event, not as a verdict about the
// package: upstream gives the two opposite treatment, so conflating them would
// permanently settle a package that was merely slow.
//
// One nanosecond, so this is cheap and independent of timing. The enforcement
// property itself -- that we return AT the budget rather than after the work --
// is measured in TestOracleVerifyBudgetKillsTheChild.
func TestOracleVerifyBudgetRejectionIsDistinct(t *testing.T) {
	o := newTestOracle(t)
	o.cfg.verifyBudget = time.Nanosecond

	err := o.verify(context.Background(), validTestPackage())
	require.Error(t, err)
	require.ErrorIs(t, err, errVerifyBudget,
		"a timeout must be distinguishable from a rejection")
}

// TestOraclePreprocessCatchesWhatTypecheckCannot is the justification for
// running preprocess at all.
//
// A pure package declaring a crossing function (a `cur realm` first argument)
// is well-typed as far as go/types is concerned -- `realm` is just a type and
// `cur` just a parameter -- but the preprocessor rejects it, because "crossing
// functions only exist in realms" is a Gno rule with no Go equivalent. The
// validator runs both stages at MsgEnablePackage, so an oracle that ran only
// the typecheck would approve this and be contradicted by the chain.
//
// It also pins the mechanism: the preprocessor reports errors by panicking, so
// this exercises verifyPackage's recover as the error path it actually is
// rather than as a safety net.
func TestOraclePreprocessCatchesWhatTypecheckCannot(t *testing.T) {
	v := newTestVerifier(t)

	const path = "gno.land/p/test/crossinpure"
	mpkg := &std.MemPackage{
		Name: "crossinpure",
		Path: path,
		Type: gno.MPUserAll,
		Files: []*std.MemFile{
			{Name: "a.gno", Body: "package crossinpure\n\nfunc F(cur realm) {}\n"},
			{Name: "gnomod.toml", Body: gno.GenGnoModLatest(path)},
		},
	}
	err := v.verifyPackage(mpkg)
	require.Error(t, err, "a crossing function in a pure package must be refused")
	assert.Contains(t, err.Error(), "non-realm package",
		"the refusal must come from the preprocessor's own rule")

	// And it is a rejection, not a budget overrun: the two get opposite
	// treatment upstream, so they must stay distinguishable.
	assert.NotErrorIs(t, err, errVerifyBudget)
}

// TestOracleHandleCandidateOverBudgetCap covers the seen/overBudget
// bookkeeping, which decides whether a package can ever be looked at again.
//
// The two directions matter for opposite reasons. Not recording an overrun is
// what makes a retry possible at all -- handleCandidate used to mark a path
// seen BEFORE verifying, so any failure blacklisted it for the life of the
// daemon. But never recording it lets one path consume a full budget per
// resubmission indefinitely, and resubmitting is cheap under "inert". So:
// pending for a while, then given up on.
//
// It bounds repeat spend on ONE path, which is all it claims; an attacker using
// fresh paths gets a fresh allowance for each.
//
// A budget of one nanosecond guarantees an overrun without needing a slow
// package, so this tests the bookkeeping and not the typechecker.
func TestOracleHandleCandidateOverBudgetCap(t *testing.T) {
	o := newTestOracle(t)
	o.cfg.verifyBudget = time.Nanosecond

	mpkg := validTestPackage()
	// Keyed on the bytes, not the path -- see candidateKey.
	path := candidateKey(mpkg)

	// Every attempt before the cap leaves the package pending, so a restart or
	// a resubmission can still reach it.
	for i := 1; i < maxOverBudgetAttempts; i++ {
		o.handleCandidate(context.Background(), mpkg)
		assert.NotContains(t, o.seen, path,
			"attempt %d of %d must leave the package pending", i, maxOverBudgetAttempts)
		assert.Equal(t, i, o.overBudget[path], "overrun count after attempt %d", i)
	}

	// At the cap the oracle stops paying for it.
	o.handleCandidate(context.Background(), mpkg)
	assert.Contains(t, o.seen, path,
		"the package must be given up on once it has burned %d budgets", maxOverBudgetAttempts)

	// And having given up, it is not verified again: the count stays put
	// because handleCandidate now returns on the seen check.
	o.handleCandidate(context.Background(), mpkg)
	assert.Equal(t, maxOverBudgetAttempts, o.overBudget[path],
		"a given-up package must not be re-verified")
}

// validTestPackage is the well-typed package shared by the tests below. Type
// must be set: verifyPackage parses as MPUserProd, and a zero Type makes the
// store's import processing panic with an internal "should not happen" rather
// than reporting anything useful.
func validTestPackage() *std.MemPackage {
	const path = "gno.land/r/test/budget"
	return &std.MemPackage{
		Name: "budget",
		Path: path,
		Type: gno.MPUserAll,
		// Sorted by name: AddMemPackage validates the order (the typecheck
		// path does not, so an unsorted fixture only fails once preprocess
		// runs).
		Files: []*std.MemFile{
			{Name: "budget.gno", Body: `package budget

import "strings"

func Shout(cur realm, s string) string {
	return strings.ToUpper(s)
}`},
			{Name: "gnomod.toml", Body: gno.GenGnoModLatest(path)},
		},
	}
}

// TestOracleVerifyBudgetKillsTheChild is the test that justifies the subprocess.
//
// It replaces one that measured only elapsed time, and was vacuous: an
// implementation that starts a goroutine and ABANDONS it at the deadline also
// returns at the budget, so it passed. That is exactly the design the
// subprocess exists to rule out, and the test meant to exclude it could not
// tell the two apart. Timing proves the call came back; only the absence of
// further progress proves the work stopped.
//
// So the child is put into a heartbeat loop and the file is checked twice: once
// when verify() returns, and again well after. A killed child leaves the count
// unchanged; an abandoned one keeps counting.
func TestOracleVerifyBudgetKillsTheChild(t *testing.T) {
	beat := filepath.Join(t.TempDir(), "beat")
	t.Setenv(spinEnv, beat) // reaches the child via childEnvAllowed

	o := newTestOracle(t)
	// Generous on purpose. The child must get far enough to emit at least one
	// heartbeat or the anti-vacuity guard below trips, and process spawn plus
	// Go runtime init is slow under -race and slower still when the rest of the
	// suite is competing for CPU. This bounds startup, not the work.
	o.cfg.verifyBudget = 2 * time.Second

	start := time.Now()
	err := o.verify(context.Background(), validTestPackage())
	require.ErrorIs(t, err, errVerifyBudget,
		"an overrun must be reported as a budget event, not a verdict")
	assert.Less(t, time.Since(start), 15*time.Second,
		"the call must return at the budget, not run to completion")

	beats := func() int {
		b, readErr := os.ReadFile(beat)
		require.NoError(t, readErr)
		return len(b)
	}

	// Guard against the test silently becoming vacuous: if the child never ran,
	// there is nothing to prove was killed.
	atReturn := beats()
	require.Greater(t, atReturn, 0,
		"the child never started, so this test would prove nothing")

	// Long enough for ~50 more beats had it survived.
	time.Sleep(time.Second)
	assert.Equal(t, atReturn, beats(),
		"the child kept working after the budget expired: the deadline abandoned "+
			"the work instead of killing it")
}

// TestVerifierWithoutRemoteDoesNotPanic pins a regression from the subprocess
// refactor.
//
// newVerifier leaves rpc nil when no remote is configured, and two call sites
// dereferenced it unconditionally — so a package importing anything the disk
// store cannot supply crashed the verifier instead of being reported. A crash
// is not a verdict: it surfaces as a non-zero exit and would be recorded as a
// rejection of a package that may be perfectly fine.
//
// The import below is deliberately a chain path that does not exist locally, so
// it reaches the RPC fallback that is absent.
func TestVerifierWithoutRemoteDoesNotPanic(t *testing.T) {
	v, err := newVerifier(gnoenv.RootDir(), "" /* no remote */, io.Discard)
	require.NoError(t, err)

	const path = "gno.land/r/test/needsremote"
	mpkg := &std.MemPackage{
		Name: "needsremote",
		Path: path,
		Type: gno.MPUserAll,
		Files: []*std.MemFile{
			{Name: "a.gno", Body: "package needsremote\n\nimport \"gno.land/r/nobody/nothing\"\n\nfunc F(cur realm) { _ = nothing.X }\n"},
			{Name: "gnomod.toml", Body: gno.GenGnoModLatest(path)},
		},
	}

	// An unresolvable import must be reported, not panicked on. Either a plain
	// typecheck error or the recovered form is acceptable; what is not
	// acceptable is the nil-pointer message that the bug produced.
	err = v.verifyPackage(mpkg)
	require.Error(t, err, "an unresolvable import must be reported")
	assert.NotContains(t, err.Error(), "nil pointer",
		"a missing remote must degrade to an unresolved import, not a crash")
}

// TestUnreachableRemoteIsNotAVerdict draws the triage boundary at the
// evidence, not the process.
//
// The child runs fine; the network under it fails; the typechecker reports the
// unfetchable import as unresolved; the child exits 1 -- and the parent read
// any non-negative exit as a verdict about the PACKAGE. handleCandidate then
// recorded statusRejected and marked the content seen, so resubmitting
// identical bytes was a silent no-op forever. The submitter was told their
// code is bad because the operator's network hiccuped.
//
// No chain anywhere: an unreachable remote is the whole reproduction.
func TestUnreachableRemoteIsNotAVerdict(t *testing.T) {
	o := newTestOracle(t)
	o.cfg.remote = "http://127.0.0.1:1" // nothing listens; connection refused, fast
	o.cfg.verifyBudget = time.Minute

	mpkg := packageImportingChainOnly()
	err := o.verify(context.Background(), mpkg)
	require.Error(t, err)
	require.ErrorIs(t, err, errVerifyUnavailable,
		"the child ran and the network under it failed; that says nothing about the package")
	assert.ErrorContains(t, err, "import resolver unavailable")

	// And the consequence the submitter feels: the content is NOT settled, so
	// a resubmission (or a restart) gets a fresh look once the fault clears.
	o.handleCandidate(context.Background(), mpkg)
	assert.NotContains(t, o.seen, candidateKey(mpkg),
		"a fault that was never about the bytes must not retire them")
	assert.Equal(t, statusPending, o.status.get(mpkg.Path).Status)
}

// packageImportingChainOnly imports a chain path that exists nowhere on disk,
// so with a remote configured the RPC getter is the only possible resolver.
func packageImportingChainOnly() *std.MemPackage {
	const path = "gno.land/r/test/needschain"
	return &std.MemPackage{
		Name: "needschain",
		Path: path,
		Type: gno.MPUserAll,
		Files: []*std.MemFile{
			{Name: "gnomod.toml", Body: gno.GenGnoModLatest(path)},
			{Name: "needschain.gno", Body: "package needschain\n\nimport \"gno.land/p/nobody/nothing\"\n\nfunc F(cur realm) { _ = nothing.X }\n"},
		},
	}
}

// TestVerifierAcceptsPackageWithTestFiles pins the package type the child uses.
//
// Real packages contain _test.gno files. AddPackage stamps the stored package
// MPUserAll, so that is what EnablePackage reads back and what this oracle must
// mirror. An earlier revision used MPUserProd here, whose validation rejects
// test files outright — so the child panicked, the parent read the non-zero exit
// as a REJECTION, and handleCandidate recorded it in `seen`. The oracle would
// have refused nearly every real package and told the operator the code was bad.
//
// Every other fixture in this file is test-file-free, which is exactly why that
// shipped. This one exists to have a test file.
func TestVerifierAcceptsPackageWithTestFiles(t *testing.T) {
	v := newTestVerifier(t)

	require.NoError(t, v.verifyPackage(packageWithTestFiles()),
		"a package with test files must verify; the chain stores these as MPUserAll")
}

// TestVerifyChildAcceptsPackageWithTestFiles is the same property one layer up,
// through a real spawned child.
//
// It has to be separate: the child RESTATES Type after decoding (the JSON
// envelope drops it), so the in-process test above cannot reach that line at
// all. Getting it wrong there is what made the oracle reject every package with
// a test file, and no budget test catches it because their fixture has none.
func TestVerifyChildAcceptsPackageWithTestFiles(t *testing.T) {
	o := newTestOracle(t)
	o.cfg.verifyBudget = time.Minute

	require.NoError(t, o.verify(context.Background(), packageWithTestFiles()),
		"the child must accept a package with test files")
}

// packageWithTestFiles is a realistic submission: production code plus a
// _test.gno. Every other fixture here is test-file-free, which is precisely why
// the MPUserProd bug shipped.
func packageWithTestFiles() *std.MemPackage {
	const path = "gno.land/r/test/withtests"
	return &std.MemPackage{
		Name: "withtests",
		Path: path,
		Type: gno.MPUserAll,
		// Sorted by name.
		Files: []*std.MemFile{
			{Name: "gnomod.toml", Body: gno.GenGnoModLatest(path)},
			{Name: "w.gno", Body: "package withtests\n\nfunc Add(a, b int) int { return a + b }\n"},
			{Name: "w_test.gno", Body: "package withtests\n\nimport \"testing\"\n\nfunc TestAdd(t *testing.T) {\n\tif Add(1, 2) != 3 {\n\t\tt.Fatal(\"bad\")\n\t}\n}\n"},
		},
	}
}

// TestOracleReVerifiesChangedBytesAtTheSamePath pins that a rejection settles
// the CODE, not the path.
//
// Keying on the path alone meant a submitter whose package was rejected could
// fix it, resubmit, and be ignored for the life of the process -- so "your
// package does not compile" became "this path is dead until someone restarts
// the daemon". It also meant bytes that changed after a verdict were assumed
// settled, which matters because the submitter may replace parked bytes at any
// time.
func TestOracleReVerifiesChangedBytesAtTheSamePath(t *testing.T) {
	o := newTestOracle(t)
	o.cfg.verifyBudget = time.Nanosecond // guarantees an overrun, cheaply

	first := validTestPackage()
	second := validTestPackage()
	// Same path, different content.
	second.Files = append([]*std.MemFile{}, second.Files...)
	second.Files[0] = &std.MemFile{
		Name: second.Files[0].Name,
		Body: second.Files[0].Body + "\n// a fix\n",
	}
	require.Equal(t, first.Path, second.Path)
	require.NotEqual(t, candidateKey(first), candidateKey(second),
		"different bytes must be a different candidate")

	// Burn the first candidate's whole allowance so it is given up on.
	for range maxOverBudgetAttempts {
		o.handleCandidate(context.Background(), first)
	}
	require.Contains(t, o.seen, candidateKey(first))

	// The corrected resubmission at the same path gets a fresh look: it is not
	// short-circuited by the first one's verdict.
	o.handleCandidate(context.Background(), second)
	assert.Equal(t, 1, o.overBudget[candidateKey(second)],
		"the corrected package must be verified rather than skipped")
	assert.NotContains(t, o.seen, candidateKey(second),
		"and it starts with its own allowance, not the rejected one's")
}
