package stdlibs

import (
	"reflect"
	"strconv"

	"github.com/gnolang/gno"
)

func InjectPackage(store gno.Store, pn *gno.PackageNode, pv *gno.PackageValue) {
	switch pv.PkgPath {
	case "strconv":
		pn.DefineGoNativeFunc("Itoa", strconv.Itoa)
		pn.DefineGoNativeFunc("Atoi", strconv.Atoi)
		pn.PrepareNewValues(pv)
	case "std":
		// NOTE: pkgs/sdk/vm/VMKeeper also
		// injects more like .Send, .GetContext.
		pn.DefineNative("Hash",
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
		pn.DefineNative("IsOriginCall",
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
		pn.PrepareNewValues(pv)
	}
}
