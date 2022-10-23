package vm

// DONTCOVER

import (
	"github.com/gnolang/gno/gnoland/types"
	bft "github.com/gnolang/gno/pkgs/bft/types"
	dbm "github.com/gnolang/gno/pkgs/db"
	"github.com/gnolang/gno/pkgs/log"

	"github.com/gnolang/gno/pkgs/sdk"
	authm "github.com/gnolang/gno/pkgs/sdk/auth"
	bankm "github.com/gnolang/gno/pkgs/sdk/bank"
	"github.com/gnolang/gno/pkgs/store"
	"github.com/gnolang/gno/pkgs/store/dbadapter"
	"github.com/gnolang/gno/pkgs/store/iavl"
)

type testEnv struct {
	ctx  sdk.Context
	vmk  *VMKeeper
	bank bankm.BankKeeper
	acck authm.AccountKeeper
}

func setupTestEnv() testEnv {
	db := dbm.NewMemDB()

	baseCapKey := store.NewStoreKey("baseCapKey")
	iavlCapKey := store.NewStoreKey("iavlCapKey")

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(baseCapKey, dbadapter.StoreConstructor, db)
	ms.MountStoreWithDB(iavlCapKey, iavl.StoreConstructor, db)
	ms.LoadLatestVersion()

	ctx := sdk.NewContext(sdk.RunTxModeDeliver, ms, &bft.Header{ChainID: "test-chain-id"}, log.NewNopLogger())
	acck := authm.NewAccountKeeper(iavlCapKey, types.ProtoGnoAccount)
	bank := bankm.NewBankKeeper(acck)
	vmk := NewVMKeeper(baseCapKey, iavlCapKey, acck, bank, "../../../stdlibs")

	vmk.Initialize(ms.MultiCacheWrap())

	return testEnv{ctx: ctx, vmk: vmk, bank: bank, acck: acck}
}
