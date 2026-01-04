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
	pm "github.com/gnolang/gno/tm2/pkg/sdk/params"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/iavl"
)

type testEnv struct {
	ctx   sdk.Context
	vmk   *VMKeeper
	bankk bankm.BankKeeper
	acck  authm.AccountKeeper
	prmk  pm.ParamsKeeper
	vmh   vmHandler
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

	ctx := sdk.NewContext(sdk.RunTxModeDeliver, ms, &bft.Header{ChainID: "test-chain-id", Height: 42}, log.NewNoopLogger())

	prmk := pm.NewParamsKeeper(iavlCapKey)
	acck := authm.NewAccountKeeper(iavlCapKey, prmk.ForModule(authm.ModuleName), std.ProtoBaseAccount)
	bankk := bankm.NewBankKeeper(acck, prmk.ForModule(bankm.ModuleName))
	vmk := NewVMKeeper(baseCapKey, iavlCapKey, acck, bankk, prmk)

	prmk.Register(authm.ModuleName, acck)
	prmk.Register(bankm.ModuleName, bankk)
	prmk.Register(ModuleName, vmk)
	vmk.SetParams(ctx, DefaultParams())

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
	vmh := NewHandler(vmk)

	return testEnv{ctx: ctx, vmk: vmk, bankk: bankk, acck: acck, prmk: prmk, vmh: vmh}
}
