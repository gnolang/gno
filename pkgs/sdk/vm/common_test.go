package vm

// DONTCOVER

import (
	bft "github.com/gnolang/gno/pkgs/bft/types"
	dbm "github.com/gnolang/gno/pkgs/db"
	"github.com/gnolang/gno/pkgs/log"

	"github.com/gnolang/gno/pkgs/sdk"
	authm "github.com/gnolang/gno/pkgs/sdk/auth"
	bankm "github.com/gnolang/gno/pkgs/sdk/bank"
	"github.com/gnolang/gno/pkgs/std"
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

	ctx := sdk.NewContext(ms, &bft.Header{ChainID: "test-chain-id"}, false, log.NewNopLogger())
	acck := authm.NewAccountKeeper(iavlCapKey, std.ProtoBaseAccount)
	bank := bankm.NewBankKeeper(acck)
	vmk := NewVMKeeper(baseCapKey, iavlCapKey, acck, bank)

	return testEnv{ctx: ctx, vmk: vmk, bank: bank, acck: acck}
}
