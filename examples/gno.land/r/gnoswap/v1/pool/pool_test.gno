package pool

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"

	"gno.land/p/gnoswap/consts"

	"gno.land/r/onbloc/bar"
	"gno.land/r/onbloc/baz"
)

func TestMint(t *testing.T) {
	token0Path := "test_token0"
	token1Path := "test_token1"
	fee := uint32(3000)
	recipient := testutils.TestAddress("recipient")
	tickLower := int32(-100)
	tickUpper := int32(100)
	liquidityAmount := "100000"

	t.Run("unauthorized caller mint should fail", func(t *testing.T) {
		unauthorized := testutils.TestAddress("unauthorized")
		defer func() {
			if r := recover(); r == nil {
				t.Error("unauthorized caller mint should fail")
			}
		}()

		Mint(token0Path, token1Path, fee, recipient, tickLower, tickUpper, liquidityAmount, unauthorized)
	})

	t.Run("mint with 0 liquidity should fail", func(t *testing.T) {
		authorized := consts.POSITION_ADDR
		defer func() {
			if r := recover(); r == nil {
				t.Error("mint with 0 liquidity should fail")
			}
		}()

		Mint(token0Path, token1Path, fee, recipient, tickLower, tickUpper, "0", authorized)
	})

	t.Run("successful mint", func(t *testing.T) {
		// create pool first
		std.TestSetRealm(adminRealm)
		SetPoolCreationFeeByAdmin(0)
		CreatePool(
			"gno.land/r/onbloc/bar",
			"gno.land/r/onbloc/baz",
			fee100,
			"79228162514264337593543950336",
		)

		// approve enough token to mint position
		bar.Approve(consts.POOL_ADDR, consts.UINT64_MAX)
		baz.Approve(consts.POOL_ADDR, consts.UINT64_MAX)

		std.TestSetRealm(std.NewCodeRealm(consts.POSITION_PATH))
		mintAmount0, mintAmount1 := Mint(
			"gno.land/r/onbloc/bar",
			"gno.land/r/onbloc/baz",
			fee100,
			consts.POSITION_ADDR,
			int32(-100),
			int32(100),
			"10000000000",
			consts.ADMIN,
		)
		uassert.Equal(t, mintAmount0, "49872721")
		uassert.Equal(t, mintAmount1, "49872721")
	})
}

func TestBurn(t *testing.T) {
	tests := []struct {
		name            string
		liquidityAmount string
		tickLower       int32
		tickUpper       int32
		expectedAmount0 string
		expectedAmount1 string
		expectPanic     bool
	}{
		{
			name:            "successful burn",
			liquidityAmount: "10000000000", // burn all liquidity
			tickLower:       -100,
			tickUpper:       100,
			expectedAmount0: "49872720",
			expectedAmount1: "49872720",
		},
		{
			name:            "zero liquidity",
			liquidityAmount: "0",
			tickLower:       -100,
			tickUpper:       100,
			expectPanic:     true,
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

			std.TestSetRealm(std.NewCodeRealm(consts.POSITION_PATH))
			amount0, amount1 := Burn(
				barPath,
				bazPath,
				fee100,
				tt.tickLower,
				tt.tickUpper,
				tt.liquidityAmount,
			)

			if !tt.expectPanic {
				if amount0 != tt.expectedAmount0 {
					t.Errorf("expected amount0 %s, got %s", tt.expectedAmount0, amount0)
				}
				if amount1 != tt.expectedAmount1 {
					t.Errorf("expected amount1 %s, got %s", tt.expectedAmount1, amount1)
				}

				pool := GetPool(barPath, bazPath, fee100)
				posKey := getPositionKey(consts.POSITION_ADDR, tt.tickLower, tt.tickUpper)
				newPosition := pool.mustGetPosition(posKey)
				if newPosition.tokensOwed0.IsZero() {
					t.Error("expected tokensOwed0 to be updated")
				}
				if newPosition.tokensOwed1.IsZero() {
					t.Error("expected tokensOwed1 to be updated")
				}
			}
		})
	}
}

func TestSetFeeProtocolInternal(t *testing.T) {
	tests := []struct {
		name         string
		feeProtocol0 uint8
		feeProtocol1 uint8
		eventName    string
	}{
		{
			name:         "set fee protocol by admin",
			feeProtocol0: 4,
			feeProtocol1: 5,
			eventName:    "SetFeeProtocolByAdmin",
		},
	}

	for _, tt := range tests {
		t.Run("set fee protocol by admin", func(t *testing.T) {
			InitialisePoolTest(t)
			pool := GetPool(wugnotPath, gnsPath, fee3000)
			SetFeeProtocolByAdmin(tt.feeProtocol0, tt.feeProtocol1)
			uassert.Equal(t, tt.feeProtocol0, pool.Slot0FeeProtocol()%16)
			uassert.Equal(t, tt.feeProtocol1, pool.Slot0FeeProtocol()>>4)
		})
	}
}
