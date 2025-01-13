package std

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func AssertOriginCall(m *gno.Machine) {
	if !IsOriginCall(m) {
		m.Panic(typedString("invalid non-origin call"))
	}
}

func IsOriginCall(m *gno.Machine) bool {
	n := m.NumFrames()
	if n == 0 {
		return false
	}
	firstPkg := m.Frames[0].LastPackage
	isMsgCall := firstPkg != nil && firstPkg.PkgPath == "main"
	return n <= 2 && isMsgCall
}

func GetChainID(m *gno.Machine) string {
	return GetContext(m).ChainID
}

func GetChainDomain(m *gno.Machine) string {
	return GetContext(m).ChainDomain
}

func GetHeight(m *gno.Machine) int64 {
	return GetContext(m).Height
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

func X_originSend(m *gno.Machine) (denoms []string, amounts []int64) {
	os := GetContext(m).OriginSend
	return ExpandCoins(os)
}

func X_originCaller(m *gno.Machine) string {
	return string(GetContext(m).OriginCaller)
}

func X_originPkgAddr(m *gno.Machine) string {
	return string(GetContext(m).OriginPkgAddr)
}

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
	return string(m.MustLastCallFrame(n).LastPackage.GetPkgAddr().Bech32())
}

func X_getRealm(m *gno.Machine, height int) (address, pkgPath string) {
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

// currentRealm retrieves the current realm's address and pkgPath.
// It's not a native binding; but is used within this package to clarify usage.
func currentRealm(m *gno.Machine) (address, pkgPath string) {
	return X_getRealm(m, 0)
}

func X_assertCallerIsRealm(m *gno.Machine) {
	frame := m.Frames[m.NumFrames()-2]
	if path := frame.LastPackage.PkgPath; !gno.IsRealmPath(path) {
		m.Panic(typedString("caller is not a realm"))
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
