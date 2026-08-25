package vm

// TODO: move most of the logic in ROOT/gno.land/...

import (
	"bytes"
	"context"
	"encoding/json"
	goerrors "errors"
	"fmt"
	"io"
	"iter"
	"log/slog"
	"maps"
	"math/big"
	"path"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/gnovm/pkg/doc"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/gnovm/stdlibs"
	"github.com/gnolang/gno/gnovm/stdlibs/chain"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/errors"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/gnolang/gno/tm2/pkg/overflow"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
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
	QueryEvalJSON(ctx sdk.Context, pkgPath string, expr string) (res string, err error)
	QueryObjectJSON(ctx sdk.Context, oidStr string) (res string, err error)
	QueryObjectBinary(ctx sdk.Context, oidStr string) (res []byte, err error)
	Run(ctx sdk.Context, msg MsgRun) (res string, err error)
	LoadStdlib(ctx sdk.Context, stdlibDir string)
	LoadStdlibCached(ctx sdk.Context, stdlibDir string)
	MakeGnoTransactionStore(ctx sdk.Context) sdk.Context
	CommitGnoTransactionStore(ctx sdk.Context)
	PopulateStdlibCache()
	PopulateStdlibCacheFrom(ms store.MultiStore)
	InitGenesis(ctx sdk.Context, data GenesisState)
}

var _ VMKeeperI = &VMKeeper{}

// getSessionAccount extracts the DelegatedAccount for the given caller
// from the SDK context, if this is a session tx.
func getSessionAccount(ctx sdk.Context, caller crypto.Address) std.DelegatedAccount {
	sa := ctx.Value(std.SessionAccountsContextKey{})
	if sa == nil {
		return nil
	}
	sessions, ok := sa.(map[crypto.Address]std.DelegatedAccount)
	if !ok {
		return nil
	}
	da, ok := sessions[caller]
	if !ok {
		return nil
	}
	return da
}

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
	typeCheckCache gno.TypeCheckCache
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
				PkgPath:            "",
				Output:             vm.Output,
				Store:              vm.gnoStore,
				BoundedPanicRender: true,
			})
		defer m2.Release()
		gno.DisableDebug()
		m2.PreprocessAllFilesAndSaveBlockNodes()
		gno.EnableDebug()

		opts := gno.TypeCheckOptions{
			Getter: vm.gnoStore,
			Mode:   gno.TCLatestStrict,
			Cache:  vm.typeCheckCache,
			// GetMemPackage returns the production blob only (test files live
			// in the #allbutprod sibling), so there is nothing here for the
			// test passes to check; stating it keeps the overlay out of every
			// keeper path.
			ProdOnly: true,
		}
		for _, stdlib := range stdlibs.InitOrder() {
			mp := vm.gnoStore.GetMemPackage(stdlib)
			pkg, err := gno.TypeCheckMemPackage(mp, opts)
			if err != nil {
				panic(fmt.Errorf("intialization error type checking %q: %w", stdlib, err))
			}
			opts.Cache[stdlib] = pkg
		}

		// Populate stdlib byte cache for gas-free stdlib reads.
		vm.gnoStore.PopulateStdlibCache(stdlibs.InitOrder())

		logger.Debug("GnoVM packages preprocessed",
			"elapsed", time.Since(start))
	}
}

// PopulateStdlibCache populates the stdlib byte cache on the gno store.
func (vm *VMKeeper) PopulateStdlibCache() {
	vm.gnoStore.PopulateStdlibCache(stdlibs.InitOrder())
}

// PopulateStdlibCacheFrom populates the stdlib byte cache by reading from
// the given multistore. Needed at genesis when the persistent gnoStore's
// baseStore doesn't have stdlib objects yet (they're in the deliver state).
func (vm *VMKeeper) PopulateStdlibCacheFrom(ms store.MultiStore) {
	baseStore := ms.GetStore(vm.baseKey)
	vm.gnoStore.PopulateStdlibCacheFrom(stdlibs.InitOrder(), baseStore)
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
			Getter:   gs,
			Mode:     gno.TCLatestStrict,
			Cache:    cachedInitTypeCheckCache,
			ProdOnly: true, // see Initialize
		}
		for _, lib := range stdlibs.InitOrder() {
			pkg, err := gno.TypeCheckMemPackage(gs.GetMemPackage(lib), opts)
			if err != nil {
				panic(fmt.Errorf("failed type checking stdlib %q: %w", lib, err))
			}
			opts.Cache[lib] = pkg
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
		Getter:   gs,
		Mode:     gno.TCLatestStrict,
		Cache:    vm.typeCheckCache,
		ProdOnly: true, // see Initialize
	}
	for _, lib := range stdlibs.InitOrder() {
		pkg, err := gno.TypeCheckMemPackage(gs.GetMemPackage(lib), opts)
		if err != nil {
			panic(fmt.Errorf("failed type checking stdlib %q: %w", lib, err))
		}
		opts.Cache[lib] = pkg
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
		PkgPath:            pkgPath,
		Store:              store,
		SkipPackage:        true,
		BoundedPanicRender: true,
	})
	defer m.Release()
	m.RunMemPackage(memPkg, true)
}

type vmkContextKey int

const (
	vmkContextKeyStore vmkContextKey = iota
	vmkContextKeyTypeCheckCache
)

func (vm *VMKeeper) newGnoTransactionStore(ctx sdk.Context) gno.TransactionStore {
	base := ctx.Store(vm.baseKey)
	iavl := ctx.Store(vm.iavlKey)
	gctx := ctx.GasContext()
	if gctx != nil {
		// Apply depth governance parameters. Write to a value-copy
		// of Config and build a fresh GasContext so we never mutate
		// a Config that a future caller (or a cached/pooled gctx)
		// might share — an in-place write there would race across
		// concurrent transactions and could break consensus.
		cfg := gctx.Config
		vm.GetParams(ctx).ApplyToGasConfig(&cfg)
		gctx = &store.GasContext{Meter: gctx.Meter, Config: cfg}
	}
	gasMeter := ctx.GasMeter()

	return vm.gnoStore.BeginTransaction(base, iavl, gctx, gasMeter)
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
var reNamespace = regexp.MustCompile(`^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}/(?:r|p)/([\.~_a-zA-Z0-9-]+)`)

// callRealmBool creates a Machine, imports pkgPath, calls funcName with args,
// and expects a single bool return value.
//
// Read-only contract: the called function MUST NOT mutate chain/params
// state. The ctx passed here is NOT seeded with a paramsAccum (that
// happens later in AddPackage/Call/Run, after this callout), so any
// chain/params.SetX from inside the realm would silently no-op the
// storage-deposit accounting while still persisting the on-disk write —
// leaving meta and reality divergent. In practice this constraint is
// only relevant to the sys realms invoked from here (sys/cla,
// sys/names) which are governance-controlled and structurally
// read-only checks. If a non-sys realm is ever invoked through this
// path, wrap Params in a read-only adapter that panics on Set/Update.
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
		SessionAccount:  getSessionAccount(ctx, creator),
	}

	preAlloc := gno.NewAllocator(maxAllocTx)
	preAlloc.SetGasMeter(ctx.GasMeter())
	store.SetPreprocessAllocator(preAlloc)
	defer store.SetPreprocessAllocator(nil)
	m := gno.NewMachineWithOptions(
		gno.MachineOptions{
			PkgPath:            "",
			Output:             vm.Output,
			Store:              store,
			Context:            msgCtx,
			Alloc:              store.GetAllocator(),
			GasMeter:           ctx.GasMeter(),
			BoundedPanicRender: true,
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

// chargePreprocessGas charges PreprocessGasPerByte gas per byte of every .gno
// source file (prod, _test, and _filetest) in mpkg: every file is parsed and
// the prod subset is type-checked and preprocessed, all otherwise unmetered.
// Charging over test bytes too is deliberately conservative — they are parsed,
// not type-checked (see TypeCheckOptions.ProdOnly). AddPackage and Run call it
// immediately before their type-check so an oversized package is rejected by
// the gas meter instead of consuming unmetered validator CPU. Params.Validate
// rejects a non-positive PreprocessGasPerByte, and GetParams defaults the field
// when reading a legacy params blob that predates it, so the charge is always
// active.
func chargePreprocessGas(ctx sdk.Context, params Params, mpkg *std.MemPackage, descriptor string) {
	var srcBytes int64
	for _, f := range mpkg.Files {
		if strings.HasSuffix(f.Name, ".gno") {
			srcBytes += int64(len(f.Body))
		}
	}
	ctx.GasMeter().ConsumeGas(overflow.Mulp(params.PreprocessGasPerByte, srcBytes), descriptor)
}

// stampGnomod writes the chain's own metadata into the package's gnomod.toml
// and re-encodes it in mpkg.
//
// Shared by AddPackage's inert and normal paths, and this one is a cross-
// function contract rather than merely shared lines: EnablePackage reads
// AddPkg.Creator back out of the stored file to decide OriginCaller. If the two
// paths ever stamped differently, the same source would initialize under a
// different identity depending on which policy was in force when it was
// submitted — with nothing to catch it. One writer makes that unrepresentable.
// maxDeposit is the creator's declared storage-deposit ceiling, and is written
// only where the charge outlives the declaring message — the inert path, where
// EnablePackage reads it back. The ordinary path passes "".
//
// Every AddPkg field is assigned unconditionally, including the empty cases.
// The section is keeper bookkeeping, but it lives in a file the submitter
// authors, so anything not overwritten here is attacker-supplied: a hand-written
// `[addpkg] max_deposit` would otherwise survive and be read at enable as though
// the message had declared it.
func stampGnomod(gm *gnomod.File, mpkg *std.MemPackage, pkgPath string, creator crypto.Address, height int64, maxDeposit string) {
	gm.Module = pkgPath // XXX: if gm.Module != msg.Package.Path { panic() }?
	gm.AddPkg.Creator = creator.String()
	gm.AddPkg.Height = int(height)
	gm.AddPkg.MaxDeposit = maxDeposit
	mpkg.SetFile("gnomod.toml", gm.WriteString())
}

// checkGnomodConstraints applies the keeper-only gnomod.toml rules that the type
// checker cannot express. Applied on all three paths that admit a package:
// AddPackage's normal and inert branches, and EnablePackage. The inert path
// runs them so a package cannot be parked to dodge a rule the normal path
// enforces, and enable runs them again because the world can move between the
// two messages -- see the call site there.
// priorPrivate reports whether a PRIVATE package already occupies pkgPath;
// height is the block the rules are evaluated against.
//
// A bool rather than the *gno.PackageValue this used to take, because
// EnablePackage cannot supply one: loading the live PackageValue populates the
// object cache and RunMemPackage then panics in SetCachePackage (see the note at
// its call site). It reads the stored blob's gnomod.toml instead, which answers
// the only question this function ever asked of the package value.
func checkGnomodConstraints(gm *gnomod.File, mpkg *std.MemPackage, pkgPath string, priorPrivate bool, height int64) error {
	// no development packages.
	if gm.HasReplaces() {
		return ErrInvalidPackage("development packages are not allowed")
	}
	if priorPrivate && !gm.Private {
		return ErrInvalidPackage("a private package cannot be overridden by a public package")
	}
	if gm.Private && !gno.IsRealmPath(pkgPath) {
		return ErrInvalidPackage("private packages must be realm packages")
	}
	if gm.Draft && height > 0 {
		return ErrInvalidPackage("draft packages can only be deployed at genesis time")
	}
	// no (deprecated) gno.mod file.
	if mpkg.GetFile("gno.mod") != nil {
		return ErrInvalidPackage("gno.mod file is deprecated and not allowed, run 'gno mod tidy' to upgrade to gnomod.toml")
	}
	return nil
}

// hasProdGnoFile reports whether mpkg contains at least one production
// (non-test) .gno file. It applies MPFProd's own per-file predicate so it
// cannot drift from what the storage split (store.go splitProdAllButProd)
// treats as prod, but without allocating a filtered copy of the package.
// FilterGno panics on non-.gno files and returns true to EXCLUDE a file, so a
// prod .gno file is a .gno file it does not exclude.
func hasProdGnoFile(mpkg *std.MemPackage) bool {
	pname := gno.Name(mpkg.Name)
	for _, f := range mpkg.Files {
		if strings.HasSuffix(f.Name, ".gno") && !gno.MPFProd.FilterGno(f, pname) {
			return true
		}
	}
	return false
}

// AddPackage adds a package with given fileset.
func (vm *VMKeeper) AddPackage(ctx sdk.Context, msg MsgAddPackage) (err error) {
	// Defense-in-depth spend check. MsgAddPackage is currently blocked
	// from session signers at the gno.land session allowlist
	// (checkSessionRestrictions in app.go), so this check is unreachable
	// for session txs today. Kept here so that if the allowlist is ever
	// relaxed, spend accounting still holds. NOTE: if AddPackage is
	// added to the session allowlist, this pre-check must be removed
	// or converted to a check-only variant — otherwise it will
	// double-count with the bank.Keeper.SendCoins session hook.
	if !msg.Send.IsZero() {
		if err := auth.CheckAndDeductSessionSpend(ctx, vm.acck, msg.Creator, msg.Send); err != nil {
			return err
		}
	}
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
	// Reject packages with no production .gno files (e.g. only _test.gno
	// files). The storage split writes no prod blob for them (store.go
	// splitProdAllButProd), so a restarted node would rebuild no PackageNode
	// while a non-restarted node still holds the deploy-time node in RAM —
	// making call gas depend on restart history.
	if !hasProdGnoFile(memPkg) {
		return ErrInvalidPackage("package has no production .gno files")
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
	// Refuse coins that could never be spent again.
	//
	// A realm can spend from its own address later, via a banker, so coins
	// attached to a realm deploy are recoverable and are allowed. A pure
	// `p/` package cannot: it has no realm identity, so it can never obtain
	// a banker, and nothing else can move coins out of its address either.
	// Crediting it would destroy the coins with no error and no way back.
	//
	// The principle is that we refuse a payment the receiver could not act
	// on, rather than accepting it and losing it silently.
	//
	// Placed after the path checks above so a bad path reports the path
	// problem rather than this one, and before the type check below so the
	// caller is not charged for compiling a package we are going to reject.
	// The checks above have already established the path is a realm or a
	// pure package, so "not a realm" here means "pure package".
	if !send.IsZero() && !gno.IsRealmPath(pkgPath) {
		return ErrUnspendableSend(fmt.Sprintf(
			"%s sent to %s, which is a pure package and can never spend it",
			send.String(), pkgPath))
	}

	// Use the parameters before executing the message, as they may change during
	// execution. The message should not fail due to parameter changes in the same
	// transaction. Read once, above the policy branch below, so the inert and the
	// normal path decide from the same snapshot.
	params := vm.GetParams(ctx)

	// If the chain is operating in "inert" submission mode, store the package
	// without typechecking or execution. It becomes callable only after an
	// approver sends MsgEnablePackage.
	//
	// Not at genesis. Genesis content is the chain's own, already reviewed by
	// whoever wrote the genesis file, and there is nobody for an approver to
	// protect it from. Parking it would produce a chain that boots with nothing
	// deployed: no r/sys/params and no govdao, so no way to propose a change,
	// and no approver able to act because the realms it would need do not exist
	// yet. The policy governs what strangers may submit to a running chain.
	//
	// The other height-sensitive rules here already work this way -- the type
	// checker drops to genesis mode at height 0, and the draft-package rule is
	// waived there too.
	//
	// Genesis REPLAY is not exempt, and reads the policy like any other
	// delivery. code_submission_policy is store-backed, so replayed history
	// carries its own: `gnogenesis fork generate` copies the source chain's vm
	// params into the fork's genesis untouched, and every historical governance
	// tx that moved the policy re-applies as it replays. A chain that parked a
	// package parks it again; one that deployed live, deploys live. Replay
	// reproduces what the source chain did by reading what the source chain
	// read, rather than by skipping the branch.
	//
	// The exception is an operator who sets a DIFFERENT policy in the fork's
	// own genesis params, which InitGenesis installs before the replay loop
	// (gno.land/pkg/gnoland/app.go). Replayed history would then run under a
	// policy that was never in force for it. To adopt "inert" at a fork, turn
	// it on with a migration tx appended after the history instead.
	if params.CodeSubmissionPolicy == CodeSubmissionPolicyInert &&
		ctx.BlockHeight() > 0 {
		// Charge the type-check and preprocess cost here, at submit, even though
		// neither runs on this path. The work is deferred, not avoided:
		// MsgEnablePackage type-checks and runs exactly these bytes later.
		// Nothing else in this branch is priced by source length, so without this
		// a submitter could park an arbitrarily large package for the price of
		// one amino write and leave the compile bill to whoever enables it.
		// Charging the submitter keeps the payer and the cause together, and
		// EnablePackage deliberately does not charge it a second time.
		chargePreprocessGas(ctx, params, memPkg, "AddInertPackagePreprocess")

		// Refuse a payment this path cannot deliver.
		//
		// On the ordinary path msg.Send is credited to the package address AND
		// presented to init() as the origin-send envelope, which is how a
		// payable deploy works: init() opens a BankerTypeOriginSend banker and
		// spends what the deployer attached.
		//
		// Here the two halves are split across two messages, and only the first
		// half can be carried. The coins would move at submit, but init() runs
		// at enable, in a message that sends nothing — so EnablePackage builds
		// its ExecContext with an empty OriginSend and no OriginSendRecipient,
		// and a payable init() does not merely see an empty envelope, it panics
		// on the recipient mismatch. The same source would deploy under
		// "permissionless" and fail under "inert", which makes the chain policy
		// change program semantics.
		//
		// Carrying the envelope through to enable was the alternative. Rejected:
		// it means stamping the amount into gnomod.toml beside the deposit
		// ceiling and reconstructing an origin-send context for coins that moved
		// in a different transaction, at a different block height, possibly
		// under a different account state — a lot of machinery to make a
		// two-phase deploy impersonate a one-phase one. Refusing is honest and
		// costs the submitter one extra transfer after activation.
		if !send.IsZero() {
			return ErrUnspendableSend(fmt.Sprintf(
				"%s sent to %s: a package submitted under the %q policy cannot "+
					"carry a payment, because init() runs in a later message and "+
					"would not see it; fund the package after it is enabled",
				send.String(), pkgPath, CodeSubmissionPolicyInert))
		}

		gm, err := gnomod.ParseMemPackage(memPkg)
		if err != nil {
			return ErrInvalidPackage(err.Error())
		}
		// Only the original submitter may replace a package already parked at
		// this path.
		//
		// The ErrPkgAlreadyExists guard above reads GetPackage, which sees only
		// the ACTIVE store; parked packages live under inert_pkg:<path> and are
		// invisible to it. Without this check anyone may overwrite anyone's
		// parked submission, and the overwrite is silent: AddInertPackage is an
		// unconditional Set. That is not merely untidy. An approver reviews the
		// source at a path and sends MsgEnablePackage; a third party who
		// front-runs the enable has their own bytes type-checked, init()ed and
		// stamped as creator under the reviewed path. Namespace permission does
		// not stand in the way either — checkNamespacePermission returns nil
		// while the names realm is undeployed, which under "inert" is exactly
		// the state a chain boots in.
		//
		// Same-submitter replacement stays allowed: it is the retry path after
		// an enable fails, and the parked bytes are the submitter's own.
		if prior := gnostore.GetInertPackage(pkgPath); prior != nil {
			priorGm, perr := gnomod.ParseMemPackage(prior)
			if perr != nil {
				return ErrInvalidPackage(fmt.Sprintf(
					"cannot read the parked package at %s: %v", pkgPath, perr))
			}
			if priorGm.AddPkg.Creator != creator.String() {
				return ErrPkgAlreadyExists(fmt.Sprintf(
					"package already awaiting approval at %s, submitted by %s",
					pkgPath, priorGm.AddPkg.Creator))
			}
		}
		// Apply the same gnomod rules as the normal path. Skipping them here
		// would make "inert" a way to park a package that no policy would ever
		// accept. EnablePackage runs them again on the stored blob; this is the
		// only chance to refuse the bytes before they are written.
		if err := checkGnomodConstraints(gm, memPkg, pkgPath, pv != nil && pv.Private, ctx.BlockHeight()); err != nil {
			return err
		}
		// Carry the creator's declared ceiling to whoever pays.
		//
		// This is the one path where the deposit is charged by a LATER message
		// than the one that declared the limit. Dropping it means EnablePackage
		// falls back to params.DefaultDeposit — so a creator who declared
		// 1000ugnot could be charged up to the chain default (100 GNOT today)
		// against a package they never got to re-approve. Recording it also
		// pins the ceiling to submit time, so a governance change to
		// DefaultDeposit between submit and enable cannot raise what the
		// creator is exposed to.
		//
		// Stamped through the same gnomod round-trip as Creator, which
		// EnablePackage already re-reads. Empty when nothing was declared, so
		// the ordinary path's stored gnomod.toml is unchanged -- and stamped
		// unconditionally either way, so a hand-written value in the
		// submitter's own file cannot stand in for a declaration.
		// Stamp the EFFECTIVE ceiling, including when none was declared.
		//
		// Leaving it empty pinned nothing for the common case: enable fell back
		// to params.DefaultDeposit read at ENABLE time, so a governance raise
		// between submit and enable widened the creator's exposure -- the exact
		// drift this stamping exists to prevent. Recording the submit-time
		// default closes that, and makes the stored value mean "what this
		// submitter agreed to" rather than "what they typed".
		declared := params.DefaultDeposit
		if !maxDeposit.IsZero() {
			// Storage deposits are denominated in the gas denom only, and
			// processStorageDeposit reads exactly that component. A ceiling
			// carrying none of it would parse cleanly at enable, contribute
			// nothing, and silently fall back to params.DefaultDeposit — read
			// at enable time, which is precisely the pinning this section
			// exists to provide. Refuse it rather than honour it in name only.
			if maxDeposit.AmountOf(ugnot.Denom) == 0 {
				return std.ErrInvalidCoins(fmt.Sprintf(
					"max_deposit %s carries no %s, so it cannot cap the storage deposit",
					maxDeposit, ugnot.Denom))
			}
			declared = maxDeposit.String()
		}
		stampGnomod(gm, memPkg, pkgPath, creator, ctx.BlockHeight(), declared)
		if err := vm.checkNamespacePermission(ctx, creator, pkgPath); err != nil {
			return err
		}
		if err := vm.checkCLASignature(ctx, creator); err != nil {
			return err
		}
		// Charge for the init() this submission defers onto someone else.
		//
		// Enable runs this package's init() on the APPROVER's transaction and
		// gas meter, and fees are flat — so the approver pays the same whatever
		// the package costs, and its exposure is that fee times the number of
		// approvals it can be induced to make. Submitting is otherwise close to
		// free, so provoking those approvals is cheap. This puts a price on it,
		// paid by the party that chose the cost.
		//
		// Deliberately flat rather than measured. Metering the deferred work and
		// refunding the remainder was tried four ways and rejected each time:
		// money derived from a gas reading makes gas a consensus input, so a
		// fork recomputes a different refund than the source chain paid. See the
		// ADR.
		//
		// Placed last, where nothing after it can fail. Everything reverts
		// together anyway — message-phase writes are discarded as a unit — but
		// the ordering makes pay-then-park legible here rather than requiring a
		// reader to know baseapp's revert semantics. It mirrors EnablePackage,
		// which puts its deposit immediately before DelInertPackage for the same
		// reason.
		//
		// SendCoins, not SendCoinsUnrestricted: this is a one-way transfer to a
		// governance-chosen address, not the refundable escrow the deposit path
		// uses, so a token lock should refuse it rather than be bypassed. It also
		// performs the session-spend check itself, so this is one call and the
		// order cannot be got wrong.
		//
		// The zero guard is load-bearing: SendCoinsUnrestricted has no zero
		// short-circuit, and even SendCoins' costs a store read. Off must cost
		// nothing, or the shipped default moves gas on every inert deploy.
		// The collector guard is the reason Params.Validate does not check it:
		// skipping the charge costs nothing, while charging the zero address
		// would burn it. applyLegacyDefaults supplies a collector on this read
		// path, so this is defence rather than an expected state.
		if params.InertSubmissionCharge != "" && !params.InertChargeCollector.IsZero() {
			charge, err := std.ParseCoins(params.InertSubmissionCharge)
			if err != nil {
				// Unreachable: Params.Validate parses this before it can be
				// stored. Panicking is what makes it safe to be wrong about that.
				panic("invalid inert_submission_charge in params: " + err.Error())
			}
			if !charge.IsZero() {
				if err := vm.bank.SendCoins(ctx, creator, params.InertChargeCollector, charge); err != nil {
					return err
				}
			}
		}
		// No SendCoins for msg.Send: a non-zero send was refused above.
		gnostore.AddInertPackage(memPkg)
		return nil
	}

	if pv != nil {
		// NOTE: reading `pv` above put this package in the object cache, and
		// RunMemPackage below panics in SetCachePackage on a cached package.
		// This path survives only because checkNamespacePermission and
		// checkCLASignature re-enter getGnoTransactionStore, whose
		// ClearObjectCache evicts it in between. That is incidental, not
		// designed: EnablePackage had the same shape, called neither, and its
		// private-redeploy branch was dead on arrival until it stopped loading
		// the package value at all. Do not reorder those checks below this
		// point without re-reading that.
		//
		// A private package is being redeployed (non-private re-adds were
		// rejected above). Clear its prior mempackage blobs first: AddMemPackage
		// stores an MP*All package as a prod blob plus a #allbutprod sibling, and
		// its conditional writes don't fully replace across both keys, so a stale
		// sibling (or stale prod blob, if redeployed prod-less) could otherwise
		// survive the re-add and be served by qfile/GetMemPackage.
		//
		// This must stay BELOW the inert branch, which returns without ever
		// calling AddMemPackage. Deleting there would strip a live package's
		// source while its realm, objects and package index survive: at boot
		// PreprocessAllFilesAndSaveBlockNodes skips the now-nil mempackage
		// silently, so a restarted node rebuilds no PackageNode and panics on
		// call, while a node that has not restarted still answers from
		// cacheNodes. That is a consensus split keyed on restart history.
		gnostore.DeleteMemPackage(pkgPath)
	}

	opts := gno.TypeCheckOptions{
		Getter: gnostore,
		Mode:   gno.TCLatestStrict,
		Cache:  vm.getTypeCheckCache(ctx),
		// Type-check production files only. Test files are still stored and
		// still parsed (a syntax error anywhere rejects the deploy), but the
		// chain can never run them, so their type-check verdict has no
		// on-chain meaning — while resolving their stdlib imports would read
		// a test-stdlib overlay off the node's local filesystem, making
		// consensus depend on node-local state. No TestGetter is supplied:
		// with ProdOnly the passes that would consult it never run.
		ProdOnly: true,
	}
	if ctx.BlockHeight() == 0 {
		opts.Mode = gno.TCGenesisStrict // genesis time, waive blocking rules for importing draft packages.
	}
	chargePreprocessGas(ctx, params, memPkg, "AddPackagePreprocess")
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
	if err := checkGnomodConstraints(gm, memPkg, pkgPath, pv != nil && pv.Private, ctx.BlockHeight()); err != nil {
		return err
	}

	// No ceiling stamped: on this path the deposit is charged in this same
	// message, so nothing needs to outlive it.
	stampGnomod(gm, memPkg, pkgPath, creator, ctx.BlockHeight(), "")

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

	// Seed per-message accumulator for chain/params byte tracking. Must
	// happen BEFORE NewSDKParams captures ctx into its struct field.
	ctx = ContextWithParamsAccum(ctx)
	// Parse and run the files, construct *PV.
	msgCtx := stdlibs.ExecContext{
		ChainID:         ctx.ChainID(),
		ChainDomain:     chainDomain,
		Height:          ctx.BlockHeight(),
		Timestamp:       ctx.BlockTime().Unix(),
		OriginCaller:    creator.Bech32(),
		OriginSend:      send,
		OriginSendSpent: new(std.Coins),
		// send was credited to pkgAddr just above; that is the only
		// address a BankerTypeOriginSend banker may spend from in this
		// message.
		OriginSendRecipient: pkgAddr.Bech32(),
		Banker:              NewSDKBanker(vm, ctx),
		Params:              NewSDKParams(vm.prmk, ctx),
		EventLogger:         ctx.EventLogger(),
		SessionAccount:      getSessionAccount(ctx, creator),
	}
	// Parse and run the files, construct *PV.
	m2 := gno.NewMachineWithOptions(
		gno.MachineOptions{
			PkgPath:            "",
			Output:             vm.Output,
			Store:              gnostore,
			Alloc:              gnostore.GetAllocator(),
			Context:            msgCtx,
			GasMeter:           ctx.GasMeter(),
			BoundedPanicRender: true,
		})
	defer m2.Release()
	defer doRecover(m2, &err)
	// Per-tx preprocess allocator: separate counter from m2.Alloc (the
	// init-phase allocator with GC). collect=nil so Allocate hard-panics
	// on maxBytes overflow rather than attempting a GC retry — GC walks
	// blocks/frames/package but not m.Values (the operand stack), and
	// would undercount in-flight preprocess values like a chained-+
	// running prefix. Closes the unbounded const-fold allocation surface
	// where preprocess sub-Machines (NewMachine(pkg, store) at
	// preprocess.go:3947, 4112, 4175, 4258) would otherwise run with
	// nil Alloc and skip both maxBytes tracking and per-allocation gas.
	//
	// The defer keeps preprocessAlloc installed for the entire handler.
	// During init phase the outer Machine (m2) uses its own m.Alloc with
	// GC for runtime ops; preprocessAlloc only takes effect for sub-
	// Machines spawned via NewMachine(pkg, store) inside Preprocess. If
	// init code re-triggers Preprocess (e.g., RunStatement on synthesized
	// init.0 calls), those sub-Machines also use preprocessAlloc — that's
	// intended so any Preprocess work during the handler is metered.
	preAlloc := gno.NewAllocator(maxAllocTx)
	preAlloc.SetGasMeter(ctx.GasMeter())
	gnostore.SetPreprocessAllocator(preAlloc)
	defer gnostore.SetPreprocessAllocator(nil)
	m2.RunMemPackage(memPkg, true)

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

// EnablePackage activates an inert package: runs the typechecker and
// initializes the package so it becomes importable and callable on-chain.
// Only addresses listed in Params.PkgApprovers may call this.
func (vm *VMKeeper) EnablePackage(ctx sdk.Context, msg MsgEnablePackage) (err error) {
	params := vm.GetParams(ctx)
	// Genesis replay reproduces a record rather than granting it again, so the
	// two authorization gates below are exempt: the fork may have moved off
	// "inert" or rotated pkg_approvers since, and must not refuse its own
	// history. Everything after them runs, so a package the source chain parked
	// is activated here as it was there -- which is the point, since AddPackage
	// parks during replay whenever the replayed policy says to.
	replay := auth.IsGenesisReplay(ctx)

	// Enable exists only to complete a submission the "inert" policy split in
	// two, so it is valid only while that policy is in force.
	//
	// Checked before the approver check so the refusal names the actual reason.
	// Without it, parked packages stay activatable forever under any later
	// policy: governance moves to "permissioned" precisely to stop strangers
	// getting code onto the chain, and every package parked during the "inert"
	// era would remain a stranger's pending deploy that one approver could still
	// land. PkgApprovers is not a substitute — it is not cleared when the policy
	// changes, and an approver's mandate was to activate what the policy of the
	// day accepted, not to carry it across a governance decision.
	//
	// This makes parked packages unactivatable once the policy moves, which is
	// the intended outcome; returning to "inert" makes them activatable again.
	// Note that nothing evicts them in the meantime — see DisablePackage.
	if !replay && params.CodeSubmissionPolicy != CodeSubmissionPolicyInert {
		return std.ErrUnauthorized(fmt.Sprintf(
			"code_submission_policy is %q, not %q: packages cannot be enabled",
			params.CodeSubmissionPolicy, CodeSubmissionPolicyInert))
	}
	if !replay && !isApprover(params.PkgApprovers, msg.Approver) {
		return std.ErrUnauthorized(fmt.Sprintf(
			"address %s is not a pkg approver", msg.Approver))
	}
	gnostore := vm.getGnoTransactionStore(ctx)
	memPkg := gnostore.GetInertPackage(msg.PkgPath)
	if memPkg == nil {
		// Nothing parked AND the package is live: the replayed submission took
		// the ordinary path, because the policy at that point in the replayed
		// history was not "inert". The enable is a genuine no-op, and every
		// genesis exported before this branch existed looks exactly like this.
		//
		// Nothing parked and nothing live means replay has already gone wrong --
		// the submission failed, so there was never anything to enable. Returning
		// nil there reports success for a package that is not on the chain, and
		// the fork comes up silently missing a realm with the enable recorded as
		// the last word on it. Refuse, so the replay report names it.
		//
		// Blob presence is the liveness test, for the reason spelled out at the
		// liveBlob probe below: GetPackage would populate the object cache and
		// make RunMemPackage panic later.
		if replay && gnostore.GetMemPackage(msg.PkgPath) != nil {
			return nil
		}
		return ErrInvalidPkgPath("no inert package at path: " + msg.PkgPath)
	}
	// Refuse to activate over a package that is already live, applying exactly
	// the rule AddPackage applies: a public package may not be replaced, a
	// private one may.
	//
	// Enable is the deferred second half of a deploy, so it has to enforce the
	// deploy's preconditions; it previously enforced none of them. Its whole
	// precondition set was "the sender is an approver" and "something is parked
	// at this path", which leaves a package takeover. A path can be parked and
	// live at once — the two live in different key spaces, and nothing clears a
	// parked blob when governance moves the policy off "inert" — so: A parks at
	// P and is never approved; the policy flips to permissionless; B deploys at
	// P through the now-open normal path; any approver then enables P and A's
	// bytes replace B's live package, running with OriginCaller = A, which is
	// what p/nt/ownable records as the owner. It does not even panic, because
	// runMemPackage takes its fresh-package branch when MachineOptions.PkgPath
	// is empty, so it silently rebuilds the node and package value over B's and
	// orphans B's realm objects.
	//
	// Deleting the prior blobs for the private case is the same requirement
	// AddPackage documents: AddMemPackage's writes are conditional across the
	// prod and #allbutprod keys and are not a full replace, so a stale sibling
	// would otherwise survive and be served by qfile.
	// Probed through the stored blob, NOT GetPackage. Loading the live
	// PackageValue populates the object cache, and RunMemPackage below then
	// panics in SetCachePackage because the package is already cached — which
	// made the private-replacement branch dead on arrival. AddPackage does read
	// the PackageValue, and escapes only incidentally: checkNamespacePermission
	// re-enters getGnoTransactionStore, whose ClearObjectCache evicts the entry
	// between the read and the run. EnablePackage calls neither, so it must not
	// create the entry in the first place.
	//
	// Blob presence is an exact liveness test here: a parked package is stored
	// under a different key prefix and is invisible to GetMemPackage, while a
	// live one always has a production blob (hasProdGnoFile guarantees it at
	// deploy). `private` comes from the same gnomod.toml the deploy stored.
	gm, err := gnomod.ParseMemPackage(memPkg)
	if err != nil {
		return ErrInvalidPackage(err.Error())
	}
	liveBlob := gnostore.GetMemPackage(msg.PkgPath)
	priorPrivate := false
	if liveBlob != nil {
		liveGm, perr := gnomod.ParseMemPackage(liveBlob)
		if perr != nil || !liveGm.Private {
			return ErrPkgAlreadyExists("package already exists: " + msg.PkgPath)
		}
		priorPrivate = true
	}

	// The full gnomod rule set, re-applied at enable.
	//
	// It ran at SUBMIT, but against the world as it was THEN, and every rule
	// whose answer can change between the two messages has to be asked again.
	// The private-override rule is the one that actually moves: for a package
	// parked before anything existed at this path, submit evaluated it against
	// nothing at all, so without this a public package parked early can be
	// activated over a private realm deployed later and flip it public,
	// retroactively exposing objects persisted under the invariant that nothing
	// outside the realm could reference them.
	//
	// The rest are stable across the split — the stored bytes cannot change, so
	// replaces, draft and gno.mod give the same answer — and are re-checked
	// anyway rather than hand-picked. Enable is the second half of a deploy;
	// enumerating which of a deploy's preconditions it may skip is how the
	// override rule went missing in the first place.
	if err := checkGnomodConstraints(gm, memPkg, msg.PkgPath, priorPrivate, ctx.BlockHeight()); err != nil {
		return err
	}
	// AddPackage wrote the creator into gnomod.toml before storing (see the
	// inert branch), so it round-trips; genesis.go reads it back the same way.
	// gm was parsed above, before the liveness probe that needs it. Parsed here
	// rather than lower down because the namespace check below needs it.
	creator, err := crypto.AddressFromBech32(gm.AddPkg.Creator)
	if err != nil {
		return ErrInvalidPackage(fmt.Sprintf(
			"invalid creator %q in stored gnomod.toml: %v", gm.AddPkg.Creator, err))
	}

	// Re-check namespace and CLA, which ran at SUBMIT against whatever was true
	// then.
	//
	// Same reasoning as the private-override rule above: a package must not reach
	// execution having cleared a weaker rule set than the deploy path enforces.
	// It matters most at bootstrap, because checkNamespacePermission returns nil
	// while r/sys/names is undeployed -- under "inert", the state a chain boots
	// in. Without this an attacker parks under a namespace nobody owns yet, and
	// an approver (typically an oracle checking only that the code type-checks)
	// activates it later under a namespace that by then belongs to someone else,
	// with OriginCaller set to the attacker.
	//
	// Placed HERE, before the type check and RunMemPackage, and not next to the
	// deposit: both of these evaluate a realm, which re-enters
	// getGnoTransactionStore and clears realmStorageDiffs. Running them after
	// RunMemPackage wipes the storage the deposit is computed from, so the
	// creator is charged nothing -- caught by the deposit tests.

	// The chain domain, the last of AddPackage's path rules. chain_domain is a
	// governance param, so a change between submit and enable would otherwise
	// let a package go live under a domain AddPackage would refuse.
	//
	// Not redundant with checkNamespacePermission, which applies the same rule
	// but returns nil early when sys_names_pkgpath is empty. Placed first, as
	// AddPackage orders it: the two checks below each evaluate a realm, so a
	// mismatch would otherwise pay for both before being refused.
	if !strings.HasPrefix(msg.PkgPath, vm.getChainDomainParam(ctx)+"/") {
		return ErrInvalidPkgPath("invalid domain: " + msg.PkgPath)
	}
	if err := vm.checkNamespacePermission(ctx, creator, msg.PkgPath); err != nil {
		return err
	}
	if err := vm.checkCLASignature(ctx, creator); err != nil {
		return err
	}
	// Typecheck the stored package.
	opts := gno.TypeCheckOptions{
		Getter: gnostore,
		// No TestGetter, and ProdOnly: mirrors AddPackage. GetMemPackage
		// returns the production blob only, and resolving test-stdlib imports
		// would make this consensus path depend on node-local state. #5888
		// predates that change on master and passed a test getter here.
		ProdOnly: true,
		Mode:     gno.TCLatestStrict,
		Cache:    vm.getTypeCheckCache(ctx),
	}
	if _, err = gno.TypeCheckMemPackage(memPkg, opts); err != nil {
		return ErrTypeCheck(err)
	}
	// The origin caller is the package's creator, not the approver.
	//
	// The approver authorises activation; they did not write the code. init()
	// runs here, and it commonly records chain.OriginCaller() as the owner
	// (p/nt/ownable's default). Passing the approver would hand every inert
	// package's ownership to whichever approver happened to sign the enable,
	// and would make ownership depend on the order approvers act in. It also
	// diverges from the non-inert path, where init() sees the deployer — so
	// the same source would initialize differently under a different policy.
	//
	// Execute and persist the package.
	ctx = ContextWithParamsAccum(ctx)
	msgCtx := stdlibs.ExecContext{
		ChainID:     ctx.ChainID(),
		ChainDomain: vm.getChainDomainParam(ctx),
		Height:      ctx.BlockHeight(),
		Timestamp:   ctx.BlockTime().Unix(),
		// Height/Timestamp are enable-time, not submit-time: init() observes
		// when it actually ran. Only the caller identity is inherited.
		OriginCaller:    creator.Bech32(),
		OriginSend:      std.Coins{},
		OriginSendSpent: new(std.Coins),
		Banker:          NewSDKBanker(vm, ctx),
		Params:          NewSDKParams(vm.prmk, ctx),
		EventLogger:     ctx.EventLogger(),
		// Keyed on the creator to stay coherent with OriginCaller. The creator
		// is normally not a signer of this tx, so this is normally nil — the
		// correct answer, since no session of theirs authorised the enable.
		SessionAccount: getSessionAccount(ctx, creator),
	}
	m2 := gno.NewMachineWithOptions(gno.MachineOptions{
		PkgPath:            "",
		Output:             vm.Output,
		Store:              gnostore,
		Alloc:              gnostore.GetAllocator(),
		Context:            msgCtx,
		GasMeter:           ctx.GasMeter(),
		BoundedPanicRender: true,
	})
	defer m2.Release()
	defer doRecover(m2, &err)
	preAlloc := gno.NewAllocator(maxAllocTx)
	preAlloc.SetGasMeter(ctx.GasMeter())
	gnostore.SetPreprocessAllocator(preAlloc)
	defer gnostore.SetPreprocessAllocator(nil)
	if liveBlob != nil {
		// Private redeploy: clear the prior blobs, as the normal path does.
		gnostore.DeleteMemPackage(msg.PkgPath)
	}
	m2.RunMemPackage(memPkg, true)

	// Take the storage deposit for the realm objects this enable just created.
	//
	// Without this they are free: the submit path cannot charge them, because
	// processStorageDeposit is driven entirely by RealmStorageDiffs() and
	// nothing has executed yet at submit time, and EnablePackage previously
	// never called it at all. So under "inert" every byte of realm state
	// created at activation escaped the deposit entirely.
	//
	// Charged to the creator, not the approver. The creator caused the storage
	// and only their own submission can lock their own funds, so this is
	// consented by the act of submitting. Charging the approver would be worse
	// than merely unfair: an approver is typically an automated oracle with a
	// hot key, so an attacker could submit large packages to bleed its balance
	// and stall approvals for everyone.
	//
	// Capped by the ceiling recorded at submit, read back from the same stamped
	// gnomod.toml the creator address came from.
	//
	// The inert path always records one -- what the submitter declared, or the
	// chain default as it stood at submit time. So the empty case below is not
	// the normal path; it covers a blob written by something other than that
	// branch, such as a genesis file, and falls back to the current default
	// exactly as an ordinary deploy would.
	//
	// This is the only path where the message that DECLARES the ceiling is not
	// the message that SPENDS against it, so the declaration has to be carried
	// or it is silently discarded. Reading params.DefaultDeposit here instead
	// would both ignore a creator who asked for a lower limit and let a
	// governance raise between submit and enable widen their exposure.
	//
	// So the split follows what each stage can know. Submit knows the source
	// length and the creator's declared limit; only execution reveals the realm
	// state, so enable charges the deposit against that limit.
	//
	// Escrowing the ceiling at submit was considered and rejected. The ceiling
	// is not a quote — the sample package in
	// TestVMKeeperEnableTakesStorageDepositFromCreator needs 210_200ugnot
	// against a 100_000_000ugnot default cap — and there is no source-bytes to
	// realm-bytes estimator to size it better (one sample is ~24x, and another
	// in examples/ is ~74x, so it is a data point, not a model). More to the
	// point, escrow would buy little: a failed enable leaves the package parked
	// and retryable once the creator is funded, so the loss it prevents is one
	// approver's gas on a transaction that can be simulated first, while the
	// cost it adds is funds locked on every submission that is never approved.
	//
	// Before DelInertPackage so the ordering reads pay-then-activate. An error
	// aborts the whole message either way, leaving the package inert and
	// retryable once the creator is funded.
	var declaredDeposit std.Coins
	if gm.AddPkg.MaxDeposit != "" {
		declaredDeposit, err = std.ParseCoins(gm.AddPkg.MaxDeposit)
		if err != nil {
			return ErrInvalidPackage(fmt.Sprintf(
				"invalid max_deposit %q in stored gnomod.toml: %v", gm.AddPkg.MaxDeposit, err))
		}
	}
	if err := vm.processStorageDeposit(ctx, creator, declaredDeposit, gnostore, params); err != nil {
		return err
	}

	// Remove from inert store now that it is active.
	gnostore.DelInertPackage(msg.PkgPath)
	return nil
}

// isApprover reports whether addr is in the approvers list.
func isApprover(approvers []crypto.Address, addr crypto.Address) bool {
	return slices.Contains(approvers, addr)
}

// DisablePackage moves an active package back to inert state.
// NOTE: full disable requires evicting executed objects from the base store,
// which is not yet implemented. This stub is provided for interface completeness.
func (vm *VMKeeper) DisablePackage(ctx sdk.Context, msg MsgDisablePackage) error {
	params := vm.GetParams(ctx)
	if !isApprover(params.PkgApprovers, msg.Approver) {
		return std.ErrUnauthorized(fmt.Sprintf(
			"address %s is not a pkg approver", msg.Approver))
	}
	// TODO: evict executed package objects from baseStore and move source back
	// to inert_pkg key. Tracked in a follow-up PR.
	return std.ErrUnknownRequest("disable_package is not yet implemented")
}

// Call calls a public Gno function (for delivertx).
func (vm *VMKeeper) Call(ctx sdk.Context, msg MsgCall) (res string, err error) {
	// Session spend on msg.Send is enforced inside bank.Keeper.SendCoins
	// (tm2/pkg/sdk/bank/keeper.go), which is where the actual coin
	// transfer happens. No pre-check needed here.
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
	var argslist strings.Builder
	for i := range msg.Args {
		if i > 0 {
			argslist.WriteString(",")
		}
		argslist.WriteString(fmt.Sprintf("arg%d", i))
	}
	var expr string
	if argslist.String() == "" {
		expr = fmt.Sprintf(`pkg.%s(cross)`, fnc)
	} else {
		expr = fmt.Sprintf(`pkg.%s(cross,%s)`, fnc, argslist.String())
	}
	// Make context.
	// NOTE: if this is too expensive,
	// could it be safely partially memoized?
	pkgAddr := gno.DerivePkgCryptoAddr(pkgPath)
	caller := msg.Caller
	send := msg.Send
	chainDomain := vm.getChainDomainParam(ctx)
	// Seed per-message accumulator before NewSDKParams captures ctx.
	ctx = ContextWithParamsAccum(ctx)
	msgCtx := stdlibs.ExecContext{
		ChainID:            ctx.ChainID(),
		ChainDomain:        chainDomain,
		Height:             ctx.BlockHeight(),
		Timestamp:          ctx.BlockTime().Unix(),
		OriginCaller:       caller.Bech32(),
		OriginSend:         send,
		OriginSendSpent:    new(std.Coins),
		OriginSendObserved: new(bool),
		// send is credited to pkgAddr (the entry realm) below; that is
		// the only address a BankerTypeOriginSend banker may spend from
		// in this message.
		OriginSendRecipient:     pkgAddr.Bech32(),
		OriginSendRecipientPath: pkgPath,
		Banker:                  NewSDKBanker(vm, ctx),
		Params:                  NewSDKParams(vm.prmk, ctx),
		EventLogger:             ctx.EventLogger(),
		SessionAccount:          getSessionAccount(ctx, caller),
	}
	preAlloc := gno.NewAllocator(maxAllocTx)
	preAlloc.SetGasMeter(ctx.GasMeter())
	gnostore.SetPreprocessAllocator(preAlloc)
	defer gnostore.SetPreprocessAllocator(nil)
	// Construct machine and evaluate.
	m := gno.NewMachineWithOptions(
		gno.MachineOptions{
			PkgPath:            "",
			Output:             vm.Output,
			Store:              gnostore,
			Context:            msgCtx,
			Alloc:              gnostore.GetAllocator(),
			GasMeter:           ctx.GasMeter(),
			BoundedPanicRender: true,
		})
	xn := m.MustParseExpr(expr)
	// Send send-coins to pkg from caller.
	err = vm.bank.SendCoins(ctx, caller, pkgAddr, send)
	if err != nil {
		return "", err
	}
	cx := xn.(*gno.CallExpr)
	// Replace the synthesized first argument at Args[0] with the
	// compiler-internal `.origin` sentinel. At preprocess this lowers to
	// the with-cross AST shape (Args[0]=nil, WithCross=true); at runtime
	// installCrossingCur routes through buildOriginRealm to mint an
	// EOA-origin cur. The dot-prefix `.origin` is unparseable from user
	// .gno source (same property as `.cur`), so it can only be introduced
	// here, by the chain-root MsgCall keeper synthesis.
	cx.Args[0] = gno.Nx(".origin")
	hasVarg := ft.HasVarg()
	// NOTE: nargs = `cur` + user's len(args)
	nargs := len(msg.Args) + 1
	var vargType gno.Type
	// If function is not variadic, it must have the same number of arguments.
	if !hasVarg {
		if nargs != len(ft.Params) {
			panic(fmt.Sprintf("wrong number of arguments in call to %s: want %d got %d", fnc, len(ft.Params), nargs))
		}
	} else {
		if nargs < len(ft.Params)-1 {
			// If function is variadic, it must have at least the number of arguments-1.
			// on the function we can simply avoid the variadic argument.
			panic(fmt.Sprintf("insufficient number of arguments in call to %s: must be at least %d, got %d", fnc, len(ft.Params)-1, nargs))
		}

		// For the variadic argument, we need to use the type of the
		// elements contained on the slice.
		vargType = ft.Params[len(ft.Params)-1].Type.(*gno.SliceType).Elt
	}

	// Convert Args to gno values.
	for i, arg := range msg.Args {
		paramIndex := i + 1
		var argType gno.Type
		if hasVarg && paramIndex >= len(ft.Params)-1 {
			argType = vargType
		} else {
			argType = ft.Params[paramIndex].Type
		}
		cx.Args[paramIndex] = &gno.ConstExpr{
			TypedValue: convertArgToGno(arg, argType),
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

	// Reject a send-envelope that nothing observed. The coins were credited
	// to pkgAddr above; if no executing code ever read them, the callee has
	// no notion of being paid and they would be stranded there. Returning an
	// error discards the whole message including that credit (msg execution
	// is cache-wrapped, tm2/pkg/sdk/baseapp.go:901).
	//
	// MsgCall only. MsgAddPackage is exempt because its envelope lands in
	// the new package's own address, recoverable later by the realm itself
	// — except for a pure `p/` package, whose address nothing can ever
	// spend from. MsgRun is exempt because pkgAddr == caller makes its
	// send a self-transfer no-op.
	if !send.IsZero() && !*msgCtx.OriginSendObserved {
		return "", ErrUnobservedSend(fmt.Sprintf(
			"%s sent to %s.%s, which never read the send-envelope",
			send.String(), pkgPath, fnc))
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
			desc := boundedString(up, 0)
			trace := gno.BoundedExceptionStacktrace(m,
				gno.MaxStacktraceFrames*gno.BoundedRenderBytes)
			*e = errors.Wrapf(
				errors.New(desc),
				"VM panic: %s\nStacktrace:\n%s\n",
				desc, trace,
			)
			return
		}
	}
	panicStr := boundedString(r, 0)
	trace := gno.BoundedStacktrace(m.Stacktrace(),
		gno.MaxStacktraceFrames*gno.BoundedRenderBytes)
	*e = errors.Wrapf(
		fmt.Errorf("%s", panicStr),
		"VM panic: %s\nStacktrace:\n%s\n",
		panicStr, trace,
	)
}

// doRecoverQueryNoMachine is like doRecoverQuery but for query paths that
// don't run a gno.Machine; uses debug.Stack() instead of Machine.Stacktrace().
func doRecoverQueryNoMachine(e *error) {
	r := recover()
	if r == nil {
		return
	}
	if err, ok := r.(error); ok {
		var oog stypes.OutOfGasError
		if goerrors.As(err, &oog) {
			*e = oog
			return
		}
	}
	*e = errors.Wrapf(
		fmt.Errorf("%v", r),
		"VM panic: %v\nStacktrace:\n%s\n",
		r, string(debug.Stack()),
	)
}

// Run executes arbitrary Gno code in the context of the caller's realm.
func (vm *VMKeeper) Run(ctx sdk.Context, msg MsgRun) (res string, err error) {
	// Session spend on msg.Send is enforced inside bank.Keeper.SendCoins.
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

	chargePreprocessGas(ctx, params, memPkg, "RunPreprocess")
	// Validate Gno syntax and type check.
	_, err = gno.TypeCheckMemPackage(memPkg, gno.TypeCheckOptions{
		Getter: gnostore,
		Mode:   gno.TCLatestRelaxed,
		Cache:  vm.getTypeCheckCache(ctx),
		// memPkg is MPUserProd here (set above) and ValidateMemPackage rejects
		// test files, so there is nothing for the test passes to check; being
		// explicit keeps the consensus path free of the test-stdlib overlay.
		ProdOnly: true,
	})
	if err != nil {
		return "", ErrTypeCheck(err)
	}

	// Send send-coins to pkg from caller.
	err = vm.bank.SendCoins(ctx, caller, pkgAddr, send)
	if err != nil {
		return "", err
	}

	// Seed per-message accumulator before NewSDKParams captures ctx.
	ctx = ContextWithParamsAccum(ctx)
	// Parse and run the files, construct *PV.
	msgCtx := stdlibs.ExecContext{
		ChainID:         ctx.ChainID(),
		ChainDomain:     chainDomain,
		Height:          ctx.BlockHeight(),
		Timestamp:       ctx.BlockTime().Unix(),
		OriginCaller:    caller.Bech32(),
		OriginSend:      send,
		OriginSendSpent: new(std.Coins),
		// No OriginSendRecipient here, deliberately. pkgAddr == caller for
		// MsgRun, so the coins move from the caller to the caller and the
		// envelope never lands anywhere. A run script cannot construct a
		// BankerTypeOriginSend banker either — its cur.Previous() is the
		// ephemeral /e/<addr>/run realm, so IsUserCall() is false. Leaving
		// the recipient empty is fail-closed: nothing can spend against an
		// envelope that never moved.
		Banker:         NewSDKBanker(vm, ctx),
		Params:         NewSDKParams(vm.prmk, ctx),
		EventLogger:    ctx.EventLogger(),
		SessionAccount: getSessionAccount(ctx, caller),
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
	// Per-tx preprocess allocator (see AddPackage for full rationale).
	// Covers both the closure-local Machine that calls RunMemPackage and
	// any subsequent Machine that re-Preprocesses; defer outlives both.
	preAlloc := gno.NewAllocator(maxAllocTx)
	preAlloc.SetGasMeter(ctx.GasMeter())
	gnostore.SetPreprocessAllocator(preAlloc)
	defer gnostore.SetPreprocessAllocator(nil)
	// Run as self-executing closure to have own function for doRecover / m.Release defers.
	pv := func() *gno.PackageValue {
		// Parse and run the files, construct *PV.
		if vm.Output != nil {
			output = io.MultiWriter(buf, vm.Output)
		}
		m := gno.NewMachineWithOptions(
			gno.MachineOptions{
				PkgPath:            "",
				Output:             output,
				Store:              gnostore,
				Alloc:              alloc,
				Context:            msgCtx,
				GasMeter:           ctx.GasMeter(),
				BoundedPanicRender: true,
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
			PkgPath:            "",
			Output:             output,
			Store:              gnostore,
			Alloc:              alloc,
			Context:            msgCtx,
			GasMeter:           ctx.GasMeter(),
			BoundedPanicRender: true,
		})
	defer m2.Release()
	m2.SetActivePackage(pv)
	defer doRecover(m2, &err)
	m2.RunMainMaybeCrossing()
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

var reUserNamespace = regexp.MustCompile(`^[~_a-zA-Z0-9/-]+$`)

// QueryPaths returns public facing function signatures.
// XXX: Implement pagination
func (vm *VMKeeper) QueryPaths(ctx sdk.Context, target string, limit int) ([]string, error) {
	if limit < 0 {
		return nil, errors.New("cannot have negative limit value")
	}

	ctx = ctx.WithGasMeter(store.NewGasMeter(maxGasQuery))
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
	ctx = ctx.WithGasMeter(store.NewGasMeter(maxGasQuery))
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
	for i := range pblock.Values {
		tv := pblock.GetPointerToInt(store, i).TV
		if tv.T.Kind() != gno.FuncKind {
			continue // must be function
		}
		fv := tv.GetFunc()
		if fv == nil {
			continue // typed-nil func variable, no signature to expose
		}
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
	err = vm.withQueryEvalMachine(ctx, pkgPath, expr, func(m *gno.Machine, rtvs []gno.TypedValue) {
		for i, rtv := range rtvs {
			res += rtv.String()
			if i < len(rtvs)-1 {
				res += "\n"
			}
		}
	})
	if err != nil {
		return "", err
	}
	return res, nil
}

// QueryEvalString evaluates a gno expression (readonly, for ABCI queries).
// The result is expected to be a single string (not a tuple).
func (vm *VMKeeper) QueryEvalString(ctx sdk.Context, pkgPath string, expr string) (res string, err error) {
	var cbErr error
	err = vm.withQueryEvalMachine(ctx, pkgPath, expr, func(m *gno.Machine, rtvs []gno.TypedValue) {
		if len(rtvs) != 1 {
			cbErr = errors.New("expected 1 string result, got %d", len(rtvs))
			return
		}
		if rtvs[0].T.Kind() != gno.StringKind {
			cbErr = errors.New("expected 1 string result, got %v", rtvs[0].T.Kind())
			return
		}
		res = rtvs[0].GetString()
	})
	if err != nil {
		return "", err
	}
	if cbErr != nil {
		return "", cbErr
	}
	return res, nil
}

// withQueryEvalMachine parses and evaluates expr under pkgPath, then calls fn
// with the live machine and its result values before releasing the machine.
// Callers that need to invoke methods on result values (e.g. call .Error() on
// an error-implementing return) must use this helper so the machine is still
// alive when fn runs.
func (vm *VMKeeper) withQueryEvalMachine(ctx sdk.Context, pkgPath string, expr string, fn func(m *gno.Machine, rtvs []gno.TypedValue)) (err error) {
	ctx = ctx.WithGasMeter(store.NewGasMeter(maxGasQuery))
	alloc := gno.NewAllocator(maxAllocQuery)
	gnostore := vm.newGnoTransactionStore(ctx) // throwaway (never committed)
	preAlloc := gno.NewAllocator(maxAllocQuery)
	preAlloc.SetGasMeter(ctx.GasMeter())
	gnostore.SetPreprocessAllocator(preAlloc)
	defer gnostore.SetPreprocessAllocator(nil)
	// Get Package.
	pv := gnostore.GetPackage(pkgPath, false)
	if pv == nil {
		return ErrInvalidPkgPath(fmt.Sprintf(
			"package not found: %s", pkgPath))
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
		Banker: NewSDKBanker(vm, ctx), // safe as long as ctx is a fork to be discarded.
		Params: NewSDKParams(vm.prmk, ctx),
		// Safe for the same reason: baseapp.go's handleQueryCustom calls
		// NewContext() per query, and tm2/pkg/sdk/context.go's NewContext
		// unconditionally allocates a fresh NewEventLogger. Any events
		// emitted by a qeval_json expression land in that per-query
		// logger and are discarded with the ctx — they never reach block
		// state (only runMsgs harvests events, on the tx path).
		EventLogger: ctx.EventLogger(),
	}
	m := gno.NewMachineWithOptions(
		gno.MachineOptions{
			PkgPath:            pkgPath,
			Output:             vm.Output,
			Store:              gnostore,
			Context:            msgCtx,
			Alloc:              alloc,
			GasMeter:           ctx.GasMeter(),
			BoundedPanicRender: true,
		})
	defer m.Release()
	defer doRecoverQuery(m, &err)
	xx, err := m.ParseExpr(expr)
	if err != nil {
		return err
	}
	// If the parsed expression is a call to a crossing function in this
	// package (e.g., `Render(cur realm, ...)` or any `Get*(cur realm, ...)`
	// getter), prepend `.cur` as the first argument. Same opt-in pattern
	// as init(cur realm) / main(cur realm): realms that don't declare a
	// crossing form are unaffected.
	m.MaybeInjectCurForEval(xx)
	fn(m, m.Eval(xx))
	return nil
}

func (vm *VMKeeper) QueryFile(ctx sdk.Context, filepath string) (res string, err error) {
	ctx = ctx.WithGasMeter(store.NewGasMeter(maxGasQuery))
	store := vm.newGnoTransactionStore(ctx) // throwaway (never committed)
	dirpath, filename := std.SplitFilepath(filepath)
	if filename != "" {
		memFile := store.GetMemFile(dirpath, filename)
		if memFile == nil {
			return "", errors.Wrapf(&InvalidFileError{}, "file %q is not available", filepath)
		}
		return memFile.Body, nil
	} else {
		// GetMemPackageAll so the file listing includes test/filetest files.
		memPkg := store.GetMemPackageAll(dirpath)
		if memPkg == nil {
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
	ctx = ctx.WithGasMeter(store.NewGasMeter(maxGasQuery))
	store := vm.newGnoTransactionStore(ctx) // throwaway (never committed)

	// GetMemPackageAll for parity with QueryFile, so doc generation sees test
	// files (e.g. for any future test-derived examples).
	memPkg := store.GetMemPackageAll(pkgPath)
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
	ctx = ctx.WithGasMeter(store.NewGasMeter(maxGasQuery))
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

// QueryEvalJSON evaluates a gno expression and returns JSON (Amino-encoded) results.
func (vm *VMKeeper) QueryEvalJSON(ctx sdk.Context, pkgPath string, expr string) (res string, err error) {
	err = vm.withQueryEvalMachine(ctx, pkgPath, expr, func(m *gno.Machine, rtvs []gno.TypedValue) {
		res = stringifyJSONResults(m, rtvs, nil)
	})
	if err != nil {
		return "", err
	}
	return res, nil
}

// exportObject retrieves and exports an object by ObjectID string.
func (vm *VMKeeper) exportObject(ctx sdk.Context, oidStr string) (gno.Value, error) {
	ctx = ctx.WithGasMeter(store.NewGasMeter(maxGasQuery))
	gnostore := vm.newGnoTransactionStore(ctx) // throwaway (never committed)

	var oid gno.ObjectID
	if err := oid.UnmarshalAmino(oidStr); err != nil {
		return nil, ErrInvalidExpr(fmt.Sprintf("invalid object id %q: %v", oidStr, err))
	}

	obj := gnostore.GetObjectSafe(oid)
	if obj == nil {
		return nil, ErrObjectNotFound(fmt.Sprintf("object not found: %s", oidStr))
	}

	return gno.ExportObject(obj), nil
}

// QueryObjectJSON retrieves an object by ObjectID and returns its Amino JSON representation.
func (vm *VMKeeper) QueryObjectJSON(ctx sdk.Context, oidStr string) (res string, err error) {
	defer doRecoverQueryNoMachine(&err)
	exported, err := vm.exportObject(ctx, oidStr)
	if err != nil {
		return "", err
	}

	jsonBytes, err := amino.MarshalJSONAny(exported)
	if err != nil {
		return "", err
	}

	envelope := struct {
		ObjectID string          `json:"objectid"`
		Value    json.RawMessage `json:"value"`
	}{ObjectID: oidStr, Value: jsonBytes}
	out, err := json.Marshal(envelope)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// QueryObjectBinary retrieves an object by ObjectID and returns its Amino binary representation.
func (vm *VMKeeper) QueryObjectBinary(ctx sdk.Context, oidStr string) (res []byte, err error) {
	defer doRecoverQueryNoMachine(&err)
	exported, err := vm.exportObject(ctx, oidStr)
	if err != nil {
		return nil, err
	}

	return amino.MarshalAny(exported)
}

// QueryPkg returns the named block variables of a package as Amino JSON.
// This is the entry point for the state explorer: given a package path,
// return variable names alongside their exported Amino JSON values.
func (vm *VMKeeper) QueryPkg(ctx sdk.Context, pkgPath string) (res string, err error) {
	defer doRecoverQueryNoMachine(&err)
	ctx = ctx.WithGasMeter(store.NewGasMeter(maxGasQuery))
	gnostore := vm.newGnoTransactionStore(ctx) // throwaway (never committed)

	pv := gnostore.GetPackage(pkgPath, false)
	if pv == nil {
		return "", ErrInvalidPkgPath(fmt.Sprintf("package not found: %s", pkgPath))
	}

	block := resolveBlock(gnostore, pv.Block)
	if block == nil {
		return "", fmt.Errorf("package block not found for %s", pkgPath)
	}

	// Resolve Source: it may be a RefNode (lazy reference to the PackageNode).
	source := resolveBlockNode(gnostore, block.Source)
	if source == nil {
		return "", fmt.Errorf("block source not found for %s", pkgPath)
	}
	sb := source.GetStaticBlock()
	names := sb.Names

	// Collect variable names and their exported values.
	varNames := make([]string, 0, len(block.Values))
	varValues := make([]gno.TypedValue, 0, len(block.Values))
	for i, tv := range block.Values {
		if i >= len(names) {
			break
		}
		name := string(names[i])
		if name == "" || name == "_" {
			continue
		}
		// Unwrap heap items. Top-level mutable vars live in dedicated
		// HeapItem cells; since #5415 the block stores them as RefValue
		// (lazy fill) so we resolve via the store before unwrapping.
		// GetObjectSafe: a stale ref must degrade to "render this var as
		// a ref" rather than 500 the whole page.
		if tv.T != nil && tv.T.Kind() == gno.HeapItemKind {
			if rv, ok := tv.V.(gno.RefValue); ok {
				if hiv, ok := gnostore.GetObjectSafe(rv.ObjectID).(*gno.HeapItemValue); ok {
					tv = hiv.Value
				}
			}
		}
		varNames = append(varNames, name)
		varValues = append(varValues, tv)
	}

	// Export values (replace persisted objects with RefValues, etc.)
	exported := gno.ExportValues(varValues)

	valuesJSON, err := amino.MarshalJSON(exported)
	if err != nil {
		return "", fmt.Errorf("failed to marshal values: %w", err)
	}
	namesJSON, err := amino.MarshalJSON(varNames)
	if err != nil {
		return "", fmt.Errorf("failed to marshal names: %w", err)
	}
	return buildPkgJSONEnvelope(namesJSON, valuesJSON), nil
}

// buildPkgJSONEnvelope assembles {"names":…,"values":…} from already-serialized
// JSON fragments. Trusts its inputs — both are produced by amino.MarshalJSON.
func buildPkgJSONEnvelope(namesJSON, valuesJSON []byte) string {
	var buf bytes.Buffer
	buf.Grow(len(namesJSON) + len(valuesJSON) + 20)
	buf.WriteString(`{"names":`)
	buf.Write(namesJSON)
	buf.WriteString(`,"values":`)
	buf.Write(valuesJSON)
	buf.WriteByte('}')
	return buf.String()
}

// QueryType retrieves a type by TypeID and returns its Amino JSON representation.
// This resolves RefType references in exported values: given a TypeID like
// "gno.land/r/demo/boards.Board", return the full type definition with field names.
func (vm *VMKeeper) QueryType(ctx sdk.Context, tidStr string) (res string, err error) {
	defer doRecoverQueryNoMachine(&err)
	ctx = ctx.WithGasMeter(store.NewGasMeter(maxGasQuery))
	gnostore := vm.newGnoTransactionStore(ctx) // throwaway (never committed)

	tid := gno.TypeID(tidStr)
	tt := gnostore.GetTypeSafe(tid)
	if tt == nil {
		return "", ErrInvalidExpr(fmt.Sprintf("type not found: %s", tidStr))
	}

	// Use a custom serializer instead of amino.MarshalJSON to avoid fatal
	// stack overflow from circular type references (e.g. time.Time).
	var buf bytes.Buffer
	marshalTypeJSON(&buf, tt, 0)
	return buildTypeJSONEnvelope(tidStr, buf.Bytes()), nil
}

// buildTypeJSONEnvelope assembles {"typeid":…,"type":…} with the TypeID
// passed through json.Marshal so any control character is JSON-escaped
// (matches the invariant Jae's original PR fixed for QueryObjectJSON).
func buildTypeJSONEnvelope(tidStr string, typeJSON []byte) string {
	tidJSON, _ := json.Marshal(tidStr)
	var buf bytes.Buffer
	buf.Grow(len(tidJSON) + len(typeJSON) + 20)
	buf.WriteString(`{"typeid":`)
	buf.Write(tidJSON)
	buf.WriteString(`,"type":`)
	buf.Write(typeJSON)
	buf.WriteByte('}')
	return buf.String()
}

// writeJSONString writes s as a JSON string literal into buf using
// encoding/json's escaping rules — never Go's `%q`.
func writeJSONString(buf *bytes.Buffer, s string) {
	b, _ := json.Marshal(s)
	buf.Write(b)
}

const maxTypeDepth = 8

// marshalTypeJSON writes a safe JSON representation of a gno.Type.
// It limits recursion depth to avoid stack overflow from circular references.
func marshalTypeJSON(buf *bytes.Buffer, t gno.Type, depth int) {
	if t == nil || depth > maxTypeDepth {
		buf.WriteString("null")
		return
	}
	switch ct := t.(type) {
	case gno.PrimitiveType:
		fmt.Fprintf(buf, `{"@type":"/gno.PrimitiveType","value":"%d"}`, int(ct))
	case *gno.PointerType:
		buf.WriteString(`{"@type":"/gno.PointerType","Elt":`)
		marshalTypeJSON(buf, ct.Elt, depth+1)
		buf.WriteByte('}')
	case *gno.ArrayType:
		fmt.Fprintf(buf, `{"@type":"/gno.ArrayType","Len":"%d","Elt":`, ct.Len)
		marshalTypeJSON(buf, ct.Elt, depth+1)
		buf.WriteByte('}')
	case *gno.SliceType:
		buf.WriteString(`{"@type":"/gno.SliceType","Elt":`)
		marshalTypeJSON(buf, ct.Elt, depth+1)
		buf.WriteByte('}')
	case *gno.StructType:
		buf.WriteString(`{"@type":"/gno.StructType","Fields":[`)
		for i, f := range ct.Fields {
			if i > 0 {
				buf.WriteByte(',')
			}
			buf.WriteString(`{"Name":`)
			writeJSONString(buf, string(f.Name))
			buf.WriteString(`,"Type":`)
			marshalTypeJSON(buf, f.Type, depth+1)
			buf.WriteByte('}')
		}
		buf.WriteString("]}")
	case *gno.MapType:
		buf.WriteString(`{"@type":"/gno.MapType","Key":`)
		marshalTypeJSON(buf, ct.Key, depth+1)
		buf.WriteString(`,"Value":`)
		marshalTypeJSON(buf, ct.Value, depth+1)
		buf.WriteByte('}')
	case *gno.FuncType:
		buf.WriteString(`{"@type":"/gno.FuncType"}`)
	case *gno.InterfaceType:
		buf.WriteString(`{"@type":"/gno.InterfaceType"}`)
	case *gno.DeclaredType:
		buf.WriteString(`{"@type":"/gno.DeclaredType","PkgPath":`)
		writeJSONString(buf, ct.PkgPath)
		buf.WriteString(`,"Name":`)
		writeJSONString(buf, string(ct.Name))
		buf.WriteString(`,"Base":`)
		marshalTypeJSON(buf, ct.Base, depth+1)
		buf.WriteByte('}')
	case *gno.PackageType:
		buf.WriteString(`{"@type":"/gno.PackageType"}`)
	case *gno.ChanType:
		buf.WriteString(`{"@type":"/gno.ChanType","Elt":`)
		marshalTypeJSON(buf, ct.Elt, depth+1)
		buf.WriteByte('}')
	default:
		// RefType or unknown — emit type ID if available
		buf.WriteString(`{"@type":"/gno.RefType","ID":`)
		writeJSONString(buf, string(t.TypeID()))
		buf.WriteByte('}')
	}
}

// resolveBlockNode resolves a BlockNode that may be a RefNode (lazy reference).
func resolveBlockNode(store gno.Store, bn gno.BlockNode) gno.BlockNode {
	if bn == nil {
		return nil
	}
	if _, ok := bn.(gno.RefNode); ok {
		loc := bn.GetLocation()
		return store.GetBlockNodeSafe(loc)
	}
	return bn
}

// resolveBlock extracts a *Block from a Value which may be a RefValue.
func resolveBlock(store gno.Store, v gno.Value) *gno.Block {
	switch cv := v.(type) {
	case *gno.Block:
		return cv
	case gno.RefValue:
		// GetObjectSafe (not GetObject): degrade a missing ref to nil for the
		// caller's guard instead of panicking. Mirrors resolveBlockNode.
		obj := store.GetObjectSafe(cv.ObjectID)
		if b, ok := obj.(*gno.Block); ok {
			return b
		}
		return nil
	default:
		return nil
	}
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
	if ctx.IsCheckTx() {
		// Defense-in-depth: baseapp already skips handler.Process in
		// CheckTx, but keep the guard so any future caller invoking
		// this directly during a non-deliver phase doesn't lock funds.
		return nil
	}
	realmDiffs := gnostore.RealmStorageDiffs()
	// Merge per-realm chain/params byte deltas accumulated on ctx.
	// See gno.land/pkg/sdk/vm/params_deposit.go.
	for path, diff := range ParamsRealmDiffs(ctx) {
		realmDiffs[path] += diff
	}
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
		if rlm == nil {
			// Should not happen: any executing realm is preprocessed
			// and materialized before it can call chain/params. Defend
			// against the rlm.Path nil-deref in lockStorageDeposit.
			allErrs = goerrors.Join(allErrs, fmt.Errorf(
				"params storage diff for unknown realm %q (size=%d) — deposit skipped",
				rlmPath, diff))
			continue
		}
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
			// Commit the per-realm meta-key only after the deposit is
			// held — keeps bank state and params meta consistent on
			// partial failure.
			FlushParamsRealmAccum(ctx, vm.prmk, rlmPath)
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
			// Proportional refund based on actual deposit ratio, not current price.
			// This ensures price governance changes don't lock or orphan deposits.
			var depositUnlocked int64
			if rlm.Storage == uint64(released) {
				// Freeing all storage, refund entire deposit (avoids rounding loss)
				depositUnlocked = int64(rlm.Deposit)
			} else {
				// Partial free: deposit * released / storage
				// Integer division truncates, so small dust amounts (< 1 ugnot per operation)
				// may accumulate in the realm's deposit over successive partial frees.
				// This is negligible in practice relative to deposit sizes.
				result := new(big.Int).SetUint64(rlm.Deposit)
				result.Mul(result, big.NewInt(released))
				result.Div(result, new(big.Int).SetUint64(rlm.Storage))
				depositUnlocked = result.Int64()
			}
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
			// Commit the per-realm meta-key only after the refund
			// transfers — symmetry with the lock branch above.
			FlushParamsRealmAccum(ctx, vm.prmk, rlmPath)
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

	// Count storage deposit against a session's SpendLimit. The transfer
	// itself uses SendCoinsUnrestricted (deposits must bypass
	// restricted-denom checks), so we perform the session deduction
	// explicitly here. Refunds via refundStorageDeposit do NOT reverse
	// SpendUsed — once budget is consumed, it stays consumed, so that
	// a compromised session cannot churn state (allocate → free →
	// allocate) to drain master beyond SpendLimit.
	if err := auth.CheckAndDeductSessionSpend(ctx, vm.acck, caller, d); err != nil {
		return fmt.Errorf("unable to lock deposit %s, %w", rlm.Path, err)
	}

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
