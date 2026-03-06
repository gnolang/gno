package runtime

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs/internal/execctx"
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
	return n <= 2 && isMsgCall
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

func typedString(s string) gno.TypedValue {
	tv := gno.TypedValue{T: gno.StringType}
	tv.SetString(gno.StringValue(s))
	return tv
}
