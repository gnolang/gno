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

// GetRealmV3a is the auto-cross-aware identity walk used by the v3a
// primitives runtime.Caller / Self / CallerN. A frame counts as a
// realm transition whenever m.Realm shifted across its push —
// detected by comparing fr.LastRealm (pre-push) against the post-push
// realm (= the higher frame's LastRealm, or m.Realm at the top).
// This catches all three borrow rules — Layer-1 (/r/-declared callee),
// Layer-2 (storage-receiver borrow), Rule 3 (closure capture-realm) —
// uniformly, because each rule writes through m.setRealm /
// setRealmAuthorityOnly in PushFrameCall.
//
// v2 primitives (CurrentRealm / PreviousRealm) keep using GetRealm
// so their established semantics remain unchanged.
func GetRealmV3a(m *gno.Machine, height int) (addr, pkgPath string) {
	var (
		ctx       = GetContext(m)
		lfr       = m.LastFrame()
		crosses   int
		postShift = m.Realm // m.Realm after the top frame's push
	)

	for i := m.NumFrames() - 1; i >= 0; i-- {
		fr := &m.Frames[i]

		if !fr.IsCall() {
			continue
		}

		shifted := !realmsEqual(postShift, fr.LastRealm)
		if !fr.WithCross && !shifted {
			lfr = fr
			postShift = fr.LastRealm
			continue
		}

		// Sanity check (only WithCross frames must have called crossing()).
		if fr.WithCross && !fr.DidCrossing {
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
		postShift = fr.LastRealm
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
				return string(ctx.OriginCaller), ""
			} else {
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

// realmsEqual reports whether two *Realm pointers refer to the same
// realm, nil-safe. Used by GetRealmV3a to detect m.Realm shifts
// across a frame's push.
func realmsEqual(a, b *gno.Realm) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.ID == b.ID
}

// CurrentRealm retrieves the current realm's address and pkgPath.
// It's not a native binding; but is used as a helper function here and
// elsewhere to clarify usage.
func CurrentRealm(m *gno.Machine) (address, pkgPath string) {
	return GetRealm(m, 0)
}
