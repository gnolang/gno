package vm

// TODO: move most of the logic in ROOT/gno.land/...

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/errors"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
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
}

var _ VMKeeperI = &VMKeeper{}

// VMKeeper holds all package code and store state.
type VMKeeper struct {
	baseKey store.StoreKey
	iavlKey store.StoreKey
	acck    auth.AccountKeeper
	bank    bank.BankKeeper

	// cached, the DeliverTx persistent state.
	gnoStore gno.Store

	maxCycles int64 // max allowed cylces on VM executions
}

// NewVMKeeper returns a new VMKeeper.
func NewVMKeeper(
	baseKey store.StoreKey,
	iavlKey store.StoreKey,
	acck auth.AccountKeeper,
	bank bank.BankKeeper,
	maxCycles int64,
) *VMKeeper {
	// TODO: create an Options struct to avoid too many constructor parameters
	vmk := &VMKeeper{
		baseKey:   baseKey,
		iavlKey:   iavlKey,
		acck:      acck,
		bank:      bank,
		maxCycles: maxCycles,
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

	alloc := gno.NewAllocator(maxAllocTx, 10000, gno.NewHeap())
	vm.gnoStore = gno.NewStore(alloc, baseStore, iavlStore)
	vm.gnoStore.SetNativeStore(stdlibs.NativeStore)

	if vm.gnoStore.NumMemPackages() > 0 {
		// for now, all mem packages must be re-run after reboot.
		// TODO remove this, and generally solve for in-mem garbage collection
		// and memory management across many objects/types/nodes/packages.
		start := time.Now()

		m2 := gno.NewMachineWithOptions(
			gno.MachineOptions{
				PkgPath: "",
				Output:  os.Stdout, // XXX
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
		gs.SetNativeStore(stdlibs.NativeStore)
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
	memPkg := gno.ReadMemPackage(stdlibPath, pkgPath)
	if memPkg.IsEmpty() {
		// no gno files are present
		panic(fmt.Sprintf("failed loading stdlib %q: not a valid MemPackage", pkgPath))
	}

	m := gno.NewMachineWithOptions(gno.MachineOptions{
		PkgPath: "gno.land/r/stdlibs/" + pkgPath,
		// PkgPath: pkgPath, XXX why?
		Output: os.Stdout,
		Store:  store,
	})
	defer m.Release()
	m.RunMemPackage(memPkg, true)
}

type gnoStoreContextKeyType struct{}

var gnoStoreContextKey gnoStoreContextKeyType

func (vm *VMKeeper) newGnoTransactionStore(ctx sdk.Context) gno.TransactionStore {
	base := ctx.Store(vm.baseKey)
	iavl := ctx.Store(vm.iavlKey)

	return vm.gnoStore.BeginTransaction(base, iavl)
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
var reNamespace = regexp.MustCompile(`^gno.land/(?:r|p)/([\.~_a-zA-Z0-9]+)`)

// checkNamespacePermission check if the user as given has correct permssion to on the given pkg path
func (vm *VMKeeper) checkNamespacePermission(ctx sdk.Context, creator crypto.Address, pkgPath string) error {
	const sysUsersPkg = "gno.land/r/sys/users"

	store := vm.getGnoTransactionStore(ctx)

	match := reNamespace.FindStringSubmatch(pkgPath)
	switch len(match) {
	case 0:
		return ErrInvalidPkgPath(pkgPath) // no match
	case 2: // ok
	default:
		panic("invalid pattern while matching pkgpath")
	}
	if len(match) != 2 {
		return ErrInvalidPkgPath(pkgPath)
	}
	username := match[1]

	// if `sysUsersPkg` does not exist -> skip validation.
	usersPkg := store.GetPackage(sysUsersPkg, false)
	if usersPkg == nil {
		return nil
	}

	// Parse and run the files, construct *PV.
	pkgAddr := gno.DerivePkgAddr(pkgPath)
	msgCtx := stdlibs.ExecContext{
		ChainID:       ctx.ChainID(),
		Height:        ctx.BlockHeight(),
		Timestamp:     ctx.BlockTime().Unix(),
		OrigCaller:    creator.Bech32(),
		OrigSendSpent: new(std.Coins),
		OrigPkgAddr:   pkgAddr.Bech32(),
		// XXX: should we remove the banker ?
		Banker:      NewSDKBanker(vm, ctx),
		EventLogger: ctx.EventLogger(),
	}

	m := gno.NewMachineWithOptions(
		gno.MachineOptions{
			PkgPath:   "",
			Output:    os.Stdout, // XXX
			Store:     store,
			Context:   msgCtx,
			Alloc:     store.GetAllocator(),
			MaxCycles: vm.maxCycles,
			GasMeter:  ctx.GasMeter(),
		})
	defer m.Release()

	// call $sysUsersPkg.IsAuthorizedAddressForName("<user>")
	// We only need to check by name here, as address have already been check
	mpv := gno.NewPackageNode("main", "main", nil).NewPackage()
	m.SetActivePackage(mpv)
	m.RunDeclaration(gno.ImportD("users", sysUsersPkg))
	x := gno.Call(
		gno.Sel(gno.Nx("users"), "IsAuthorizedAddressForName"),
		gno.Str(creator.String()),
		gno.Str(username),
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
		return ErrUnauthorizedUser(username)
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

	// Validate arguments.
	if creator.IsZero() {
		return std.ErrInvalidAddress("missing creator address")
	}
	creatorAcc := vm.acck.GetAccount(ctx, creator)
	if creatorAcc == nil {
		return std.ErrUnknownAddress(fmt.Sprintf("account %s does not exist", creator))
	}
	if err := msg.Package.Validate(); err != nil {
		return ErrInvalidPkgPath(err.Error())
	}
	if pv := gnostore.GetPackage(pkgPath, false); pv != nil {
		return ErrInvalidPkgPath("package already exists: " + pkgPath)
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
	pkgAddr := gno.DerivePkgAddr(pkgPath)

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
		ChainID:       ctx.ChainID(),
		Height:        ctx.BlockHeight(),
		Timestamp:     ctx.BlockTime().Unix(),
		Msg:           msg,
		OrigCaller:    creator.Bech32(),
		OrigSend:      deposit,
		OrigSendSpent: new(std.Coins),
		OrigPkgAddr:   pkgAddr.Bech32(),
		Banker:        NewSDKBanker(vm, ctx),
		EventLogger:   ctx.EventLogger(),
	}
	// Parse and run the files, construct *PV.
	m2 := gno.NewMachineWithOptions(
		gno.MachineOptions{
			PkgPath:   "",
			Output:    os.Stdout, // XXX
			Store:     gnostore,
			Alloc:     gnostore.GetAllocator(),
			Context:   msgCtx,
			MaxCycles: vm.maxCycles,
			GasMeter:  ctx.GasMeter(),
		})
	defer m2.Release()
	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case store.OutOfGasException: // panic in consumeGas()
				panic(r)
			default:
				err = errors.Wrap(fmt.Errorf("%v", r), "VM addpkg panic: %v\n%s\n",
					r, m2.String())
				return
			}
		}
	}()
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
	mpn := gno.NewPackageNode("main", "main", nil)
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
	expr := fmt.Sprintf(`pkg.%s(%s)`, fnc, argslist)
	xn := gno.MustParseExpr(expr)
	// Send send-coins to pkg from caller.
	pkgAddr := gno.DerivePkgAddr(pkgPath)
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
	msgCtx := stdlibs.ExecContext{
		ChainID:       ctx.ChainID(),
		Height:        ctx.BlockHeight(),
		Timestamp:     ctx.BlockTime().Unix(),
		Msg:           msg,
		OrigCaller:    caller.Bech32(),
		OrigSend:      send,
		OrigSendSpent: new(std.Coins),
		OrigPkgAddr:   pkgAddr.Bech32(),
		Banker:        NewSDKBanker(vm, ctx),
		EventLogger:   ctx.EventLogger(),
	}
	// Construct machine and evaluate.
	m := gno.NewMachineWithOptions(
		gno.MachineOptions{
			PkgPath:   "",
			Output:    os.Stdout, // XXX
			Store:     gnostore,
			Context:   msgCtx,
			Alloc:     gnostore.GetAllocator(),
			MaxCycles: vm.maxCycles,
			GasMeter:  ctx.GasMeter(),
		})
	defer m.Release()
	m.SetActivePackage(mpv)
	defer func() {
		if r := recover(); r != nil {
			switch r := r.(type) {
			case store.OutOfGasException: // panic in consumeGas()
				panic(r)
			case gno.UnhandledPanicError:
				err = errors.Wrap(fmt.Errorf("%v", r.Error()), "VM call panic: %s\nStacktrace: %s\n",
					r.Error(), m.ExceptionsStacktrace())
			default:
				err = errors.Wrap(fmt.Errorf("%v", r), "VM call panic: %v\nMachine State:%s\nStacktrace: %s\n",
					r, m.String(), m.Stacktrace().String())
				return
			}
		}
	}()
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

// Run executes arbitrary Gno code in the context of the caller's realm.
func (vm *VMKeeper) Run(ctx sdk.Context, msg MsgRun) (res string, err error) {
	caller := msg.Caller
	pkgAddr := caller
	gnostore := vm.getGnoTransactionStore(ctx)
	send := msg.Send
	memPkg := msg.Package

	// coerce path to right one.
	// the path in the message must be "" or the following path.
	// this is already checked in MsgRun.ValidateBasic
	memPkg.Path = "gno.land/r/" + msg.Caller.String() + "/run"

	// Validate arguments.
	callerAcc := vm.acck.GetAccount(ctx, caller)
	if callerAcc == nil {
		return "", std.ErrUnknownAddress(fmt.Sprintf("account %s does not exist", caller))
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
		ChainID:       ctx.ChainID(),
		Height:        ctx.BlockHeight(),
		Timestamp:     ctx.BlockTime().Unix(),
		Msg:           msg,
		OrigCaller:    caller.Bech32(),
		OrigSend:      send,
		OrigSendSpent: new(std.Coins),
		OrigPkgAddr:   pkgAddr.Bech32(),
		Banker:        NewSDKBanker(vm, ctx),
		EventLogger:   ctx.EventLogger(),
	}
	// Parse and run the files, construct *PV.
	buf := new(bytes.Buffer)
	m := gno.NewMachineWithOptions(
		gno.MachineOptions{
			PkgPath:   "",
			Output:    buf,
			Store:     gnostore,
			Alloc:     gnostore.GetAllocator(),
			Context:   msgCtx,
			MaxCycles: vm.maxCycles,
			GasMeter:  ctx.GasMeter(),
		})
	// XXX MsgRun does not have pkgPath. How do we find it on chain?
	defer m.Release()
	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case store.OutOfGasException: // panic in consumeGas()
				panic(r)
			default:
				err = errors.Wrap(fmt.Errorf("%v", r), "VM run main addpkg panic: %v\n%s\n",
					r, m.String())
				return
			}
		}
	}()

	_, pv := m.RunMemPackage(memPkg, false)

	m2 := gno.NewMachineWithOptions(
		gno.MachineOptions{
			PkgPath:   "",
			Output:    buf,
			Store:     gnostore,
			Alloc:     gnostore.GetAllocator(),
			Context:   msgCtx,
			MaxCycles: vm.maxCycles,
			GasMeter:  ctx.GasMeter(),
		})
	defer m2.Release()
	m2.SetActivePackage(pv)
	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case store.OutOfGasException: // panic in consumeGas()
				panic(r)
			default:
				err = errors.Wrap(fmt.Errorf("%v", r), "VM run main call panic: %v\n%s\n",
					r, m2.String())
				return
			}
		}
	}()
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
// TODO: modify query protocol to allow MsgEval.
// TODO: then, rename to "Eval".
func (vm *VMKeeper) QueryEval(ctx sdk.Context, pkgPath string, expr string) (res string, err error) {
	alloc := gno.NewAllocator(maxAllocQuery, 10000, gno.NewHeap())
	gnostore := vm.newGnoTransactionStore(ctx) // throwaway (never committed)
	pkgAddr := gno.DerivePkgAddr(pkgPath)
	// Get Package.
	pv := gnostore.GetPackage(pkgPath, false)
	if pv == nil {
		err = ErrInvalidPkgPath(fmt.Sprintf(
			"package not found: %s", pkgPath))
		return "", err
	}
	// Parse expression.
	xx, err := gno.ParseExpr(expr)
	if err != nil {
		return "", err
	}
	// Construct new machine.
	msgCtx := stdlibs.ExecContext{
		ChainID:   ctx.ChainID(),
		Height:    ctx.BlockHeight(),
		Timestamp: ctx.BlockTime().Unix(),
		// Msg:           msg,
		// OrigCaller:    caller,
		// OrigSend:      send,
		// OrigSendSpent: nil,
		OrigPkgAddr: pkgAddr.Bech32(),
		Banker:      NewSDKBanker(vm, ctx), // safe as long as ctx is a fork to be discarded.
		EventLogger: ctx.EventLogger(),
	}
	m := gno.NewMachineWithOptions(
		gno.MachineOptions{
			PkgPath:   pkgPath,
			Output:    os.Stdout, // XXX
			Store:     gnostore,
			Context:   msgCtx,
			Alloc:     alloc,
			MaxCycles: vm.maxCycles,
			GasMeter:  ctx.GasMeter(),
		})
	defer m.Release()
	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case store.OutOfGasException: // panic in consumeGas()
				panic(r)
			default:
				err = errors.Wrap(fmt.Errorf("%v", r), "VM query eval panic: %v\n%s\n",
					r, m.String())
				return
			}
		}
	}()
	rtvs := m.Eval(xx)
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
// TODO: modify query protocol to allow MsgEval.
// TODO: then, rename to "EvalString".
func (vm *VMKeeper) QueryEvalString(ctx sdk.Context, pkgPath string, expr string) (res string, err error) {
	alloc := gno.NewAllocator(maxAllocQuery, 10000, gno.NewHeap())
	gnostore := vm.newGnoTransactionStore(ctx) // throwaway (never committed)
	pkgAddr := gno.DerivePkgAddr(pkgPath)
	// Get Package.
	pv := gnostore.GetPackage(pkgPath, false)
	if pv == nil {
		err = ErrInvalidPkgPath(fmt.Sprintf(
			"package not found: %s", pkgPath))
		return "", err
	}
	// Parse expression.
	xx, err := gno.ParseExpr(expr)
	if err != nil {
		return "", err
	}
	// Construct new machine.
	msgCtx := stdlibs.ExecContext{
		ChainID:   ctx.ChainID(),
		Height:    ctx.BlockHeight(),
		Timestamp: ctx.BlockTime().Unix(),
		// Msg:           msg,
		// OrigCaller:    caller,
		// OrigSend:      jsend,
		// OrigSendSpent: nil,
		OrigPkgAddr: pkgAddr.Bech32(),
		Banker:      NewSDKBanker(vm, ctx), // safe as long as ctx is a fork to be discarded.
		EventLogger: ctx.EventLogger(),
	}
	m := gno.NewMachineWithOptions(
		gno.MachineOptions{
			PkgPath:   pkgPath,
			Output:    os.Stdout, // XXX
			Store:     gnostore,
			Context:   msgCtx,
			Alloc:     alloc,
			MaxCycles: vm.maxCycles,
			GasMeter:  ctx.GasMeter(),
		})
	defer m.Release()
	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case store.OutOfGasException: // panic in consumeGas()
				panic(r)
			default:
				err = errors.Wrap(fmt.Errorf("%v", r), "VM query eval string panic: %v\n%s\n",
					r, m.String())
				return
			}
		}
	}()
	rtvs := m.Eval(xx)
	if len(rtvs) != 1 {
		return "", errors.New("expected 1 string result, got %d", len(rtvs))
	} else if rtvs[0].T.Kind() != gno.StringKind {
		return "", errors.New("expected 1 string result, got %v", rtvs[0].T.Kind())
	}
	res = rtvs[0].GetString()
	return res, nil
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
