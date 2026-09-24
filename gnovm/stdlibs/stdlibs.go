package stdlibs

//go:generate go run github.com/gnolang/gno/misc/genstd

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs/internal/execctx"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// These types are aliases to the equivalent types in internal/execctx.
// The package exists to avoid an import cycle (this package imports individual
// stdlibs through generated.go; thus packages cannot import this package).
type (
	ExecContext = execctx.ExecContext

	// ExecContexter is a type capable of returning the parent [ExecContext]. When
	// using these standard libraries, m.Context should always implement this
	// interface. This can be obtained by embedding [ExecContext].
	ExecContexter = execctx.ExecContexter

	BankerInterface = execctx.BankerInterface
	ParamsInterface = execctx.ParamsInterface

	// CallCredits is the per-message realm->realm payment ledger; see the
	// ExecContext field of the same name.
	CallCredits = execctx.CallCredits
)

// NewCallCredits allocates the per-message call-credit ledger, seeded with
// the message send credited to payee (nothing for a zero send).
func NewCallCredits(payee string, send std.Coins) *execctx.CallCredits {
	return execctx.NewCallCredits(payee, send)
}

// GetContext returns the execution context.
// This is used to allow extending the exec context using interfaces,
// for instance when testing.
func GetContext(m *gno.Machine) ExecContext {
	return execctx.GetContext(m)
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

// NativeResolver is used by the GnoVM to determine if the given function,
// specified by its pkgPath and name, has a native implementation; and if so
// retrieve it.
func NativeResolver(pkgPath string, name gno.Name) func(*gno.Machine) {
	nt := FindNative(pkgPath, name)
	if nt == nil {
		return nil
	}
	return nt.f
}

func HasNativePkg(pkgPath string) bool {
	for _, nf := range nativeFuncs {
		if nf.gnoPkg == pkgPath {
			return true
		}
	}
	return false
}
