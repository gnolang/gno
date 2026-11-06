package main

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/integration"
	vm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	rpcclient "github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/commands"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// purseFee is defaultGasFee as an amount, derived rather than copied: the whole
// test is arithmetic on it, so a copy would go green on the wrong number.
var purseFee = std.MustParseCoin(defaultGasFee).Amount

// purseTxFee is what the fixture's own transactions pay, which is not the
// quantity under test.
func purseTxFee() std.Fee {
	return std.NewFee(defaultGasWanted, std.MustParseCoin(defaultGasFee))
}

// purseFixture is a chain whose approver holds a known, small number of
// approval fees, plus a separate rich key to park with.
type purseFixture struct {
	approver  crypto.Address
	submitter crypto.Address
	client    gnoclient.Client
	remote    string
	chainID   string
	gnoroot   string
}

// newPurseFixture starts an inert-policy node whose approver holds exactly
// approvals worth of fees, to the ugnot: the ante deducts tx.Fee.GasFee whole
// rather than metering it, so the next enable meets an empty account.
func newPurseFixture(t *testing.T, approvals int64) *purseFixture {
	t.Helper()
	gnoroot := gnoenv.RootDir()
	cfg := integration.TestingMinimalNodeConfig(gnoroot)
	cfg.SkipGenesisSigVerification = true

	// Index 0 is what newOracle builds from cfg.mnemonic, so it is the approver.
	key := func(i uint32) (gnoclient.Signer, crypto.Address) {
		t.Helper()
		s, err := gnoclient.SignerFromBip39(
			integration.DefaultAccount_Seed, cfg.Genesis.ChainID, "", 0, i)
		require.NoError(t, err)
		info, err := s.Info()
		require.NoError(t, err)
		return s, info.GetAddress()
	}
	_, approver := key(0)
	submitterSigner, submitter := key(1)

	ggs := cfg.Genesis.AppState.(gnoland.GnoGenesisState)
	ggs.Balances = []gnoland.Balance{
		{Address: submitter, Amount: std.NewCoins(std.NewCoin(ugnotDenom, 100_000_000_000))},
		{Address: approver, Amount: std.NewCoins(std.NewCoin(ugnotDenom, approvals*purseFee))},
	}
	ggs.VM.Params.CodeSubmissionPolicy = vm.CodeSubmissionPolicyInert
	ggs.VM.Params.PkgApprovers = []crypto.Address{approver}
	cfg.Genesis.AppState = ggs

	node, remote := integration.TestingInMemoryNode(t, log.NewNoopLogger(), cfg)
	t.Cleanup(func() { _ = node.Stop() })

	rpc, err := rpcclient.NewHTTPClient(remote)
	require.NoError(t, err)

	return &purseFixture{
		approver:  approver,
		submitter: submitter,
		client:    gnoclient.Client{Signer: submitterSigner, RPCClient: rpc},
		remote:    remote,
		chainID:   cfg.Genesis.ChainID,
		gnoroot:   gnoroot,
	}
}

func (f *purseFixture) pkg(name string) *std.MemPackage {
	return chainPackage("gno.land/r/purse/"+name,
		"package "+name+"\n\nfunc F(cur realm) string { return \""+name+"\" }\n")
}

// park submits a package and returns the height MsgEnablePackage has to name.
func (f *purseFixture) park(t *testing.T, mpkg *std.MemPackage) int64 {
	t.Helper()
	tx := std.Tx{
		Msgs: []std.Msg{vm.MsgAddPackage{Creator: f.submitter, Package: mpkg}},
		Fee:  purseTxFee(),
	}
	signed, err := f.client.SignTx(tx, 0, 0)
	require.NoError(t, err)
	res, err := f.client.BroadcastTxCommit(signed)
	require.NoError(t, err)
	require.True(t, res.CheckTx.IsOK(), "park checkTx: %v", res.CheckTx.Error)
	require.True(t, res.DeliverTx.IsOK(), "park deliverTx: %v", res.DeliverTx.Error)
	return res.Height
}

// fund tops the approver up by a plain transfer.
func (f *purseFixture) fund(t *testing.T, approvals int64) {
	t.Helper()
	tx := std.Tx{
		Msgs: []std.Msg{bank.MsgSend{
			FromAddress: f.submitter,
			ToAddress:   f.approver,
			Amount:      std.NewCoins(std.NewCoin(ugnotDenom, approvals*purseFee)),
		}},
		Fee: purseTxFee(),
	}
	signed, err := f.client.SignTx(tx, 0, 0)
	require.NoError(t, err)
	res, err := f.client.BroadcastTxCommit(signed)
	require.NoError(t, err)
	require.True(t, res.DeliverTx.IsOK(), "fund deliverTx: %v", res.DeliverTx.Error)
}

func (f *purseFixture) oracle(t *testing.T, startHeight int64) *oracle {
	t.Helper()
	o, err := newOracle(config{
		remote:   f.remote,
		chainID:  f.chainID,
		mnemonic: integration.DefaultAccount_Seed,
		gnoRoot:  f.gnoroot,
		gasFee:   defaultGasFee,
		// No bound, deliberately: these tests are about what the balance does,
		// and a bound reached first would measure that instead.
		maxSpend:      "",
		gasWanted:     defaultGasWanted,
		verifyBudget:  time.Minute,
		prepareBudget: defaultPrepareBudget,
		pollInterval:  200 * time.Millisecond,
		startHeight:   startHeight,
	}, testIO(t))
	require.NoError(t, err)
	require.Equal(t, f.approver, o.approver)
	require.Zero(t, o.maxSpend)
	// The pause cadence is the one thing here that would be spent in real time.
	o.fundsPollInterval = 200 * time.Millisecond
	return o
}

// captureStderr sends the oracle's stderr to a buffer, to be read only once its
// run has returned.
func captureStderr(t *testing.T, o *oracle) *bytes.Buffer {
	t.Helper()
	var stderr bytes.Buffer
	tio := testIO(t)
	tio.SetErr(commands.WriteNopCloser(&stderr))
	o.io = tio
	return &stderr
}

// awaitStatus waits for a verdict, and reports rather than hanging.
func awaitStatus(t *testing.T, o *oracle, path, want string) {
	t.Helper()
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		if got := o.status.get(path); got.Status == want {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("%s never reached %q; last verdict %+v", path, want, o.status.get(path))
}

// TestAnUnpayableFeeIsTypedInsufficientFunds pins the type the pause is selected
// on. Worth its own test because misreading it now waits forever instead of
// retiring one package.
func TestAnUnpayableFeeIsTypedInsufficientFunds(t *testing.T) {
	f := newPurseFixture(t, 0) // the approver cannot pay for anything

	mpkg := f.pkg("nofunds")

	height := f.park(t, mpkg)

	o := f.oracle(t, 1)
	maxGas, answered := o.queryBlockMaxGas(t.Context())
	require.True(t, answered, "a zero ceiling clamps every gas figure to zero")
	o.blockMaxGas = maxGas

	pkgHash, err := vm.PackageContentHash(mpkg)
	require.NoError(t, err)

	err = o.enable(mpkg.Path, pkgHash, height)
	require.Error(t, err)
	assert.ErrorAs(t, err, &std.InsufficientFundsError{},
		"handleCandidate pauses on exactly this type; see its enable loop")
	assert.Zero(t, o.spent,
		"a simulate refusal creates no transaction, so it must not be counted")
}

// TestAnEmptyPursePausesRatherThanRetiringPackages: out of money is a property
// of the oracle, never a verdict on a package. An unfunded approver used to walk
// the queue spending each package's three attempts and then marking it seen, and
// even without that a skipped package was gone -- heights only move forward. So
// it stops: blocked on the package in hand, no attempt, nothing seen, and the
// reader stalled through the queue. Funding resumes all of it in place.
//
// The approver holds exactly two approvals, which makes the third package the
// arm under test and the first two its control.
func TestAnEmptyPursePausesRatherThanRetiringPackages(t *testing.T) {
	const approvalsAffordable = 2
	f := newPurseFixture(t, approvalsAffordable)

	one, two, three := f.pkg("one"), f.pkg("two"), f.pkg("three")
	f.park(t, one)
	f.park(t, two)
	f.park(t, three)

	o := f.oracle(t, 1)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	go func() { _ = o.run(ctx) }()

	// ---- Control arm: everything the key can afford goes through
	awaitStatus(t, o, one.Path, statusApproved)
	awaitStatus(t, o, two.Path, statusApproved)

	// ---- The arm under test: out of money, and nothing held against the package
	awaitStatus(t, o, three.Path, statusBlocked)

	// Reading the verifier's maps is safe here only: it is parked in
	// waitForFunds, and the status mutex orders its writes before this read.
	status := o.status.get(three.Path)
	assert.Contains(t, status.Reason, "cannot pay the approval fee")
	assert.Zero(t, status.Attempt, "a funding stall must not be counted as an attempt")
	assert.Equal(t, approvalsAffordable*purseFee, o.spent,
		"only the two affordable approvals were ever sent")

	key := candidateKey(three)
	assert.NotContains(t, o.seen, key,
		"a paused package has to stay eligible, or resuming approves nothing")
	assert.NotContains(t, o.failedEnable, key,
		"an empty purse must not spend one of the three chances the package gets")

	balance, _, err := f.client.QueryBalance(f.approver)
	require.NoError(t, err)
	require.Zero(t, balance.AmountOf(ugnotDenom),
		"the purse has to be empty for this to be about an empty purse")

	// ---- Submitted DURING the stall, and still not lost
	later := f.pkg("later")
	f.park(t, later)
	require.Equal(t, statusUnknown, o.status.get(later.Path).Status,
		"a paused oracle issues no verdicts at all, not even for what it queued")

	// ---- Funding resumes everything, with no restart
	f.fund(t, 5)

	awaitStatus(t, o, three.Path, statusApproved)
	awaitStatus(t, o, later.Path, statusApproved)
}

// TestAHeavyPackageIsDeferredNotHeldAgainstTheQueue: the chain checks
// -gas-fee over each package's GasWanted against its gas price, so whether
// the fee suffices depends on the package. A heavy one refused at CheckTx is
// deferred -- no enable attempt, nothing seen, nothing spent -- and the light
// one behind it is approved.
func TestAHeavyPackageIsDeferredNotHeldAgainstTheQueue(t *testing.T) {
	f := newPurseFixture(t, 1)

	const heavyPath = "gno.land/r/purse/heavy"
	heavy := chainPackage(heavyPath, "package heavy\n\nvar n int\n\n"+
		"func init() { for i := 0; i < 300000; i++ { n++ } }\n\n"+
		"func F(cur realm) int { return n }\n")
	light := f.pkg("light")
	f.park(t, heavy)
	f.park(t, light)

	o := f.oracle(t, 1)
	// At this chain's 1ugnot per 1000 gas, this covers 1e8 gas: the light
	// package, not the heavy one's init.
	o.gasFee = std.NewCoin(ugnotDenom, 100_000)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = o.run(ctx)
	}()

	awaitStatus(t, o, light.Path, statusApproved)
	cancel()
	<-done

	status := o.status.get(heavyPath)
	assert.Equal(t, statusDeferred, status.Status)
	assert.Contains(t, status.Reason, "below the chain's gas price")
	assert.Equal(t, 1, status.Attempt)
	key := candidateKey(heavy)
	assert.NotContains(t, o.seen, key)
	assert.NotContains(t, o.failedEnable, key,
		"an underpriced fee must not spend one of the package's enable attempts")
	require.Len(t, o.deferred, 1)
	assert.Equal(t, heavyPath, o.deferred[0].path)
	assert.Equal(t, o.gasFee.Amount, o.spent,
		"only the light approval was charged; the CheckTx refusal was free")
}

// newDeferringOracle parks a package and returns an oracle whose -gas-fee is
// below what the chain asks for it, ready to drive synchronously.
func newDeferringOracle(t *testing.T) (*oracle, *std.MemPackage, int64) {
	t.Helper()
	f := newPurseFixture(t, 1)
	mpkg := f.pkg("underpriced")
	height := f.park(t, mpkg)

	o := f.oracle(t, 1)
	maxGas, answered := o.queryBlockMaxGas(t.Context())
	require.True(t, answered)
	o.blockMaxGas = maxGas
	// The test chain prices gas at 1ugnot per 1000, so this covers 1000 gas.
	o.gasFee = std.NewCoin(ugnotDenom, 1)
	return o, mpkg, height
}

// TestADeferredApprovalIsRetried: once the fee covers the package -- here by
// raising it, standing in for a price that fell back -- the next pass approves
// it, with no restart and nothing resubmitted.
func TestADeferredApprovalIsRetried(t *testing.T) {
	o, mpkg, height := newDeferringOracle(t)

	o.handleCandidate(t.Context(), candidate{mpkg: mpkg, height: height})
	require.Equal(t, statusDeferred, o.status.get(mpkg.Path).Status)

	o.gasFee = std.MustParseCoin(defaultGasFee)
	o.retryDeferred(t.Context())

	assert.Equal(t, statusApproved, o.status.get(mpkg.Path).Status)
	assert.Empty(t, o.deferred)
	assert.Contains(t, o.seen, candidateKey(mpkg))
}

// TestADeferredApprovalIsBlockedOnceItsRetriesRunOut: a package the fee never
// covers frees its slot after maxDeferredRetries, blocked with the height a
// restart needs.
func TestADeferredApprovalIsBlockedOnceItsRetriesRunOut(t *testing.T) {
	o, mpkg, height := newDeferringOracle(t)
	stderr := captureStderr(t, o)

	o.handleCandidate(t.Context(), candidate{mpkg: mpkg, height: height})
	for range maxDeferredRetries - 1 {
		o.retryDeferred(t.Context())
		require.Equal(t, statusDeferred, o.status.get(mpkg.Path).Status)
	}
	o.retryDeferred(t.Context())

	status := o.status.get(mpkg.Path)
	assert.Equal(t, statusBlocked, status.Status)
	assert.Contains(t, status.Reason, "retry window ran out")
	assert.Equal(t, maxDeferredRetries+1, status.Attempt)
	assert.Empty(t, o.deferred)
	assert.NotContains(t, o.seen, candidateKey(mpkg),
		"blocked is left unseen, so a restart can still approve it")
	assert.Zero(t, o.spent, "every refusal was CheckTx's, and free")
	assert.Contains(t, stderr.String(), fmt.Sprintf("-start-height %d", height))
}

// TestReachingMaxSpendPausesRatherThanDroppingPackages: the bound only moves on
// a restart, and heights only move forward, so skipping past the package in
// hand would lose it and everything behind. It holds like an empty purse.
func TestReachingMaxSpendPausesRatherThanDroppingPackages(t *testing.T) {
	f := newPurseFixture(t, 10)

	one, two, three := f.pkg("one"), f.pkg("two"), f.pkg("three")
	f.park(t, one)
	twoHeight := f.park(t, two)
	f.park(t, three)

	o := f.oracle(t, 1)
	// Room for exactly one approval, with the balance for ten.
	o.maxSpend = purseFee
	stderr := captureStderr(t, o)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = o.run(ctx)
	}()

	awaitStatus(t, o, one.Path, statusApproved)
	awaitStatus(t, o, two.Path, statusBlocked)
	// Long enough for a walk past the held package to show.
	time.Sleep(time.Second)
	cancel()
	<-done

	status := o.status.get(two.Path)
	assert.Contains(t, status.Reason, "spending limit")
	assert.Zero(t, status.Attempt, "reaching the bound must not be counted as an attempt")
	assert.Equal(t, statusUnknown, o.status.get(three.Path).Status,
		"the package behind has to wait, not be dropped the same way")
	key := candidateKey(two)
	assert.NotContains(t, o.seen, key)
	assert.NotContains(t, o.failedEnable, key)
	assert.Equal(t, purseFee, o.spent, "only the one approval inside the bound was sent")
	assert.Contains(t, stderr.String(), fmt.Sprintf("-start-height %d", twoHeight),
		"a restart from anywhere later would lose the held package")
}
