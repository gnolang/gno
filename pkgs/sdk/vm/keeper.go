package vm

import (
	"fmt"
	"path"
	"path/filepath"
	"reflect"

	"github.com/gnolang/gno"
	"github.com/gnolang/gno/pkgs/crypto"
	dbm "github.com/gnolang/gno/pkgs/db"
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
	Eval(ctx sdk.Context, msg MsgEval) (string, error)
}

var _ VMKeeperI = VMKeeper{}

// VMKeeper holds all package code and store state.
type VMKeeper struct {
	key  store.StoreKey
	acck auth.AccountKeeper
	bank bank.BankKeeper

	// TODO: remove these and fully implement persistence.
	// For now, the whole chain must be re-run with each reboot.
	fs    *dbm.FSDB // XXX hack -- not immutable store.
	store gno.Store // XXX hack -- in mem only.
}

// NewVMKeeper returns a new VMKeeper.
func NewVMKeeper(key store.StoreKey, acck auth.AccountKeeper, bank bank.BankKeeper) VMKeeper {
	fs := dbm.NewFSDB("_testdata")  // XXX hack
	store := gno.NewCacheStore(nil) // XXX hack

	vmk := VMKeeper{
		key:   key,
		acck:  acck,
		bank:  bank,
		fs:    fs,
		store: store,
	}
	// initialize built-in packages.
	vmk.initBuiltinPackages(store)
	return vmk
}

func (vmk VMKeeper) initBuiltinPackages(store gno.Store) {
	{ // std
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
			reflect.TypeOf((*EvalContext)(nil)).Elem())
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
				ctx := m.Context.(EvalContext)
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
				"ctx", "EvalContext",
			),
			func(m *gno.Machine) {
				ctx := m.Context.(EvalContext)
				res0 := gno.Go2GnoValue(
					reflect.ValueOf(ctx),
				)
				m.PushValue(res0)
			},
		)
		store.SetPackage(pkg.NewPackage())
	}
}

// AddPackage adds a package with given fileset.
func (vm VMKeeper) AddPackage(ctx sdk.Context, msg MsgAddPackage) error {
	creator := msg.Creator
	pkgPath := msg.PkgPath
	files := msg.Files
	deposit := msg.Deposit

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
	if pv := vm.store.GetPackage(pkgPath); pv != nil {
		// TODO: return error instead of panicking?
		panic("package already exists: " + pkgPath)
	}
	// Pay deposit from creator.
	pkgAddr := DerivePkgAddr(pkgPath)
	err := vm.bank.SendCoins(ctx, creator, pkgAddr, deposit)
	if err != nil {
		return err
	}
	// Add files to global. NOTE: hack
	for _, file := range files {
		name := file.Name
		body := file.Body
		fpath := path.Join(pkgPath, name)
		vm.fs.Set([]byte(fpath), []byte(body))
	}
	// Parse and run the files, construct *PV.
	pkgName := gno.Name("")
	fnodes := []*gno.FileNode{}
	for i, file := range files {
		if filepath.Ext(file.Name) != ".go" {
			continue
		}
		fnode := gno.MustParseFile(file.Name, file.Body)
		if i == 0 {
			pkgName = fnode.PkgName
		} else if fnode.PkgName != pkgName {
			panic(fmt.Sprintf(
				"expected package name %q but got %v",
				pkgName,
				fnode.PkgName))
		}
		fnodes = append(fnodes, fnode)
	}
	pkg := gno.NewPackageNode(pkgName, pkgPath, nil)
	pv := pkg.NewPackage()
	m2 := gno.NewMachineWithOptions(
		gno.MachineOptions{
			Package: pv,
			Output:  nil, // XXX
			Store:   vm.store,
		})
	m2.RunFiles(fnodes...)
	// Set package to store.
	vm.store.SetPackage(pv)
	return nil
}

// Eval evaluates gno expression (for delivertx).
func (vm VMKeeper) Eval(ctx sdk.Context, msg MsgEval) (res string, err error) {
	pkgPath := msg.PkgPath
	expr := msg.Expr
	// Get Package.
	pv := vm.store.GetPackage(pkgPath)
	if pv == nil {
		err = ErrInvalidPkgPath("package not found")
		return "", err
	}
	// Parse expression.
	xx, err := gno.ParseExpr(expr)
	if err != nil {
		return "", err
	}
	// Send send-coins to pkg from caller.
	pkgAddr := DerivePkgAddr(pkgPath)
	caller := msg.Caller
	send := msg.Send
	err = vm.bank.SendCoins(ctx, caller, pkgAddr, send)
	if err != nil {
		return "", err
	}

	// Construct new machine.
	msgCtx := EvalContext{
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
			Store:   vm.store,
			Context: msgCtx,
		})
	rtv := m.Eval(xx)
	res = rtv.String()
	return res, nil
	// TODO pay for gas? TODO see context?
}

// QueryEval evaluates gno expression (readonly, for ABCI queries).
func (vm VMKeeper) QueryEval(ctx sdk.Context, pkgPath string, expr string) (res string, err error) {
	// Get Package.
	pv := vm.store.GetPackage(pkgPath)
	if pv == nil {
		err = ErrInvalidPkgPath("package not found")
		return "", err
	}
	// Parse expression.
	xx, err := gno.ParseExpr(expr)
	if err != nil {
		return "", err
	}
	// Construct new machine.
	msgCtx := EvalContext{
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
			Store:   vm.store,
			Context: msgCtx,
		})
	rtv := m.Eval(xx)
	res = rtv.String()
	return res, nil
}

//----------------------------------------

// For keeping record of package & realm coins.
func DerivePkgAddr(pkgPath string) crypto.Address {
	// NOTE: must not collide with pubkey addrs.
	return crypto.AddressFromPreimage([]byte("pkgPath:" + pkgPath))
}
