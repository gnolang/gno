package execctx

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

func GetRealm(m *gno.Machine, height int) (address, pkgPath string) {
	// NOTE: keep in sync with test/stdlibs/std.getRealm

	var (
		ctx           = GetContext(m)
		currentCaller crypto.Bech32Address
		// Keeps track of the number of times currentCaller
		// has changed.
		changes int
	)

	for i := m.NumFrames() - 1; i >= 0; i-- {
		fr := m.Frames[i]
		if fr.LastPackage == nil || !fr.LastPackage.IsRealm() {
			continue
		}

		// LastPackage is a realm. Get caller and pkgPath, and compare against
		// current* values.
		caller := fr.LastPackage.GetPkgAddr().Bech32()
		pkgPath := fr.LastPackage.PkgPath
		if caller != currentCaller {
			if changes == height {
				return string(caller), pkgPath
			}
			currentCaller = caller
			changes++
		}
	}

	// Fallback case: return OriginCaller.
	return string(ctx.OriginCaller), ""
}

// CurrentRealm retrieves the current realm's address and pkgPath.
// It's not a native binding; but is used as a helper function here and
// elsewhere to clarify usage.
func CurrentRealm(m *gno.Machine) (address, pkgPath string) {
	return GetRealm(m, 0)
}
