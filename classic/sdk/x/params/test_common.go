// nolint: deadcode unused
package params

import (
	abci "github.com/tendermint/classic/abci/types"
	dbm "github.com/tendermint/classic/db"
	"github.com/tendermint/classic/libs/log"
	"github.com/tendermint/go-amino-x"

	"github.com/tendermint/classic/sdk/store"
	sdk "github.com/tendermint/classic/sdk/types"
)

type invalid struct{}

type s struct {
	I int
}

func defaultContext(key sdk.StoreKey, tkey sdk.StoreKey) sdk.Context {
	db := dbm.NewMemDB()
	cms := store.NewCommitMultiStore(db)
	cms.MountStoreWithDB(key, sdk.StoreTypeIAVL, db)
	cms.MountStoreWithDB(tkey, sdk.StoreTypeTransient, db)
	err := cms.LoadLatestVersion()
	if err != nil {
		panic(err)
	}
	ctx := sdk.NewContext(cms, abci.Header{}, false, log.NewNopLogger())
	return ctx
}

func testComponents() (sdk.Context, sdk.StoreKey, sdk.StoreKey, Keeper) {
	mkey := sdk.NewKVStoreKey("test")
	tkey := sdk.NewTransientStoreKey("transient_test")
	ctx := defaultContext(mkey, tkey)
	keeper := NewKeeper(mkey, tkey, DefaultCodespace)

	return ctx, mkey, tkey, keeper
}
