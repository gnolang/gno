package std

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/bech32"
	"github.com/gnolang/gno/tm2/pkg/crypto"
)

func AssertOriginCall(m *gno.Machine) {
	if !IsOriginCall(m) {
		m.Panic(typedString("invalid non-origin call"))
	}
}

func IsOriginCall(m *gno.Machine) bool {
	return len(m.Frames) == 2
}

func CurrentRealmPath(m *gno.Machine) string {
	if m.Realm != nil {
		return m.Realm.Path
	}
	return ""
}

func GetChainID(m *gno.Machine) string {
	return m.Context.(ExecContext).ChainID
}

func GetHeight(m *gno.Machine) int64 {
	return m.Context.(ExecContext).Height
}

func X_origSend(m *gno.Machine) (denoms []string, amounts []int64) {
	os := m.Context.(ExecContext).OrigSend
	denoms = make([]string, len(os))
	amounts = make([]int64, len(os))
	for i, coin := range os {
		denoms[i] = coin.Denom
		amounts[i] = coin.Amount
	}
	return denoms, amounts
}

func X_origCaller(m *gno.Machine) string {
	return string(m.Context.(ExecContext).OrigCaller)
}

func X_origPkgAddr(m *gno.Machine) string {
	return string(m.Context.(ExecContext).OrigPkgAddr)
}

func X_callerAt(m *gno.Machine, n int) string {
	if n <= 0 {
		m.Panic(typedString("GetCallerAt requires positive arg"))
		return ""
	}
	if n > m.NumFrames() {
		// NOTE: the last frame's LastPackage
		// is set to the original non-frame
		// package, so need this check.
		m.Panic(typedString("frame not found"))
		return ""
	}
	if n == m.NumFrames() {
		// This makes it consistent with GetOrigCaller.
		ctx := m.Context.(ExecContext)
		return string(ctx.OrigCaller)
	}
	return string(m.LastCallFrame(n).LastPackage.GetPkgAddr().Bech32())
}

func X_getRealm(m *gno.Machine, height int) (address string, pkgPath string) {
	var (
		ctx           = m.Context.(ExecContext)
		currentCaller crypto.Bech32Address
		// Keeps track of the number of times currentCaller
		// has changed.
		changes int
	)

	for i := m.NumFrames() - 1; i > 0; i-- {
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

	// Fallback case: return OrigCaller.
	return string(ctx.OrigCaller), ""
}

func X_derivePkgAddr(pkgPath string) string {
	return string(gno.DerivePkgAddr(pkgPath).Bech32())
}

func X_encodeBech32(prefix string, bytes [20]byte) string {
	b32, err := bech32.ConvertAndEncode(prefix, bytes[:])
	if err != nil {
		panic(err) // should not happen
	}
	return b32
}

func X_decodeBech32(addr string) (prefix string, bytes [20]byte, ok bool) {
	prefix, bz, err := bech32.Decode(addr)
	if err != nil || len(bz) != 20 {
		return "", [20]byte{}, false
	}
	return prefix, [20]byte(bz), true
}

func typedString(s gno.StringValue) gno.TypedValue {
	tv := gno.TypedValue{T: gno.StringType}
	tv.SetString(s)
	return tv
}
