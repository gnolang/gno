package router

import (
	"std"
	"strings"
	"testing"
	"time"

	"gno.land/p/demo/uassert"

	pl "gno.land/r/gnoswap/v1/pool"
	pn "gno.land/r/gnoswap/v1/position"

	"gno.land/r/gnoswap/v1/gns"
	"gno.land/r/onbloc/bar"
	"gno.land/r/onbloc/foo"

	u256 "gno.land/p/gnoswap/uint256"
)

func TestExactOutSwapOperation(t *testing.T) {
	testing.SetRealm(routerRealm)
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
		uassert.AbortsWithMessage(t, "invalid fee: format", func() {
			ExactOutSwapRoute(
				cross,
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
	})

	t.Run("exact out swap with expired deadline", func(t *testing.T) {
		uassert.AbortsWithMessage(t, "[GNOSWAP-ROUTER-004] transaction expired || transaction too old, now(1234567890) > deadline(1234564290)", func() {
			ExactOutSwapRoute(
				cross,
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
	})
}

func TestExactOutSwapBoundary(t *testing.T) {
	t.Skip("run this test separately")
	testing.SetRealm(routerRealm)
	setupTestPool(t)

	testing.SetRealm(std.NewUserRealm(adminAddr))
	TokenFaucet2(t, fooPath, routerAddr)
	TokenFaucet2(t, barPath, routerAddr)

	testing.SetRealm(std.NewUserRealm(routerAddr))
	foo.Approve(cross, poolAddr, maxApprove)
	bar.Approve(cross, poolAddr, maxApprove)

	// create position
	// a. Position 1: (-50%+50%)
	pn.Mint(
		cross,
		barPath,
		fooPath,
		3000,
		-6960,
		6960,
		"3000000000",
		"3000000000",
		"0",
		"0",
		max_timeout,
		routerAddr,
		routerAddr,
		"",
	)

	// b. Position 2: (-10%+10%)
	pn.Mint(
		cross,
		barPath,
		fooPath,
		3000,
		-1080,
		960,
		"1000000000",
		"1000000000",
		"0",
		"0",
		max_timeout,
		routerAddr,
		routerAddr,
		"",
	)

	// c. Position 3: (full range)
	pn.Mint(
		cross,
		barPath,
		fooPath,
		3000,
		-887220,
		887220,
		"5000000000",
		"5000000000",
		"0",
		"0",
		max_timeout,
		routerAddr,
		routerAddr,
		"",
	)

	// d. Position 4: (lower out-range)
	pn.Mint(
		cross,
		barPath,
		fooPath,
		3000,
		-44280,
		-19080,
		"1000000000",
		"1000000",
		"0",
		"0",
		max_timeout,
		routerAddr,
		routerAddr,
		"",
	)

	// e. Position 5: (upper out-range)
	pn.Mint(
		cross,
		barPath,
		fooPath,
		3000,
		62160,
		69060,
		"1000000",
		"1000000000",
		"0",
		"0",
		max_timeout,
		routerAddr,
		routerAddr,
		"",
	)

	t.Run("successful exact out swap at boundary", func(t *testing.T) {
		amountIn, amountOut := ExactOutSwapRoute(
			cross,
			barPath,
			fooPath,
			"1000000",
			"gno.land/r/onbloc/foo:gno.land/r/onbloc/bar:3000",
			"100",
			"1100000",
			time.Now().Add(time.Hour).Unix(),
			"",
		)

		uassert.NotEqual(t, amountIn, "0")
		uassert.NotEqual(t, amountOut, "0")
	})

	t.Run("failed exact out swap at boundary", func(t *testing.T) {
		poolPath := pl.GetPoolPath(barPath, fooPath, 3000)
		balance0 := pl.PoolGetBalanceToken0(poolPath)
		balance1 := pl.PoolGetBalanceToken1(poolPath)

		token0Balance := u256.MustFromDecimal(balance0)
		token1Balance := u256.MustFromDecimal(balance1)
		println("token1Balance", token1Balance.ToString())

		overflowAmount := u256.Zero().Add(token0Balance, u256.One()).ToString()

		defer func() {
			if r := recover(); r == nil {
				if !strings.Contains(r.(string), "overflow") {
					t.Error("Expected panic for overflow at boundary")
				}
			}
		}()

		// request amount that slightly exceeds the pool balance
		ExactOutSwapRoute(
			cross,
			barPath,
			fooPath,
			overflowAmount,
			"gno.land/r/onbloc/foo:gno.land/r/onbloc/bar:3000",
			"100",
			"50101690",
			time.Now().Add(time.Hour).Unix(),
			"",
		)
	})
}

func setupTestPool(t *testing.T) {
	t.Helper()
	testing.SetRealm(adminRealm)
	gns.Approve(cross, poolAddr, pl.GetPoolCreationFee())

	foo.Approve(cross, poolAddr, maxApprove)
	bar.Approve(cross, poolAddr, maxApprove)

	pl.CreatePool(cross, barPath, fooPath, 3000, "79228162514264337593543950336") // encodePriceSqrt(1, 1)

	_, _, _, _ = pn.Mint(
		cross,
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
