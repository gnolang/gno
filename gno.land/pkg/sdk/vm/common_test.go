package vm

// DONTCOVER

import (
	"path/filepath"

	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	authm "github.com/gnolang/gno/tm2/pkg/sdk/auth"
	bankm "github.com/gnolang/gno/tm2/pkg/sdk/bank"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
)

type testEnv struct {
	ctx  sdk.Context
	vmk  *VMKeeper
	bank bankm.BankKeeper
	acck authm.AccountKeeper
}

func setupTestEnv() testEnv {
	return _setupTestEnv(true)
}

func setupTestEnvCold() testEnv {
	return _setupTestEnv(false)
}

func _setupTestEnv(cacheStdlibs bool) testEnv {
	db := memdb.NewMemDB()

	baseCapKey := store.NewStoreKey("baseCapKey")
	iavlCapKey := store.NewStoreKey("iavlCapKey")

	// Mount db store and iavlstore
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(baseCapKey, dbadapter.StoreConstructor, db)
	ms.MountStoreWithDB(iavlCapKey, iavl.StoreConstructor, db)
	ms.LoadLatestVersion()

	ctx := sdk.NewContext(sdk.RunTxModeDeliver, ms, &bft.Header{ChainID: "test-chain-id"}, log.NewNoopLogger())
	acck := authm.NewAccountKeeper(iavlCapKey, std.ProtoBaseAccount)
	bank := bankm.NewBankKeeper(acck)
	vmk := NewVMKeeper(baseCapKey, iavlCapKey, acck, bank, 100_000_000)

	mcw := ms.MultiCacheWrap()
	vmk.Initialize(log.NewNoopLogger(), mcw)
	stdlibCtx := vmk.MakeGnoTransactionStore(ctx.WithMultiStore(mcw))
	stdlibsDir := filepath.Join("..", "..", "..", "..", "gnovm", "stdlibs")
	if cacheStdlibs {
		vmk.LoadStdlibCached(stdlibCtx, stdlibsDir)
	} else {
		vmk.LoadStdlib(stdlibCtx, stdlibsDir)
	}
	vmk.CommitGnoTransactionStore(stdlibCtx)
	mcw.MultiWrite()

	return testEnv{ctx: ctx, vmk: vmk, bank: bank, acck: acck}
}
