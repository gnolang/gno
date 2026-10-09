package runtime

import (
	"strings"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs/internal/execctx"
	"github.com/gnolang/gno/tm2/pkg/overflow"
)

// X_payGas records pkgPath's gas commitment. PayGas has already checked that
// pkgPath came from the current realm's capability.
func X_payGas(m *gno.Machine, pkgPath string, maxFee int64) {
	if maxFee <= 0 {
		m.Panic(typedString("PayGas: maxFee must be positive"))
		return
	}
	// Excludes packages, MsgRun's ephemeral realm and sub-realm tokens. A
	// sub-realm token ("host#sub") is refused before the regexp: a failing
	// match can backtrack past what this native's gas row, fitted on
	// matches, charges.
	if strings.IndexByte(pkgPath, '#') >= 0 || !gno.IsRealmPath(pkgPath) {
		m.Panic(typedString("PayGas: rlm is not a realm"))
		return
	}

	ctx := execctx.GetContext(m)
	pgi := ctx.PayGasInfo // shared with the SDK context, so writes reach settlement
	if pgi == nil || !pgi.Eligible {
		// Not a sponsored tx: the signer pays its own fee (or this is a query
		// or a test), so there is nothing to commit to.
		return
	}
	if pgi.MaxFee > 0 {
		m.Panic(typedString("PayGas: already called in this transaction"))
		return
	}

	gp := ctx.GasPrice
	if gp.Gas <= 0 || gp.Price.Amount <= 0 || gp.Price.Denom == "" {
		m.Panic(typedString("PayGas: gas price not set"))
		return
	}
	product, ok := overflow.Mul(maxFee, gp.Gas)
	if !ok {
		m.Panic(typedString("PayGas: maxFee * gas price overflows"))
		return
	}
	limit := product / gp.Price.Amount

	// Tighten the meter to what maxFee buys, never past the credit window the
	// ante granted. An infinite meter (source-gas replay) reports Limit() == 0
	// and is left alone: it has no cap to tighten, and enforcing today's budget
	// could refuse to replay a historically successful tx.
	if curLimit := m.GasMeter.Limit(); curLimit > 0 {
		limit = min(limit, curLimit)
		if limit < m.GasMeter.GasConsumed() {
			m.Panic(typedString("PayGas: maxFee budget exceeded at current gas price"))
			return
		}
		m.GasMeter.SetLimit(limit)
	}

	pgi.RealmAddr = gno.DerivePkgCryptoAddr(pkgPath)
	pgi.MaxFee = maxFee
}
