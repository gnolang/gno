package std

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/bech32"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
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

// getPrevFunctionNameFromTarget returns the last called function name (identifier) from the call stack.
func getPrevFunctionNameFromTarget(m *gno.Machine, targetFunc string) string {
	targetIndex := findTargetFuncIndex(m, targetFunc)
	if targetIndex == -1 {
		return ""
	}
	return findPrevFuncName(m, targetIndex)
}

// findTargetFuncIndex finds and returns the index of the target function in the call stack.
func findTargetFuncIndex(m *gno.Machine, targetFunc string) int {
	for i := len(m.Frames) - 1; i >= 0; i-- {
		currFunc := m.Frames[i].Func
		if currFunc != nil && currFunc.Name == gno.Name(targetFunc) {
			return i
		}
	}
	return -1
}

// findPrevFuncName returns the function name before the given index in the call stack.
func findPrevFuncName(m *gno.Machine, targetIndex int) string {
	for i := targetIndex - 1; i >= 0; i-- {
		currFunc := m.Frames[i].Func
		if currFunc != nil {
			return string(currFunc.Name)
		}
	}
	panic("function name not found")
}

func X_origSend(m *gno.Machine) (denoms []string, amounts []int64) {
	os := m.Context.(ExecContext).OrigSend
	return ExpandCoins(os)
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
	// Add 1 to n to account for the GetCallerAt (gno fn) frame.
	n++
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
	return string(m.MustLastCallFrame(n).LastPackage.GetPkgAddr().Bech32())
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
