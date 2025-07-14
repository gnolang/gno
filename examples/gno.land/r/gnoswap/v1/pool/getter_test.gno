package pool

import (
	"testing"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/uassert"

	i256 "gno.land/p/gnoswap/int256"
	u256 "gno.land/p/gnoswap/uint256"
)

// create mock data for getters
func testInitData(t *testing.T) {
	t.Helper()
	resetObject(t)

	mockPool := &Pool{
		token0Path: "token0",
		token1Path: "token1",
		balances: Balances{
			token0: u256.NewUint(1000),
			token1: u256.NewUint(2000),
		},
		fee:                 3000,
		tickSpacing:         10,
		maxLiquidityPerTick: u256.NewUint(1000000),
		slot0: Slot0{
			sqrtPriceX96: u256.NewUint(1000000000000000000),
			tick:         5,
			feeProtocol:  6,
			unlocked:     false,
		},
		feeGrowthGlobal0X128: u256.NewUint(1000000000000000000),
		feeGrowthGlobal1X128: u256.NewUint(2000000000000000000),
		protocolFees: ProtocolFees{
			token0: u256.NewUint(1000),
			token1: u256.NewUint(2000),
		},
		liquidity: u256.NewUint(1000000),
	}

	mockTicks := avl.NewTree()
	mockTicks.Set("0", TickInfo{
		liquidityGross:                 u256.NewUint(1000000),
		liquidityNet:                   i256.NewInt(2000000),
		feeGrowthOutside0X128:          u256.NewUint(3000000),
		feeGrowthOutside1X128:          u256.NewUint(4000000),
		tickCumulativeOutside:          5,
		secondsPerLiquidityOutsideX128: u256.NewUint(6000000),
		secondsOutside:                 7,
		initialized:                    true,
	})
	mockPool.ticks = mockTicks

	mockPositions := avl.NewTree()
	mockPositions.Set("test_position", PositionInfo{
		liquidity:                u256.NewUint(1000000),
		feeGrowthInside0LastX128: u256.NewUint(2000000),
		feeGrowthInside1LastX128: u256.NewUint(3000000),
		tokensOwed0:              u256.NewUint(4000000),
		tokensOwed1:              u256.NewUint(5000000),
	})
	mockPool.positions = mockPositions

	pools.Set("token0:token1:3000", mockPool)
}

func TestPoolGetters(t *testing.T) {
	testInitData(t)

	t.Run("get pool list", func(t *testing.T) {
		poolList := PoolGetPoolList()
		if len(poolList) != 1 || poolList[0] != "token0:token1:3000" {
			t.Errorf("Expected pool list [%s], got %v", "token0:token1:3000", poolList)
		}
	})

	t.Run("valid pool path", func(t *testing.T) {
		validPoolPath := "token0:token1:3000"

		t.Run("get token0 path", func(t *testing.T) {
			token0Path := PoolGetToken0Path(validPoolPath)
			if token0Path != "token0" {
				t.Errorf("Expected token0 path [%s], got %s", "token0", token0Path)
			}
		})

		t.Run("get token1 path", func(t *testing.T) {
			token1Path := PoolGetToken1Path(validPoolPath)
			if token1Path != "token1" {
				t.Errorf("Expected token1 path [%s], got %s", "token1", token1Path)
			}
		})

		t.Run("get fee", func(t *testing.T) {
			fee := PoolGetFee(validPoolPath)
			if fee != 3000 {
				t.Errorf("Expected fee [%d], got %d", 3000, fee)
			}
		})

		t.Run("get balance token0", func(t *testing.T) {
			balanceToken0 := PoolGetBalanceToken0(validPoolPath)
			if balanceToken0 != "1000" {
				t.Errorf("Expected balance token0 [%s], got %s", "1000", balanceToken0)
			}
		})

		t.Run("get balance token1", func(t *testing.T) {
			balanceToken1 := PoolGetBalanceToken1(validPoolPath)
			if balanceToken1 != "2000" {
				t.Errorf("Expected balance token1 [%s], got %s", "2000", balanceToken1)
			}
		})

		t.Run("get tick spacing", func(t *testing.T) {
			tickSpacing := PoolGetTickSpacing(validPoolPath)
			if tickSpacing != 10 {
				t.Errorf("Expected tick spacing [%d], got %d", 10, tickSpacing)
			}
		})

		t.Run("get max liquidity per tick", func(t *testing.T) {
			maxLiquidityPerTick := PoolGetMaxLiquidityPerTick(validPoolPath)
			if maxLiquidityPerTick != "1000000" {
				t.Errorf("Expected max liquidity per tick [%s], got %s", "1000000", maxLiquidityPerTick)
			}
		})

		t.Run("get slot0 sqrt price x96", func(t *testing.T) {
			sqrtPriceX96 := PoolGetSlot0SqrtPriceX96(validPoolPath)
			if sqrtPriceX96 != "1000000000000000000" {
				t.Errorf("Expected sqrt price x96 [%s], got %s", "1000000000000000000", sqrtPriceX96)
			}
		})

		t.Run("get slot0 tick", func(t *testing.T) {
			tick := PoolGetSlot0Tick(validPoolPath)
			if tick != 5 {
				t.Errorf("Expected tick [%d], got %d", 5, tick)
			}
		})

		t.Run("get slot0 fee protocol", func(t *testing.T) {
			feeProtocol := PoolGetSlot0FeeProtocol(validPoolPath)
			if feeProtocol != 6 {
				t.Errorf("Expected fee protocol [%d], got %d", 6, feeProtocol)
			}
		})

		t.Run("get slot0 unlocked", func(t *testing.T) {
			unlocked := PoolGetSlot0Unlocked(validPoolPath)
			if unlocked != false {
				t.Errorf("Expected unlocked [%t], got %t", false, unlocked)
			}
		})

		t.Run("get fee growth global0x128", func(t *testing.T) {
			feeGrowthGlobal0X128 := PoolGetFeeGrowthGlobal0X128(validPoolPath)
			if feeGrowthGlobal0X128 != "1000000000000000000" {
				t.Errorf("Expected fee growth global 0 x128 [%s], got %s", "1000000000000000000", feeGrowthGlobal0X128)
			}
		})

		t.Run("get fee growth global1x128", func(t *testing.T) {
			feeGrowthGlobal1X128 := PoolGetFeeGrowthGlobal1X128(validPoolPath)
			if feeGrowthGlobal1X128 != "2000000000000000000" {
				t.Errorf("Expected fee growth global 1 x128 [%s], got %s", "2000000000000000000", feeGrowthGlobal1X128)
			}
		})

		t.Run("get protocol fees token0", func(t *testing.T) {
			protocolFeesToken0 := PoolGetProtocolFeesToken0(validPoolPath)
			if protocolFeesToken0 != "1000" {
				t.Errorf("Expected protocol fees token0 [%s], got %s", "1000", protocolFeesToken0)
			}
		})

		t.Run("get protocol fees token1", func(t *testing.T) {
			protocolFeesToken1 := PoolGetProtocolFeesToken1(validPoolPath)
			if protocolFeesToken1 != "2000" {
				t.Errorf("Expected protocol fees token1 [%s], got %s", "2000", protocolFeesToken1)
			}
		})

		t.Run("get liquidity", func(t *testing.T) {
			liquidity := PoolGetLiquidity(validPoolPath)
			if liquidity != "1000000" {
				t.Errorf("Expected liquidity [%s], got %s", "1000000", liquidity)
			}
		})
	})

	t.Run("invalid pool path - should panic", func(t *testing.T) {
		invalidPoolPath := "invalid/pool/path"

		t.Run("get token0 path", func(t *testing.T) {
			shouldPanic(t, func() {
				PoolGetToken0Path(invalidPoolPath)
			})
		})

		t.Run("get token1 path", func(t *testing.T) {
			shouldPanic(t, func() {
				PoolGetToken1Path(invalidPoolPath)
			})
		})

		t.Run("get fee", func(t *testing.T) {
			shouldPanic(t, func() {
				PoolGetFee(invalidPoolPath)
			})
		})

		t.Run("get balance token0", func(t *testing.T) {
			shouldPanic(t, func() {
				PoolGetBalanceToken0(invalidPoolPath)
			})
		})

		t.Run("get balance token1", func(t *testing.T) {
			shouldPanic(t, func() {
				PoolGetBalanceToken1(invalidPoolPath)
			})
		})

		t.Run("get tick spacing", func(t *testing.T) {
			shouldPanic(t, func() {
				PoolGetTickSpacing(invalidPoolPath)
			})
		})

		t.Run("get liquidity", func(t *testing.T) {
			shouldPanic(t, func() {
				PoolGetLiquidity(invalidPoolPath)
			})
		})

		t.Run("get max liquidity per tick", func(t *testing.T) {
			shouldPanic(t, func() {
				PoolGetMaxLiquidityPerTick(invalidPoolPath)
			})
		})

		t.Run("get slot0 sqrt price x96", func(t *testing.T) {
			shouldPanic(t, func() {
				PoolGetSlot0SqrtPriceX96(invalidPoolPath)
			})
		})

		t.Run("get slot0 tick", func(t *testing.T) {
			shouldPanic(t, func() {
				PoolGetSlot0Tick(invalidPoolPath)
			})
		})

		t.Run("get slot0 fee protocol", func(t *testing.T) {
			shouldPanic(t, func() {
				PoolGetSlot0FeeProtocol(invalidPoolPath)
			})
		})

		t.Run("get slot0 unlocked", func(t *testing.T) {
			shouldPanic(t, func() {
				PoolGetSlot0Unlocked(invalidPoolPath)
			})
		})

		t.Run("get fee growth global0x128", func(t *testing.T) {
			shouldPanic(t, func() {
				PoolGetFeeGrowthGlobal0X128(invalidPoolPath)
			})
		})

		t.Run("get fee growth global1x128", func(t *testing.T) {
			shouldPanic(t, func() {
				PoolGetFeeGrowthGlobal1X128(invalidPoolPath)
			})
		})

		t.Run("get protocol fees token0", func(t *testing.T) {
			shouldPanic(t, func() {
				PoolGetProtocolFeesToken0(invalidPoolPath)
			})
		})

		t.Run("get protocol fees token1", func(t *testing.T) {
			shouldPanic(t, func() {
				PoolGetProtocolFeesToken1(invalidPoolPath)
			})
		})

		t.Run("get liquidity", func(t *testing.T) {
			shouldPanic(t, func() {
				PoolGetLiquidity(invalidPoolPath)
			})
		})
	})
}

func TestPositionGetters(t *testing.T) {
	tests := []struct {
		name                             string
		poolPath                         string
		positionKey                      string
		expectedLiquidity                string
		expectedFeeGrowthInside0LastX128 string
		expectedFeeGrowthInside1LastX128 string
		expectedTokensOwed0              string
		expectedTokensOwed1              string
		expectedHasPanic                 bool
		expectedPanicMessage             string
	}{
		{
			name:                             "position getters is success by valid position",
			poolPath:                         "token0:token1:3000",
			positionKey:                      "test_position",
			expectedLiquidity:                "1000000",
			expectedFeeGrowthInside0LastX128: "2000000",
			expectedFeeGrowthInside1LastX128: "3000000",
			expectedTokensOwed0:              "4000000",
			expectedTokensOwed1:              "5000000",
			expectedHasPanic:                 false,
		},
		{
			name:                             "position getters is failed by invalid position",
			poolPath:                         "token0:token1:3000",
			positionKey:                      "invalid_position",
			expectedLiquidity:                "0",
			expectedFeeGrowthInside0LastX128: "0",
			expectedFeeGrowthInside1LastX128: "0",
			expectedTokensOwed0:              "0",
			expectedTokensOwed1:              "0",
			expectedHasPanic:                 true,
			expectedPanicMessage:             "[GNOSWAP-POOL-008] requested data not found || positionKey(invalid_position) does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testInitData(t)

			var liquidity string
			var feeGrowthInside0LastX128 string
			var feeGrowthInside1LastX128 string
			var tokensOwed0 string
			var tokensOwed1 string

			if tt.expectedHasPanic {
				uassert.PanicsWithMessage(t, tt.expectedPanicMessage, func() {
					PoolGetPositionLiquidity(tt.poolPath, tt.positionKey)
				})
				uassert.PanicsWithMessage(t, tt.expectedPanicMessage, func() {
					PoolGetPositionFeeGrowthInside0LastX128(tt.poolPath, tt.positionKey)
				})
				uassert.PanicsWithMessage(t, tt.expectedPanicMessage, func() {
					PoolGetPositionFeeGrowthInside1LastX128(tt.poolPath, tt.positionKey)
				})
				uassert.PanicsWithMessage(t, tt.expectedPanicMessage, func() {
					PoolGetPositionTokensOwed0(tt.poolPath, tt.positionKey)
				})
				uassert.PanicsWithMessage(t, tt.expectedPanicMessage, func() {
					PoolGetPositionTokensOwed1(tt.poolPath, tt.positionKey)
				})

				return
			} else {
				liquidity = PoolGetPositionLiquidity(tt.poolPath, tt.positionKey)
				feeGrowthInside0LastX128 = PoolGetPositionFeeGrowthInside0LastX128(tt.poolPath, tt.positionKey)
				feeGrowthInside1LastX128 = PoolGetPositionFeeGrowthInside1LastX128(tt.poolPath, tt.positionKey)
				tokensOwed0 = PoolGetPositionTokensOwed0(tt.poolPath, tt.positionKey)
				tokensOwed1 = PoolGetPositionTokensOwed1(tt.poolPath, tt.positionKey)
			}

			uassert.Equal(t, tt.expectedLiquidity, liquidity)
			uassert.Equal(t, tt.expectedFeeGrowthInside0LastX128, feeGrowthInside0LastX128)
			uassert.Equal(t, tt.expectedFeeGrowthInside1LastX128, feeGrowthInside1LastX128)
			uassert.Equal(t, tt.expectedTokensOwed0, tokensOwed0)
			uassert.Equal(t, tt.expectedTokensOwed1, tokensOwed1)
		})
	}

	t.Run("invalid position", func(t *testing.T) {
		validPoolPath := "token0:token1:3000"        // pool must valid
		invalidPositionKey := "invalid/position/key" // but position must not

		t.Run("get position liquidity", func(t *testing.T) {
			shouldPanic(t, func() {
				PoolGetPositionLiquidity(validPoolPath, invalidPositionKey)
			})
		})

		t.Run("get position fee growth inside0 last x128", func(t *testing.T) {
			shouldPanic(t, func() {
				PoolGetPositionFeeGrowthInside0LastX128(validPoolPath, invalidPositionKey)
			})
		})

		t.Run("get position fee growth inside1 last x128", func(t *testing.T) {
			shouldPanic(t, func() {
				PoolGetPositionFeeGrowthInside1LastX128(validPoolPath, invalidPositionKey)
			})
		})

		t.Run("get position tokens owed0", func(t *testing.T) {
			shouldPanic(t, func() {
				PoolGetPositionTokensOwed0(validPoolPath, invalidPositionKey)
			})
		})

		t.Run("get position tokens owed1", func(t *testing.T) {
			shouldPanic(t, func() {
				PoolGetPositionTokensOwed1(validPoolPath, invalidPositionKey)
			})
		})
	})
}

func TestTickGetters(t *testing.T) {
	testInitData(t)
	t.Run("valid tick", func(t *testing.T) {
		validPoolPath := "token0:token1:3000"
		validTick := int32(0)

		t.Run("get tick liquidity gross", func(t *testing.T) {
			liquidityGross := PoolGetTickLiquidityGross(validPoolPath, validTick)
			if liquidityGross != "1000000" {
				t.Errorf("Expected liquidity gross [%s], got %s", "1000000", liquidityGross)
			}
		})

		t.Run("get tick liquidity net", func(t *testing.T) {
			liquidityNet := PoolGetTickLiquidityNet(validPoolPath, validTick)
			if liquidityNet != "2000000" {
				t.Errorf("Expected liquidity net [%s], got %s", "2000000", liquidityNet)
			}
		})

		t.Run("get tick fee growth outside0 x128", func(t *testing.T) {
			feeGrowthOutside0X128 := PoolGetTickFeeGrowthOutside0X128(validPoolPath, validTick)
			if feeGrowthOutside0X128 != "3000000" {
				t.Errorf("Expected fee growth outside0 x128 [%s], got %s", "3000000", feeGrowthOutside0X128)
			}
		})

		t.Run("get tick fee growth outside1 x128", func(t *testing.T) {
			feeGrowthOutside1X128 := PoolGetTickFeeGrowthOutside1X128(validPoolPath, validTick)
			if feeGrowthOutside1X128 != "4000000" {
				t.Errorf("Expected fee growth outside1 x128 [%s], got %s", "4000000", feeGrowthOutside1X128)
			}
		})

		t.Run("get tick cumulative outside", func(t *testing.T) {
			tickCumulativeOutside := PoolGetTickCumulativeOutside(validPoolPath, validTick)
			if tickCumulativeOutside != 5 {
				t.Errorf("Expected tick cumulative outside [%d], got %d", 5, tickCumulativeOutside)
			}
		})

		t.Run("get tick seconds per liquidity outside x128", func(t *testing.T) {
			secondsPerLiquidityOutsideX128 := PoolGetTickSecondsPerLiquidityOutsideX128(validPoolPath, validTick)
			if secondsPerLiquidityOutsideX128 != "6000000" {
				t.Errorf("Expected seconds per liquidity outside x128 [%s], got %s", "6000000", secondsPerLiquidityOutsideX128)
			}
		})

		t.Run("get tick seconds outside", func(t *testing.T) {
			secondsOutside := PoolGetTickSecondsOutside(validPoolPath, validTick)
			if secondsOutside != 7 {
				t.Errorf("Expected seconds outside [%d], got %d", 7, secondsOutside)
			}
		})

		t.Run("get tick initialized", func(t *testing.T) {
			initialized := PoolGetTickInitialized(validPoolPath, validTick)
			if initialized != true {
				t.Errorf("Expected initialized [%t], got %t", true, initialized)
			}
		})
	})

	t.Run("invalid tick", func(t *testing.T) {
		validPoolPath := "token0:token1:3000" // pool must valid
		invalidTick := int32(-1)              // but tick must not

		t.Run("get tick liquidity gross", func(t *testing.T) {
			shouldPanic(t, func() {
				PoolGetTickLiquidityGross(validPoolPath, invalidTick)
			})
		})

		t.Run("get tick liquidity net", func(t *testing.T) {
			shouldPanic(t, func() {
				PoolGetTickLiquidityNet(validPoolPath, invalidTick)
			})
		})

		t.Run("get tick fee growth outside0 x128", func(t *testing.T) {
			shouldPanic(t, func() {
				PoolGetTickFeeGrowthOutside0X128(validPoolPath, invalidTick)
			})
		})

		t.Run("get tick cumulative outside", func(t *testing.T) {
			shouldPanic(t, func() {
				PoolGetTickCumulativeOutside(validPoolPath, invalidTick)
			})
		})

		t.Run("get tick seconds per liquidity outside x128", func(t *testing.T) {
			shouldPanic(t, func() {
				PoolGetTickSecondsPerLiquidityOutsideX128(validPoolPath, invalidTick)
			})
		})

		t.Run("get tick seconds outside", func(t *testing.T) {
			shouldPanic(t, func() {
				PoolGetTickSecondsOutside(validPoolPath, invalidTick)
			})
		})

		t.Run("get tick initialized", func(t *testing.T) {
			shouldPanic(t, func() {
				PoolGetTickInitialized(validPoolPath, invalidTick)
			})
		})
	})
}

func shouldPanic(t *testing.T, fn func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic")
		}
	}()
	fn()
}
