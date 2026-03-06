package vm

// DONTCOVER

import (
	"path/filepath"
	"strings"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
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

// examplesDir returns the path to the examples directory relative to this test file.
func examplesDir() string {
	return filepath.Join("..", "..", "..", "..", "examples", "gno.land")
}

// loadExamplePackage reads a package from the examples directory and returns MemFiles.
// pkgPath is the full package path (e.g., "gno.land/p/nt/avl").
// The package files are read from examples/gno.land/{p,r}/...
func loadExamplePackage(pkgPath string) []*std.MemFile {
	// Extract the relative path from pkgPath (e.g., "gno.land/p/nt/avl" -> "p/nt/avl")
	const prefix = "gno.land/"
	if len(pkgPath) <= len(prefix) {
		panic("invalid package path: " + pkgPath)
	}
	relPath := pkgPath[len(prefix):]
	dir := filepath.Join(examplesDir(), relPath)

	memPkg, err := gno.ReadMemPackage(dir, pkgPath, gno.MPUserProd)
	if err != nil {
		panic("failed to read example package " + pkgPath + ": " + err.Error())
	}
	return memPkg.Files
}

// deployExamplePackage deploys a package from the examples directory.
// It reads the package from disk and deploys it using the provided VMKeeper.
func deployExamplePackage(env testEnv, ctx sdk.Context, deployer crypto.Address, pkgPath string) error {
	files := loadExamplePackage(pkgPath)
	msg := NewMsgAddPackage(deployer, pkgPath, files)
	return env.vmk.AddPackage(ctx, msg)
}

// deployExamplePackageWithPatch deploys a package from the examples directory,
// applying string replacements to the source files before deployment.
// patches is a map of old string -> new string replacements to apply.
func deployExamplePackageWithPatch(env testEnv, ctx sdk.Context, deployer crypto.Address, pkgPath string, patches map[string]string) error {
	files := loadExamplePackage(pkgPath)
	// Apply patches to all files
	for i, f := range files {
		body := f.Body
		for old, new := range patches {
			body = strings.ReplaceAll(body, old, new)
		}
		if body != f.Body {
			files[i] = &std.MemFile{Name: f.Name, Body: body}
		}
	}
	msg := NewMsgAddPackage(deployer, pkgPath, files)
	return env.vmk.AddPackage(ctx, msg)
}
