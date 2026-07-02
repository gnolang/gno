package execctx

import (
	"fmt"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

func GetRealm(m *gno.Machine, height int) (addr, pkgPath string) {
	// NOTE: keep in sync with test/stdlibs/std.getRealm

	// Identity-chain walk (presented identities): start from the
	// innermost crossing frame's Cur and follow .prev height times.
	// Each prev slot holds the realm value presented at that crossing —
	// cross(rlm) stores rlm verbatim, including sub-realm tokens, which
	// is what keeps unsafe.{Current,Previous}Realm in agreement with
	// cur/cur.Previous() for sub-identities. For ordinary crosses the
	// presented identity coincides with the crossed-from context, so
	// answers match the legacy context-chain walk below at every
	// height. The origin-shaped terminal (own prev truly-nil) and
	// stacks with no captured Cur fall through to the legacy walk,
	// which serves the height==crosses and crosses+1 boundary answers
	// (stage-dependent) unchanged.
	if cur, ok := innermostCrossingCur(m); ok {
		v := cur
		for h := 0; h <= height; h++ {
			a, p, prev, ok := gno.RealmValueParts(v)
			if !ok || prev.T == nil {
				break // terminal or non-realm shape: legacy fallback
			}
			if h == height {
				return a, p
			}
			v = prev
		}
	}

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

// innermostCrossingCur returns the topmost crossing frame's captured
// Cur, if any. Crossing functions entered without cross inherit their
// caller's Cur (pointer-identical), so the first crossing frame with a
// captured Cur anchors the presented-identity chain.
func innermostCrossingCur(m *gno.Machine) (gno.TypedValue, bool) {
	for i := m.NumFrames() - 1; i >= 0; i-- {
		fr := &m.Frames[i]
		if !fr.IsCall() {
			continue
		}
		if !(fr.WithCross || fr.DidCrossing) {
			continue
		}
		if fr.Cur.T == nil {
			continue
		}
		return fr.Cur, true
	}
	return gno.TypedValue{}, false
}

// CurrentRealm retrieves the current realm's address and pkgPath.
// It's not a native binding; but is used as a helper function here and
// elsewhere to clarify usage.
func CurrentRealm(m *gno.Machine) (address, pkgPath string) {
	return GetRealm(m, 0)
}
