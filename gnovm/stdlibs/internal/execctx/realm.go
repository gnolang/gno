package execctx

import (
	"fmt"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func GetRealm(m *gno.Machine, height int) (addr, pkgPath string) {
	// NOTE: keep in sync with test/stdlibs/std.getRealm

	var (
		ctx     = GetContext(m)
		lfr     = m.LastFrame() // last call frame
		crosses int             // track realm crosses
	)

	for i := m.NumFrames() - 1; i >= 0; i-- {
		fr := &m.Frames[i]

		// Skip over (non-realm) non-crosses.
		if !fr.IsCall() {
			continue
		}
		if !fr.WithCross {
			lfr = fr
			continue
		}

		// Sanity check
		if !fr.DidCrossing {
			panic(fmt.Sprintf(
				"call to cross(fn) did not call crossing : %s",
				fr.Func.String()))
		}

		crosses++
		if crosses > height {
			currlm := lfr.LastRealm
			caller, rlmPath := gno.DerivePkgBech32Addr(currlm.Path), currlm.Path
			return string(caller), rlmPath
		}
		lfr = fr
	}

	switch m.Stage {
	case gno.StageAdd:
		switch height {
		case crosses:
			fr := m.Frames[0]
			path := fr.LastPackage.PkgPath
			return string(gno.DerivePkgBech32Addr(path)), path
		case crosses + 1:
			return string(ctx.OriginCaller), ""
		default:
			m.PanicString("frame not found")
			return "", ""
		}
	case gno.StageRun:
		switch height {
		case crosses:
			fr := m.Frames[0]
			path := fr.LastPackage.PkgPath
			if path == "" {
				// e.g. MsgCall, cross-call a public realm function
				return string(ctx.OriginCaller), ""
			} else {
				// e.g. MsgRun, non-cross-call main()
				return string(gno.DerivePkgBech32Addr(path)), path
			}
		case crosses + 1:
			return string(ctx.OriginCaller), ""
		default:
			m.PanicString("frame not found")
			return "", ""
		}
	default:
		panic("exec kind unspecified")
	}
}

// CurrentRealm retrieves the current realm's address and pkgPath.
// It's not a native binding; but is used as a helper function here and
// elsewhere to clarify usage.
func CurrentRealm(m *gno.Machine) (address, pkgPath string) {
	return GetRealm(m, 0)
}
