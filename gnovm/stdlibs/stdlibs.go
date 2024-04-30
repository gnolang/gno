package stdlibs

//go:generate go run github.com/gnolang/gno/misc/genstd

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	libsstd "github.com/gnolang/gno/gnovm/stdlibs/std"
)

type ExecContext = libsstd.ExecContext

func NativeStore(pkgPath string, name gno.Name) func(*gno.Machine) {
	for _, nf := range nativeFuncs {
		if nf.gnoPkg == pkgPath && name == nf.gnoFunc {
			return nf.f
		}
	}
	return nil
}
