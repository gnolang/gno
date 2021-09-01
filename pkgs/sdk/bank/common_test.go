package bank

// DONTCOVER

import (
	bft "github.com/gnolang/gno/pkgs/bft/types"
	dbm "github.com/gnolang/gno/pkgs/db"
	"github.com/gnolang/gno/pkgs/log"

	"github.com/gnolang/gno/pkgs/sdk"
	"github.com/gnolang/gno/pkgs/sdk/auth"
	"github.com/gnolang/gno/pkgs/std"
	"github.com/gnolang/gno/pkgs/store"
	"github.com/gnolang/gno/pkgs/store/iavl"
)

type testInput struct {
	ctx  sdk.Context
	bank BankKeeper
	acck auth.AccountKeeper
}

func setupTestInput() testInput {
	db := dbm.NewMemDB()

	authCapKey := store.NewStoreKey("authCapKey")

	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(authCapKey, iavl.StoreConstructor, db)
	ms.LoadLatestVersion()

	ctx := sdk.NewContext(ms, &bft.Header{ChainID: "test-chain-id"}, false, log.NewNopLogger())
	acck := auth.NewAccountKeeper(
		authCapKey, std.ProtoBaseAccount,
	)

	bank := NewBankKeeper(acck)

	return testInput{ctx: ctx, bank: bank, acck: acck}
}
