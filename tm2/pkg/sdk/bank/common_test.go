package bank

// DONTCOVER

import (
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/log"

	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/params"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
)

type testEnv struct {
	ctx    sdk.Context
	bank   BankKeeper
	acck   auth.AccountKeeper
	paramk params.ParamsKeeper
}

func setupTestEnv() testEnv {
	db := memdb.NewMemDB()

	authCapKey := store.NewStoreKey("authCapKey")

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(authCapKey, iavl.StoreConstructor, db)
	ms.LoadLatestVersion()
	ctx := sdk.NewContext(sdk.RunTxModeDeliver, ms, &bft.Header{ChainID: "test-chain-id"}, log.NewNoopLogger())

	paramk := params.NewParamsKeeper(authCapKey)
	acck := auth.NewAccountKeeper(
		authCapKey, paramk, std.ProtoBaseAccount,
	)
	bank := NewBankKeeper(acck, paramk)

	paramk.Register(acck.GetParamfulKey(), acck)
	paramk.Register(bank.GetParamfulKey(), bank)

	return testEnv{ctx: ctx, bank: bank, acck: acck, paramk: paramk}
}
