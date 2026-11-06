package runtime

import (
	"math"
	"math/bits"
	"strings"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs/internal/execctx"
	"github.com/gnolang/gno/tm2/pkg/std"
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
	limit := gasBought(maxFee, gp)

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

// gasBought is the gas maxFee buys at gp (both positive). maxFee * gp.Gas is
// computed in 128 bits, so it may exceed int64, and a quotient too large for
// int64 saturates at MaxInt64, above any meter limit. Refusing either would
// only show on chain, since PayGas returns before this outside a sponsored tx,
// as under gno test.
func gasBought(maxFee int64, gp std.GasPrice) int64 {
	hi, lo := bits.Mul64(uint64(maxFee), uint64(gp.Gas))
	price := uint64(gp.Price.Amount)
	if hi >= price {
		return math.MaxInt64
	}
	q, _ := bits.Div64(hi, lo, price)
	return int64(min(q, math.MaxInt64))
}
