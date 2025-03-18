package position

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	u256 "gno.land/p/gnoswap/uint256"

	"gno.land/r/gnoswap/v1/gns"
	pl "gno.land/r/gnoswap/v1/pool"
	"gno.land/r/onbloc/bar"
	"gno.land/r/onbloc/baz"
)

func TestMintInternal_PositionMint(t *testing.T) {
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
				tickLower:      -120,
				tickUpper:      120,
				amount0Desired: u256.MustFromDecimal("100000"),
				amount1Desired: u256.MustFromDecimal("100000"),
				amount0Min:     u256.MustFromDecimal("90000"),
				amount1Min:     u256.MustFromDecimal("90000"),
				caller:         alice,
				mintTo:         alice,
			},
			expectedSuccess: true,
		},
		{
			name: "zero liquidity should fail",
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
			std.TestSetRealm(adminRealm)
			TokenApprove(t, gnsPath, alice, poolAddr, maxApprove)
			poolPath := pl.GetPoolPath(tt.params.token0, tt.params.token1, tt.params.fee)
			if !pl.DoesPoolPathExist(poolPath) {
				pl.CreatePool(tt.params.token0, tt.params.token1, tt.params.fee, "79228162514264337593543950336")
			}

			std.TestSetRealm(adminRealm)
			TokenFaucet(t, tt.params.token0, alice)
			TokenFaucet(t, tt.params.token1, alice)

			std.TestSetRealm(std.NewUserRealm(alice))
			bar.Approve(poolAddr, maxApprove)
			baz.Approve(poolAddr, maxApprove)
			std.TestSetOriginCaller(alice)
			id, liquidity, amount0, amount1 := mint(tt.params)

			if tt.expectedSuccess {
				uassert.Equal(t, uint64(23), id)
				uassert.Equal(t, "16717549", liquidity.ToString())
				uassert.Equal(t, "100000", amount0.ToString())
				uassert.Equal(t, "100000", amount1.ToString())
			}
		})
	}
}
