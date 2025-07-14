package position

import (
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	u256 "gno.land/p/gnoswap/uint256"

	pl "gno.land/r/gnoswap/v1/pool"
	"gno.land/r/onbloc/bar"
	"gno.land/r/onbloc/baz"
)

func TestMintInternal_PositionMint(t *testing.T) {
	tests := []struct {
		name              string
		params            MintParams
		expectPanic       bool
		expectedMsg       string
		expectedId        uint64
		expectedLiquidity string
		expectedAmount0   string
		expectedAmount1   string
	}{
		{
			name: "mint is success by successful mint",
			params: MintParams{
				token0:         "gno.land/r/onbloc/bar",
				token1:         "gno.land/r/onbloc/baz",
				fee:            3000,
				tickLower:      -120,
				tickUpper:      120,
				amount0Desired: u256.MustFromDecimal("100000"),
				amount1Desired: u256.MustFromDecimal("100000"),
				amount0Min:     u256.MustFromDecimal("90000"),
				amount1Min:     u256.MustFromDecimal("90000"),
				caller:         alice,
				mintTo:         alice,
			},
			expectPanic:       false,
			expectedId:        1,
			expectedLiquidity: "16717549",
			expectedAmount0:   "100000",
			expectedAmount1:   "100000",
		},
		{
			name: "mint is failed by zero liquidity",
			params: MintParams{
				token0:         "gno.land/r/onbloc/bar",
				token1:         "gno.land/r/onbloc/baz",
				fee:            3000,
				tickLower:      -120,
				tickUpper:      120,
				amount0Desired: u256.MustFromDecimal("0"),
				amount1Desired: u256.MustFromDecimal("0"),
				amount0Min:     u256.MustFromDecimal("0"),
				amount1Min:     u256.MustFromDecimal("0"),
				caller:         testutils.TestAddress("caller"),
				mintTo:         testutils.TestAddress("recipient"),
			},
			expectPanic: true,
			expectedMsg: "[GNOSWAP-POOL-010] zero liquidity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initPositionTest(t)

			testing.SetRealm(adminRealm)
			pl.SetPoolCreationFeeByAdmin(cross, 0)
			CreatePool(t, tt.params.token0, tt.params.token1, tt.params.fee, "79228162514264337593543950336", adminAddr)

			TokenFaucet(t, tt.params.token0, alice)
			TokenFaucet(t, tt.params.token1, alice)

			testing.SetRealm(posRealm)
			TokenApprove(t, tt.params.token0, alice, poolAddr, maxApprove)
			TokenApprove(t, tt.params.token1, alice, poolAddr, maxApprove)

			bar.Approve(cross, poolAddr, maxApprove)
			baz.Approve(cross, poolAddr, maxApprove)

			if tt.expectPanic {
				uassert.AbortsWithMessage(t, tt.expectedMsg, func() {
					mint(tt.params)
				})
			} else {
				id, liquidity, amount0, amount1 := mint(tt.params)
				uassert.Equal(t, tt.expectedId, id)
				uassert.Equal(t, tt.expectedLiquidity, liquidity.ToString())
				uassert.Equal(t, tt.expectedAmount0, amount0.ToString())
				uassert.Equal(t, tt.expectedAmount1, amount1.ToString())
			}
		})
	}
}
