package gnoland

import (
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/sdk/params"

	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
)

type testEnv struct {
	ctx   sdk.Context
	acck  auth.AccountKeeper
	bankk bank.BankKeeper
}

func setupTestEnv() testEnv {
	db := memdb.NewMemDB()

	authCapKey := store.NewStoreKey("authCapKey")

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(authCapKey, iavl.StoreConstructor, db)
	ms.LoadLatestVersion()
	prmk := params.NewParamsKeeper(authCapKey)
	acck := auth.NewAccountKeeper(authCapKey, prmk.ForModule(auth.ModuleName), ProtoGnoAccount)
	bankk := bank.NewBankKeeper(acck, prmk.ForModule(bank.ModuleName))
	prmk.Register(auth.ModuleName, acck)
	prmk.Register(bank.ModuleName, bankk)

	ctx := sdk.NewContext(sdk.RunTxModeDeliver, ms, &bft.Header{Height: 1, ChainID: "test-chain-id"}, log.NewNoopLogger())

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

	return testEnv{ctx: ctx, acck: acck, bankk: bankk}
}
