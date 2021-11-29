package stdlibs

import (
	"reflect"
	"strconv"

	"github.com/gnolang/gno"
)

func InjectNatives(store gno.Store, pv *gno.PackageValue) {
	// First load the package node.
	pkg := store.GetBlockNode(gno.PackageNodeLocation(pv.PkgPath)).(*gno.PackageNode)
	// Override loaded package and update pv.
	switch pv.PkgPath {
	case "strconv":
		pkg.DefineGoNativeFunc("Itoa", strconv.Itoa)
		pkg.DefineGoNativeFunc("Atoi", strconv.Atoi)
		pkg.PrepareNewValues(pv)
	case "std":
		// Native functions.
		// NOTE: pkgs/sdk/vm/VMKeeper also
		// injects more like .Send, .GetContext.
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
		pkg.DefineNative("IsOriginCall",
			gno.Flds( // params
			),
			gno.Flds( // results
				"isOrigin", "bool",
			),
			func(m *gno.Machine) {
				isOrigin := len(m.Frames) == 1
				res0 := gno.TypedValue{T: gno.BoolType}
				res0.SetBool(isOrigin)
				m.PushValue(res0)
			},
		)
		pkg.PrepareNewValues(pv)
	default:
		// nothing to do
	}
}
