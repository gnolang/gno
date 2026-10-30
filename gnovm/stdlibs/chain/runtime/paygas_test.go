package runtime

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/gnovm/stdlibs/internal/execctx"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store"
)

// TestPayGasLargeMaxFee: a maxFee whose product with the gas price does not fit
// int64 still commits, at the gas it buys. PayGas returns before this
// arithmetic outside a sponsored tx (gno test), so refusing such a maxFee would
// only show on chain.
func TestPayGasLargeMaxFee(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name      string
		maxFee    int64
		price     int64 // ugnot per 1000 gas
		meter     int64 // the credit window the ante set
		wantLimit int64
	}{
		// Buys far more than the window, which stays the limit.
		{"no cap", math.MaxInt64, 1, 10_000_000, 10_000_000},
		// 1e16 * 1000 overflows int64; the quotient, 1e7, is below the window.
		{"exact quotient", 1e16, 1e12, 20_000_000, 10_000_000},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			pgi := &sdk.PayGasInfo{Eligible: true}
			m := &gno.Machine{
				Context: execctx.ExecContext{
					PayGasInfo: pgi,
					GasPrice:   std.GasPrice{Gas: 1000, Price: std.Coin{Denom: "ugnot", Amount: tc.price}},
				},
				GasMeter: store.NewGasMeter(tc.meter),
			}
			X_payGas(m, "gno.land/r/test/sponsor", tc.maxFee)
			require.Equal(t, tc.maxFee, pgi.MaxFee, "PayGas did not commit")
			require.Equal(t, tc.wantLimit, m.GasMeter.Limit())
		})
	}
}
