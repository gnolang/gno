package stdlibs

//go:generate go run github.com/gnolang/gno/misc/genstd

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	libsstd "github.com/gnolang/gno/gnovm/stdlibs/std"
)

type ExecContext = libsstd.ExecContext

func GetContext(m *gno.Machine) ExecContext {
	return libsstd.GetContext(m)
}

// FindNative returns the NativeFunc associated with the given pkgPath+name
// combination. If there is none, FindNative returns nil.
func FindNative(pkgPath string, name gno.Name) *NativeFunc {
	for i, nf := range nativeFuncs {
		if nf.gnoPkg == pkgPath && name == nf.gnoFunc {
			return &nativeFuncs[i]
		}
	}
	return nil
}

// NativeStore is used by the GnoVM to determine if the given function,
// specified by its pkgPath and name, has a native implementation; and if so
// retrieve it.
func NativeStore(pkgPath string, name gno.Name) func(*gno.Machine) {
	nt := FindNative(pkgPath, name)
	if nt == nil {
		return nil
	}
	return nt.f
}
