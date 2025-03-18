package pool

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/gnoswap/consts"
	u256 "gno.land/p/gnoswap/uint256"

	"gno.land/r/demo/wugnot"
	"gno.land/r/gnoswap/v1/common"
	"gno.land/r/gnoswap/v1/gns"
	pn "gno.land/r/gnoswap/v1/position"
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
	authorized := positionAddr
	barTokenPath := "gno.land/r/onbloc/bar"
	bazTokenPath := "gno.land/r/onbloc/baz"

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
		defer func() {
			if r := recover(); r == nil {
				t.Error("mint with 0 liquidity should fail")
			}
		}()

		Mint(token0Path, token1Path, fee, recipient, tickLower, tickUpper, "0", authorized)
	})

	t.Run("invalid tick range should fail", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("invalid tick range should fail")
			}
		}()
		Mint(token0Path, token1Path, fee, recipient, 100, 100, liquidityAmount, authorized)
	})

	t.Run("invalid fee amount should fail", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("invalid fee amount should fail")
			}
		}()
		Mint(token0Path, token1Path, 2000, recipient, -100, 100, liquidityAmount, authorized)
	})

	t.Run("non-existent pool should fail", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("non-existent pool should fail")
			}
		}()
		Mint("nonexistent_token0", "nonexistent_token1", fee, recipient, -100, 100, liquidityAmount, authorized)
	})

	t.Run("successful mint", func(t *testing.T) {
		// create pool first
		std.TestSetRealm(adminRealm)
		SetPoolCreationFeeByAdmin(0)
		CreatePool(
			barTokenPath,
			bazTokenPath,
			FeeTier100,
			"79228162514264337593543950336",
		)

		// approve enough token to mint position
		bar.Approve(poolAddr, consts.UINT64_MAX)
		baz.Approve(poolAddr, consts.UINT64_MAX)

		beforeBalance0 := bar.BalanceOf(poolAddr)
		beforeBalance1 := baz.BalanceOf(poolAddr)

		std.TestSetRealm(std.NewCodeRealm(consts.POSITION_PATH))
		mintAmount0, mintAmount1 := Mint(
			barTokenPath,
			bazTokenPath,
			FeeTier100,
			positionAddr,
			int32(-100),
			int32(100),
			"10000000000",
			adminAddr,
		)
		uassert.Equal(t, mintAmount0, "49872721")
		uassert.Equal(t, mintAmount1, "49872721")

		afterBalance0 := bar.BalanceOf(poolAddr)
		afterBalance1 := baz.BalanceOf(poolAddr)
		uassert.Equal(t, uint64(49872721), afterBalance0-beforeBalance0)
		uassert.Equal(t, uint64(49872721), afterBalance1-beforeBalance1)
	})
}

func TestBurn(t *testing.T) {
	barTokenPath := "gno.land/r/onbloc/bar"
	bazTokenPath := "gno.land/r/onbloc/baz"

	tests := []struct {
		name            string
		liquidityAmount string
		tickLower       int32
		tickUpper       int32
		fee             uint32
		expectedAmount0 string
		expectedAmount1 string
		expectPanic     bool
	}{
		{
			name:            "successful burn",
			liquidityAmount: "10000000000", // burn all liquidity
			tickLower:       -100,
			tickUpper:       100,
			fee:             FeeTier100,
			expectedAmount0: "49872720",
			expectedAmount1: "49872720",
		},
		{
			name:            "zero liquidity",
			liquidityAmount: "0",
			tickLower:       -100,
			tickUpper:       100,
			fee:             FeeTier100,
			expectPanic:     true,
		},
		{
			name:            "invalid tick range",
			liquidityAmount: "100000",
			tickLower:       200,
			tickUpper:       100,
			expectPanic:     true,
		},
		{
			name:            "invalid fee amount",
			liquidityAmount: "100000",
			tickLower:       -100,
			tickUpper:       100,
			fee:             2000,
			expectPanic:     true,
		},
		{
			name:            "non-existent pool",
			liquidityAmount: "100000",
			tickLower:       -100,
			tickUpper:       100,
			fee:             FeeTier500,
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
				barTokenPath,
				bazTokenPath,
				tt.fee,
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

				pool := GetPool(barTokenPath, bazTokenPath, FeeTier100)
				posKey := getPositionKey(positionAddr, tt.tickLower, tt.tickUpper)
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

func TestCollect(t *testing.T) {
	positionOwner := adminAddr

	tests := []struct {
		name             string
		tokenPath0       string
		tokenPath1       string
		feeTier          uint32
		recipient        std.Address
		tickLower        int32
		tickUpper        int32
		amount0Requested string
		amount1Requested string
		expectedAmount0  string
		expectedAmount1  string
		expectPanic      bool
	}{
		{
			name:             "successful collect full amount",
			tokenPath0:       wugnotPath,
			tokenPath1:       gnsPath,
			feeTier:          FeeTier10000,
			recipient:        positionOwner,
			tickLower:        -200,
			tickUpper:        200,
			amount0Requested: "100",
			amount1Requested: "100",
			expectedAmount0:  "0",
			expectedAmount1:  "0",
		},
		{
			name:             "partial collect due to pool balance",
			tokenPath0:       wugnotPath,
			tokenPath1:       gnsPath,
			feeTier:          FeeTier10000,
			recipient:        positionOwner,
			tickLower:        -200,
			tickUpper:        200,
			amount0Requested: "1000",
			amount1Requested: "0",
			expectedAmount0:  "1000",
			expectedAmount1:  "0",
		},
		{
			name:             "partial collect due to tokensOwed limit",
			tokenPath0:       wugnotPath,
			tokenPath1:       gnsPath,
			feeTier:          FeeTier10000,
			recipient:        positionOwner,
			tickLower:        -200,
			tickUpper:        200,
			amount0Requested: "18998",
			amount1Requested: "600",
			expectedAmount0:  "18998",
			expectedAmount1:  "0",
		},
		{
			name:             "collect with zero requested amount",
			tokenPath0:       wugnotPath,
			tokenPath1:       gnsPath,
			feeTier:          FeeTier10000,
			recipient:        positionOwner,
			tickLower:        -200,
			tickUpper:        200,
			amount0Requested: "0",
			amount1Requested: "0",
			expectedAmount0:  "0",
			expectedAmount1:  "0",
		},
		{
			name:             "panic when position does not exist",
			tokenPath0:       wugnotPath,
			tokenPath1:       gnsPath,
			feeTier:          FeeTier10000,
			recipient:        positionOwner,
			tickLower:        -400,
			tickUpper:        400,
			amount0Requested: "500",
			amount1Requested: "300",
			expectPanic:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Pool Creation
			std.TestSetRealm(std.NewUserRealm(positionOwner))
			SetPoolCreationFeeByAdmin(0)
			if !DoesPoolPathExist(GetPoolPath(tt.tokenPath0, tt.tokenPath1, tt.feeTier)) {
				CreatePool(tt.tokenPath0, tt.tokenPath1, tt.feeTier, "79228162514264337593543950336")
			}

			if !tt.expectPanic {
				std.TestSetOriginCaller(positionOwner)
				newCoins := std.Coins{{"ugnot", int64(10000000000)}}
				std.TestIssueCoins(positionOwner, newCoins)
				std.TestSetOriginSend(newCoins, nil)
				banker := std.NewBanker(std.BankerTypeRealmSend)
				banker.SendCoins(positionOwner, consts.WUGNOT_ADDR, newCoins)
				wugnot.Deposit()

				wugnot.Approve(poolAddr, maxApprove)
				gns.Approve(poolAddr, maxApprove)

				teller := common.GetTokenTeller(wugnotPath)
				teller.Approve(poolAddr, maxApprove)

				std.TestSetRealm(std.NewUserRealm(positionOwner))
				// Position Creation
				pn.Mint(
					tt.tokenPath0,
					tt.tokenPath1,
					tt.feeTier,
					tt.tickLower,
					tt.tickUpper,
					"100000000",
					"100000000",
					"0",
					"0",
					max_timeout,
					tt.recipient,
					tt.recipient,
					"",
				)

				pool := GetPool(tt.tokenPath0, tt.tokenPath1, tt.feeTier)
				posKey := getPositionKey(positionAddr, tt.tickLower, tt.tickUpper)
				pool.mustGetPosition(posKey)

				std.TestSkipHeights(1)

				// Swap
				std.TestSetOriginCaller(positionOwner)
				wugnot.Transfer(routerAddr, 20000000)
				gns.Transfer(routerAddr, 20000000)

				std.TestSetRealm(rouRealm)
				wugnot.Approve(poolAddr, maxApprove)

				amount0, amount1 := Swap(
					tt.tokenPath0,
					tt.tokenPath1,
					tt.feeTier,
					tt.recipient,
					true,
					"1000000",
					"3945129629379410362911094632",
					tt.recipient,
				)
				std.TestSkipHeights(1)
			}

			defer func() {
				if tt.expectPanic {
					if r := recover(); r == nil {
						t.Errorf("expected panic but got none")
					}
				}
			}()

			// Collect
			std.TestSetRealm(posRealm)
			collectedAmount0, collectedAmount1 := Collect(
				tt.tokenPath0,
				tt.tokenPath1,
				tt.feeTier,
				tt.recipient,
				tt.tickLower,
				tt.tickUpper,
				tt.amount0Requested,
				tt.amount1Requested,
			)
			if !tt.expectPanic {
				uassert.Equal(t, tt.expectedAmount0, collectedAmount0)
				uassert.Equal(t, tt.expectedAmount1, collectedAmount1)
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
			pool := GetPool(wugnotPath, gnsPath, FeeTier3000)
			SetFeeProtocolByAdmin(tt.feeProtocol0, tt.feeProtocol1)
			uassert.Equal(t, tt.feeProtocol0, pool.Slot0FeeProtocol()%16)
			uassert.Equal(t, tt.feeProtocol1, pool.Slot0FeeProtocol()>>4)
		})
	}
}

func TestCollectProtocolByAdmin(t *testing.T) {
	positionOwner := adminAddr

	tests := []struct {
		name             string
		tokenPath0       string
		tokenPath1       string
		feeTier          uint32
		recipient        std.Address
		tickLower        int32
		tickUpper        int32
		amount0Requested string
		amount1Requested string
		feeProtocol0     uint8
		feeProtocol1     uint8
		expectedAmount0  string
		expectedAmount1  string
		expectPanic      bool
	}{
		{
			name:             "failure - not admin address",
			tokenPath0:       wugnotPath,
			tokenPath1:       gnsPath,
			feeTier:          FeeTier500,
			recipient:        alice,
			tickLower:        -200,
			tickUpper:        200,
			amount0Requested: "100",
			amount1Requested: "100",
			feeProtocol0:     10,
			feeProtocol1:     10,
			expectedAmount0:  "0",
			expectedAmount1:  "0",
			expectPanic:      true,
		},
		{
			name:             "failure - collect fee amount is zero",
			tokenPath0:       wugnotPath,
			tokenPath1:       gnsPath,
			feeTier:          FeeTier500,
			recipient:        positionOwner,
			tickLower:        -200,
			tickUpper:        200,
			amount0Requested: "0",
			amount1Requested: "0",
			expectedAmount0:  "0",
			expectedAmount1:  "0",
		},
		{
			name:             "successful collect fee amount",
			tokenPath0:       wugnotPath,
			tokenPath1:       gnsPath,
			feeTier:          FeeTier500,
			recipient:        positionOwner,
			tickLower:        -200,
			tickUpper:        200,
			amount0Requested: "100",
			amount1Requested: "100",
			expectedAmount0:  "100",
			expectedAmount1:  "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Pool Creation
			std.TestSetRealm(std.NewUserRealm(positionOwner))
			SetPoolCreationFeeByAdmin(0)
			if !DoesPoolPathExist(GetPoolPath(tt.tokenPath0, tt.tokenPath1, tt.feeTier)) {
				CreatePool(tt.tokenPath0, tt.tokenPath1, tt.feeTier, "79228162514264337593543950336")
			}

			if tt.feeProtocol0 != 0 || tt.feeProtocol1 != 0 {
				pool := GetPool(tt.tokenPath0, tt.tokenPath1, tt.feeTier)
				SetFeeProtocolByAdmin(tt.feeProtocol0, tt.feeProtocol1)
				uassert.Equal(t, tt.feeProtocol0, pool.Slot0FeeProtocol()%16)
				uassert.Equal(t, tt.feeProtocol1, pool.Slot0FeeProtocol()>>4)
			}
			std.TestSetOriginCaller(positionOwner)
			newCoins := std.Coins{{"ugnot", int64(10000000000)}}
			std.TestIssueCoins(positionOwner, newCoins)
			std.TestSetOriginSend(newCoins, nil)
			banker := std.NewBanker(std.BankerTypeRealmSend)
			banker.SendCoins(positionOwner, consts.WUGNOT_ADDR, newCoins)
			wugnot.Deposit()
			if tt.recipient != positionOwner {
				wugnot.Transfer(tt.recipient, 10000000000)
			}

			std.TestSetRealm(std.NewUserRealm(tt.recipient))
			wugnot.Approve(poolAddr, maxApprove)
			gns.Approve(poolAddr, maxApprove)

			teller := common.GetTokenTeller(wugnotPath)
			teller.Approve(poolAddr, maxApprove)

			// Position Creation
			pn.Mint(
				tt.tokenPath0,
				tt.tokenPath1,
				tt.feeTier,
				tt.tickLower,
				tt.tickUpper,
				"10000000",
				"10000000",
				"0",
				"0",
				max_timeout,
				tt.recipient,
				tt.recipient,
				"",
			)

			pool := GetPool(tt.tokenPath0, tt.tokenPath1, tt.feeTier)
			posKey := getPositionKey(positionAddr, tt.tickLower, tt.tickUpper)
			pool.mustGetPosition(posKey)

			std.TestSkipHeights(1)

			// Swap
			std.TestSetOriginCaller(tt.recipient)
			wugnot.Transfer(routerAddr, 20000000)
			gns.Transfer(routerAddr, 20000000)

			std.TestSetRealm(rouRealm)
			wugnot.Approve(poolAddr, maxApprove)

			amount0, amount1 := Swap(
				tt.tokenPath0,
				tt.tokenPath1,
				tt.feeTier,
				tt.recipient,
				true,
				"1000000",
				"3945129629379410362911094632",
				tt.recipient,
			)
			std.TestSkipHeights(1)

			defer func() {
				if r := recover(); r != nil {
					if tt.expectPanic {
						if errMsg, ok := r.(string); ok {
							uassert.Equal(t, "not authorized", errMsg)
						}
					} else {
						t.Errorf("expected panic but got none")
					}
				}
			}()

			// Collect
			pool = GetPool(tt.tokenPath0, tt.tokenPath1, tt.feeTier)
			beforePoolBalance0 := pool.BalanceToken0()
			beforePoolBalance1 := pool.BalanceToken1()
			std.TestSetRealm(std.NewUserRealm(tt.recipient))
			collectedAmount0, collectedAmount1 := CollectProtocolByAdmin(
				tt.tokenPath0,
				tt.tokenPath1,
				tt.feeTier,
				tt.recipient,
				tt.amount0Requested,
				tt.amount1Requested,
			)
			if !tt.expectPanic {
				uassert.Equal(t, tt.expectedAmount0, collectedAmount0)
				uassert.Equal(t, tt.expectedAmount1, collectedAmount1)

				pool = GetPool(tt.tokenPath0, tt.tokenPath1, tt.feeTier)
				afterPoolBalance0 := pool.BalanceToken0()
				afterPoolBalance1 := pool.BalanceToken1()

				uassert.Equal(t, u256.Zero().Sub(beforePoolBalance0, afterPoolBalance0).ToString(), collectedAmount0)
				uassert.Equal(t, u256.Zero().Sub(beforePoolBalance1, afterPoolBalance1).ToString(), collectedAmount1)
			}
		})
	}
}
