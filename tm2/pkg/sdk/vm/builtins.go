package vm

import (
	"os"
	"path/filepath"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	osm "github.com/gnolang/gno/tm2/pkg/os"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func (vm *VMKeeper) initBuiltinPackagesAndTypes(store gno.Store) {
	// NOTE: native functions/methods added here must be quick operations,
	// or account for gas before operation.
	// TODO: define criteria for inclusion, and solve gas calculations.
	getPackage := func(pkgPath string) (pn *gno.PackageNode, pv *gno.PackageValue) {
		// otherwise, built-in package value.
		// first, load from filepath.
		stdlibPath := filepath.Join(vm.stdlibsDir, pkgPath)
		if !osm.DirExists(stdlibPath) {
			// does not exist.
			return nil, nil
		}
		memPkg := gno.ReadMemPackage(stdlibPath, pkgPath)
		m2 := gno.NewMachineWithOptions(gno.MachineOptions{
			PkgPath: "gno.land/r/stdlibs/" + pkgPath,
			// PkgPath: pkgPath,
			Output: os.Stdout,
			Store:  store,
		})
		defer m2.Release()
		return m2.RunMemPackage(memPkg, true)
	}
	store.SetPackageGetter(getPackage)
	store.SetPackageInjector(vm.packageInjector)
	stdlibs.InjectNativeMappings(store)
}

func (vm *VMKeeper) packageInjector(store gno.Store, pn *gno.PackageNode) {
	// Also inject stdlibs native functions.
	stdlibs.InjectPackage(store, pn)
	// vm (this package) specific injections:
	switch pn.PkgPath {
	case "std":
		/* XXX deleteme
		// Also see stdlibs/InjectPackage.
		pn.DefineNative("AssertOriginCall",
			gno.Flds( // params
			),
			gno.Flds( // results
			),
			func(m *gno.Machine) {
				isOrigin := len(m.Frames) == 2
				if !isOrigin {
					panic("invalid non-origin call")
				}
			},
		)
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
		*/
	}
}

// ----------------------------------------
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

func (bnk *SDKBanker) GetCoins(b32addr crypto.Bech32Address) (dst std.Coins) {
	addr := crypto.MustAddressFromString(string(b32addr))
	coins := bnk.vmk.bank.GetCoins(bnk.ctx, addr)
	return coins
}

func (bnk *SDKBanker) SendCoins(b32from, b32to crypto.Bech32Address, amt std.Coins) {
	from := crypto.MustAddressFromString(string(b32from))
	to := crypto.MustAddressFromString(string(b32to))
	err := bnk.vmk.bank.SendCoins(bnk.ctx, from, to, amt)
	if err != nil {
		panic(err)
	}
}

func (bnk *SDKBanker) TotalCoin(denom string) int64 {
	panic("not yet implemented")
}

func (bnk *SDKBanker) IssueCoin(b32addr crypto.Bech32Address, denom string, amount int64) {
	addr := crypto.MustAddressFromString(string(b32addr))
	_, err := bnk.vmk.bank.AddCoins(bnk.ctx, addr, std.Coins{std.Coin{denom, amount}})
	if err != nil {
		panic(err)
	}
}

func (bnk *SDKBanker) RemoveCoin(b32addr crypto.Bech32Address, denom string, amount int64) {
	addr := crypto.MustAddressFromString(string(b32addr))
	_, err := bnk.vmk.bank.SubtractCoins(bnk.ctx, addr, std.Coins{std.Coin{denom, amount}})
	if err != nil {
		panic(err)
	}
}
