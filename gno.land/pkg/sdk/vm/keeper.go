package vm

// TODO: move most of the logic in ROOT/gno.land/...

import (
	"bytes"
	"context"
	goerrors "errors"
	"fmt"
	"io"
	"iter"
	"log/slog"
	"maps"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/gnovm/pkg/doc"
	"github.com/gnolang/gno/gnovm/pkg/gnoenv"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/gnovm/stdlibs"
	"github.com/gnolang/gno/gnovm/stdlibs/chain"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/errors"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/gnolang/gno/tm2/pkg/overflow"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	stypes "github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/gnolang/gno/tm2/pkg/telemetry"
	"github.com/gnolang/gno/tm2/pkg/telemetry/metrics"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	maxAllocTx    = 500_000_000
	maxAllocQuery = 1_500_000_000 // higher limit for queries
	maxGasQuery   = 3_000_000_000 // same as max block gas
)

// vm.VMKeeperI defines a module interface that supports Gno
// smart contracts programming (scripting).
type VMKeeperI interface {
	AddPackage(ctx sdk.Context, msg MsgAddPackage) error
	Call(ctx sdk.Context, msg MsgCall) (res string, err error)
	QueryEval(ctx sdk.Context, pkgPath string, expr string) (res string, err error)
	Run(ctx sdk.Context, msg MsgRun) (res string, err error)
	LoadStdlib(ctx sdk.Context, stdlibDir string)
	LoadStdlibCached(ctx sdk.Context, stdlibDir string)
	MakeGnoTransactionStore(ctx sdk.Context) sdk.Context
	CommitGnoTransactionStore(ctx sdk.Context)
	InitGenesis(ctx sdk.Context, data GenesisState)
}

var _ VMKeeperI = &VMKeeper{}

// VMKeeper holds all package code and store state.
type VMKeeper struct {
	// Needs to be explicitly set, like in the case of gnodev.
	Output io.Writer

	baseKey store.StoreKey
	iavlKey store.StoreKey
	acck    AccountKeeperI
	bank    BankKeeperI
	prmk    ParamsKeeperI

	// cached, the DeliverTx persistent state.
	gnoStore gno.Store
	// committed typecheck cache
	typeCheckCache  gno.TypeCheckCache
	testStdlibCache testStdlibCache
}

// NewVMKeeper returns a new VMKeeper.
// NOTE: prmk must be the root ParamsKeeper such that
// ExecContext.Params may set any module's parameter.
func NewVMKeeper(
	baseKey store.StoreKey,
	iavlKey store.StoreKey,
	acck AccountKeeperI,
	bank BankKeeperI,
	prmk ParamsKeeperI,
) *VMKeeper {
	vmk := &VMKeeper{
		baseKey:        baseKey,
		iavlKey:        iavlKey,
		acck:           acck,
		bank:           bank,
		prmk:           prmk,
		typeCheckCache: gno.TypeCheckCache{},
		testStdlibCache: testStdlibCache{
			rootDir: gnoenv.RootDir(),
			cache:   map[string]*std.MemPackage{},
		},
	}

	return vmk
}

func (vm *VMKeeper) Initialize(
	logger *slog.Logger,
	ms store.MultiStore,
) {
	if vm.gnoStore != nil {
		panic("should not happen")
	}
	baseStore := ms.GetStore(vm.baseKey)
	iavlStore := ms.GetStore(vm.iavlKey)

	alloc := gno.NewAllocator(maxAllocTx)
	vm.gnoStore = gno.NewStore(alloc, baseStore, iavlStore)
	vm.gnoStore.SetNativeResolver(stdlibs.NativeResolver)

	if vm.gnoStore.NumMemPackages() > 0 {
		// for now, all mem packages must be re-run after reboot.
		// TODO remove this, and generally solve for in-mem garbage collection
		// and memory management across many objects/types/nodes/packages.
		start := time.Now()

		m2 := gno.NewMachineWithOptions(
			gno.MachineOptions{
				PkgPath: "",
				Output:  vm.Output,
				Store:   vm.gnoStore,
			})
		defer m2.Release()
		gno.DisableDebug()
		m2.PreprocessAllFilesAndSaveBlockNodes()
		gno.EnableDebug()

		opts := gno.TypeCheckOptions{
			Getter:     vm.gnoStore,
			TestGetter: vm.testStdlibCache.memPackageGetter(vm.gnoStore),
			Mode:       gno.TCLatestStrict,
			Cache:      vm.typeCheckCache,
		}
		for _, stdlib := range stdlibs.InitOrder() {
			mp := vm.gnoStore.GetMemPackage(stdlib)
			_, err := gno.TypeCheckMemPackage(mp, opts)
			if err != nil {
				panic(fmt.Errorf("intialization error type checking %q: %w", stdlib, err))
			}
		}

		logger.Debug("GnoVM packages preprocessed",
			"elapsed", time.Since(start))
	}
}

type stdlibCache struct {
	dir  string
	base store.Store
	iavl store.Store
	gno  gno.Store
}

var (
	cachedStdlibOnce         sync.Once
	cachedStdlib             stdlibCache
	cachedInitTypeCheckCache gno.TypeCheckCache
)

// LoadStdlibCached loads the Gno standard library into the given store.
//
// This works differently from [VMKeeper.LoadStdlib] as it performs an initial
// loading of the stdlib, which is then copied for future use.
//
// LoadStdlibCached is more efficient for programs which have to load a fresh
// keeper many times (including tests and gnodev). For normal node execution,
// LoadStdlib should be used instead, for lower memory consumption and faster
// cold start.
func (vm *VMKeeper) LoadStdlibCached(ctx sdk.Context, stdlibDir string) {
	cachedStdlibOnce.Do(func() {
		cachedStdlib = stdlibCache{
			dir:  stdlibDir,
			base: dbadapter.StoreConstructor(memdb.NewMemDB(), stypes.StoreOptions{}),
			iavl: dbadapter.StoreConstructor(memdb.NewMemDB(), stypes.StoreOptions{}),
		}

		gs := gno.NewStore(nil, cachedStdlib.base, cachedStdlib.iavl)
		gs.SetNativeResolver(stdlibs.NativeResolver)
		loadStdlib(gs, stdlibDir)
		cachedInitTypeCheckCache = make(gno.TypeCheckCache)
		opts := gno.TypeCheckOptions{
			Getter:     gs,
			TestGetter: vm.testStdlibCache.memPackageGetter(gs),
			Mode:       gno.TCLatestStrict,
			Cache:      cachedInitTypeCheckCache,
		}
		for _, lib := range stdlibs.InitOrder() {
			_, err := gno.TypeCheckMemPackage(gs.GetMemPackage(lib), opts)
			if err != nil {
				panic(fmt.Errorf("failed type checking stdlib %q: %w", lib, err))
			}
		}
		cachedStdlib.gno = gs
	})

	if stdlibDir != cachedStdlib.dir {
		panic(fmt.Sprintf(
			"cannot load cached stdlib: cached stdlib is in dir %q; wanted to load stdlib in dir %q",
			cachedStdlib.dir, stdlibDir,
		))
	}

	gs := vm.getGnoTransactionStore(ctx)
	gno.CopyFromCachedStore(gs, cachedStdlib.gno, cachedStdlib.base, cachedStdlib.iavl)
	vm.typeCheckCache = maps.Clone(cachedInitTypeCheckCache)
}

// LoadStdlib loads the Gno standard library into the given store. It will
// additionally execute type checking on the mempackages in the standard
// library.
func (vm *VMKeeper) LoadStdlib(ctx sdk.Context, stdlibDir string) {
	gs := vm.getGnoTransactionStore(ctx)
	loadStdlib(gs, stdlibDir)
	opts := gno.TypeCheckOptions{
		Getter:     gs,
		TestGetter: vm.testStdlibCache.memPackageGetter(gs),
		Mode:       gno.TCLatestStrict,
		Cache:      vm.getTypeCheckCache(ctx),
	}
	for _, lib := range stdlibs.InitOrder() {
		_, err := gno.TypeCheckMemPackage(gs.GetMemPackage(lib), opts)
		if err != nil {
			panic(fmt.Errorf("failed type checking stdlib %q: %w", lib, err))
		}
	}
}

func loadStdlib(store gno.Store, stdlibDir string) {
	stdlibInitList := stdlibs.InitOrder()
	for _, lib := range stdlibInitList {
		loadStdlibPackage(lib, stdlibDir, store)
	}
}

func loadStdlibPackage(pkgPath, stdlibDir string, store gno.Store) {
	stdlibPath := filepath.Join(stdlibDir, pkgPath)
	if !osm.DirExists(stdlibPath) {
		// does not exist.
		panic(fmt.Errorf("failed loading stdlib %q: does not exist", pkgPath))
	}
	memPkg, err := gno.ReadMemPackage(stdlibPath, pkgPath, gno.MPStdlibAll)
	if err != nil {
		// no gno files are present
		panic(fmt.Errorf("failed loading stdlib %q: %w", pkgPath, err))
	}

	m := gno.NewMachineWithOptions(gno.MachineOptions{
		// XXX: gno.land, vm.domain, other?
		PkgPath:     pkgPath,
		Store:       store,
		SkipPackage: true,
	})
	defer m.Release()
	m.RunMemPackage(memPkg, true)
}

type testStdlibCache struct {
	rootDir  string
	cache    map[string]*std.MemPackage // nil = no test package, use source; otherwise result from test stdlib
	cacheMtx sync.RWMutex
}

type testStdlibGetter struct {
	*testStdlibCache
	source gno.MemPackageGetter
}

func (tsc *testStdlibCache) memPackageGetter(source gno.Store) gno.MemPackageGetter {
	return testStdlibGetter{testStdlibCache: tsc, source: source}
}

func (tsg testStdlibGetter) GetMemPackage(pkgPath string) *std.MemPackage {
	// Only stdlibs have alternative versions.
	if !gno.IsStdlib(pkgPath) {
		return tsg.source.GetMemPackage(pkgPath)
	}

	tsg.cacheMtx.RLock()
	res, ok := tsg.cache[pkgPath]
	tsg.cacheMtx.RUnlock()
	// fast path: if cache was hit, return the mempackage from tsg.source (if
	// nil) or
	if ok {
		if res == nil {
			return tsg.source.GetMemPackage(pkgPath)
		}
		return res
	}

	// Cache miss: load package, and join it with the base package if necessary.
	sourceMpkg := tsg.source.GetMemPackage(pkgPath)
	// load from directory. NOTE: pkgPath is validated by `!gno.IsStdlib`,
	// hence it cannot contain path traversals like `../`.
	dir := filepath.Join(tsg.rootDir, "gnovm", "tests", "stdlibs", pkgPath)
	testMpkg, err := gno.ReadMemPackage(dir, pkgPath, gno.MPStdlibTest)
	if err != nil {
		tsg.cacheMtx.Lock()
		tsg.cache[pkgPath] = nil
		tsg.cacheMtx.Unlock()
		return sourceMpkg
	}
	if sourceMpkg != nil {
		testMpkg.Files = slices.Concat(sourceMpkg.Files, testMpkg.Files)
	}

	tsg.cacheMtx.Lock()
	tsg.cache[pkgPath] = testMpkg
	tsg.cacheMtx.Unlock()
	return testMpkg
}

type vmkContextKey int

const (
	vmkContextKeyStore vmkContextKey = iota
	vmkContextKeyTypeCheckCache
)

func (vm *VMKeeper) newGnoTransactionStore(ctx sdk.Context) gno.TransactionStore {
	base := ctx.Store(vm.baseKey)
	iavl := ctx.Store(vm.iavlKey)
	gasMeter := ctx.GasMeter()

	return vm.gnoStore.BeginTransaction(base, iavl, gasMeter)
}

func (vm *VMKeeper) MakeGnoTransactionStore(ctx sdk.Context) sdk.Context {
	return ctx.
		WithValue(vmkContextKeyTypeCheckCache, maps.Clone(vm.typeCheckCache)).
		WithValue(vmkContextKeyStore, vm.newGnoTransactionStore(ctx))
}

func (vm *VMKeeper) CommitGnoTransactionStore(ctx sdk.Context) {
	vm.getGnoTransactionStore(ctx).Write()
}

func (vm *VMKeeper) getTypeCheckCache(ctx sdk.Context) gno.TypeCheckCache {
	return ctx.Value(vmkContextKeyTypeCheckCache).(gno.TypeCheckCache)
}

func (vm *VMKeeper) getGnoTransactionStore(ctx sdk.Context) gno.TransactionStore {
	txStore := ctx.Value(vmkContextKeyStore).(gno.TransactionStore)
	txStore.ClearObjectCache()
	return txStore
}

// Namespace can be either a user or crypto address.
var reNamespace = regexp.MustCompile(`^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}/(?:r|p)/([\.~_a-zA-Z0-9]+)`)

// callRealmBool creates a Machine, imports pkgPath, calls funcName with args,
// and expects a single bool return value.
func (vm *VMKeeper) callRealmBool(
	ctx sdk.Context,
	creator crypto.Address,
	pkgPath, importAlias, funcName string,
	args ...any,
) (result bool, err error) {
	chainDomain := vm.getChainDomainParam(ctx)
	store := vm.getGnoTransactionStore(ctx)

	msgCtx := stdlibs.ExecContext{
		ChainID:         ctx.ChainID(),
		ChainDomain:     chainDomain,
		Height:          ctx.BlockHeight(),
		Timestamp:       ctx.BlockTime().Unix(),
		OriginCaller:    creator.Bech32(),
		OriginSendSpent: new(std.Coins),
		Banker:          NewSDKBanker(vm, ctx),
		Params:          NewSDKParams(vm.prmk, ctx),
		EventLogger:     ctx.EventLogger(),
	}

	m := gno.NewMachineWithOptions(
		gno.MachineOptions{
			PkgPath:  "",
			Output:   vm.Output,
			Store:    store,
			Context:  msgCtx,
			Alloc:    store.GetAllocator(),
			GasMeter: ctx.GasMeter(),
		})
	defer m.Release()
	defer doRecover(m, &err)

	mpv := gno.NewPackageNode("main", "main", nil).NewPackage(m.Alloc)
	m.SetActivePackage(mpv)
	m.RunDeclaration(gno.ImportD(importAlias, pkgPath))
	x := gno.Call(
		gno.Sel(gno.Nx(importAlias), funcName),
		args...,
	)

	ret := m.Eval(x)
	if len(ret) != 1 {
		return false, fmt.Errorf("callRealmBool: expected 1 return value, got %d", len(ret))
	}
	if ret[0].T.Kind() != gno.BoolKind {
		return false, fmt.Errorf("callRealmBool: expected bool return value, got %s", ret[0].T.Kind())
	}

	return ret[0].GetBool(), nil
}

// checkNamespacePermission check if the user as given has correct permssion to on the given pkg path
func (vm *VMKeeper) checkNamespacePermission(ctx sdk.Context, creator crypto.Address, pkgPath string) error {
	sysNamesPkg := vm.getSysNamesPkgParam(ctx)
	if sysNamesPkg == "" {
		return nil
	}
	chainDomain := vm.getChainDomainParam(ctx)

	store := vm.getGnoTransactionStore(ctx)

	if !strings.HasPrefix(pkgPath, chainDomain+"/") {
		return ErrInvalidPkgPath(pkgPath) // no match
	}

	match := reNamespace.FindStringSubmatch(pkgPath)
	switch len(match) {
	case 0:
		return ErrInvalidPkgPath(pkgPath) // no match
	case 2: // ok
	default:
		panic("invalid pattern while matching pkgpath")
	}
	namespace := match[1]

	// if `sysUsersPkg` does not exist -> skip validation.
	usersPkg := store.GetPackage(sysNamesPkg, false)
	if usersPkg == nil {
		return nil
	}

	result, err := vm.callRealmBool(ctx, creator, sysNamesPkg, "names",
		"IsAuthorizedAddressForNamespace",
		gno.Str(creator.String()), gno.Str(namespace))
	if err != nil {
		return err
	}

	if !result {
		return ErrUnauthorizedUser(
			fmt.Sprintf("%s is not authorized to deploy packages to namespace `%s`",
				creator.String(),
				namespace,
			))
	}

	return nil
}

// checkCLASignature verifies the creator has signed the required CLA.
// Returns nil if:
//   - SysCLAPkgPath parameter is empty (CLA enforcement disabled)
//   - CLA realm is not deployed yet (needed for bootstrap: the CLA realm
//     itself must be deployable before it exists on-chain)
//   - Creator has a valid CLA signature
func (vm *VMKeeper) checkCLASignature(ctx sdk.Context, creator crypto.Address) error {
	sysCLAPkg := vm.getSysCLAPkgParam(ctx)
	if sysCLAPkg == "" {
		return nil // CLA enforcement disabled
	}

	store := vm.getGnoTransactionStore(ctx)

	// If CLA realm does not exist -> skip validation.
	// This is required for bootstrap: the CLA realm itself needs to be
	// deployable before it exists on-chain. Once deployed, all subsequent
	// deployments will be checked.
	claPkg := store.GetPackage(sysCLAPkg, false)
	if claPkg == nil {
		return nil
	}

	result, err := vm.callRealmBool(ctx, creator, sysCLAPkg, "cla",
		"HasValidSignature",
		gno.Str(creator.String()))
	if err != nil {
		return err
	}

	if !result {
		return ErrUnauthorizedUser(
			fmt.Sprintf("address %s has not signed the required CLA",
				creator.String()))
	}

	return nil
}

// AddPackage adds a package with given fileset.
func (vm *VMKeeper) AddPackage(ctx sdk.Context, msg MsgAddPackage) (err error) {
	creator := msg.Creator
	pkgPath := msg.Package.Path
	memPkg := msg.Package
	send := msg.Send
	maxDeposit := msg.MaxDeposit
	gnostore := vm.getGnoTransactionStore(ctx)
	chainDomain := vm.getChainDomainParam(ctx)

	memPkg.Type = gno.MPUserAll

	// Validate arguments.
	if creator.IsZero() {
		return std.ErrInvalidAddress("missing creator address")
	}
	creatorAcc := vm.acck.GetAccount(ctx, creator)
	if creatorAcc == nil {
		return std.ErrUnknownAddress(fmt.Sprintf("account %s does not exist, it must receive coins to be created", creator))
	}
	if err := gno.ValidateMemPackageAny(msg.Package); err != nil {
		return ErrInvalidPkgPath(err.Error())
	}

	if !strings.HasPrefix(pkgPath, chainDomain+"/") {
		return ErrInvalidPkgPath("invalid domain: " + pkgPath)
	}

	pv := gnostore.GetPackage(pkgPath, false)
	if pv != nil && !pv.Private {
		return ErrPkgAlreadyExists("package already exists: " + pkgPath)
	}

	if !gno.IsRealmPath(pkgPath) && !gno.IsPPackagePath(pkgPath) {
		return ErrInvalidPkgPath("package path must be valid realm or p package path")
	}
	if strings.HasSuffix(pkgPath, "_test") || strings.HasSuffix(pkgPath, "_filetest") {
		return ErrInvalidPkgPath("package path must not end with _test or _filetest")
	}
	if _, ok := gno.IsGnoRunPath(pkgPath); ok {
		return ErrInvalidPkgPath("reserved package name: " + pkgPath)
	}
	opts := gno.TypeCheckOptions{
		Getter:     gnostore,
		TestGetter: vm.testStdlibCache.memPackageGetter(gnostore),
		Mode:       gno.TCLatestStrict,
		Cache:      vm.getTypeCheckCache(ctx),
	}
	if ctx.BlockHeight() == 0 {
		opts.Mode = gno.TCGenesisStrict // genesis time, waive blocking rules for importing draft packages.
	}
	// Validate Gno syntax and type check.
	_, err = gno.TypeCheckMemPackage(memPkg, opts)
	if err != nil {
		return ErrTypeCheck(err)
	}

	// Extra keeper-only checks.
	gm, err := gnomod.ParseMemPackage(memPkg)
	if err != nil {
		return ErrInvalidPackage(err.Error())
	}
	// no development packages.
	if gm.HasReplaces() {
		return ErrInvalidPackage("development packages are not allowed")
	}
	if pv != nil && pv.Private && !gm.Private {
		return ErrInvalidPackage("a private package cannot be overridden by a public package")
	}
	if gm.Private && !gno.IsRealmPath(pkgPath) {
		return ErrInvalidPackage("private packages must be realm packages")
	}
	if gm.Draft && ctx.BlockHeight() > 0 {
		return ErrInvalidPackage("draft packages can only be deployed at genesis time")
	}
	// no (deprecated) gno.mod file.
	if memPkg.GetFile("gno.mod") != nil {
		return ErrInvalidPackage("gno.mod file is deprecated and not allowed, run 'gno mod tidy' to upgrade to gnomod.toml")
	}

	// Patch gnomod.toml metadata
	gm.Module = pkgPath // XXX: if gm.Module != msg.Package.Path { panic() }?
	gm.AddPkg.Creator = creator.String()
	gm.AddPkg.Height = int(ctx.BlockHeight())
	// Re-encode gnomod.toml in memPkg
	memPkg.SetFile("gnomod.toml", gm.WriteString())

	// Pay deposit from creator.
	pkgAddr := gno.DerivePkgCryptoAddr(pkgPath)

	// TODO: ACLs.
	// - if r/system/names does not exists -> skip validation.
	// - loads r/system/names data state.
	if err := vm.checkNamespacePermission(ctx, creator, pkgPath); err != nil {
		return err
	}

	// Check CLA signature
	if err := vm.checkCLASignature(ctx, creator); err != nil {
		return err
	}

	err = vm.bank.SendCoins(ctx, creator, pkgAddr, send)
	if err != nil {
		return err
	}

	// Parse and run the files, construct *PV.
	msgCtx := stdlibs.ExecContext{
		ChainID:         ctx.ChainID(),
		ChainDomain:     chainDomain,
		Height:          ctx.BlockHeight(),
		Timestamp:       ctx.BlockTime().Unix(),
		OriginCaller:    creator.Bech32(),
		OriginSend:      send,
		OriginSendSpent: new(std.Coins),
		Banker:          NewSDKBanker(vm, ctx),
		Params:          NewSDKParams(vm.prmk, ctx),
		EventLogger:     ctx.EventLogger(),
	}
	// Parse and run the files, construct *PV.
	m2 := gno.NewMachineWithOptions(
		gno.MachineOptions{
			PkgPath:  "",
			Output:   vm.Output,
			Store:    gnostore,
			Alloc:    gnostore.GetAllocator(),
			Context:  msgCtx,
			GasMeter: ctx.GasMeter(),
		})
	defer m2.Release()
	defer doRecover(m2, &err)
	params := vm.GetParams(ctx)
	m2.RunMemPackage(memPkg, true)

	// use the parameters before executing the message, as they may change during execution.
	// The message should not fail due to parameter changes in the same transaction.
	err = vm.processStorageDeposit(ctx, creator, maxDeposit, gnostore, params)
	if err != nil {
		return err
	}
	// Log the telemetry
	logTelemetry(
		m2.GasMeter.GasConsumed(),
		m2.Cycles,
		attribute.KeyValue{
			Key:   "operation",
			Value: attribute.StringValue("m_addpkg"),
		},
	)

	return nil
}

// Call calls a public Gno function (for delivertx).
func (vm *VMKeeper) Call(ctx sdk.Context, msg MsgCall) (res string, err error) {
	params := vm.GetParams(ctx)
	pkgPath := msg.PkgPath // to import
	fnc := msg.Func
	gnostore := vm.getGnoTransactionStore(ctx)
	// Get the package and function type.
	pv := gnostore.GetPackage(pkgPath, false)
	pl := gno.PackageNodeLocation(pkgPath)
	pn := gnostore.GetBlockNode(pl).(*gno.PackageNode)
	ft := pn.GetStaticTypeOf(gnostore, gno.Name(fnc)).(*gno.FuncType)
	if len(ft.Params) == 0 || ft.Params[0].Type.String() != ".uverse.realm" {
		panic(fmt.Sprintf("function %s is non-crossing and cannot be called with MsgCall; query with vm/qeval or use MsgRun", fnc))
	}

	// Make main Package with imports.
	mpn := gno.NewPackageNode("main", "", nil)
	mpn.Define("pkg", gno.TypedValue{T: &gno.PackageType{}, V: pv})
	mpv := mpn.NewPackage(gnostore.GetAllocator())
	// Parse expression.
	argslist := ""
	for i := range msg.Args {
		if i > 0 {
			argslist += ","
		}
		argslist += fmt.Sprintf("arg%d", i)
	}
	var expr string
	if argslist == "" {
		expr = fmt.Sprintf(`pkg.%s(cross)`, fnc)
	} else {
		expr = fmt.Sprintf(`pkg.%s(cross,%s)`, fnc, argslist)
	}
	// Make context.
	// NOTE: if this is too expensive,
	// could it be safely partially memoized?
	pkgAddr := gno.DerivePkgCryptoAddr(pkgPath)
	caller := msg.Caller
	send := msg.Send
	chainDomain := vm.getChainDomainParam(ctx)
	msgCtx := stdlibs.ExecContext{
		ChainID:         ctx.ChainID(),
		ChainDomain:     chainDomain,
		Height:          ctx.BlockHeight(),
		Timestamp:       ctx.BlockTime().Unix(),
		OriginCaller:    caller.Bech32(),
		OriginSend:      send,
		OriginSendSpent: new(std.Coins),
		Banker:          NewSDKBanker(vm, ctx),
		Params:          NewSDKParams(vm.prmk, ctx),
		EventLogger:     ctx.EventLogger(),
	}
	// Construct machine and evaluate.
	m := gno.NewMachineWithOptions(
		gno.MachineOptions{
			PkgPath:  "",
			Output:   vm.Output,
			Store:    gnostore,
			Context:  msgCtx,
			Alloc:    gnostore.GetAllocator(),
			GasMeter: ctx.GasMeter(),
		})
	xn := m.MustParseExpr(expr)
	// Send send-coins to pkg from caller.
	err = vm.bank.SendCoins(ctx, caller, pkgAddr, send)
	if err != nil {
		return "", err
	}
	// Convert Args to gno values.
	cx := xn.(*gno.CallExpr)
	if cx.Varg {
		panic("variadic calls not yet supported")
	}
	if nargs := len(msg.Args) + 1; nargs != len(ft.Params) { // NOTE: nargs = `cur` + user's len(args)
		panic(fmt.Sprintf("wrong number of arguments in call to %s: want %d got %d", fnc, len(ft.Params), nargs))
	}
	for i, arg := range msg.Args {
		argType := ft.Params[i+1].Type
		atv := convertArgToGno(arg, argType)
		cx.Args[i+1] = &gno.ConstExpr{
			TypedValue: atv,
		}
	}
	defer m.Release()
	m.SetActivePackage(mpv)
	defer doRecover(m, &err)
	rtvs := m.Eval(xn)
	for i, rtv := range rtvs {
		res = res + rtv.String()
		if i < len(rtvs)-1 {
			res += "\n"
		}
	}

	// Use parameters before executing the message, as they may change during execution.
	// Parameter changes take effect only after the message has executed successfully.
	err = vm.processStorageDeposit(ctx, caller, msg.MaxDeposit, gnostore, params)
	if err != nil {
		return "", err
	}
	// Log the telemetry
	logTelemetry(
		m.GasMeter.GasConsumed(),
		m.Cycles,
		attribute.KeyValue{
			Key:   "operation",
			Value: attribute.StringValue("m_call"),
		},
	)

	res += "\n\n" // use `\n\n` as separator to separate results for single tx with multi msgs

	return res, nil
	// TODO pay for gas? TODO see context?
}

func doRecover(m *gno.Machine, e *error) {
	r := recover()

	// On normal transaction execution, out of gas panics are handled in the
	// BaseApp, so repanic here.
	const repanicOutOfGas = true
	doRecoverInternal(m, e, r, repanicOutOfGas)
}

func doRecoverQuery(m *gno.Machine, e *error) {
	r := recover()
	const repanicOutOfGas = false
	doRecoverInternal(m, e, r, repanicOutOfGas)
}

func doRecoverInternal(m *gno.Machine, e *error, r any, repanicOutOfGas bool) {
	if r == nil {
		return
	}
	if err, ok := r.(error); ok {
		var oog stypes.OutOfGasError
		if goerrors.As(err, &oog) {
			if repanicOutOfGas {
				panic(oog)
			}
			*e = oog
			return
		}
		var up gno.UnhandledPanicError
		if goerrors.As(err, &up) {
			// Common unhandled panic error, skip machine state.
			*e = errors.Wrapf(
				errors.New(up.Descriptor),
				"VM panic: %s\nStacktrace:\n%s\n",
				up.Descriptor, m.ExceptionStacktrace(),
			)
			return
		}
	}
	*e = errors.Wrapf(
		fmt.Errorf("%v", r),
		"VM panic: %v\nStacktrace:\n%s\n",
		r, m.Stacktrace().String(),
	)
}

// Run executes arbitrary Gno code in the context of the caller's realm.
func (vm *VMKeeper) Run(ctx sdk.Context, msg MsgRun) (res string, err error) {
	caller := msg.Caller
	pkgAddr := caller
	gnostore := vm.getGnoTransactionStore(ctx)
	send := msg.Send
	memPkg := msg.Package
	chainDomain := vm.getChainDomainParam(ctx)
	params := vm.GetParams(ctx)

	memPkg.Type = gno.MPUserProd

	// coerce path to right one.
	// the path in the message must be "" or the following path.
	// this is already checked in MsgRun.ValidateBasic
	memPkg.Path = chainDomain + "/e/" + msg.Caller.String() + "/run"

	// Validate arguments.
	callerAcc := vm.acck.GetAccount(ctx, caller)
	if callerAcc == nil {
		return "", std.ErrUnknownAddress(fmt.Sprintf("account %s does not exist, it must receive coins to be created", caller))
	}
	if err := gno.ValidateMemPackage(memPkg); err != nil {
		return "", ErrInvalidPkgPath(err.Error())
	}

	// Validate Gno syntax and type check.
	_, err = gno.TypeCheckMemPackage(memPkg, gno.TypeCheckOptions{
		Getter:     gnostore,
		TestGetter: vm.testStdlibCache.memPackageGetter(gnostore),
		Mode:       gno.TCLatestRelaxed,
		Cache:      vm.getTypeCheckCache(ctx),
	})
	if err != nil {
		return "", ErrTypeCheck(err)
	}

	// Send send-coins to pkg from caller.
	err = vm.bank.SendCoins(ctx, caller, pkgAddr, send)
	if err != nil {
		return "", err
	}

	// Parse and run the files, construct *PV.
	msgCtx := stdlibs.ExecContext{
		ChainID:         ctx.ChainID(),
		ChainDomain:     chainDomain,
		Height:          ctx.BlockHeight(),
		Timestamp:       ctx.BlockTime().Unix(),
		OriginCaller:    caller.Bech32(),
		OriginSend:      send,
		OriginSendSpent: new(std.Coins),
		Banker:          NewSDKBanker(vm, ctx),
		Params:          NewSDKParams(vm.prmk, ctx),
		EventLogger:     ctx.EventLogger(),
	}

	buf := new(bytes.Buffer)
	output := io.Writer(buf)

	// XXX: see reason of private for run msg here: https://github.com/gnolang/gno/pull/4594
	gm := new(gnomod.File)
	gm.Module = memPkg.Path
	gm.Gno = gno.GnoVerLatest
	gm.Private = true
	memPkg.SetFile("gnomod.toml", gm.WriteString())

	alloc := gnostore.GetAllocator()
	// Run as self-executing closure to have own function for doRecover / m.Release defers.
	pv := func() *gno.PackageValue {
		// Parse and run the files, construct *PV.
		if vm.Output != nil {
			output = io.MultiWriter(buf, vm.Output)
		}
		m := gno.NewMachineWithOptions(
			gno.MachineOptions{
				PkgPath:  "",
				Output:   output,
				Store:    gnostore,
				Alloc:    alloc,
				Context:  msgCtx,
				GasMeter: ctx.GasMeter(),
			})
		defer m.Release()
		defer doRecover(m, &err)

		_, pv := m.RunMemPackage(memPkg, false)
		return pv
	}()
	if err != nil {
		// handle any errors happened within pv generation.
		return
	}

	m2 := gno.NewMachineWithOptions(
		gno.MachineOptions{
			PkgPath:  "",
			Output:   output,
			Store:    gnostore,
			Alloc:    alloc,
			Context:  msgCtx,
			GasMeter: ctx.GasMeter(),
		})
	defer m2.Release()
	m2.SetActivePackage(pv)
	defer doRecover(m2, &err)
	m2.RunMain()
	res = buf.String()
	// Use parameters before executing the message, as they may change during execution.
	// Parameter changes take effect only after the message has executed successfully.
	err = vm.processStorageDeposit(ctx, caller, msg.MaxDeposit, gnostore, params)
	if err != nil {
		return "", err
	}
	// Log the telemetry
	logTelemetry(
		m2.GasMeter.GasConsumed(),
		m2.Cycles,
		attribute.KeyValue{
			Key:   "operation",
			Value: attribute.StringValue("m_run"),
		},
	)

	return res, nil
}

var reUserNamespace = regexp.MustCompile(`^[~_a-zA-Z0-9/]+$`)

// QueryPaths returns public facing function signatures.
// XXX: Implement pagination
func (vm *VMKeeper) QueryPaths(ctx sdk.Context, target string, limit int) ([]string, error) {
	if limit < 0 {
		return nil, errors.New("cannot have negative limit value")
	}

	// Determine effective limit to return
	store := vm.newGnoTransactionStore(ctx) // throwaway (never committed)

	// Handle case where no name is specified (general prefix lookup)
	if !strings.HasPrefix(target, "@") {
		return collectWithLimit(store.FindPathsByPrefix(target), limit), nil
	}

	// Extract name and sub-subPrefix from target
	name, subPrefix, hasSubPrefix := strings.Cut(target[1:], "/")
	if !reUserNamespace.MatchString(name) {
		return nil, errors.New("invalid username format")
	}

	// Handle reserved name
	if name == "stdlibs" || name == "std" {
		// XXX: Keep it simple here for now. If we have more reserved names at
		// some point, we should consider centralizing it somewhere.
		path := path.Join("_", subPrefix)
		return collectWithLimit(store.FindPathsByPrefix(path), limit), nil
	}
	// Lookup for both `/r` & `/p` paths of the namespace
	ctxDomain := vm.getChainDomainParam(ctx)
	rpath := path.Join(ctxDomain, "r", name, subPrefix)
	ppath := path.Join(ctxDomain, "p", name, subPrefix)

	// Add trailing slash if no subname is specified
	if !hasSubPrefix {
		rpath += "/"
		ppath += "/"
	}

	// Collect both paths
	return collectWithLimit(joinIters(
		store.FindPathsByPrefix(ppath),
		store.FindPathsByPrefix(rpath),
	), limit), nil
}

// joinIters joins the given iterators in a single iterator.
func joinIters[T any](seqs ...iter.Seq[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, seq := range seqs {
			for v := range seq {
				if !yield(v) {
					return
				}
			}
		}
	}
}

// like slices.Collect, but limits the slice size to the given limit.
func collectWithLimit[T any](seq iter.Seq[T], limit int) []T {
	s := []T{}
	for v := range seq {
		s = append(s, v)
		if len(s) >= limit {
			return s
		}
	}
	return s
}

// QueryFuncs returns public facing function signatures.
func (vm *VMKeeper) QueryFuncs(ctx sdk.Context, pkgPath string) (fsigs FunctionSignatures, err error) {
	store := vm.newGnoTransactionStore(ctx) // throwaway (never committed)
	// Ensure pkgPath is realm.
	if !gno.IsRealmPath(pkgPath) {
		err = ErrInvalidPkgPath(fmt.Sprintf(
			"package is not realm: %s", pkgPath))
		return nil, err
	}
	// Get Package.
	pv := store.GetPackage(pkgPath, false)
	if pv == nil {
		err = ErrInvalidPkgPath(fmt.Sprintf(
			"package not found: %s", pkgPath))
		return nil, err
	}
	// Iterate over public functions.
	pblock := pv.GetBlock(store)
	for _, tv := range pblock.Values {
		if tv.T.Kind() != gno.FuncKind {
			continue // must be function
		}
		fv := tv.GetFunc()
		if fv.IsMethod {
			continue // cannot be method
		}
		fname := string(fv.Name)
		first := fname[0:1]
		if strings.ToUpper(first) != first {
			continue // must be exposed
		}
		fsig := FunctionSignature{
			FuncName: fname,
		}
		ft := fv.Type.(*gno.FuncType)
		for _, param := range ft.Params {
			pname := string(param.Name)
			if pname == "" {
				pname = "_"
			}
			ptype := gno.BaseOf(param.Type).String()
			fsig.Params = append(fsig.Params,
				NamedType{Name: pname, Type: ptype},
			)
		}
		for _, result := range ft.Results {
			rname := string(result.Name)
			if rname == "" {
				rname = "_"
			}
			rtype := gno.BaseOf(result.Type).String()
			fsig.Results = append(fsig.Results,
				NamedType{Name: rname, Type: rtype},
			)
		}
		fsigs = append(fsigs, fsig)
	}
	return fsigs, nil
}

// QueryEval evaluates a gno expression (readonly, for ABCI queries).
func (vm *VMKeeper) QueryEval(ctx sdk.Context, pkgPath string, expr string) (res string, err error) {
	rtvs, err := vm.queryEvalInternal(ctx, pkgPath, expr)
	if err != nil {
		return "", err
	}
	res = ""
	for i, rtv := range rtvs {
		res += rtv.String()
		if i < len(rtvs)-1 {
			res += "\n"
		}
	}
	return res, nil
}

// QueryEvalString evaluates a gno expression (readonly, for ABCI queries).
// The result is expected to be a single string (not a tuple).
func (vm *VMKeeper) QueryEvalString(ctx sdk.Context, pkgPath string, expr string) (res string, err error) {
	rtvs, err := vm.queryEvalInternal(ctx, pkgPath, expr)
	if err != nil {
		return "", err
	}
	if len(rtvs) != 1 {
		return "", errors.New("expected 1 string result, got %d", len(rtvs))
	} else if rtvs[0].T.Kind() != gno.StringKind {
		return "", errors.New("expected 1 string result, got %v", rtvs[0].T.Kind())
	}
	res = rtvs[0].GetString()
	return res, nil
}

func (vm *VMKeeper) queryEvalInternal(ctx sdk.Context, pkgPath string, expr string) (rtvs []gno.TypedValue, err error) {
	ctx = ctx.WithGasMeter(store.NewGasMeter(maxGasQuery))
	alloc := gno.NewAllocator(maxAllocQuery)
	gnostore := vm.newGnoTransactionStore(ctx) // throwaway (never committed)
	// Get Package.
	pv := gnostore.GetPackage(pkgPath, false)
	if pv == nil {
		err = ErrInvalidPkgPath(fmt.Sprintf(
			"package not found: %s", pkgPath))
		return nil, err
	}
	// Construct new machine.
	chainDomain := vm.getChainDomainParam(ctx)
	msgCtx := stdlibs.ExecContext{
		ChainID:     ctx.ChainID(),
		ChainDomain: chainDomain,
		Height:      ctx.BlockHeight(),
		Timestamp:   ctx.BlockTime().Unix(),
		// OrigCaller:    caller,
		// OrigSend:      send,
		// OrigSendSpent: nil,
		Banker:      NewSDKBanker(vm, ctx), // safe as long as ctx is a fork to be discarded.
		Params:      NewSDKParams(vm.prmk, ctx),
		EventLogger: ctx.EventLogger(),
	}
	m := gno.NewMachineWithOptions(
		gno.MachineOptions{
			PkgPath:  pkgPath,
			Output:   vm.Output,
			Store:    gnostore,
			Context:  msgCtx,
			Alloc:    alloc,
			GasMeter: ctx.GasMeter(),
		})
	defer m.Release()
	defer doRecoverQuery(m, &err)
	// Parse expression.
	xx, err := m.ParseExpr(expr)
	if err != nil {
		return nil, err
	}
	return m.Eval(xx), err
}

func (vm *VMKeeper) QueryFile(ctx sdk.Context, filepath string) (res string, err error) {
	store := vm.newGnoTransactionStore(ctx) // throwaway (never committed)
	dirpath, filename := std.SplitFilepath(filepath)
	if filename != "" {
		memFile := store.GetMemFile(dirpath, filename)
		if memFile == nil {
			// TODO: XSS protection
			return "", errors.Wrapf(&InvalidFileError{}, "file %q is not available", filepath)
		}
		return memFile.Body, nil
	} else {
		memPkg := store.GetMemPackage(dirpath)
		if memPkg == nil {
			// TODO: XSS protection
			return "", errors.Wrapf(&InvalidPackageError{}, "package %q is not available", dirpath)
		}
		for i, memfile := range memPkg.Files {
			if i > 0 {
				res += "\n"
			}
			res += memfile.Name
		}
		return res, nil
	}
}

func (vm *VMKeeper) QueryDoc(ctx sdk.Context, pkgPath string) (*doc.JSONDocumentation, error) {
	store := vm.newGnoTransactionStore(ctx) // throwaway (never committed)

	memPkg := store.GetMemPackage(pkgPath)
	if memPkg == nil {
		err := ErrInvalidPkgPath(fmt.Sprintf(
			"package not found: %s", pkgPath))
		return nil, err
	}
	d, err := doc.NewDocumentableFromMemPkg(memPkg, true, "", "")
	if err != nil {
		return nil, err
	}
	return d.WriteJSONDocumentation(nil)
}

// QueryStorage returns storage and deposit for a realm.
func (vm *VMKeeper) QueryStorage(ctx sdk.Context, pkgPath string) (string, error) {
	store := vm.newGnoTransactionStore(ctx) // throwaway (never committed)
	rlm := store.GetPackageRealm(pkgPath)
	if rlm == nil {
		err := ErrInvalidPkgPath(fmt.Sprintf(
			"realm not found: %s", pkgPath))
		return "", err
	}
	res := fmt.Sprintf("storage: %d, deposit: %d", rlm.Storage, rlm.Deposit)

	return res, nil
}

// processStorageDeposit processes storage deposit adjustments for package realms based on
// storage size changes tracked within the gnoStore.
//
// For each realm, it:
// - Charges the caller a deposit proportional to newly used storage (positive size difference).
// - Returns the deposit to the caller for released storage (negative size difference).
//
// Returns an aggregated error if any realm processing fails due to insufficient deposit,
// transfer errors.

func (vm *VMKeeper) processStorageDeposit(ctx sdk.Context, caller crypto.Address, deposit std.Coins, gnostore gno.Store, params Params) error {
	realmDiffs := gnostore.RealmStorageDiffs()
	depositAmt := deposit.AmountOf(ugnot.Denom)
	if depositAmt == 0 {
		depositAmt = std.MustParseCoin(params.DefaultDeposit).Amount
	}
	price := std.MustParseCoin(params.StoragePrice)

	// Sort paths for determinism
	sortedRealm := make([]string, 0, len(realmDiffs))
	for path := range realmDiffs {
		sortedRealm = append(sortedRealm, path)
	}
	slices.SortFunc(sortedRealm, strings.Compare)

	var allErrs error
	for _, rlmPath := range sortedRealm {
		diff := realmDiffs[rlmPath]
		if diff == 0 {
			continue
		}
		rlm := gnostore.GetPackageRealm(rlmPath)
		if diff > 0 {
			// lock deposit for the additional storage used.
			requiredDeposit := overflow.Mulp(diff, price.Amount)
			if depositAmt < requiredDeposit {
				allErrs = goerrors.Join(allErrs, fmt.Errorf(
					"not enough deposit to cover the storage usage: requires %d%s for %d bytes",
					requiredDeposit, ugnot.Denom, diff))
				continue
			}
			err := vm.lockStorageDeposit(ctx, caller, rlm, requiredDeposit, diff)
			if err != nil {
				allErrs = goerrors.Join(allErrs, fmt.Errorf(
					"lockStorageDeposit failed for realm %s: %w",
					rlmPath, err))
				continue
			}
			depositAmt -= requiredDeposit
			// Emit event for storage deposit lock
			d := std.Coin{Denom: ugnot.Denom, Amount: requiredDeposit}
			evt := chain.StorageDepositEvent{
				BytesDelta: diff,
				FeeDelta:   d,
				PkgPath:    rlmPath,
			}
			ctx.EventLogger().EmitEvent(evt)
		} else {
			// release storage used and return deposit
			released := -diff
			if rlm.Storage < uint64(released) {
				panic(fmt.Sprintf(
					"not enough storage to be released for realm %s, realm storage %d bytes; requested release: %d bytes",
					rlmPath, rlm.Storage, released))
			}
			depositUnlocked := overflow.Mulp(released, price.Amount)
			if rlm.Deposit < uint64(depositUnlocked) {
				panic(fmt.Sprintf(
					"not enough deposit to be unlocked for realm %s, realm deposit %d%s; required to unlock: %d%s",
					rlmPath, rlm.Deposit, ugnot.Denom, depositUnlocked, ugnot.Denom))
			}

			isRestricted := slices.Contains(vm.bank.RestrictedDenoms(ctx), ugnot.Denom)
			receiver := caller
			if isRestricted {
				// If gnot tokens are locked, sent them to the storageFeeCollector address
				// If unlocked, sent them to memory releaser
				receiver = params.StorageFeeCollector
			}

			err := vm.refundStorageDeposit(ctx, receiver, rlm, depositUnlocked, released)
			if err != nil {
				return err
			}
			d := std.Coin{Denom: ugnot.Denom, Amount: depositUnlocked}
			evt := chain.StorageUnlockEvent{
				// For unlock, BytesDelta is negative
				BytesDelta:     diff,
				FeeRefund:      d,
				PkgPath:        rlmPath,
				RefundWithheld: isRestricted,
			}
			ctx.EventLogger().EmitEvent(evt)
		}
		gnostore.SetPackageRealm(rlm)
	}
	if allErrs != nil {
		return fmt.Errorf("storage deposit processing encountered one or more errors: %w", allErrs)
	}
	return nil
}

func (vm *VMKeeper) lockStorageDeposit(ctx sdk.Context, caller crypto.Address, rlm *gno.Realm, requiredDeposit int64, diff int64) error {
	storageDepositAddr := gno.DeriveStorageDepositCryptoAddr(rlm.Path)

	d := std.Coins{std.Coin{Denom: ugnot.Denom, Amount: requiredDeposit}}
	err := vm.bank.SendCoinsUnrestricted(ctx, caller, storageDepositAddr, d)
	if err != nil {
		return fmt.Errorf("unable to transfer deposit %s, %w", rlm.Path, err)
	}

	rlm.Deposit = overflow.Addp(rlm.Deposit, uint64(requiredDeposit))

	rlm.Storage = overflow.Addp(rlm.Storage, uint64(diff))
	return nil
}

func (vm *VMKeeper) refundStorageDeposit(ctx sdk.Context, refundReceiver crypto.Address, rlm *gno.Realm, depositUnlocked int64, released int64) error {
	storageDepositAddr := gno.DeriveStorageDepositCryptoAddr(rlm.Path)
	d := std.Coins{std.Coin{Denom: ugnot.Denom, Amount: depositUnlocked}}

	err := vm.bank.SendCoinsUnrestricted(ctx, storageDepositAddr, refundReceiver, d)
	if err != nil {
		return fmt.Errorf("unable to return deposit %s, %w", rlm.Path, err)
	}
	rlm.Deposit = overflow.Subp(rlm.Deposit, uint64(depositUnlocked))
	rlm.Storage = overflow.Subp(rlm.Storage, uint64(released))

	return nil
}

// logTelemetry logs the VM processing telemetry
func logTelemetry(
	gasUsed int64,
	cpuCycles int64,
	attributes ...attribute.KeyValue,
) {
	if !telemetry.MetricsEnabled() {
		return
	}

	// Record the operation frequency
	metrics.VMExecMsgFrequency.Add(
		context.Background(),
		1,
		metric.WithAttributes(attributes...),
	)

	// Record the CPU cycles
	metrics.VMCPUCycles.Record(
		context.Background(),
		cpuCycles,
		metric.WithAttributes(attributes...),
	)

	// Record the gas used
	metrics.VMGasUsed.Record(
		context.Background(),
		gasUsed,
		metric.WithAttributes(attributes...),
	)
}
