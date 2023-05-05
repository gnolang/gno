package vmk

import (
	"fmt"
	"os"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/crypto"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs"
	"github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/sdk/bank"
	vmh "github.com/gnolang/gno/tm2/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
)

const (
	maxAllocTx    = 500 * 1000 * 1000
	maxAllocQuery = 1500 * 1000 * 1000 // higher limit for queries
)

// VMKeeper holds all package code and store state.
type VMKeeper struct {
	baseKey store.StoreKey
	iavlKey store.StoreKey
	acck    auth.AccountKeeper
	bank    bank.BankKeeper
	// dispatcher *Dispatcher
	stdlibsDir string

	// cached, the DeliverTx persistent state.
	gnoStore         gno.Store
	ctx              sdk.Context     // share in a chained call
	callStack        []*vmh.MsgCall  // call stack, typically, call1<-callback1, call2<-callback2,... the whole call graph
	internalMsgQueue chan vmh.GnoMsg // receive msg from contract
	ibcMsgQueue      chan string
	ibcResponseQueue chan string
	IBCChannelKeeper *IBCChannelKeeper // mock
}

// NewVMKeeper returns a new VMKeeper.
func NewVMKeeper(baseKey store.StoreKey, iavlKey store.StoreKey, acck auth.AccountKeeper, bank bank.BankKeeper, stdlibsDir string) *VMKeeper {
	vmk := &VMKeeper{
		baseKey:          baseKey,
		iavlKey:          iavlKey,
		acck:             acck,
		bank:             bank,
		stdlibsDir:       stdlibsDir,
		internalMsgQueue: make(chan vmh.GnoMsg),
		ibcMsgQueue:      make(chan string),
		ibcResponseQueue: make(chan string),
	}
	return vmk
}

func (vmk *VMKeeper) SubmitTxFee(ctx sdk.Context, fromAddr crypto.Address, toAddr crypto.Address, amt std.Coins) error {
	err := vmk.bank.SendCoins(ctx, fromAddr, toAddr, amt)
	return err
}

func (vmk *VMKeeper) PushCall(call *vmh.MsgCall) {
	vmk.callStack = append(vmk.callStack, call)
}

func (vmk *VMKeeper) printCallStack() {
	for _, m := range vmk.callStack {
		println("msgCall: Caller, PkgPath, Func", m.Caller.String(), m.PkgPath, m.Func)
	}
}

// get origCaller from callstack
func (vmk *VMKeeper) GetOrigCaller() crypto.Address {
	if len(vmk.callStack) == 0 {
		panic("should not happen")
	}
	return vmk.callStack[0].Caller
}

func (vmk *VMKeeper) PopCall() (call *vmh.MsgCall) {
	if len(vmk.callStack) == 0 {
		return nil
	}
	lastIndex := len(vmk.callStack) - 1
	e := vmk.callStack[lastIndex]
	vmk.callStack = vmk.callStack[:lastIndex]
	return e
}

func (vmk *VMKeeper) Release() {
	vmk.callStack = vmk.callStack[:0]
	// copy(vmk.callStack, vmk.callStack[:0])
}

func (vmk *VMKeeper) Initialize(ms store.MultiStore) {
	if vmk.gnoStore != nil {
		panic("should not happen")
	}
	alloc := gno.NewAllocator(maxAllocTx)
	baseSDKStore := ms.GetStore(vmk.baseKey)
	iavlSDKStore := ms.GetStore(vmk.iavlKey)
	vmk.gnoStore = gno.NewStore(alloc, baseSDKStore, iavlSDKStore)
	vmk.initBuiltinPackagesAndTypes(vmk.gnoStore)
	if vmk.gnoStore.NumMemPackages() > 0 {
		// for now, all mem packages must be re-run after reboot.
		// TODO remove this, and generally solve for in-mem garbage collection
		// and memory management across many objects/types/nodes/packages.
		m2 := gno.NewMachineWithOptions(
			gno.MachineOptions{
				PkgPath: "",
				Output:  os.Stdout, // XXX
				Store:   vmk.gnoStore,
			})
		gno.DisableDebug()
		m2.PreprocessAllFilesAndSaveBlockNodes()
		gno.EnableDebug()
	}
}

func (vmk *VMKeeper) getGnoStore(ctx sdk.Context) gno.Store {
	// construct main gnoStore if nil.
	if vmk.gnoStore == nil {
		panic("VMKeeper must first be initialized")
	}
	switch ctx.Mode() {
	case sdk.RunTxModeDeliver:
		// swap sdk store of existing gnoStore.
		// this is needed due to e.g. gas wrappers.
		baseSDKStore := ctx.Store(vmk.baseKey)
		iavlSDKStore := ctx.Store(vmk.iavlKey)
		vmk.gnoStore.SwapStores(baseSDKStore, iavlSDKStore)
		// clear object cache for every transaction.
		// NOTE: this is inefficient, but simple.
		// in the future, replace with more advanced caching strategy.
		vmk.gnoStore.ClearObjectCache()
		return vmk.gnoStore
	case sdk.RunTxModeCheck:
		// For query??? XXX Why not RunTxModeQuery?
		simStore := vmk.gnoStore.Fork()
		baseSDKStore := ctx.Store(vmk.baseKey)
		iavlSDKStore := ctx.Store(vmk.iavlKey)
		simStore.SwapStores(baseSDKStore, iavlSDKStore)
		return simStore
	case sdk.RunTxModeSimulate:
		// always make a new store for simulate for isolation.
		simStore := vmk.gnoStore.Fork()
		baseSDKStore := ctx.Store(vmk.baseKey)
		iavlSDKStore := ctx.Store(vmk.iavlKey)
		simStore.SwapStores(baseSDKStore, iavlSDKStore)
		return simStore
	default:
		panic("should not happen")
	}
}

// input
func (vmk *VMKeeper) DispatchInternalMsg(msg vmh.GnoMsg) {
	vmk.internalMsgQueue <- msg
}

func (vmk *VMKeeper) StartEventLoop() {
	for {
		select {
		case msg := <-vmk.internalMsgQueue:
			go vmk.HandleMsg(msg)
		}
	}
}

// initially called somewhere, call? which act like a main routine
func (vmk *VMKeeper) HandleMsg(msg vmh.GnoMsg) {
	println("------HandleMsg, routine spawned ------")
	// prepare call
	msgCall, isLocal, response, err := vmk.preprocessMessage(msg)
	if err != nil {
		panic(err.Error())
	}
	// do the call
	if isLocal {
		println("in VM call")
		println("msgCall: ", msgCall.Caller.String(), msgCall.PkgPath, msgCall.Func, msgCall.Args[0])

		r, err := vmk.Call(vmk.ctx, msgCall)
		println("call finished, res: ", r)
		// have an return
		if err == nil {
			response <- r
		}
	} else { // IBC call
		println("IBC call")
		// send IBC packet, waiting for OnRecv
		vmk.ibcResponseQueue = response
		vmk.SendIBCMsg(vmk.ctx, msgCall)
	}
}

// AddPackage adds a package with given fileset.
func (vm *VMKeeper) AddPackage(ctx sdk.Context, msg vmh.MsgAddPackage) error {
	creator := msg.Creator
	pkgPath := msg.Package.Path
	memPkg := msg.Package
	deposit := msg.Deposit
	store := vm.getGnoStore(ctx)

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
	if pv := store.GetPackage(pkgPath, false); pv != nil {
		// TODO: return error instead of panicking?
		panic("package already exists: " + pkgPath)
	}
	// Pay deposit from creator.
	pkgAddr := gno.DerivePkgAddr(pkgPath)

	// TODO: ACLs.
	// - if r/system/names does not exists -> skip validation.
	// - loads r/system/names data state.
	// - lookup r/system/names.namespaces for `{r,p}/NAMES`.
	// - check if caller is in Admins or Editors.
	// - check if namespace is not in pause.

	err := vm.bank.SendCoins(ctx, creator, pkgAddr, deposit)
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
	}
	// Parse and run the files, construct *PV.
	m2 := gno.NewMachineWithOptions(
		gno.MachineOptions{
			PkgPath:   "",
			Output:    os.Stdout, // XXX
			Store:     store,
			Alloc:     store.GetAllocator(),
			Context:   msgCtx,
			MaxCycles: 10 * 1000 * 1000, // 10M cycles // XXX
		})
	m2.RunMemPackage(memPkg, true)
	fmt.Println("CPUCYCLES addpkg", m2.Cycles)
	return nil
}

// Calls calls a public Gno function (for delivertx).
func (vm *VMKeeper) Call(ctx sdk.Context, msg vmh.MsgCall) (res string, err error) {
	// println("vmk call, msg.Caller: ", msg.Caller.String())
	vm.ctx = ctx
	vm.PushCall(&msg)

	pkgPath := msg.PkgPath // to import
	fnc := msg.Func
	store := vm.getGnoStore(ctx)
	// Get the package and function type.
	pv := store.GetPackage(pkgPath, false)
	pl := gno.PackageNodeLocation(pkgPath)
	pn := store.GetBlockNode(pl).(*gno.PackageNode)
	ft := pn.GetStaticTypeOf(store, gno.Name(fnc)).(*gno.FuncType)
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
	// caller := msg.Caller
	caller := vm.GetOrigCaller()
	println("caller: ", caller.String())
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
	}
	// Construct machine and evaluate.
	m := gno.NewMachineWithOptions(
		gno.MachineOptions{
			PkgPath:   "",
			Output:    os.Stdout, // XXX
			Store:     store,
			Context:   msgCtx,
			Alloc:     store.GetAllocator(),
			MaxCycles: 10 * 1000 * 1000, // 10M cycles // XXX
		})
	m.SetActivePackage(mpv)
	defer func() {
		if r := recover(); r != nil {
			err = errors.Wrap(fmt.Errorf("%v", r), "VM call panic: %v\n%s\n",
				r, m.String())
			return
		}
		m.Release()
		// vm.Release()
		// vm.PopCall()
	}()
	rtvs := m.Eval(xn)
	fmt.Println("CPUCYCLES call", m.Cycles)
	for i, rtv := range rtvs {
		res = res + rtv.String()
		if i < len(rtvs)-1 {
			res += "\n"
		}
	}
	return res, nil
	// TODO pay for gas? TODO see context?
}

// QueryFuncs returns public facing function signatures.
func (vm *VMKeeper) QueryFuncs(ctx sdk.Context, pkgPath string) (fsigs vmh.FunctionSignatures, err error) {
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
		fsig := vmh.FunctionSignature{
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
				vmh.NamedType{Name: pname, Type: ptype},
			)
		}
		for _, result := range ft.Results {
			rname := string(result.Name)
			if rname == "" {
				rname = "_"
			}
			rtype := gno.BaseOf(result.Type).String()
			fsig.Results = append(fsig.Results,
				vmh.NamedType{Name: rname, Type: rtype},
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
	store := vm.getGnoStore(ctx)
	pkgAddr := gno.DerivePkgAddr(pkgPath)
	// Get Package.
	pv := store.GetPackage(pkgPath, false)
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
	}
	m := gno.NewMachineWithOptions(
		gno.MachineOptions{
			PkgPath:   pkgPath,
			Output:    os.Stdout, // XXX
			Store:     store,
			Context:   msgCtx,
			Alloc:     alloc,
			MaxCycles: 10 * 1000 * 1000, // 10M cycles // XXX
		})
	defer func() {
		if r := recover(); r != nil {
			err = errors.Wrap(fmt.Errorf("%v", r), "VM query eval panic: %v\n%s\n",
				r, m.String())
			return
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
	store := vm.getGnoStore(ctx)
	pkgAddr := gno.DerivePkgAddr(pkgPath)
	// Get Package.
	pv := store.GetPackage(pkgPath, false)
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
	}
	m := gno.NewMachineWithOptions(
		gno.MachineOptions{
			PkgPath:   pkgPath,
			Output:    os.Stdout, // XXX
			Store:     store,
			Context:   msgCtx,
			Alloc:     alloc,
			MaxCycles: 10 * 1000 * 1000, // 10M cycles // XXX
		})
	defer func() {
		if r := recover(); r != nil {
			err = errors.Wrap(fmt.Errorf("%v", r), "VM query eval string panic: %v\n%s\n",
				r, m.String())
			return
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
		for i, memfile := range memPkg.Files {
			if i > 0 {
				res += "\n"
			}
			res += memfile.Name
		}
		return res, nil
	}
}

func (vmk *VMKeeper) preprocessMessage(gnoMsg vmh.GnoMsg) (vmh.MsgCall, bool, chan string, error) {
	var isLocal bool
	if gnoMsg.ChainID == vmk.ctx.ChainID() {
		isLocal = true
	}

	msgCall := convertMsg(gnoMsg)
	// using the origCaller
	msgCall.Caller = vmk.GetOrigCaller()

	return msgCall, isLocal, gnoMsg.Response, nil
}

// send IBC packet, only support gnovm type call (MsgCall) for now, gnoVM <-> gnoVM
func (vmk *VMKeeper) SendIBCMsg(ctx sdk.Context, msg vmh.MsgCall) {
	// set callback map
	// this simulates a IBC call, using a chan to loop back
	go vmk.IBCChannelKeeper.SendPacket(msg)
}
