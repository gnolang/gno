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

func findNative(pkgPath string, name gno.Name) nativeFunc {
	for _, nf := range nativeFuncs {
		if nf.gnoPkg == pkgPath && name == nf.gnoFunc {
			return nf
		}
	}
	return nativeFunc{}
}

// NativeStore is used by the GnoVM to determine if the given function,
// specified by its pkgPath and name, has a native implementation; and if so
// retrieve it.
func NativeStore(pkgPath string, name gno.Name) func(*gno.Machine) {
	return findNative(pkgPath, name).f
}

// HasNativeBinding determines if the function specified by the given pkgPath
// and name is a native binding.
func HasNativeBinding(pkgPath string, name gno.Name) bool {
	return findNative(pkgPath, name).f != nil
}

// HasMachineParam determines if the function specified by the given pkgPath
// and name contains a machine parameter; ie., its native implementation is
// prefixed with the parameter `m *gno.Machine`.
func HasMachineParam(pkgPath string, name gno.Name) bool {
	return findNative(pkgPath, name).hasMachine
}
