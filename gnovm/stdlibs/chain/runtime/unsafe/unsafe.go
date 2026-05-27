// Package unsafe quarantines the stack-walking and tx-origin
// primitives. See unsafe.gno for the security rationale.
package unsafe

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs/internal/execctx"
)

func X_getRealm(m *gno.Machine, height int) (address, pkgPath string) {
	return execctx.GetRealm(m, height)
}

func X_originCaller(m *gno.Machine) string {
	return string(execctx.GetContext(m).OriginCaller)
}

// X_originSend mirrors the implementation that lived at
// gnovm/stdlibs/chain/banker.X_originSend. The OriginSend envelope is
// a tx-origin primitive (see Class-2 in docs/resources/gno-security.md),
// so its native binding belongs with the other tx-origin natives.
func X_originSend(m *gno.Machine) (denoms []string, amounts []int64) {
	os := execctx.GetContext(m).OriginSend
	denoms = make([]string, len(os))
	amounts = make([]int64, len(os))
	for i, coin := range os {
		denoms[i] = coin.Denom
		amounts[i] = coin.Amount
	}
	return denoms, amounts
}
