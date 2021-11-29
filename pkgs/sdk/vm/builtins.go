package vm

import (
	"reflect"
	"strconv"

	"github.com/gnolang/gno"
	"github.com/gnolang/gno/pkgs/crypto"
	"github.com/gnolang/gno/pkgs/std"
)

func (vmk *VMKeeper) initBuiltinPackages(store gno.Store) {
	// NOTE: native functions/methods added here must be quick operations,
	// or account for gas before operation.
	// TODO: define criteria for inclusion, and solve gas calculations.
	getPackage := func(pkgPath string) (pv *gno.PackageValue) {
		// otherwise, built-in package value.
		switch pkgPath {
		case "strconv":
			pkg := gno.NewPackageNode("strconv", "strconv", nil)
			pkg.DefineGoNativeFunc("Itoa", strconv.Itoa)
			pkg.DefineGoNativeFunc("Atoi", strconv.Atoi)
			return pkg.NewPackage()
		case "std":
			// TODO: probably, convert these to Gno types.
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
			// Native functions.
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
			pkg.DefineNative("Hash",
				gno.Flds( // params
					"bz", "[]byte",
				),
				gno.Flds( // results
					"hash", "[20]byte",
				),
				func(m *gno.Machine) {
					arg0 := m.LastBlock().GetParams1().TV
					bz := []byte(nil)
					if arg0.V != nil {
						slice := arg0.V.(*gno.SliceValue)
						array := slice.GetBase(m.Store)
						bz = array.GetReadonlyBytes()
					}
					hash := gno.HashBytes(bz)
					res0 := gno.Go2GnoValue(
						reflect.ValueOf([20]byte(hash)),
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
