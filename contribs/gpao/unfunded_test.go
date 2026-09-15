package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/gnoclient"
	"github.com/gnolang/gno/gno.land/pkg/gnoland"
	"github.com/gnolang/gno/gno.land/pkg/integration"
	vm "github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/client"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
)

const purseFee = int64(1_000_000) // defaultGasFee, as an amount

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
	key := func(i uint32) crypto.Address {
		t.Helper()
		s, err := gnoclient.SignerFromBip39(
			integration.DefaultAccount_Seed, cfg.Genesis.ChainID, "", 0, i)
		require.NoError(t, err)
		info, err := s.Info()
		require.NoError(t, err)
		return info.GetAddress()
	}
	approver, submitter := key(0), key(1)
	require.NotEqual(t, approver, submitter,
		"the two arms have to be different accounts or the purse cannot be emptied")

	ggs := cfg.Genesis.AppState.(gnoland.GnoGenesisState)
	ggs.Balances = []gnoland.Balance{
		{Address: submitter, Amount: std.NewCoins(std.NewCoin(ugnotDenom, 100_000_000_000))},
		{Address: approver, Amount: std.NewCoins(std.NewCoin(ugnotDenom, approvals*purseFee))},
	}
	ggs.VM.Params.CodeSubmissionPolicy = "inert"
	ggs.VM.Params.PkgApprovers = []crypto.Address{approver}
	cfg.Genesis.AppState = ggs

	node, remote := integration.TestingInMemoryNode(t, log.NewNoopLogger(), cfg)
	t.Cleanup(func() { _ = node.Stop() })

	rpc, err := client.NewHTTPClient(remote)
	require.NoError(t, err)
	submitterSigner, err := gnoclient.SignerFromBip39(
		integration.DefaultAccount_Seed, cfg.Genesis.ChainID, "", 0, 1)
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
	path := "gno.land/r/purse/" + name
	return &std.MemPackage{
		Name: name,
		Path: path,
		Files: []*std.MemFile{
			{Name: "gnomod.toml", Body: gno.GenGnoModLatest(path)},
			{Name: name + ".gno", Body: "package " + name + "\n\nfunc F(cur realm) string { return \"" + name + "\" }\n"},
		},
	}
}

// park submits a package and returns the height MsgEnablePackage has to name.
func (f *purseFixture) park(t *testing.T, mpkg *std.MemPackage) int64 {
	t.Helper()
	tx := std.Tx{
		Msgs: []std.Msg{vm.MsgAddPackage{Creator: f.submitter, Package: mpkg}},
		Fee:  std.NewFee(20_000_000, std.MustParseCoin("1000000ugnot")),
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
		Fee: std.NewFee(20_000_000, std.MustParseCoin("1000000ugnot")),
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
		remote:        f.remote,
		chainID:       f.chainID,
		mnemonic:      integration.DefaultAccount_Seed,
		gnoRoot:       f.gnoroot,
		gasFee:        defaultGasFee,
		maxSpend:      defaultMaxSpend,
		gasWanted:     defaultGasWanted,
		verifyBudget:  time.Minute,
		prepareBudget: defaultPrepareBudget,
		pollInterval:  200 * time.Millisecond,
		startHeight:   startHeight,
	}, testIO(t))
	require.NoError(t, err)
	require.Equal(t, f.approver, o.approver)
	require.Zero(t, o.maxSpend,
		"the default has to leave the balance as the only bound, or these tests "+
			"measure the bound instead of the balance")
	return o
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

	mpkg := f.pkg("nofunds") // sorts after gnomod.toml; unsorted files are refused

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
