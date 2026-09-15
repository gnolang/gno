package main

import (
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

// TestAnEmptyPurseBlocksRatherThanRetiringPackages pins what happens when the
// approver runs out of money, which is the state the default -max-spend used to
// hide: a run stopped at its bound while the key was still full, so nobody had
// to find out what an actually empty key did.
//
// It does something bad. An enable whose fee cannot be paid fails at simulate,
// and every failed enable spends one of the package's three attempts before the
// third marks it seen -- so an unfunded approver walks through the queue
// retiring, for the life of the run, packages that nothing is wrong with, each
// under a message saying a human is needed. Nothing on chain contradicts it: a
// parked package reports "waiting for a package approver" whether the oracle is
// broke or the package landed a second ago.
//
// So the rule is: out of money is a property of the ORACLE, never a verdict on
// the package. It records statusBlocked, it burns no attempt, it marks nothing
// seen, and it clears itself the moment the key is funded -- without a restart,
// because the cure happens in a different process entirely.
//
// Every figure below is read off the chain rather than assumed. The approver is
// funded with exactly two approvals' worth, which is what makes the third
// package the one under test and the first two its control arm.
func TestAnEmptyPurseBlocksRatherThanRetiringPackages(t *testing.T) {
	gnoroot := gnoenv.RootDir()
	cfg := integration.TestingMinimalNodeConfig(gnoroot)
	cfg.SkipGenesisSigVerification = true

	// Two keys off one seed. Index 0 is what newOracle builds from cfg.mnemonic,
	// so it has to be the approver; index 1 pays for the submissions, because a
	// key poor enough to be the subject of this test cannot afford to park
	// anything.
	approverSigner, err := gnoclient.SignerFromBip39(
		integration.DefaultAccount_Seed, cfg.Genesis.ChainID, "", 0, 0)
	require.NoError(t, err)
	approverInfo, err := approverSigner.Info()
	require.NoError(t, err)
	approver := approverInfo.GetAddress()

	submitterSigner, err := gnoclient.SignerFromBip39(
		integration.DefaultAccount_Seed, cfg.Genesis.ChainID, "", 0, 1)
	require.NoError(t, err)
	submitterInfo, err := submitterSigner.Info()
	require.NoError(t, err)
	submitter := submitterInfo.GetAddress()
	require.NotEqual(t, approver, submitter,
		"the two arms have to be different accounts or the purse cannot be emptied")

	const fee = int64(1_000_000) // defaultGasFee, as an amount
	const approvalsAffordable = 2

	ggs := cfg.Genesis.AppState.(gnoland.GnoGenesisState)
	ggs.Balances = []gnoland.Balance{
		{
			Address: submitter,
			Amount:  std.NewCoins(std.NewCoin(ugnotDenom, 100_000_000_000)),
		},
		{
			// Exactly two approvals, to the ugnot. The ante deducts
			// tx.Fee.GasFee whole rather than metering it, so the third enable
			// meets an account holding nothing at all.
			Address: approver,
			Amount:  std.NewCoins(std.NewCoin(ugnotDenom, approvalsAffordable*fee)),
		},
	}
	ggs.VM.Params.CodeSubmissionPolicy = "inert"
	ggs.VM.Params.PkgApprovers = []crypto.Address{approver}
	cfg.Genesis.AppState = ggs

	node, remote := integration.TestingInMemoryNode(t, log.NewNoopLogger(), cfg)
	defer node.Stop()

	rpc, err := client.NewHTTPClient(remote)
	require.NoError(t, err)
	submitterClient := gnoclient.Client{Signer: submitterSigner, RPCClient: rpc}

	pkg := func(name string) *std.MemPackage {
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
	park := func(t *testing.T, mpkg *std.MemPackage) {
		t.Helper()
		tx := std.Tx{
			Msgs: []std.Msg{vm.MsgAddPackage{Creator: submitter, Package: mpkg}},
			Fee:  std.NewFee(20_000_000, std.MustParseCoin("1000000ugnot")),
		}
		signed, err := submitterClient.SignTx(tx, 0, 0)
		require.NoError(t, err)
		res, err := submitterClient.BroadcastTxCommit(signed)
		require.NoError(t, err)
		require.True(t, res.CheckTx.IsOK(), "park checkTx: %v", res.CheckTx.Error)
		require.True(t, res.DeliverTx.IsOK(), "park deliverTx: %v", res.DeliverTx.Error)
	}

	one, two, three := pkg("one"), pkg("two"), pkg("three")
	park(t, one)
	park(t, two)
	park(t, three)

	ocfg := config{
		remote:        remote,
		chainID:       cfg.Genesis.ChainID,
		mnemonic:      integration.DefaultAccount_Seed,
		gnoRoot:       gnoroot,
		gasFee:        defaultGasFee,
		maxSpend:      defaultMaxSpend,
		gasWanted:     defaultGasWanted,
		verifyBudget:  time.Minute,
		prepareBudget: defaultPrepareBudget,
	}
	o, err := newOracle(ocfg, testIO(t))
	require.NoError(t, err)
	require.Equal(t, approver, o.approver)
	require.Zero(t, o.maxSpend,
		"the default has to leave the balance as the only bound, or this test "+
			"measures the bound instead of the balance")
	maxGas, answered := o.queryBlockMaxGas(t.Context())
	require.True(t, answered, "a zero ceiling clamps every gas figure to zero and the enable is refused")
	o.blockMaxGas = maxGas

	// ---- Control arm: everything it can afford goes through
	for _, mpkg := range []*std.MemPackage{one, two} {
		o.handleCandidate(t.Context(), candidate{mpkg: mpkg})
		require.Equal(t, statusApproved, o.status.get(mpkg.Path).Status,
			"%s is affordable and has to be approved", mpkg.Path)
	}
	assert.Equal(t, approvalsAffordable*fee, o.spent)

	balance, _, err := submitterClient.QueryBalance(approver)
	require.NoError(t, err)
	require.Zero(t, balance.AmountOf(ugnotDenom),
		"the purse has to be empty for the arm under test to be about an empty purse")

	// ---- The arm under test: out of money, and nothing held against the package
	o.handleCandidate(t.Context(), candidate{mpkg: three})

	status := o.status.get(three.Path)
	assert.Equal(t, statusBlocked, status.Status,
		"an unpayable fee is the oracle's problem; rejected or gave_up would "+
			"tell the submitter their code is at fault")
	assert.Contains(t, status.Reason, "cannot pay the approval fee")
	assert.Zero(t, status.Attempt, "a funding stall must not be counted as an attempt")

	key := candidateKey(three)
	assert.NotContains(t, o.seen, key,
		"a blocked package has to stay eligible, or funding the key fixes nothing "+
			"without a restart")
	assert.NotContains(t, o.failedEnable, key,
		"an empty purse must not spend one of the three chances the package gets")

	// The boundary the branch above is selected on, pinned against a real node.
	// auth.DeductFees is the only thing in the tree that raises this type, and
	// MsgEnablePackage names the approver as its sole signer, so it cannot mean
	// anyone else's shortfall -- the creator's storage deposit fails on this
	// same message as InsufficientCoinsError, from the VM handler. If that ever
	// changed class, handleCandidate would go back to burning attempts with no
	// other test noticing.
	pkgHash, err := vm.PackageContentHash(three)
	require.NoError(t, err)
	enableErr := o.enable(three.Path, pkgHash, 0)
	require.Error(t, enableErr)
	assert.ErrorAs(t, enableErr, &std.InsufficientFundsError{},
		"blockedOnFunds is chosen on this type; see handleCandidate")
	assert.Zero(t, o.spent-approvalsAffordable*fee,
		"a simulate refusal must not be counted as spend")

	// ---- Funding the key resumes the run, with no restart
	//
	// A plain transfer, which is what an operator actually does about this, and
	// nothing else: no restart, no flag, no resubmission of the package.
	fund := std.Tx{
		Msgs: []std.Msg{bank.MsgSend{
			FromAddress: submitter,
			ToAddress:   approver,
			Amount:      std.NewCoins(std.NewCoin(ugnotDenom, 5*fee)),
		}},
		Fee: std.NewFee(20_000_000, std.MustParseCoin("1000000ugnot")),
	}
	signedFund, err := submitterClient.SignTx(fund, 0, 0)
	require.NoError(t, err)
	fundRes, err := submitterClient.BroadcastTxCommit(signedFund)
	require.NoError(t, err)
	require.True(t, fundRes.CheckTx.IsOK(), "fund checkTx: %v", fundRes.CheckTx.Error)
	require.True(t, fundRes.DeliverTx.IsOK(), "fund deliverTx: %v", fundRes.DeliverTx.Error)

	o.handleCandidate(t.Context(), candidate{mpkg: three})
	assert.Equal(t, statusApproved, o.status.get(three.Path).Status,
		"the package was never at fault; a funded key has to reach it without a restart")
}
