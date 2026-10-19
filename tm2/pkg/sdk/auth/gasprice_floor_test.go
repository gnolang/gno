package auth

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/require"
)

// The initial gas price is the floor the block price decays to. Both it and the
// block price are ratios -- Price per Gas units -- and governance sets the
// initial one on its own, so it can name a ratio over a different number of gas
// units than the stored price uses.
//
// Comparing the two amounts alone compares numerators over unequal
// denominators. A floor of 10ugnot per 1 gas then reads as 10ugnot per 1000
// gas, and the price decays a thousandfold under the floor it is supposed to
// stop at.
//
// Asserted with IsGTE, the ratio comparison every other gas price check uses.
func TestCalcBlockGasPriceFloorHoldsAcrossGasUnits(t *testing.T) {
	t.Parallel()

	gk := GasPriceKeeper{}
	const maxGas = int64(3_000_000_000)

	// The floor: 10ugnot for every 1 gas.
	params := Params{
		TargetGasRatio:            70,
		GasPricesChangeCompressor: 10,
		InitialGasPrice:           std.GasPrice{Gas: 1, Price: std.Coin{Amount: 10, Denom: "ugnot"}},
	}

	atOrAboveFloor := func(t *testing.T, gp std.GasPrice) {
		t.Helper()
		ok, err := gp.IsGTE(params.InitialGasPrice)
		require.NoError(t, err)
		require.True(t, ok,
			"%d/%dgas is below the floor of %d/%dgas",
			gp.Price.Amount, gp.Gas,
			params.InitialGasPrice.Price.Amount, params.InitialGasPrice.Gas)
	}

	cases := []struct {
		name string
		last std.GasPrice
	}{
		{
			// Already under the floor: 0.1ugnot per gas against a floor of 10.
			"stored price below the floor",
			std.GasPrice{Gas: 1000, Price: std.Coin{Amount: 100, Denom: "ugnot"}},
		},
		{
			// Above the floor: 100ugnot per gas. It may fall, but only to 10.
			"stored price above the floor",
			std.GasPrice{Gas: 1000, Price: std.Coin{Amount: 100_000, Denom: "ugnot"}},
		},
		{
			// Same ratio as the floor, written over different units.
			"stored price equal to the floor",
			std.GasPrice{Gas: 1000, Price: std.Coin{Amount: 10_000, Denom: "ugnot"}},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			atOrAboveFloor(t, gk.calcBlockGasPrice(tt.last, 0, maxGas, params))

			// Idle blocks push the price down every time. However many of them
			// arrive, it must stop at the floor.
			next := tt.last
			for range 200 {
				next = gk.calcBlockGasPrice(next, 0, maxGas, params)
			}
			atOrAboveFloor(t, next)
		})
	}
}
