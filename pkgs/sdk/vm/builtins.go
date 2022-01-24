package vm

import (
	"os"
	"path/filepath"

	"github.com/gnolang/gno"
	"github.com/gnolang/gno/pkgs/crypto"
	osm "github.com/gnolang/gno/pkgs/os"
	"github.com/gnolang/gno/pkgs/sdk"
	"github.com/gnolang/gno/pkgs/std"
	"github.com/gnolang/gno/stdlibs"
)

func (vmk *VMKeeper) initBuiltinPackages(store gno.Store) {
	// NOTE: native functions/methods added here must be quick operations,
	// or account for gas before operation.
	// TODO: define criteria for inclusion, and solve gas calculations.
	getPackage := func(pkgPath string) (pn *gno.PackageNode, pv *gno.PackageValue) {
		// otherwise, built-in package value.
		// first, load from filepath.
		stdlibPath := filepath.Join(vmk.stdlibsDir, pkgPath)
		if !osm.DirExists(stdlibPath) {
			// does not exist.
			return nil, nil
		}
		memPkg := gno.ReadMemPackage(stdlibPath, pkgPath)
		m2 := gno.NewMachineWithOptions(gno.MachineOptions{
			Package: nil,
			Output:  os.Stdout,
			Store:   store,
		})
		return m2.RunMemPackage(memPkg, true)
	}
	store.SetPackageGetter(getPackage)
	store.SetPackageInjector(vmk.packageInjector)
}

func (vmk *VMKeeper) packageInjector(store gno.Store, pn *gno.PackageNode) {
	// Also inject stdlibs native functions.
	stdlibs.InjectPackage(store, pn)
	// vm (this package) specific injections:
	switch pn.PkgPath {
	case "std":
		// Also see stdlibs/InjectPackage.
		pn.DefineNative("IsOriginCall",
			gno.Flds( // params
			),
			gno.Flds( // results
				"isOrigin", "bool",
			),
			func(m *gno.Machine) {
				isOrigin := len(m.Frames) == 2
				res0 := gno.TypedValue{T: gno.BoolType}
				res0.SetBool(isOrigin)
				m.PushValue(res0)
			},
		)
	}
}

//----------------------------------------
// SDKBanker

type SDKBanker struct {
	vmk *VMKeeper
	ctx sdk.Context
}

func NewSDKBanker(vmk *VMKeeper, ctx sdk.Context) *SDKBanker {
	return &SDKBanker{
		vmk: vmk,
		ctx: ctx,
	}
}

func (bnk *SDKBanker) GetCoins(addr crypto.Address, dst *std.Coins) {
	coins := bnk.vmk.bank.GetCoins(bnk.ctx, addr)
	*dst = coins
}
func (bnk *SDKBanker) SendCoins(from, to crypto.Address, amt std.Coins) {
	err := bnk.vmk.bank.SendCoins(bnk.ctx, from, to, amt)
	if err != nil {
		panic(err)
	}
}
func (bnk *SDKBanker) TotalCoin(denom string) int64 {
	panic("not yet implemented")
}
func (bnk *SDKBanker) IssueCoin(addr crypto.Address, denom string, amount int64) {
	_, err := bnk.vmk.bank.AddCoins(bnk.ctx, addr, std.Coins{std.Coin{denom, amount}})
	if err != nil {
		panic(err)
	}
}
func (bnk *SDKBanker) RemoveCoin(addr crypto.Address, denom string, amount int64) {
	_, err := bnk.vmk.bank.SubtractCoins(bnk.ctx, addr, std.Coins{std.Coin{denom, amount}})
	if err != nil {
		panic(err)
	}
}
