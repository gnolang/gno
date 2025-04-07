package params

import (
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk"

	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
)

type testEnv struct {
	ctx    sdk.Context
	store  store.Store
	keeper ParamsKeeper
}

func setupTestEnv() testEnv {
	db := memdb.NewMemDB()
	paramsCapKey := store.NewStoreKey("paramsCapKey")
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(paramsCapKey, iavl.StoreConstructor, db)
	ms.LoadLatestVersion()

	prmk := NewParamsKeeper(paramsCapKey)
	dk := NewDummyKeeper(prmk.ForModule(dummyModuleName))
	prmk.Register(dummyModuleName, dk)

	ctx := sdk.NewContext(sdk.RunTxModeDeliver, ms, &bft.Header{Height: 1, ChainID: "test-chain-id"}, log.NewNoopLogger())
	// XXX: context key?
	ctx = ctx.WithConsensusParams(&abci.ConsensusParams{
		Block: &abci.BlockParams{
			MaxTxBytes:    1024,
			MaxDataBytes:  1024 * 100,
			MaxBlockBytes: 1024 * 100,
			MaxGas:        10 * 1000 * 1000,
			TimeIotaMS:    10,
		},
		Validator: &abci.ValidatorParams{
			PubKeyTypeURLs: []string{}, // XXX
		},
	})

	stor := ctx.Store(paramsCapKey)
	return testEnv{ctx: ctx, store: stor, keeper: prmk}
}

const dummyModuleName = "params_test"

type DummyKeeper struct {
	prmk ParamsKeeperI
}

func NewDummyKeeper(prmk ParamsKeeperI) DummyKeeper {
	return DummyKeeper{
		prmk: prmk,
	}
}

func (dk DummyKeeper) WillSetParam(ctx sdk.Context, key string, value any) {
	// do nothing
}
