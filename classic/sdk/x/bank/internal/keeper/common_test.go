package keeper

// DONTCOVER

import (
	abci "github.com/tendermint/classic/abci/types"
	dbm "github.com/tendermint/classic/db"
	"github.com/tendermint/classic/libs/log"
	"github.com/tendermint/go-amino-x"

	"github.com/tendermint/classic/sdk/store"
	sdk "github.com/tendermint/classic/sdk/types"
	"github.com/tendermint/classic/sdk/x/auth"
	"github.com/tendermint/classic/sdk/x/bank/internal/types"
	"github.com/tendermint/classic/sdk/x/params"
)

type testInput struct {
	ctx sdk.Context
	k   Keeper
	ak  auth.AccountKeeper
	pk  params.Keeper
}

func setupTestInput() testInput {
	db := dbm.NewMemDB()

	authCapKey := sdk.NewKVStoreKey("authCapKey")
	keyParams := sdk.NewKVStoreKey("params")
	tkeyParams := sdk.NewTransientStoreKey("transient_params")

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(authCapKey, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(keyParams, sdk.StoreTypeIAVL, db)
	ms.MountStoreWithDB(tkeyParams, sdk.StoreTypeTransient, db)
	ms.LoadLatestVersion()

	blacklistedAddrs := make(map[string]bool)
	blacklistedAddrs[sdk.AccAddress([]byte("moduleAcc")).String()] = true

	pk := params.NewKeeper(keyParams, tkeyParams, params.DefaultCodespace)

	ak := auth.NewAccountKeeper(
		authCapKey, pk.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount,
	)
	ctx := sdk.NewContext(ms, abci.Header{ChainID: "test-chain-id"}, false, log.NewNopLogger())

	ak.SetParams(ctx, auth.DefaultParams())

	bankKeeper := NewBaseKeeper(ak, pk.Subspace(types.DefaultParamspace), types.DefaultCodespace, blacklistedAddrs)
	bankKeeper.SetSendEnabled(ctx, true)

	return testInput{ctx: ctx, k: bankKeeper, ak: ak, pk: pk}
}
