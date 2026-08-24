package gnoland

import (
	"testing"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// queryAccount reads an account through the public query interface, i.e. the
// same way anything outside the node would see it.
//
// Callers take the account number and sequence from here rather than assuming
// them. Assuming is not merely fragile: a wrong number fails the tx in the ante
// handler, which would let a "this must be rejected" assertion pass on a
// signature error while the behavior under test never ran.
func queryAccount(t *testing.T, app abci.Application, addr crypto.Address) GnoAccount {
	t.Helper()

	qr := app.Query(abci.RequestQuery{Path: "auth/accounts/" + addr.String()})
	require.True(t, qr.IsOK(), "account query failed: %v", qr.ResponseBase.Error)

	// The auth handler answers in Data, and marshals the concrete account, so
	// there is no interface envelope to decode through.
	var acc GnoAccount
	require.NoError(t, amino.UnmarshalJSON(qr.Data, &acc))
	require.Equal(t, addr, acc.Address, "no account for %s", addr)
	return acc
}

// ugnotBalance reads an address's gas-denom balance.
func ugnotBalance(t *testing.T, app abci.Application, addr crypto.Address) int64 {
	t.Helper()
	return queryAccount(t, app, addr).Coins.AmountOf("ugnot")
}

// TestInertPackageLifecycleEndToEnd drives submit -> parked -> enable through
// real transactions, which is the only way to cover what the keeper's own unit
// tests cannot: the ante handler, the transaction boundary, and the committed
// store.
//
// It pins the three behaviors that distinguish "inert" from the other policies
// and that are otherwise only asserted against direct keeper calls:
//
//   - a submitted package is PARKED: it exists, but calling it fails;
//   - enabling runs init() as the CREATOR, not as the approver who signed the
//     MsgEnablePackage -- the identity is recovered from the gnomod.toml the
//     keeper stamped at submit time;
//   - the storage deposit is charged to the CREATOR at enable, so the approver
//     pays only its own gas.
//
// The submitter, the approver, and the two callers are four distinct accounts,
// so the balance assertions can attribute a cost to one party rather than to
// "whoever signed".
func TestInertPackageLifecycleEndToEnd(t *testing.T) {
	t.Parallel()

	const (
		chainID = "test-chain"
		path    = "gno.land/r/demo/inertlife"
		// init() records who the VM believes is running it. Under "inert" that
		// happens at enable time, in a transaction the approver signs -- so a
		// naive implementation records the approver, and this realm is what
		// tells the two apart.
		body = `package inertlife

import "chain/runtime/unsafe"

var origin string

func init() {
	origin = string(unsafe.OriginCaller())
}

func Origin(cur realm) string { return origin }
`
	)

	keys := getDummyKeys(t, 4)
	creator, approver, prober, caller := keys[0], keys[1], keys[2], keys[3]
	creatorAddr := creator.PubKey().Address()
	approverAddr := approver.PubKey().Address()

	// Genesis: the chain accepts submissions from anyone but activates nothing
	// on its own; only approverAddr can enable.
	vmGen := vm.DefaultGenesisState()
	vmGen.Params.CodeSubmissionPolicy = "inert"
	vmGen.Params.PkgApprovers = []crypto.Address{approverAddr}

	balances := make([]Balance, 0, len(keys))
	for _, k := range keys {
		balances = append(balances, Balance{
			Address: k.PubKey().Address(),
			Amount:  std.NewCoins(std.NewCoin("ugnot", 100_000_000)),
		})
	}

	app, err := NewAppWithOptions(TestAppOptions(memdb.NewMemDB()))
	require.NoError(t, err)

	app.InitChain(abci.RequestInitChain{
		ChainID: chainID,
		Time:    time.Now(),
		ConsensusParams: &abci.ConsensusParams{
			Block:     defaultBlockParams(),
			Validator: &abci.ValidatorParams{PubKeyTypeURLs: []string{}},
		},
		AppState: GnoGenesisState{
			Balances: balances,
			Auth:     auth.DefaultGenesisState(),
			Bank:     bank.DefaultGenesisState(),
			VM:       vmGen,
		},
	})
	// Commit genesis, so the accounts seeded above are visible to the queries
	// the signing helper makes.
	app.Commit()

	// Each transaction gets its own committed block. Committing is not
	// incidental: the account queries below read committed state, and enable
	// must see what submit actually persisted rather than an uncommitted cache.
	height := int64(0)
	deliver := func(t *testing.T, msgs []std.Msg, key crypto.PrivKey) abci.ResponseDeliverTx {
		t.Helper()
		signer := queryAccount(t, app, key.PubKey().Address())
		height++
		app.BeginBlock(abci.RequestBeginBlock{Header: &bft.Header{
			ChainID: chainID, Height: height, Time: time.Now(),
		}})
		tx := createAndSignTxWithAccSeq(t, msgs, chainID, key,
			signer.AccountNumber, signer.Sequence)
		raw, err := amino.Marshal(tx)
		require.NoError(t, err)
		resp := app.DeliverTx(abci.RequestDeliverTx{Tx: raw})
		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()
		return resp
	}

	// 1. Submit. Accepted, but nothing runs yet.
	addResp := deliver(t, []std.Msg{vm.MsgAddPackage{
		Creator: creatorAddr,
		Package: &std.MemPackage{
			Name: "inertlife",
			Path: path,
			// Sorted by name: AddMemPackage rejects an unsorted file list.
			Files: []*std.MemFile{
				{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(path)},
				{Name: "inertlife.gno", Body: body},
			},
		},
	}}, creator)
	require.True(t, addResp.IsOK(), "submit under inert should be accepted: %s", addResp.Log)

	// 2. Parked: not callable.
	//
	// This step alone cannot tell "parked" from "was never stored" -- both fail
	// here, and identically, because a call to an absent package currently dies
	// on a generic VM panic ("unexpected node with location <path>:0:0") rather
	// than a package-not-found error. Verified not to be inert-specific: a path
	// that was never submitted produces the same message.
	//
	// Step 3 is what closes that gap. EnablePackage looks the package up in the
	// inert key space and refuses with "no inert package at path" when it is
	// missing, so the enable below succeeding is the proof that submit really
	// stored it. Keep these two steps together; either alone proves little.
	parkedResp := deliver(t, []std.Msg{vm.MsgCall{
		Caller: prober.PubKey().Address(), PkgPath: path, Func: "Origin",
	}}, prober)
	require.False(t, parkedResp.IsOK(),
		"a submitted-but-unenabled package must not be callable; got %q", parkedResp.Data)

	creatorBefore := ugnotBalance(t, app, creatorAddr)
	approverBefore := ugnotBalance(t, app, approverAddr)

	// 3. Enable, signed by the approver.
	enableResp := deliver(t, []std.Msg{vm.MsgEnablePackage{
		Approver: approverAddr, PkgPath: path,
	}}, approver)
	require.True(t, enableResp.IsOK(), "approver should be able to enable: %s", enableResp.Log)

	// 4. Now callable -- and init() ran as the creator, not the approver.
	callResp := deliver(t, []std.Msg{vm.MsgCall{
		Caller: caller.PubKey().Address(), PkgPath: path, Func: "Origin",
	}}, caller)
	require.True(t, callResp.IsOK(), "enabled package should be callable: %s", callResp.Log)
	assert.Contains(t, string(callResp.Data), creatorAddr.String(),
		"init() must run as the creator recorded at submit time, not as the approver who signed the enable")
	assert.NotContains(t, string(callResp.Data), approverAddr.String(),
		"the approver's identity must never reach the package being enabled")

	// 5. The creator funds the package it submitted, even though it signed
	//    nothing here; the approver is out only its own gas.
	creatorAfter := ugnotBalance(t, app, creatorAddr)
	approverAfter := ugnotBalance(t, app, approverAddr)

	assert.Less(t, creatorAfter, creatorBefore,
		"the storage deposit must come from the creator at enable time")
	assert.Equal(t, approverBefore-2_000_000, approverAfter,
		"the approver must pay its gas fee and nothing else -- no deposit, no storage cost")
}
