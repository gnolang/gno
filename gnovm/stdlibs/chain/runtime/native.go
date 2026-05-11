package runtime

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs/internal/execctx"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func AssertOriginCall(m *gno.Machine) {
	if !isOriginCall(m) {
		m.Panic(typedString("invalid non-origin call"))
	}
}

func isOriginCall(m *gno.Machine) bool {
	n := m.NumFrames()
	if n == 0 {
		return false
	}
	firstPkg := m.Frames[0].LastPackage
	isMsgCall := firstPkg != nil && firstPkg.PkgPath == ""
	if !isMsgCall {
		return false
	}
	// Count only actual function call frames (excludes closures
	// and control-flow basic frames like for/range/switch).
	return m.NumCallFrames() <= 2
}

func ChainID(m *gno.Machine) string {
	return execctx.GetContext(m).ChainID
}

func ChainDomain(m *gno.Machine) string {
	return execctx.GetContext(m).ChainDomain
}

func ChainHeight(m *gno.Machine) int64 {
	return execctx.GetContext(m).Height
}

func X_originCaller(m *gno.Machine) string {
	return string(execctx.GetContext(m).OriginCaller)
}

func X_getRealm(m *gno.Machine, height int) (address, pkgPath string) {
	return execctx.GetRealm(m, height)
}

// pathRestricted is satisfied by GnoSessionAccount without importing gno.land.
// If the session account type doesn't implement this, allowPaths will be nil
// (meaning "unrestricted"), which is the correct semantic — the session has
// no path restrictions configured at the protocol level.
type pathRestricted interface{ GetAllowPaths() []string }

func X_getSessionInfo(m *gno.Machine) (pubKeyAddr string, expiresAt int64, allowPaths []string, isSession bool) {
	ctx := execctx.GetContext(m)
	if ctx.SessionAccount == nil {
		return "", 0, nil, false
	}
	da := ctx.SessionAccount
	addr := da.(std.Account).GetAddress()
	var paths []string
	if pr, ok := da.(pathRestricted); ok {
		paths = pr.GetAllowPaths()
	}
	return addr.String(), da.GetExpiresAt(), paths, true
}

func typedString(s string) gno.TypedValue {
	tv := gno.TypedValue{T: gno.StringType}
	tv.SetString(gno.StringValue(s))
	return tv
}
