package std

import (
	"fmt"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func AssertOriginCall(m *gno.Machine) {
	if !isOriginCall(m) {
		m.Panic(typedString("invalid non-origin call"))
		return
	}
}

func isOriginCall(m *gno.Machine) bool {
	n := m.NumFrames()
	if n == 0 {
		return false
	}
	firstPkg := m.Frames[0].LastPackage
	isMsgCall := firstPkg != nil && firstPkg.PkgPath == "main"
	return n <= 2 && isMsgCall
}

func ChainID(m *gno.Machine) string {
	return GetContext(m).ChainID
}

func ChainDomain(m *gno.Machine) string {
	return GetContext(m).ChainDomain
}

func ChainHeight(m *gno.Machine) int64 {
	return GetContext(m).Height
}

func X_originSend(m *gno.Machine) (denoms []string, amounts []int64) {
	os := GetContext(m).OriginSend
	return ExpandCoins(os)
}

func X_originCaller(m *gno.Machine) string {
	return string(GetContext(m).OriginCaller)
}

/* See comment in stdlibs/std/native.gno
func X_callerAt(m *gno.Machine, n int) string {
	if n <= 0 {
		m.Panic(typedString("CallerAt requires positive arg"))
		return ""
	}
	// Add 1 to n to account for the CallerAt (gno fn) frame.
	n++
	if n > m.NumFrames() {
		// NOTE: the last frame's LastPackage
		// is set to the original non-frame
		// package, so need this check.
		m.Panic(typedString("frame not found"))
		return ""
	}
	if n == m.NumFrames() {
		// This makes it consistent with OriginCaller.
		ctx := GetContext(m)
		return string(ctx.OriginCaller)
	}
	return string(m.MustPeekCallFrame(n).LastPackage.GetPkgAddr().Bech32())
}
*/

func X_getRealm(m *gno.Machine, height int) (address, pkgPath string) {
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
			m.Panic(typedString("frame not found"))
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
			m.Panic(typedString("frame not found"))
			return "", ""
		}
	default:
		panic("exec kind unspecified")
	}
}

// currentPkgPath retrieves the current package's pkgPath.
// It's not a native binding; but is used within this package to clarify usage.
func currentPkgPath(m *gno.Machine) (pkgPath string) {
	return m.MustPeekCallFrame(2).LastPackage.PkgPath
}

// currentRealm retrieves the current realm's address and pkgPath.
// It's not a native binding; but is used within this package to clarify usage.
func currentRealm(m *gno.Machine) (address, pkgPath string) {
	return X_getRealm(m, 0)
}

func X_assertCallerIsRealm(m *gno.Machine) {
	fr := &m.Frames[m.NumFrames()-2]
	if path := fr.LastPackage.PkgPath; !gno.IsRealmPath(path) {
		m.Panic(typedString("caller is not a realm"))
		return
	}
}

func typedString(s string) gno.TypedValue {
	tv := gno.TypedValue{T: gno.StringType}
	tv.SetString(gno.StringValue(s))
	return tv
}

func ExpandCoins(c std.Coins) (denoms []string, amounts []int64) {
	denoms = make([]string, len(c))
	amounts = make([]int64, len(c))
	for i, coin := range c {
		denoms[i] = coin.Denom
		amounts[i] = coin.Amount
	}
	return denoms, amounts
}

func CompactCoins(denoms []string, amounts []int64) std.Coins {
	coins := make(std.Coins, len(denoms))
	for i := range coins {
		coins[i] = std.Coin{Denom: denoms[i], Amount: amounts[i]}
	}
	return coins
}
