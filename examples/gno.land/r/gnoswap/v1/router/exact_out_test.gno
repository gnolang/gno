package router

import (
	"std"
	"testing"
	"time"

	"gno.land/p/demo/uassert"
	"gno.land/p/gnoswap/consts"

	pl "gno.land/r/gnoswap/v1/pool"
	pn "gno.land/r/gnoswap/v1/position"

	"gno.land/r/gnoswap/v1/gns"
	"gno.land/r/onbloc/bar"
	"gno.land/r/onbloc/foo"
)

func TestExactOutSwapOperation(t *testing.T) {
	std.TestSetRealm(routerRealm)
	setupTestPool(t)

	t.Run("validate exact out params", func(t *testing.T) {
		op := NewExactOutSwapOperation(ExactOutParams{
			BaseSwapParams: BaseSwapParams{
				InputToken:  barPath,
				OutputToken: fooPath,
				RouteArr:    "gno.land/r/onbloc/foo:gno.land/r/onbloc/bar:3000",
				QuoteArr:    "100",
			},
			AmountOut:   "100",
			AmountInMax: "150",
		})

		err := op.Validate()
		uassert.Equal(t, err, nil)

		// Invalid amount out
		invalidOp := NewExactOutSwapOperation(ExactOutParams{
			BaseSwapParams: BaseSwapParams{
				InputToken:  barPath,
				OutputToken: fooPath,
				RouteArr:    "gno.land/r/onbloc/foo:gno.land/r/onbloc/bar:3000",
				QuoteArr:    "100",
			},
			AmountOut:   "-100", // negative amount
			AmountInMax: "150",
		})

		err = invalidOp.Validate()
		uassert.NotEqual(t, err, nil)
	})

	t.Run("exact out swap with invalid route", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for invalid route")
			}
		}()

		ExactOutSwapRoute(
			barPath,
			fooPath,
			"3",
			"invalid:route:format", // invalid route format
			"100",
			"5",
			time.Now().Add(time.Hour).Unix(),
			"",
		)
	})

	t.Run("exact out swap with expired deadline", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for expired deadline")
			}
		}()

		ExactOutSwapRoute(
			barPath,
			fooPath,
			"3",
			"gno.land/r/onbloc/foo:gno.land/r/onbloc/bar:3000",
			"100",
			"5",
			time.Now().Add(-time.Hour).Unix(), // expired deadline
			"",
		)
	})
}

// Helper function to setup test pool
func setupTestPool(t *testing.T) {
	t.Helper()
	std.TestSetRealm(adminRealm)
	gns.Approve(poolAddr, pl.GetPoolCreationFee())

	foo.Approve(poolAddr, consts.UINT64_MAX)
	bar.Approve(poolAddr, consts.UINT64_MAX)

	pl.CreatePool(barPath, fooPath, 3000, "79228162514264337593543950336") // encodePriceSqrt(1, 1)

	_, _, _, _ = pn.Mint(
		barPath,
		fooPath,
		3000,
		minTick,
		maxTick,
		"1000000",
		"1000000",
		"0",
		"0",
		max_timeout,
		adminAddr,
		adminAddr,
		"",
	)

	pool := pl.GetPool(barPath, fooPath, FEE_MEDIUM)
	uassert.Equal(t, pool.Liquidity().ToString(), "1000000")
}
