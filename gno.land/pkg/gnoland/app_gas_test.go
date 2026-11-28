package gnoland

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/sdk/params"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/cachemulti"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/types"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
)

func TestDeliverTxVMParamsReadDoesNotConsumeBlockGas(t *testing.T) {
	app, err := NewAppWithOptions(TestAppOptions(memdb.NewMemDB()))
	require.NoError(t, err)

	initRes := app.InitChain(abci.RequestInitChain{
		ChainID: "test-chain",
		ConsensusParams: &abci.ConsensusParams{
			Block: &abci.BlockParams{MaxGas: 1},
		},
		AppState: DefaultGenState(),
	})
	require.True(t, initRes.ResponseBase.IsOK(), "InitChain failed: %v", initRes.ResponseBase.Error)

	app.BeginBlock(abci.RequestBeginBlock{
		Header: &bft.Header{ChainID: "test-chain", Height: 1},
	})

	from := crypto.AddressFromPreimage([]byte("from"))
	to := crypto.AddressFromPreimage([]byte("to"))
	txBytes := amino.MustMarshal(std.Tx{
		Msgs: []std.Msg{bank.NewMsgSend(from, to, std.NewCoins(std.NewCoin("ugnot", 1)))},
		Fee:  std.Fee{GasWanted: 2, GasFee: std.NewCoin("ugnot", 1)},
	})

	for range 2 {
		res := app.DeliverTx(abci.RequestDeliverTx{Tx: txBytes})
		require.IsType(t, std.InvalidGasWantedError{}, res.Error, "unexpected response: %+v", res)
		require.Equal(t, int64(0), res.GasUsed)
		require.Contains(t, res.Log, "invalid gas-wanted; got: 2 block-max-gas: 1")
	}
}

func TestDeliverTxVMParamsReadDoesNotConsumeBlockHeadroom(t *testing.T) {
	app, err := NewAppWithOptions(TestAppOptions(memdb.NewMemDB()))
	require.NoError(t, err)

	key := getDummyKey(t)
	from := key.PubKey().Address()
	state := DefaultGenState()
	// Keep the valid transaction's governed store costs below the block limit;
	// the ante-site params read still uses the default gas config before
	// applying these. Zeroing Fixed*/Min* makes cacheStore fall back to the
	// tree estimate, which for this store is far below the pinned defaults --
	// without it the transaction OOGs and the test cannot run.
	// Named vmParams, not params: the latter shadows the params package import.
	vmParams := state.VM.Params
	vmParams.MinGetReadDepth100, vmParams.FixedGetReadDepth100 = 0, 0
	vmParams.MinSetReadDepth100, vmParams.FixedSetReadDepth100 = 0, 0
	vmParams.MinWriteDepth100, vmParams.FixedWriteDepth100 = 0, 0
	vmParams.IterNextCostFlat = 1
	state.VM = vm.NewGenesisState(vmParams)
	// Keep headroom just below the measured read cost so the pre-fix metered
	// read OOGs. Measured, not pinned, to survive DefaultGasConfig changes.
	// The helper measures against a flat store, so this is a lower bound on
	// what the depth-estimating production store charges -- the safe direction.
	anteReadCost := measureVMParamsReadGas(t, vmParams)
	gasLimit := anteReadCost - 1
	state.Balances = []Balance{{
		Address: from,
		Amount:  std.NewCoins(std.NewCoin("ugnot", 20_000_000)),
	}}
	initRes := app.InitChain(abci.RequestInitChain{
		ChainID: "test-chain",
		ConsensusParams: &abci.ConsensusParams{
			Block: &abci.BlockParams{MaxGas: gasLimit},
		},
		AppState: state,
	})
	require.True(t, initRes.ResponseBase.IsOK(), "InitChain failed: %v", initRes.ResponseBase.Error)

	app.BeginBlock(abci.RequestBeginBlock{
		Header: &bft.Header{ChainID: "test-chain", Height: 1},
	})

	tx := createAndSignTx(t, []std.Msg{
		bank.NewMsgSend(from, crypto.AddressFromPreimage([]byte("to")), std.NewCoins(std.NewCoin("ugnot", 1))),
	}, "test-chain", key)
	tx.Fee.GasWanted = gasLimit
	signBytes, err := tx.GetSignBytes("test-chain", 0, 0)
	require.NoError(t, err)
	tx.Signatures[0].Signature, err = key.Sign(signBytes)
	require.NoError(t, err)

	res := app.DeliverTx(abci.RequestDeliverTx{Tx: amino.MustMarshal(tx)})
	// IsOK is the actual guard: GasWanted == gasLimit, so a metered ante-site
	// read pushes the transaction over its own meter and it fails here.
	require.True(t, res.IsOK(), "DeliverTx failed: %+v", res)
	t.Logf("reported GasUsed=%d, block MaxGas=%d", res.GasUsed, gasLimit)
	// Bounded by GasWanted for any successful tx; kept as a sanity check.
	require.Less(t, res.GasUsed, gasLimit)
}

func measureVMParamsReadGas(t *testing.T, vmParams vm.Params) int64 {
	t.Helper()

	key := types.NewStoreKey("params")
	backing := dbadapter.Store{DB: memdb.NewMemDB()}
	newStore := func() types.MultiStore {
		return cachemulti.NewFromStores(map[types.StoreKey]types.Store{key: backing}, nil)
	}
	newContext := func(ms types.MultiStore) sdk.Context {
		return sdk.NewContext(sdk.RunTxModeDeliver, ms, &bft.Header{ChainID: "test-chain"}, log.NewNoopLogger())
	}

	prmk := params.NewParamsKeeper(key)
	vmk := vm.NewVMKeeper(nil, key, nil, nil, prmk)
	prmk.Register(vm.ModuleName, vmk)
	seed := newStore()
	require.NoError(t, vmk.SetParams(newContext(seed), vmParams))
	seed.MultiWrite()

	meter := types.NewGasMeter(1 << 62)
	ctx := newContext(newStore()).WithGasMeter(meter).WithGasConfig(types.DefaultGasConfig())
	require.Equal(t, vmParams, vmk.GetParams(ctx))
	require.Positive(t, meter.GasConsumed())
	return meter.GasConsumed()
}
