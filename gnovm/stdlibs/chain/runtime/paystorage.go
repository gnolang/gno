package runtime

import (
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs/internal/execctx"
)

// X_payStorage records pkgPath's storage commitment. PayStorage has already
// checked that pkgPath came from the current realm's capability.
func X_payStorage(m *gno.Machine, pkgPath string, maxDeposit int64) {
	if maxDeposit <= 0 {
		m.Panic(typedString("PayStorage: maxDeposit must be positive"))
		return
	}
	// Excludes packages, MsgRun's ephemeral realm and sub-realm tokens.
	if !gno.IsRealmPath(pkgPath) {
		m.Panic(typedString("PayStorage: rlm is not a realm"))
		return
	}

	psi := execctx.GetContext(m).PayStorageInfo
	if psi == nil || psi.Entry != pkgPath {
		// Not a sponsored tx, or not the realm this message calls: storage
		// stays on the caller.
		return
	}
	if psi.MaxDeposit > 0 {
		m.Panic(typedString("PayStorage: already called in this transaction"))
		return
	}

	psi.RealmPkgPath = pkgPath
	psi.RealmAddr = gno.DerivePkgCryptoAddr(pkgPath)
	psi.MaxDeposit = maxDeposit
}
