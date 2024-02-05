package stdlibs

//go:generate go run github.com/gnolang/gno/misc/genstd

import (
	"reflect"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	libsstd "github.com/gnolang/gno/gnovm/stdlibs/std"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type ExecContext = libsstd.ExecContext

func InjectNativeMappings(store gno.Store) {
	store.AddGo2GnoMapping(reflect.TypeOf(crypto.Bech32Address("")), "std", "Address")
	store.AddGo2GnoMapping(reflect.TypeOf(std.Coins{}), "std", "Coins")
	store.AddGo2GnoMapping(reflect.TypeOf(std.Coin{}), "std", "Coin")
	store.AddGo2GnoMapping(reflect.TypeOf(libsstd.Realm{}), "std", "Realm")
}

func NativeStore(pkgPath string, name gno.Name) func(*gno.Machine) {
	for _, nf := range nativeFuncs {
		if nf.gnoPkg == pkgPath && name == nf.gnoFunc {
			return nf.f
		}
	}
	return nil
}
