package vm

// TODO: move most of the logic in ROOT/gno.land/...

import (
	"bytes"
	"context"
	goerrors "errors"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gnolang/gno/gnovm/pkg/doc"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/errors"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
	"github.com/gnolang/gno/tm2/pkg/store/dbadapter"
	"github.com/gnolang/gno/tm2/pkg/store/types"
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
		baseKey: baseKey,
		iavlKey: iavlKey,
		acck:    acck,
		bank:    bank,
		prmk:    prmk,
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
	cachedStdlibOnce sync.Once
	cachedStdlib     stdlibCache
)

// LoadStdlib loads the Gno standard library into the given store.
func (vm *VMKeeper) LoadStdlibCached(ctx sdk.Context, stdlibDir string) {
	cachedStdlibOnce.Do(func() {
		cachedStdlib = stdlibCache{
			dir:  stdlibDir,
			base: dbadapter.StoreConstructor(memdb.NewMemDB(), types.StoreOptions{}),
			iavl: dbadapter.StoreConstructor(memdb.NewMemDB(), types.StoreOptions{}),
		}

		gs := gno.NewStore(nil, cachedStdlib.base, cachedStdlib.iavl)
		gs.SetNativeResolver(stdlibs.NativeResolver)
		loadStdlib(gs, stdlibDir)
		cachedStdlib.gno = gs
	})

	if stdlibDir != cachedStdlib.dir {
		panic(fmt.Sprintf(
			"cannot load cached stdlib: cached stdlib is in dir %q; wanted to load stdlib in dir %q",
			cachedStdlib.dir, stdlibDir))
	}

	gs := vm.getGnoTransactionStore(ctx)
	gno.CopyFromCachedStore(gs, cachedStdlib.gno, cachedStdlib.base, cachedStdlib.iavl)
}

// LoadStdlib loads the Gno standard library into the given store.
func (vm *VMKeeper) LoadStdlib(ctx sdk.Context, stdlibDir string) {
	gs := vm.getGnoTransactionStore(ctx)
	loadStdlib(gs, stdlibDir)
}

func loadStdlib(store gno.Store, stdlibDir string) {
	stdlibInitList := stdlibs.InitOrder()
	for _, lib := range stdlibInitList {
		if lib == "testing" {
			// XXX: testing is skipped for now while it uses testing-only packages
			// like fmt and encoding/json
			continue
		}
		loadStdlibPackage(lib, stdlibDir, store)
	}
}

func loadStdlibPackage(pkgPath, stdlibDir string, store gno.Store) {
	stdlibPath := filepath.Join(stdlibDir, pkgPath)
	if !osm.DirExists(stdlibPath) {
		// does not exist.
		panic(fmt.Sprintf("failed loading stdlib %q: does not exist", pkgPath))
	}
	memPkg := gno.MustReadMemPackage(stdlibPath, pkgPath)
	if memPkg.IsEmpty() {
		// no gno files are present
		panic(fmt.Sprintf("failed loading stdlib %q: not a valid MemPackage", pkgPath))
	}

	m := gno.NewMachineWithOptions(gno.MachineOptions{
		// XXX: gno.land, vm.domain, other?
		PkgPath: "gno.land/r/stdlibs/" + pkgPath,
		// PkgPath: pkgPath, XXX why?
		Store: store,
	})
	defer m.Release()
	m.RunMemPackage(memPkg, true)
}

type gnoStoreContextKeyType struct{}

var gnoStoreContextKey gnoStoreContextKeyType

func (vm *VMKeeper) newGnoTransactionStore(ctx sdk.Context) gno.TransactionStore {
	base := ctx.Store(vm.baseKey)
	iavl := ctx.Store(vm.iavlKey)
	gasMeter := ctx.GasMeter()

	return vm.gnoStore.BeginTransaction(base, iavl, gasMeter)
}

func (vm *VMKeeper) MakeGnoTransactionStore(ctx sdk.Context) sdk.Context {
	return ctx.WithValue(gnoStoreContextKey, vm.newGnoTransactionStore(ctx))
}

func (vm *VMKeeper) CommitGnoTransactionStore(ctx sdk.Context) {
	vm.getGnoTransactionStore(ctx).Write()
}

func (vm *VMKeeper) getGnoTransactionStore(ctx sdk.Context) gno.TransactionStore {
	txStore := ctx.Value(gnoStoreContextKey).(gno.TransactionStore)
	txStore.ClearObjectCache()
	return txStore
}

// Namespace can be either a user or crypto address.
var reNamespace = regexp.MustCompile(`^[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}/(?:r|p)/([\.~_a-zA-Z0-9]+)`)

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

	// Parse and run the files, construct *PV.
	msgCtx := stdlibs.ExecContext{
		ChainID:         ctx.ChainID(),
		ChainDomain:     chainDomain,
		Height:          ctx.BlockHeight(),
		Timestamp:       ctx.BlockTime().Unix(),
		OriginCaller:    creator.Bech32(),
		OriginSendSpent: new(std.Coins),
		// XXX: should we remove the banker ?
		Banker:      NewSDKBanker(vm, ctx),
		Params:      NewSDKParams(vm.prmk, ctx),
		EventLogger: ctx.EventLogger(),
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

	// call sysNamesPkg.IsAuthorizedAddressForName("<user>")
	// We only need to check by name here, as addresses have already been checked
	mpv := gno.NewPackageNode("main", "main", nil).NewPackage()
	m.SetActivePackage(mpv)
	m.RunDeclaration(gno.ImportD("names", sysNamesPkg))
	x := gno.Call(
		gno.Sel(gno.Nx("names"), "IsAuthorizedAddressForNamespace"),
		gno.Str(creator.String()),
		gno.Str(namespace),
	)

	ret := m.Eval(x)
	if len(ret) == 0 {
		panic("call: invalid response length")
	}

	useraddress := ret[0]
	if useraddress.T.Kind() != gno.BoolKind {
		panic("call: invalid response kind")
	}

	if isAuthorized := useraddress.GetBool(); !isAuthorized {
		return ErrUnauthorizedUser(
			fmt.Sprintf("%s is not authorized to deploy packages to namespace `%s`",
				creator.String(),
				namespace,
			))
	}

	return nil
}

// AddPackage adds a package with given fileset.
func (vm *VMKeeper) AddPackage(ctx sdk.Context, msg MsgAddPackage) (err error) {
	creator := msg.Creator
	pkgPath := msg.Package.Path
	memPkg := msg.Package
	deposit := msg.Deposit
	gnostore := vm.getGnoTransactionStore(ctx)
	chainDomain := vm.getChainDomainParam(ctx)

	// Validate arguments.
	if creator.IsZero() {
		return std.ErrInvalidAddress("missing creator address")
	}
	creatorAcc := vm.acck.GetAccount(ctx, creator)
	if creatorAcc == nil {
		return std.ErrUnknownAddress(fmt.Sprintf("account %s does not exist, it must receive coins to be created", creator))
	}
	if err := msg.Package.Validate(); err != nil {
		return ErrInvalidPkgPath(err.Error())
	}
	if !strings.HasPrefix(pkgPath, chainDomain+"/") {
		return ErrInvalidPkgPath("invalid domain: " + pkgPath)
	}
	if pv := gnostore.GetPackage(pkgPath, false); pv != nil {
		return ErrPkgAlreadyExists("package already exists: " + pkgPath)
	}
	if !gno.IsRealmPath(pkgPath) && !gno.IsPPackagePath(pkgPath) {
		return ErrInvalidPkgPath("package path must be valid realm or p package path")
	}
	if strings.HasSuffix(pkgPath, "_test") || strings.HasSuffix(pkgPath, "_filetest") {
		return ErrInvalidPkgPath("package path must not end with _test or _filetest")
	}
	if gno.ReGnoRunPath.MatchString(pkgPath) {
		return ErrInvalidPkgPath("reserved package name: " + pkgPath)
	}

	// Validate Gno syntax and type check.
	format := true
	if err := gno.TypeCheckMemPackage(memPkg, gnostore, format); err != nil {
		return ErrTypeCheck(err)
	}

	// Pay deposit from creator.
	pkgAddr := gno.DerivePkgCryptoAddr(pkgPath)

	// TODO: ACLs.
	// - if r/system/names does not exists -> skip validation.
	// - loads r/system/names data state.
	if err := vm.checkNamespacePermission(ctx, creator, pkgPath); err != nil {
		return err
	}

	err = vm.bank.SendCoins(ctx, creator, pkgAddr, deposit)
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
		OriginSend:      deposit,
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
	m2.RunMemPackage(memPkg, true)

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
	pkgPath := msg.PkgPath // to import
	fnc := msg.Func
	gnostore := vm.getGnoTransactionStore(ctx)
	// Get the package and function type.
	pv := gnostore.GetPackage(pkgPath, false)
	pl := gno.PackageNodeLocation(pkgPath)
	pn := gnostore.GetBlockNode(pl).(*gno.PackageNode)
	ft := pn.GetStaticTypeOf(gnostore, gno.Name(fnc)).(*gno.FuncType)
	// Make main Package with imports.
	mpn := gno.NewPackageNode("main", "", nil)
	mpn.Define("pkg", gno.TypedValue{T: &gno.PackageType{}, V: pv})
	mpv := mpn.NewPackage()
	// Parse expression.
	argslist := ""
	for i := range msg.Args {
		if i > 0 {
			argslist += ","
		}
		argslist += fmt.Sprintf("arg%d", i)
	}
	expr := fmt.Sprintf(`cross(pkg.%s)(%s)`, fnc, argslist)
	xn := gno.MustParseExpr(expr)
	// Send send-coins to pkg from caller.
	pkgAddr := gno.DerivePkgCryptoAddr(pkgPath)
	caller := msg.Caller
	send := msg.Send
	err = vm.bank.SendCoins(ctx, caller, pkgAddr, send)
	if err != nil {
		return "", err
	}
	// Convert Args to gno values.
	cx := xn.(*gno.CallExpr)
	if cx.Varg {
		panic("variadic calls not yet supported")
	}
	if len(msg.Args) != len(ft.Params) {
		panic(fmt.Sprintf("wrong number of arguments in call to %s: want %d got %d", fnc, len(ft.Params), len(msg.Args)))
	}
	for i, arg := range msg.Args {
		argType := ft.Params[i].Type
		atv := convertArgToGno(arg, argType)
		cx.Args[i] = &gno.ConstExpr{
			TypedValue: atv,
		}
	}
	// Make context.
	// NOTE: if this is too expensive,
	// could it be safely partially memoized?
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
		var oog types.OutOfGasError
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
		"VM panic: %v\nMachine State:%s\nStacktrace:\n%s\n",
		r, m.String(), m.Stacktrace().String(),
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

	// coerce path to right one.
	// the path in the message must be "" or the following path.
	// this is already checked in MsgRun.ValidateBasic
	memPkg.Path = chainDomain + "/r/" + msg.Caller.String() + "/run"

	// Validate arguments.
	callerAcc := vm.acck.GetAccount(ctx, caller)
	if callerAcc == nil {
		return "", std.ErrUnknownAddress(fmt.Sprintf("account %s does not exist, it must receive coins to be created", caller))
	}
	if err := msg.Package.Validate(); err != nil {
		return "", ErrInvalidPkgPath(err.Error())
	}

	// Validate Gno syntax and type check.
	format := false
	if err = gno.TypeCheckMemPackage(memPkg, gnostore, format); err != nil {
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
				Alloc:    gnostore.GetAllocator(),
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
			Alloc:    gnostore.GetAllocator(),
			Context:  msgCtx,
			GasMeter: ctx.GasMeter(),
		})
	defer m2.Release()
	m2.SetActivePackage(pv)
	defer doRecover(m2, &err)
	m2.RunMain()
	res = buf.String()

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
	// Parse expression.
	xx, err := gno.ParseExpr(expr)
	if err != nil {
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
	return m.Eval(xx), err
}

func (vm *VMKeeper) QueryFile(ctx sdk.Context, filepath string) (res string, err error) {
	store := vm.newGnoTransactionStore(ctx) // throwaway (never committed)
	dirpath, filename := std.SplitFilepath(filepath)
	if filename != "" {
		memFile := store.GetMemFile(dirpath, filename)
		if memFile == nil {
			return "", fmt.Errorf("file %q is not available", filepath) // TODO: XSS protection
		}
		return memFile.Body, nil
	} else {
		memPkg := store.GetMemPackage(dirpath)
		if memPkg == nil {
			return "", fmt.Errorf("package %q is not available", dirpath) // TODO: XSS protection
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
	return d.WriteJSONDocumentation()
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
