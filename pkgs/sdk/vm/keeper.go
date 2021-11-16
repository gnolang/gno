package vm

import (
	"fmt"
	"reflect"
	"strconv"

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
	Exec(ctx sdk.Context, msg MsgExec) error
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
}

// NewVMKeeper returns a new VMKeeper.
func NewVMKeeper(baseKey store.StoreKey, iavlKey store.StoreKey, acck auth.AccountKeeper, bank bank.BankKeeper) *VMKeeper {
	vmk := &VMKeeper{
		baseKey: baseKey,
		iavlKey: iavlKey,
		acck:    acck,
		bank:    bank,
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
						Output:  nil, // XXX
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
		panic("should not happen")
	case sdk.RunTxModeSimulate:
		// always make a new store for simualte for isolation.
		baseSDKStore := ctx.Store(vmk.baseKey)
		iavlSDKStore := ctx.Store(vmk.iavlKey)
		simStore := gno.NewStore(baseSDKStore, iavlSDKStore)
		vmk.initBuiltinPackages(simStore)
		return simStore
	default:
		panic("should not happen")
	}
}

func (vmk *VMKeeper) initBuiltinPackages(store gno.Store) {
	// NOTE: native functions/methods added here must be quick operations.
	// TODO: define criteria for inclusion, and solve gas calculations.
	getPackage := func(pkgPath string) (pv *gno.PackageValue) {
		// otherwise, built-in package value.
		switch pkgPath {
		case "strconv":
			pkg := gno.NewPackageNode("strconv", "strconv", nil)
			pkg.DefineGoNativeFunc("Itoa", strconv.Itoa)
			return pkg.NewPackage()
		case "std":
			pkg := gno.NewPackageNode("std", "std", nil)
			pkg.DefineGoNativeType(
				reflect.TypeOf((*std.Coin)(nil)).Elem())
			pkg.DefineGoNativeType(
				reflect.TypeOf((*std.Coins)(nil)).Elem())
			pkg.DefineGoNativeType(
				reflect.TypeOf((*crypto.Address)(nil)).Elem())
			pkg.DefineGoNativeType(
				reflect.TypeOf((*crypto.PubKey)(nil)).Elem())
			pkg.DefineGoNativeType(
				reflect.TypeOf((*crypto.PrivKey)(nil)).Elem())
			pkg.DefineGoNativeType(
				reflect.TypeOf((*std.Msg)(nil)).Elem())
			pkg.DefineGoNativeType(
				reflect.TypeOf((*ExecContext)(nil)).Elem())
			pkg.DefineNative("Send",
				gno.Flds( // params
					"toAddr", "Address",
					"coins", "Coins",
				),
				gno.Flds( // results
					"err", "error",
				),
				func(m *gno.Machine) {
					if m.ReadOnly {
						panic("cannot send -- readonly")
					}
					arg0, arg1 := m.LastBlock().GetParams2()
					toAddr := arg0.TV.V.(*gno.NativeValue).Value.Interface().(crypto.Address)
					send := arg1.TV.V.(*gno.NativeValue).Value.Interface().(std.Coins)
					//toAddr := arg0.TV.V.
					ctx := m.Context.(ExecContext)
					err := vmk.bank.SendCoins(
						ctx.sdkCtx,
						ctx.PkgAddr,
						toAddr,
						send,
					)
					if err != nil {
						res0 := gno.Go2GnoValue(
							reflect.ValueOf(err),
						)
						m.PushValue(res0)
					} else {
						m.PushValue(gno.TypedValue{})
					}
				},
			)
			pkg.DefineNative("GetContext",
				gno.Flds( // params
				),
				gno.Flds( // results
					"ctx", "ExecContext",
				),
				func(m *gno.Machine) {
					ctx := m.Context.(ExecContext)
					res0 := gno.Go2GnoValue(
						reflect.ValueOf(ctx),
					)
					m.PushValue(res0)
				},
			)
			return pkg.NewPackage()
		default:
			return nil // does not exist.
		}
	}
	store.SetPackageGetter(getPackage)
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
			Output:  nil, // XXX
			Store:   store,
		})
	m2.RunMemPackage(memPkg, true)
	return nil
}

// Exec executes a Gno statement (for delivertx).
func (vm *VMKeeper) Exec(ctx sdk.Context, msg MsgExec) (err error) {
	pkgPath := msg.PkgPath // to import
	stmt := msg.Stmt
	store := vm.getGnoStore(ctx)
	// Make blank main Package.
	pn := gno.NewPackageNode("main", "main", nil)
	pv := pn.NewPackage()
	// Make and parse file.
	// NOTE: this is temporary until we can optimize.
	// Optimization requires go/parser.ParseStmt.
	fbody := fmt.Sprintf(`package main
import pkg %q

func main() {
	pkg.%s
}`, pkgPath, stmt)
	file := gno.MustParseFile("exec_main.go", fbody)
	// Send send-coins to pkg from caller.
	pkgAddr := DerivePkgAddr(pkgPath)
	caller := msg.Caller
	send := msg.Send
	err = vm.bank.SendCoins(ctx, caller, pkgAddr, send)
	if err != nil {
		return err
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
			Package: pv,
			Output:  nil,
			Store:   store,
			Context: msgCtx,
		})
	m.RunFiles(file)
	m.RunMain()
	return nil
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
			Output:  nil,
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
