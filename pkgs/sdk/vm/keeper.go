package vm

import (
	"fmt"
	"os"

	"github.com/gnolang/gno"
	"github.com/gnolang/gno/pkgs/crypto"
	"github.com/gnolang/gno/pkgs/sdk"
	"github.com/gnolang/gno/pkgs/sdk/auth"
	"github.com/gnolang/gno/pkgs/sdk/bank"
	"github.com/gnolang/gno/pkgs/std"
	"github.com/gnolang/gno/pkgs/store"
)

// vm.VMKeeperI defines a module interface that supports Gno
// smart contracts programming (scripting).
type VMKeeperI interface {
	AddPackage(ctx sdk.Context, msg MsgAddPackage) error
	Call(ctx sdk.Context, msg MsgCall) (res string, err error)
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
}

// NewVMKeeper returns a new VMKeeper.
func NewVMKeeper(baseKey store.StoreKey, iavlKey store.StoreKey, acck auth.AccountKeeper, bank bank.BankKeeper, stdlibsDir string) *VMKeeper {
	vmk := &VMKeeper{
		baseKey:    baseKey,
		iavlKey:    iavlKey,
		acck:       acck,
		bank:       bank,
		stdlibsDir: stdlibsDir,
	}
	return vmk
}

func (vmk *VMKeeper) getGnoStore(ctx sdk.Context) gno.Store {
	switch ctx.Mode() {
	case sdk.RunTxModeDeliver:
		// construct gnoStore if nil.
		if vmk.gnoStore == nil {
			baseSDKStore := ctx.Store(vmk.baseKey)
			iavlSDKStore := ctx.Store(vmk.iavlKey)
			vmk.gnoStore = gno.NewStore(baseSDKStore, iavlSDKStore)
			vmk.initBuiltinPackages(vmk.gnoStore)
			if vmk.gnoStore.NumMemPackages() > 0 {
				// for now, all mem packages must be re-run after reboot.
				// TODO remove this, and generally solve for in-mem garbage collection
				// and memory management across many objects/types/nodes/packages.
				m2 := gno.NewMachineWithOptions(
					gno.MachineOptions{
						Package: nil,
						Output:  os.Stdout, // XXX
						Store:   vmk.gnoStore,
					})
				m2.PreprocessAllFilesAndSaveBlockNodes()
			}
		} else {
			// otherwise, swap sdk store of existing gnoStore.
			// this is needed due to e.g. gas wrappers.
			baseStore := ctx.Store(vmk.baseKey)
			iavlStore := ctx.Store(vmk.iavlKey)
			vmk.gnoStore.SwapStores(baseStore, iavlStore)
		}
		return vmk.gnoStore
	case sdk.RunTxModeCheck:
		// For query??? XXX Why not RunTxModeQuery?
		baseSDKStore := ctx.Store(vmk.baseKey)
		iavlSDKStore := ctx.Store(vmk.iavlKey)
		simStore := gno.NewStore(baseSDKStore, iavlSDKStore)
		vmk.initBuiltinPackages(simStore)
		// XXX This is crazy, there has to be a better way.
		if simStore.NumMemPackages() > 0 {
			m2 := gno.NewMachineWithOptions(
				gno.MachineOptions{
					Package: nil,
					Output:  nil, // XXX
					Store:   simStore,
				})
			m2.PreprocessAllFilesAndSaveBlockNodes()
		}
		return simStore
	case sdk.RunTxModeSimulate:
		// always make a new store for simualte for isolation.
		baseSDKStore := ctx.Store(vmk.baseKey)
		iavlSDKStore := ctx.Store(vmk.iavlKey)
		simStore := gno.NewStore(baseSDKStore, iavlSDKStore)
		vmk.initBuiltinPackages(simStore)
		// XXX This is crazy, there has to be a better way.
		if simStore.NumMemPackages() > 0 {
			m2 := gno.NewMachineWithOptions(
				gno.MachineOptions{
					Package: nil,
					Output:  nil, // XXX
					Store:   simStore,
				})
			m2.PreprocessAllFilesAndSaveBlockNodes()
		}
		return simStore
	default:
		panic("should not happen")
	}
}

// AddPackage adds a package with given fileset.
func (vm *VMKeeper) AddPackage(ctx sdk.Context, msg MsgAddPackage) error {
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
	if pkgPath == "" {
		return ErrInvalidPkgPath("missing package path")
	}
	if pv := store.GetPackage(pkgPath); pv != nil {
		// TODO: return error instead of panicking?
		panic("package already exists: " + pkgPath)
	}
	// Pay deposit from creator.
	pkgAddr := DerivePkgAddr(pkgPath)
	err := vm.bank.SendCoins(ctx, creator, pkgAddr, deposit)
	if err != nil {
		return err
	}
	// Parse and run the files, construct *PV.
	m2 := gno.NewMachineWithOptions(
		gno.MachineOptions{
			Package: nil,
			Output:  os.Stdout, // XXX
			Store:   store,
		})
	m2.RunMemPackage(memPkg, true)
	return nil
}

// Calls calls a public Gno function (for delivertx).
func (vm *VMKeeper) Call(ctx sdk.Context, msg MsgCall) (res string, err error) {
	pkgPath := msg.PkgPath // to import
	fnc := msg.Func
	store := vm.getGnoStore(ctx)
	// Get the package and function type.
	pv := store.GetPackage(pkgPath)
	pl := gno.PackageNodeLocation(pkgPath)
	pn := store.GetBlockNode(pl).(*gno.PackageNode)
	ft := pn.GetStaticTypeOf(store, gno.Name(fnc)).(*gno.FuncType)
	// Make main Package with imports.
	mpn := gno.NewPackageNode("main", "main", nil)
	mpn.Define("pkg", gno.TypedValue{T: &gno.PackageType{}, V: pv})
	mpv := mpn.NewPackage()
	// Parse expression.
	argslist := ""
	for i, _ := range msg.Args {
		if i > 0 {
			argslist += ","
		}
		argslist += fmt.Sprintf("arg%d", i)
	}
	expr := fmt.Sprintf(`pkg.%s(%s)`, fnc, argslist)
	xn := gno.MustParseExpr(expr)
	// Send send-coins to pkg from caller.
	pkgAddr := DerivePkgAddr(pkgPath)
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
	for i, arg := range msg.Args {
		argType := ft.Params[i].Type
		atv := convertArgToGno(arg, argType)
		cx.Args[i] = &gno.ConstExpr{
			TypedValue: atv,
		}
	}
	// Construct new machine.
	msgCtx := ExecContext{
		ChainID: ctx.ChainID(),
		Height:  ctx.BlockHeight(),
		Msg:     msg,
		PkgAddr: pkgAddr,
		sdkCtx:  ctx,
	}
	m := gno.NewMachineWithOptions(
		gno.MachineOptions{
			Package: mpv,
			Output:  os.Stdout, // XXX
			Store:   store,
			Context: msgCtx,
		})
	rtvs := m.Eval(xn)
	for i, rtv := range rtvs {
		res = res + rtv.String()
		if i < len(rtvs)-1 {
			res += "\n"
		}
	}
	return res, nil
	// TODO pay for gas? TODO see context?
}

// QueryEval evaluates gno expression (readonly, for ABCI queries).
// TODO: modify query protocol to allow MsgEval.
// TODO: then, rename to "Eval".
func (vm *VMKeeper) QueryEval(ctx sdk.Context, pkgPath string, expr string) (res string, err error) {
	store := vm.getGnoStore(ctx)
	// Get Package.
	pv := store.GetPackage(pkgPath)
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
	msgCtx := ExecContext{
		ChainID: ctx.ChainID(),
		Height:  ctx.BlockHeight(),
		//Msg:     msg,
		//PkgAddr: pkgAddr,
		sdkCtx: ctx,
	}
	m := gno.NewMachineWithOptions(
		gno.MachineOptions{
			Package: pv,
			Output:  os.Stdout, // XXX
			Store:   store,
			Context: msgCtx,
		})
	rtvs := m.Eval(xx)
	for i, rtv := range rtvs {
		res = res + rtv.String()
		if i < len(rtvs)-1 {
			res += "\n"
		}
	}
	return res, nil
}

//----------------------------------------

// For keeping record of package & realm coins.
func DerivePkgAddr(pkgPath string) crypto.Address {
	// NOTE: must not collide with pubkey addrs.
	return crypto.AddressFromPreimage([]byte("pkgPath:" + pkgPath))
}
