package position

import (
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	u256 "gno.land/p/gnoswap/uint256"
)

func TestMintInternal(t *testing.T) {
	tests := []struct {
		name            string
		params          MintParams
		expectPanic     bool
		expectedSuccess bool
	}{
		{
			name: "successful mint",
			params: MintParams{
				token0:         "gno.land/r/onbloc/bar",
				token1:         "gno.land/r/onbloc/baz",
				fee:            3000,
				tickLower:      -100,
				tickUpper:      100,
				amount0Desired: u256.MustFromDecimal("100000"),
				amount1Desired: u256.MustFromDecimal("100000"),
				amount0Min:     u256.MustFromDecimal("90000"),
				amount1Min:     u256.MustFromDecimal("90000"),
				caller:         testutils.TestAddress("caller"),
				mintTo:         testutils.TestAddress("recipient"),
			},
			expectedSuccess: true,
		},
		{
			name: "zero liquidity should fail",
			params: MintParams{
				token0:         "gno.land/r/onbloc/bar",
				token1:         "gno.land/r/onbloc/baz",
				fee:            3000,
				tickLower:      -100,
				tickUpper:      100,
				amount0Desired: u256.MustFromDecimal("0"),
				amount1Desired: u256.MustFromDecimal("0"),
				amount0Min:     u256.MustFromDecimal("0"),
				amount1Min:     u256.MustFromDecimal("0"),
				caller:         testutils.TestAddress("caller"),
				mintTo:         testutils.TestAddress("recipient"),
			},
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("expected panic but got none")
					}
				}()
			}

			id, liquidity, amount0, amount1 := mint(tt.params)

			if tt.expectedSuccess {
				uassert.Equal(t, id, "Position ID should be generated")
				uassert.Equal(t, amount0, "Amount0 should not be nil")
				uassert.Equal(t, liquidity, "Liquidity should not be nil")
				uassert.Equal(t, amount1, "Amount1 should not be nil")
			}
		})
	}
}
