package vm

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	authm "github.com/gnolang/gno/tm2/pkg/sdk/auth"
	bankm "github.com/gnolang/gno/tm2/pkg/sdk/bank"
	paramsm "github.com/gnolang/gno/tm2/pkg/sdk/params"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
	storebptree "github.com/gnolang/gno/tm2/pkg/store/bptree"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/stretchr/testify/require"
)

type paramsGasTestEnv struct {
	ctx   sdk.Context
	vmk   *VMKeeper
	prmk  paramsm.ParamsKeeper
	acck  authm.AccountKeeper
	bankk bankm.BankKeeper
	ms    store.CommitMultiStore
}

func newParamsGasTestEnv(t *testing.T, db *memdb.MemDB, genesis bool) paramsGasTestEnv {
	t.Helper()

	baseKey := store.NewStoreKey("baseCapKey")
	mainKey := store.NewStoreKey("mainCapKey")
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(baseKey, dbadapter.StoreConstructor, db)
	ms.MountStoreWithDB(mainKey, storebptree.FastStoreConstructor, db)
	require.NoError(t, ms.LoadLatestVersion())

	ctx := sdk.NewContext(sdk.RunTxModeDeliver, ms,
		&bft.Header{ChainID: "params-bundle-gas-test", Height: 42},
		log.NewNoopLogger(),
	)
	prmk := paramsm.NewParamsKeeper(mainKey)
	acck := authm.NewAccountKeeper(mainKey, prmk.ForModule(authm.ModuleName), std.ProtoBaseAccount, std.ProtoBaseSessionAccount)
	bankk := bankm.NewBankKeeper(acck, prmk.ForModule(bankm.ModuleName))
	vmk := NewVMKeeper(baseKey, mainKey, acck, bankk, prmk)
	prmk.Register(authm.ModuleName, acck)
	prmk.Register(bankm.ModuleName, bankk)
	prmk.Register(ModuleName, vmk)

	if genesis {
		acck.SetParams(ctx, authm.DefaultParams())
		bankk.SetParams(ctx, bankm.DefaultParams())
		require.NoError(t, vmk.SetParams(ctx, DefaultParams()))
	}

	mcw := ms.MultiCacheWrap()
	vmk.Initialize(log.NewNoopLogger(), mcw)
	if genesis {
		stdlibCtx := vmk.MakeGnoTransactionStore(ctx.WithMultiStore(mcw))
		vmk.LoadStdlibCached(stdlibCtx, filepath.Join("..", "..", "..", "..", "gnovm", "stdlibs"))
		vmk.CommitGnoTransactionStore(stdlibCtx)
	}
	mcw.MultiWrite()
	vmk.PopulateStdlibCache()

	return paramsGasTestEnv{ctx: ctx, vmk: vmk, prmk: prmk, acck: acck, bankk: bankk, ms: ms}
}

var paramsGasImportPattern = regexp.MustCompile(`"(gno\.land/[pr]/[\w/.\-]+)"`)

func deployParamsGasPackage(t *testing.T, env paramsGasTestEnv, ctx sdk.Context, deployer crypto.Address, pkgPath string, files []*std.MemFile) {
	t.Helper()

	deployed := make(map[string]bool)
	var deployDependencies func(string)
	deployDependencies = func(importPath string) {
		if deployed[importPath] {
			return
		}
		deployed[importPath] = true

		rel := strings.TrimPrefix(importPath, "gno.land/")
		info, err := os.Stat(filepath.Join(examplesDir(), rel))
		if err != nil || !info.IsDir() {
			return
		}
		dependencyFiles := loadExamplePackage(importPath)
		for _, file := range dependencyFiles {
			for _, match := range paramsGasImportPattern.FindAllStringSubmatch(file.Body, -1) {
				deployDependencies(match[1])
			}
		}
		require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(deployer, importPath, dependencyFiles)))
	}

	for _, file := range files {
		for _, match := range paramsGasImportPattern.FindAllStringSubmatch(file.Body, -1) {
			deployDependencies(match[1])
		}
	}
	require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(deployer, pkgPath, files)))
}

type paramsGasMeter struct {
	store.GasMeter
	counts map[string]int64
	gas    map[string]int64
}

func newParamsGasMeter() *paramsGasMeter {
	return &paramsGasMeter{
		GasMeter: store.NewGasMeter(1_000_000_000),
		counts:   make(map[string]int64),
		gas:      make(map[string]int64),
	}
}

func (m *paramsGasMeter) ConsumeGas(amount int64, descriptor string) {
	m.counts[descriptor]++
	m.gas[descriptor] += amount
	m.GasMeter.ConsumeGas(amount, descriptor)
}

type coldTransferGas struct {
	total          int64
	params         int64
	paramReadCount int64
	paramReadGas   int64
}

func measureColdGRC20Transfer(t *testing.T, db *memdb.MemDB, legacy bool, caller crypto.Address, pkgPath string) coldTransferGas {
	t.Helper()

	env := newParamsGasTestEnv(t, db, false)
	cache := env.ms.MultiCacheWrap()
	ctx := env.ctx.WithMultiStore(cache)
	if legacy {
		// The old implementation had no bundle lookup. Warm only the known-
		// absent compatibility key without a meter so every other store access
		// remains cold while this baseline charges the former 14-field path.
		_, ok := env.vmk.getParamsBundle(ctx.WithGasMeter(nil))
		require.False(t, ok)
	}

	meter := newParamsGasMeter()
	ctx = ctx.WithGasMeter(meter)
	params := env.vmk.GetParams(ctx)
	paramGas := meter.GasConsumed()
	paramReadCount := meter.counts["DepthReadFlat"]
	paramReadGas := meter.gas["DepthReadFlat"]

	gasConfig := store.DefaultGasConfig()
	params.ApplyToGasConfig(&gasConfig)
	ctx = ctx.WithGasConfig(gasConfig)
	callCtx := env.vmk.MakeGnoTransactionStore(ctx)
	_, err := env.vmk.Call(callCtx, NewMsgCall(caller, nil, pkgPath, "Transfer", nil))
	require.NoError(t, err)

	return coldTransferGas{
		total:          meter.GasConsumed(),
		params:         paramGas,
		paramReadCount: paramReadCount,
		paramReadGas:   paramReadGas,
	}
}

func TestParamsBundleColdGRC20TransferGas(t *testing.T) {
	db := memdb.NewMemDB()
	env := newParamsGasTestEnv(t, db, true)
	deployer := crypto.AddressFromPreimage([]byte("params-bundle-gas-deployer"))
	env.acck.SetAccount(env.ctx, env.acck.NewAccountWithAddress(env.ctx, deployer))
	require.NoError(t, env.bankk.SetCoins(env.ctx, deployer,
		std.NewCoins(std.NewCoin("ugnot", 1_000_000_000_000)),
	))

	const pkgPath = "gno.land/r/test/paramsbundlegrc20"
	files := []*std.MemFile{
		{Name: "gnomod.toml", Body: "module = \"" + pkgPath + "\"\ngno = \"0.9\"\n"},
		{Name: "token.gno", Body: `package paramsbundlegrc20

import (
	"chain"
	"gno.land/p/demo/tokens/grc20"
)

var ledger *grc20.PrivateLedger

func init(cur realm) {
	_, ledger = grc20.NewToken(0, cur, "Bundle Gas", "BGS", 0)
	checkErr(ledger.Mint(chain.PackageAddress("holder"), 1000))
}

func Transfer(cur realm) {
	checkErr(ledger.Transfer(
		chain.PackageAddress("holder"),
		chain.PackageAddress("recipient"),
		1,
	))
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
`},
	}
	deployCtx := env.vmk.MakeGnoTransactionStore(env.ctx)
	deployParamsGasPackage(t, env, deployCtx, deployer, pkgPath, files)
	env.vmk.CommitGnoTransactionStore(deployCtx)
	env.ms.Commit()

	bundled := measureColdGRC20Transfer(t, db, false, deployer, pkgPath)

	// Produce the same committed token state with the legacy params layout.
	legacyStore := store.NewCommitMultiStore(db)
	legacyBaseKey := store.NewStoreKey("baseCapKey")
	legacyMainKey := store.NewStoreKey("mainCapKey")
	legacyStore.MountStoreWithDB(legacyBaseKey, dbadapter.StoreConstructor, db)
	legacyStore.MountStoreWithDB(legacyMainKey, storebptree.FastStoreConstructor, db)
	require.NoError(t, legacyStore.LoadLatestVersion())
	legacyCtx := sdk.NewContext(sdk.RunTxModeDeliver, legacyStore,
		&bft.Header{ChainID: "params-bundle-gas-test", Height: 42},
		log.NewNoopLogger(),
	)
	paramsm.NewParamsKeeper(legacyMainKey).SetBytes(legacyCtx, paramsBundleKey, nil)
	legacyStore.Commit()

	legacy := measureColdGRC20Transfer(t, db, true, deployer, pkgPath)

	require.Equal(t, int64(1), bundled.paramReadCount)
	require.Equal(t, int64(reflect.TypeFor[Params]().NumField()), legacy.paramReadCount)
	require.Equal(t, legacy.total-legacy.params, bundled.total-bundled.params,
		"the mint/deploy/transfer inputs and all non-param gas must stay identical")

	perReadGas := bundled.paramReadGas
	require.Equal(t, perReadGas, legacy.paramReadGas/legacy.paramReadCount,
		"the live B+ depth must be the same for both layouts")

	params := DefaultParams()
	bundleBytes := len(amino.MustMarshal(params))
	legacyBytes := 0
	rv := reflect.ValueOf(params)
	for i := range rv.NumField() {
		legacyBytes += len(amino.MustMarshalJSON(rv.Field(i).Interface()))
	}
	wantSaving := int64(reflect.TypeFor[Params]().NumField()-1)*perReadGas +
		int64(legacyBytes-bundleBytes)*store.DefaultGasConfig().ReadCostPerByte
	require.Equal(t, wantSaving, legacy.total-bundled.total,
		"cold transfer saving must equal 13 removed reads plus the encoding-byte delta")
	t.Logf("cold transfer gas: legacy=%d bundled=%d saving=%d params_reads=%d->%d read_gas=%d",
		legacy.total, bundled.total, legacy.total-bundled.total,
		legacy.paramReadCount, bundled.paramReadCount, perReadGas)
}
