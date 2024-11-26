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
	bank   *BankKeeper
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
	acck := auth.NewAccountKeeper(
		authCapKey, std.ProtoBaseAccount,
	)
	km := params.NewPrefixKeyMapper()
	km.RegisterPrefix(ParamsPrefixKey)
	paramk := params.NewParamsKeeper(authCapKey, km)
	bank := NewBankKeeper(acck, paramk)

	return testEnv{ctx: ctx, bank: bank, acck: acck, paramk: paramk}
}
