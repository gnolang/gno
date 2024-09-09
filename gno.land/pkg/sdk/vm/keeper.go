package vm

// TODO: move most of the logic in ROOT/gno.land/...

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
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
	maxAllocTx               = 500 * 1000 * 1000
	maxAllocQuery            = 1500 * 1000 * 1000 // higher limit for queries
	maxMetaFields            = 10                 // maximum number of package metadata fields
	maxMetaFieldValueSize    = 1_000_000          // maximum size for package metadata field values in bytes
	maxToolDescriptionLenght = 100
)

var reToolName = regexp.MustCompile(`^[a-z]+[_a-z0-9]{5,16}$`)

// vm.VMKeeperI defines a module interface that supports Gno
// smart contracts programming (scripting).
type VMKeeperI interface {
	AddPackage(ctx sdk.Context, msg MsgAddPackage) error
	Call(ctx sdk.Context, msg MsgCall) (res string, err error)
	QueryEval(ctx sdk.Context, pkgPath string, expr string) (res string, err error)
	Run(ctx sdk.Context, msg MsgRun) (res string, err error)
}

var _ VMKeeperI = &VMKeeper{}

// VMKeeper holds all package code and store state.
type VMKeeper struct {
	baseKey    store.StoreKey
	iavlKey    store.StoreKey
	acck       auth.AccountKeeper
	bank       bank.BankKeeper
	stdlibsDir string

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
	stdlibsDir string,
	maxCycles int64,
) *VMKeeper {
	// TODO: create an Options struct to avoid too many constructor parameters
	vmk := &VMKeeper{
		baseKey:    baseKey,
		iavlKey:    iavlKey,
		acck:       acck,
		bank:       bank,
		stdlibsDir: stdlibsDir,
		maxCycles:  maxCycles,
	}
	return vmk
}

func (vm *VMKeeper) Initialize(
	logger *slog.Logger,
	ms store.MultiStore,
	cacheStdlibLoad bool,
) {
	if vm.gnoStore != nil {
		panic("should not happen")
	}
	baseSDKStore := ms.GetStore(vm.baseKey)
	iavlSDKStore := ms.GetStore(vm.iavlKey)

	if cacheStdlibLoad {
		// Testing case (using the cache speeds up starting many nodes)
		vm.gnoStore = cachedStdlibLoad(vm.stdlibsDir, baseSDKStore, iavlSDKStore)
	} else {
		// On-chain case
		vm.gnoStore = uncachedPackageLoad(logger, vm.stdlibsDir, baseSDKStore, iavlSDKStore)
	}
}

func uncachedPackageLoad(
	logger *slog.Logger,
	stdlibsDir string,
	baseStore, iavlStore store.Store,
) gno.Store {
	alloc := gno.NewAllocator(maxAllocTx)
	gnoStore := gno.NewStore(alloc, baseStore, iavlStore)
	gnoStore.SetNativeStore(stdlibs.NativeStore)
	if gnoStore.NumMemPackages() == 0 {
		// No packages in the store; set up the stdlibs.
		start := time.Now()

		loadStdlib(stdlibsDir, gnoStore)

		// XXX Quick and dirty to make this function work on non-validator nodes
		iter := iavlStore.Iterator(nil, nil)
		for ; iter.Valid(); iter.Next() {
			baseStore.Set(append(iavlBackupPrefix, iter.Key()...), iter.Value())
		}
		iter.Close()

		logger.Debug("Standard libraries initialized",
			"elapsed", time.Since(start))
	} else {
		// for now, all mem packages must be re-run after reboot.
		// TODO remove this, and generally solve for in-mem garbage collection
		// and memory management across many objects/types/nodes/packages.
		start := time.Now()

		// XXX Quick and dirty to make this function work on non-validator nodes
		if isStoreEmpty(iavlStore) {
			iter := baseStore.Iterator(iavlBackupPrefix, nil)
			for ; iter.Valid(); iter.Next() {
				if !bytes.HasPrefix(iter.Key(), iavlBackupPrefix) {
					break
				}
				iavlStore.Set(iter.Key()[len(iavlBackupPrefix):], iter.Value())
			}
			iter.Close()
		}

		m2 := gno.NewMachineWithOptions(
			gno.MachineOptions{
				PkgPath: "",
				Output:  os.Stdout, // XXX
				Store:   gnoStore,
			})
		defer m2.Release()
		gno.DisableDebug()
		m2.PreprocessAllFilesAndSaveBlockNodes()
		gno.EnableDebug()

		logger.Debug("GnoVM packages preprocessed",
			"elapsed", time.Since(start))
	}
	return gnoStore
}

var iavlBackupPrefix = []byte("init_iavl_backup:")

func isStoreEmpty(st store.Store) bool {
	iter := st.Iterator(nil, nil)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		return false
	}
	return true
}

func cachedStdlibLoad(stdlibsDir string, baseStore, iavlStore store.Store) gno.Store {
	cachedStdlibOnce.Do(func() {
		cachedStdlibBase = memdb.NewMemDB()
		cachedStdlibIavl = memdb.NewMemDB()

		cachedGnoStore = gno.NewStore(nil,
			dbadapter.StoreConstructor(cachedStdlibBase, types.StoreOptions{}),
			dbadapter.StoreConstructor(cachedStdlibIavl, types.StoreOptions{}))
		cachedGnoStore.SetNativeStore(stdlibs.NativeStore)
		loadStdlib(stdlibsDir, cachedGnoStore)
	})

	itr := cachedStdlibBase.Iterator(nil, nil)
	for ; itr.Valid(); itr.Next() {
		baseStore.Set(itr.Key(), itr.Value())
	}

	itr = cachedStdlibIavl.Iterator(nil, nil)
	for ; itr.Valid(); itr.Next() {
		iavlStore.Set(itr.Key(), itr.Value())
	}

	alloc := gno.NewAllocator(maxAllocTx)
	gs := gno.NewStore(alloc, baseStore, iavlStore)
	gs.SetNativeStore(stdlibs.NativeStore)
	gno.CopyCachesFromStore(gs, cachedGnoStore)
	return gs
}

var (
	cachedStdlibOnce sync.Once
	cachedStdlibBase *memdb.MemDB
	cachedStdlibIavl *memdb.MemDB
	cachedGnoStore   gno.Store
)

func loadStdlib(stdlibsDir string, store gno.Store) {
	stdlibInitList := stdlibs.InitOrder()
	for _, lib := range stdlibInitList {
		if lib == "testing" {
			// XXX: testing is skipped for now while it uses testing-only packages
			// like fmt and encoding/json
			continue
		}
		loadStdlibPackage(lib, stdlibsDir, store)
	}
}

func loadStdlibPackage(pkgPath, stdlibsDir string, store gno.Store) {
	stdlibPath := filepath.Join(stdlibsDir, pkgPath)
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

func (vm *VMKeeper) getGnoStore(ctx sdk.Context) gno.Store {
	// construct main store if nil.
	if vm.gnoStore == nil {
		panic("VMKeeper must first be initialized")
	}
	switch ctx.Mode() {
	case sdk.RunTxModeDeliver:
		// swap sdk store of existing store.
		// this is needed due to e.g. gas wrappers.
		baseSDKStore := ctx.Store(vm.baseKey)
		iavlSDKStore := ctx.Store(vm.iavlKey)
		vm.gnoStore.SwapStores(baseSDKStore, iavlSDKStore)
		// clear object cache for every transaction.
		// NOTE: this is inefficient, but simple.
		// in the future, replace with more advanced caching strategy.
		vm.gnoStore.ClearObjectCache()
		return vm.gnoStore
	case sdk.RunTxModeCheck:
		// For query??? XXX Why not RunTxModeQuery?
		simStore := vm.gnoStore.Fork()
		baseSDKStore := ctx.Store(vm.baseKey)
		iavlSDKStore := ctx.Store(vm.iavlKey)
		simStore.SwapStores(baseSDKStore, iavlSDKStore)
		return simStore
	case sdk.RunTxModeSimulate:
		// always make a new store for simulate for isolation.
		simStore := vm.gnoStore.Fork()
		baseSDKStore := ctx.Store(vm.baseKey)
		iavlSDKStore := ctx.Store(vm.iavlKey)
		simStore.SwapStores(baseSDKStore, iavlSDKStore)
		return simStore
	default:
		panic("should not happen")
	}
}

// Namespace can be either a user or crypto address.
var reNamespace = regexp.MustCompile(`^gno.land/(?:r|p)/([\.~_a-zA-Z0-9]+)`)

// checkNamespacePermission check if the user as given has correct permssion to on the given pkg path
func (vm *VMKeeper) checkNamespacePermission(ctx sdk.Context, creator crypto.Address, pkgPath string) error {
	const sysUsersPkg = "gno.land/r/sys/users"

	store := vm.getGnoStore(ctx)

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
	gnostore := vm.getGnoStore(ctx)

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
	gnostore := vm.getGnoStore(ctx)
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
	gnostore := vm.getGnoStore(ctx)
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

// SetMeta sets or updates the metadata of a package.
func (vm *VMKeeper) SetMeta(ctx sdk.Context, msg MsgSetMeta) error {
	store := vm.getGnoStore(ctx)
	if p := store.GetPackage(msg.PkgPath, false); p == nil {
		return ErrInvalidPkgPath("package doesn't exists: " + msg.PkgPath)
	}

	// Only the package creator is allowed to change its metadata.
	// TODO: When gno.land/r/sys/users is not available check silently passes. How to check effectibly?
	if err := vm.checkNamespacePermission(ctx, msg.Caller, msg.PkgPath); err != nil {
		return err
	}

	// Get current metadata field names
	fields := make(map[string]struct{})
	store.IteratePackageMeta(msg.PkgPath, func(name string, _ []byte) bool {
		fields[name] = struct{}{}
		return false
	})

	// Make sure that total number of fields is still valid
	var toolsValue []byte
	for _, f := range msg.Fields {
		// The tools field can be setted only once, after initialization the field is inmutable
		if f.Name == "tools" {
			if _, exists := fields[f.Name]; exists {
				return ErrInvalidPkgMeta("tools package metadata field is already initialized")
			}

			// Get the tools field value so it can be validated
			toolsValue = f.Value
		}

		// Fields with empty values are removed, otherwise they are added
		if len(f.Value) == 0 {
			delete(fields, f.Name)
		} else if _, exists := fields[f.Name]; !exists {
			fields[f.Name] = struct{}{}
		}
	}

	if len(fields) > maxMetaFields {
		return ErrInvalidPkgMeta("maximum number of package metadata fields exceeded")
	}

	// Validate tools metadata field value
	if toolsValue != nil {
		if err := validateToolsMetaField(toolsValue); err != nil {
			return ErrInvalidPkgMeta(err.Error())
		}
	}

	// Update metadata fields, deleting fields with empty values
	for _, f := range msg.Fields {
		if len(f.Value) == 0 {
			store.DeletePackageMetaField(msg.PkgPath, f.Name)
		} else {
			store.SetPackageMetaField(msg.PkgPath, f.Name, f.Value)
		}
	}
	return nil
}

// QueryFuncs returns public facing function signatures.
func (vm *VMKeeper) QueryFuncs(ctx sdk.Context, pkgPath string) (fsigs FunctionSignatures, err error) {
	store := vm.getGnoStore(ctx)
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
	alloc := gno.NewAllocator(maxAllocQuery)
	gnostore := vm.getGnoStore(ctx)
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
	alloc := gno.NewAllocator(maxAllocQuery)
	gnostore := vm.getGnoStore(ctx)
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
	store := vm.getGnoStore(ctx)
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

func (vm *VMKeeper) QueryMeta(ctx sdk.Context, pkgPath, name string) ([]byte, error) {
	var (
		res   []byte
		store = vm.getGnoStore(ctx)
	)

	found := store.IteratePackageMeta(pkgPath, func(field string, value []byte) bool {
		if field == name {
			res = make([]byte, base64.StdEncoding.EncodedLen(len(value)))
			base64.StdEncoding.Encode(res, value)
			return true
		}
		return false
	})

	if !found {
		return nil, fmt.Errorf("metadata field for package %s not found: %s", pkgPath, name) // TODO: XSS protection
	}
	return res, nil
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

func validateToolsMetaField(value []byte) error {
	var v struct {
		Tools []struct {
			Name        string  `json:"name"`
			Weight      float64 `json:"weight"`
			Description string  `json:"description,omitempty"`
		} `json:"tools,omitempty"`
	}

	if err := json.Unmarshal(value, &v); err != nil {
		return fmt.Errorf("invalid tools field format: %s", err)
	}

	if len(v.Tools) == 0 {
		return errors.New("tools list is empty")
	}

	var totalWeight float64
	for _, t := range v.Tools {
		if len(t.Description) > maxToolDescriptionLenght {
			return fmt.Errorf("maximum length for tool description is %d", maxToolDescriptionLenght)
		}

		if !reToolName.MatchString(t.Name) {
			return fmt.Errorf(
				"invalid tool name: %s (minimum 6 chars, lowercase alphanumeric with underscore)",
				t.Name,
			)
		}

		totalWeight += t.Weight
	}

	if totalWeight != 1 {
		return errors.New("the sum of all tool weights must be 1")
	}
	return nil
}
