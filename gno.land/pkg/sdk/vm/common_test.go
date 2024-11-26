package vm

// DONTCOVER

import (
	"fmt"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	authm "github.com/gnolang/gno/tm2/pkg/sdk/auth"
	bankm "github.com/gnolang/gno/tm2/pkg/sdk/bank"
	paramsm "github.com/gnolang/gno/tm2/pkg/sdk/params"
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
	vmh  vmHandler
}

func setupTestEnv() testEnv {
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
	prmk := paramsm.NewParamsKeeper(iavlCapKey, "params")
	vmk := NewVMKeeper(baseCapKey, iavlCapKey, acck, bank, prmk)

	mcw := ms.MultiCacheWrap()
	vmk.Initialize(log.NewNoopLogger(), mcw)
	stdlibCtx := vmk.MakeGnoTransactionStore(ctx.WithMultiStore(mcw))
	loadStdlibs(vmk.gnoStore)
	vmk.CommitGnoTransactionStore(stdlibCtx)
	mcw.MultiWrite()
	vmh := NewHandler(vmk)

	return testEnv{ctx: ctx, vmk: vmk, bank: bank, acck: acck, vmh: vmh}
}

func loadStdlibs(store gno.Store) {
	stdlibInitList := stdlibs.InitOrder()
	for _, lib := range stdlibInitList {
		if lib == "testing" {
			// XXX: testing is skipped for now while it uses testing-only packages
			// like fmt and encoding/json
			continue
		}
		loadStdlibPackage(lib, store)
	}
}

func loadStdlibPackage(pkgPath string, store gno.Store) {
	memPkg := stdlibs.EmbeddedMemPackage(pkgPath)
	if memPkg == nil || memPkg.IsEmpty() {
		// no gno files are present
		panic(fmt.Sprintf("failed loading stdlib %q: not a valid MemPackage", pkgPath))
	}

	m := gno.NewMachineWithOptions(gno.MachineOptions{
		PkgPath: "gno.land/r/stdlibs/" + pkgPath,
		// PkgPath: pkgPath, XXX why?
		Store: store,
	})
	defer m.Release()
	m.RunMemPackage(memPkg, true)
}
