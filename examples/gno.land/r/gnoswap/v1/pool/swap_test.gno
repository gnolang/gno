package pool

import (
	"std"
	"testing"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/uassert"

	"gno.land/p/gnoswap/consts"
	i256 "gno.land/p/gnoswap/int256"
	u256 "gno.land/p/gnoswap/uint256"
	"gno.land/r/demo/wugnot"
	"gno.land/r/gnoswap/v1/common"
	"gno.land/r/gnoswap/v1/gns"
)

func TestSaveProtocolFees(t *testing.T) {
	tests := []struct {
		name     string
		pool     *Pool
		amount0  *u256.Uint
		amount1  *u256.Uint
		want0    *u256.Uint
		want1    *u256.Uint
		wantFee0 *u256.Uint
		wantFee1 *u256.Uint
	}{
		{
			name: "normal fee deduction",
			pool: &Pool{
				protocolFees: ProtocolFees{
					token0: u256.NewUint(1000),
					token1: u256.NewUint(2000),
				},
			},
			amount0:  u256.NewUint(500),
			amount1:  u256.NewUint(1000),
			want0:    u256.NewUint(500),
			want1:    u256.NewUint(1000),
			wantFee0: u256.NewUint(500),
			wantFee1: u256.NewUint(1000),
		},
		{
			name: "exact fee deduction (1 deduction)",
			pool: &Pool{
				protocolFees: ProtocolFees{
					token0: u256.NewUint(1000),
					token1: u256.NewUint(2000),
				},
			},
			amount0:  u256.NewUint(999),
			amount1:  u256.NewUint(1999),
			want0:    u256.NewUint(999),
			want1:    u256.NewUint(1999),
			wantFee0: u256.NewUint(1),
			wantFee1: u256.NewUint(1),
		},
		{
			name: "0 fee deduction",
			pool: &Pool{
				protocolFees: ProtocolFees{
					token0: u256.NewUint(1000),
					token1: u256.NewUint(2000),
				},
			},
			amount0:  u256.NewUint(0),
			amount1:  u256.NewUint(0),
			want0:    u256.NewUint(0),
			want1:    u256.NewUint(0),
			wantFee0: u256.NewUint(1000),
			wantFee1: u256.NewUint(2000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got0, got1 := tt.pool.saveProtocolFees(tt.amount0, tt.amount1)

			uassert.Equal(t, got0.ToString(), tt.want0.ToString())
			uassert.Equal(t, got1.ToString(), tt.want1.ToString())
			uassert.Equal(t, tt.pool.protocolFees.token0.ToString(), tt.wantFee0.ToString())
			uassert.Equal(t, tt.pool.protocolFees.token1.ToString(), tt.wantFee1.ToString())
		})
	}
}

func TestShouldContinueSwap(t *testing.T) {
	tests := []struct {
		name              string
		state             SwapState
		sqrtPriceLimitX96 *u256.Uint
		expected          bool
	}{
		{
			name: "Should continue - amount remaining and price not at limit",
			state: SwapState{
				amountSpecifiedRemaining: i256.MustFromDecimal("1000"),
				sqrtPriceX96:             u256.MustFromDecimal("1000000"),
			},
			sqrtPriceLimitX96: u256.MustFromDecimal("900000"),
			expected:          true,
		},
		{
			name: "Should stop - no amount remaining",
			state: SwapState{
				amountSpecifiedRemaining: i256.Zero(),
				sqrtPriceX96:             u256.MustFromDecimal("1000000"),
			},
			sqrtPriceLimitX96: u256.MustFromDecimal("900000"),
			expected:          false,
		},
		{
			name: "Should stop - price at limit",
			state: SwapState{
				amountSpecifiedRemaining: i256.MustFromDecimal("1000"),
				sqrtPriceX96:             u256.MustFromDecimal("900000"),
			},
			sqrtPriceLimitX96: u256.MustFromDecimal("900000"),
			expected:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldContinueSwap(tt.state, tt.sqrtPriceLimitX96)
			uassert.Equal(t, tt.expected, result)
		})
	}
}

func TestUpdateAmounts(t *testing.T) {
	tests := []struct {
		name          string
		step          StepComputations
		state         SwapState
		exactInput    bool
		expectedState SwapState
	}{
		{
			name: "Exact input update",
			step: StepComputations{
				amountIn:  u256.MustFromDecimal("100"),
				amountOut: u256.MustFromDecimal("97"),
				feeAmount: u256.MustFromDecimal("3"),
			},
			state: SwapState{
				amountSpecifiedRemaining: i256.MustFromDecimal("1000"),
				amountCalculated:         i256.Zero(),
			},
			exactInput: true,
			expectedState: SwapState{
				amountSpecifiedRemaining: i256.MustFromDecimal("897"), // 1000 - (100 + 3)
				amountCalculated:         i256.MustFromDecimal("-97"),
			},
		},
		{
			name: "Exact output update",
			step: StepComputations{
				amountIn:  u256.MustFromDecimal("100"),
				amountOut: u256.MustFromDecimal("97"),
				feeAmount: u256.MustFromDecimal("3"),
			},
			state: SwapState{
				amountSpecifiedRemaining: i256.MustFromDecimal("-1000"),
				amountCalculated:         i256.Zero(),
			},
			exactInput: false,
			expectedState: SwapState{
				amountSpecifiedRemaining: i256.MustFromDecimal("-903"), // -1000 + 97
				amountCalculated:         i256.MustFromDecimal("103"),  // 100 + 3
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := updateAmounts(tt.step, tt.state, tt.exactInput)

			uassert.True(t, tt.expectedState.amountSpecifiedRemaining.Eq(result.amountSpecifiedRemaining))
			uassert.True(t, tt.expectedState.amountCalculated.Eq(result.amountCalculated))
		})
	}
}

func TestComputeSwap(t *testing.T) {
	mockPool := &Pool{
		token0Path:  "token0",
		token1Path:  "token1",
		fee:         3000, // 0.3%
		tickSpacing: 60,
		slot0: Slot0{
			sqrtPriceX96: u256.MustFromDecimal("1000000000000000000"), // 1.0
			tick:         0,
			feeProtocol:  0,
			unlocked:     true,
		},
		liquidity: u256.MustFromDecimal("1000000000000000000"), // 1.0
		protocolFees: ProtocolFees{
			token0: u256.Zero(),
			token1: u256.Zero(),
		},
		feeGrowthGlobal0X128: u256.Zero(),
		feeGrowthGlobal1X128: u256.Zero(),
		tickBitmaps:          avl.NewTree(),
		ticks:                avl.NewTree(),
		positions:            avl.NewTree(),
	}

	wordPos, _ := tickBitmapPosition(0)
	mockPool.setTickBitmap(wordPos, u256.NewUint(1))

	t.Run("basic swap", func(t *testing.T) {
		comp := SwapComputation{
			AmountSpecified:   i256.MustFromDecimal("1000000"),             // 1.0 token
			SqrtPriceLimitX96: u256.MustFromDecimal("1100000000000000000"), // 1.1
			ZeroForOne:        true,
			ExactInput:        true,
			InitialState: SwapState{
				amountSpecifiedRemaining: i256.MustFromDecimal("1000000"),
				amountCalculated:         i256.Zero(),
				sqrtPriceX96:             mockPool.slot0.sqrtPriceX96,
				tick:                     mockPool.slot0.tick,
				feeGrowthGlobalX128:      mockPool.feeGrowthGlobal0X128,
				protocolFee:              u256.Zero(),
				liquidity:                mockPool.liquidity,
			},
			Cache: SwapCache{
				feeProtocol:    0,
				liquidityStart: mockPool.liquidity,
			},
		}

		result, err := computeSwap(mockPool, comp)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if result.Amount0.IsZero() {
			t.Error("expected non-zero amount0")
		}
		if result.Amount1.IsZero() {
			t.Error("expected non-zero amount1")
		}
	})

	t.Run("swap with zero liquidity", func(t *testing.T) {
		mockPoolZeroLiq := *mockPool
		mockPoolZeroLiq.liquidity = u256.Zero()

		comp := SwapComputation{
			AmountSpecified:   i256.MustFromDecimal("1000000"),
			SqrtPriceLimitX96: u256.MustFromDecimal("1100000000000000000"),
			ZeroForOne:        true,
			ExactInput:        true,
			InitialState: SwapState{
				amountSpecifiedRemaining: i256.MustFromDecimal("1000000"),
				amountCalculated:         i256.Zero(),
				sqrtPriceX96:             mockPoolZeroLiq.slot0.sqrtPriceX96,
				tick:                     mockPoolZeroLiq.slot0.tick,
				feeGrowthGlobalX128:      mockPoolZeroLiq.feeGrowthGlobal0X128,
				protocolFee:              u256.Zero(),
				liquidity:                mockPoolZeroLiq.liquidity,
			},
			Cache: SwapCache{
				feeProtocol:    0,
				liquidityStart: mockPoolZeroLiq.liquidity,
			},
		}

		result, err := computeSwap(&mockPoolZeroLiq, comp)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if !result.Amount0.IsZero() || !result.Amount1.IsZero() {
			t.Error("expected zero amounts for zero liquidity")
		}
	})
}

func TestDrySwap_Failures(t *testing.T) {
	testing.SetRealm(adminRealm)
	SetPoolCreationFeeByAdmin(cross, 0)
	CreatePool(cross, barPath, bazPath, fee100, "79228162514264337593543950336")

	tests := []struct {
		name              string
		token0Path        string
		token1Path        string
		fee               uint32
		zeroForOne        bool
		amountSpecified   string
		sqrtPriceLimitX96 string
		expectAmount0     string
		expectAmount1     string
		expectSuccess     bool
	}{
		{
			name:              "zero amount token0 to token1",
			token0Path:        barPath,
			token1Path:        bazPath,
			fee:               fee100,
			zeroForOne:        true,
			amountSpecified:   "0",
			sqrtPriceLimitX96: "79228162514264337593543950337",
			expectAmount0:     "0",
			expectAmount1:     "0",
			expectSuccess:     false,
		},
		{
			name:              "insufficient balance",
			token0Path:        barPath,
			token1Path:        bazPath,
			fee:               fee100,
			zeroForOne:        false,
			amountSpecified:   "2000000000",
			sqrtPriceLimitX96: "79228162514264337593543950337",
			expectAmount0:     "0",
			expectAmount1:     "0",
			expectSuccess:     false,
		},
		{
			name:              "insufficient balance token1 to token0",
			token0Path:        barPath,
			token1Path:        bazPath,
			fee:               fee100,
			zeroForOne:        false,
			amountSpecified:   "3000000000",
			sqrtPriceLimitX96: "79228162514264337593543950337",
			expectAmount0:     "0",
			expectAmount1:     "0",
			expectSuccess:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amount0, amount1, success := DrySwap(
				cross,
				tt.token0Path,
				tt.token1Path,
				tt.fee,
				tt.zeroForOne,
				tt.amountSpecified,
				tt.sqrtPriceLimitX96,
			)

			uassert.Equal(t, success, tt.expectSuccess)
			uassert.Equal(t, amount0, tt.expectAmount0)
			uassert.Equal(t, amount1, tt.expectAmount1)
		})
	}
}

func TestSwap_Failures(t *testing.T) {
	tests := []struct {
		name              string
		setupFn           func(t *testing.T)
		token0Path        string
		token1Path        string
		fee               uint32
		recipient         std.Address
		zeroForOne        bool
		amountSpecified   string
		sqrtPriceLimitX96 string
		payer             std.Address
		expectedAmount0   string
		expectedAmount1   string
		expectedMsg       string
	}{
		{
			name: "locked pool",
			setupFn: func(t *testing.T) {
				InitialisePoolTest(t)
				pool := GetPool(wugnotPath, gnsPath, fee3000)
				pool.slot0.unlocked = false
			},
			token0Path:        wugnotPath,
			token1Path:        gnsPath,
			fee:               fee3000,
			recipient:         routerAddr,
			zeroForOne:        true,
			amountSpecified:   "100",
			sqrtPriceLimitX96: "79228162514264337593543950336",
			payer:             routerAddr,
			expectedMsg:       "[GNOSWAP-POOL-018] can't swap while pool is locked",
		},
		{
			name: "zero amount",
			setupFn: func(t *testing.T) {
				InitialisePoolTest(t)
				// unlock pool
				pool := GetPool(wugnotPath, gnsPath, fee3000)
				pool.slot0.unlocked = true
			},
			token0Path:        wugnotPath,
			token1Path:        gnsPath,
			fee:               fee3000,
			recipient:         alice,
			zeroForOne:        true,
			amountSpecified:   "0",
			sqrtPriceLimitX96: "79228162514264337593543950336",
			payer:             alice,
			expectedMsg:       "[GNOSWAP-POOL-015] invalid swap amount || amountSpecified == 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetObject(t)
			burnTokens(t)

			if tt.setupFn != nil {
				tt.setupFn(t)
			}

			testing.SetRealm(rouRealm)
			uassert.AbortsWithMessage(t, tt.expectedMsg, func() {
				Swap(
					cross,
					tt.token0Path,
					tt.token1Path,
					tt.fee,
					tt.recipient,
					tt.zeroForOne,
					tt.amountSpecified,
					tt.sqrtPriceLimitX96,
					tt.payer,
				)
			})
		})
	}
}

func TestSwapAndDrySwapComparison(t *testing.T) {
	addr := routerAddr

	tests := []struct {
		name              string
		setupFn           func(t *testing.T)
		token0Path        string
		token1Path        string
		fee               uint32
		zeroForOne        bool
		amountSpecified   string
		sqrtPriceLimitX96 string
	}{
		{
			name: "normal swap",
			setupFn: func(t *testing.T) {
				InitialisePoolTest(t)
				TokenFaucet(t, gnsPath, addr)
				TokenApprove(t, gnsPath, addr, poolAddr, 1000)
			},
			token0Path:        wugnotPath,
			token1Path:        gnsPath,
			fee:               fee3000,
			zeroForOne:        false,
			amountSpecified:   "100",
			sqrtPriceLimitX96: maxSqrtPriceLimitX96,
		},
		{
			name: "swap - request to swap amount over total liquidty",
			setupFn: func(t *testing.T) {
				InitialisePoolTest(t)
				MintPositionAll(t, adminAddr)
				TokenFaucet(t, gnsPath, addr)
				TokenApprove(t, gnsPath, addr, poolAddr, maxApprove)
			},
			token0Path:        wugnotPath,
			token1Path:        gnsPath,
			fee:               fee3000,
			zeroForOne:        false,
			amountSpecified:   "2000000000",
			sqrtPriceLimitX96: maxSqrtPriceLimitX96,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetObject(t)
			burnTokens(t)

			if tt.setupFn != nil {
				tt.setupFn(t)
			}

			dryAmount0, dryAmount1, drySuccess := DrySwap(
				cross,
				tt.token0Path,
				tt.token1Path,
				tt.fee,
				tt.zeroForOne,
				tt.amountSpecified,
				tt.sqrtPriceLimitX96,
			)

			testing.SetOriginCaller(routerAddr)
			actualAmount0, actualAmount1 := Swap(
				cross,
				tt.token0Path,
				tt.token1Path,
				tt.fee,
				addr,
				tt.zeroForOne,
				tt.amountSpecified,
				tt.sqrtPriceLimitX96,
				addr,
			)

			if !drySuccess {
				t.Error("DrySwap failed but actual Swap succeeded")
			}

			uassert.NotEqual(t, dryAmount0, "0", "amount0 should not be zero")
			uassert.NotEqual(t, dryAmount1, "0", "amount1 should not be zero")
			uassert.NotEqual(t, actualAmount0, "0", "amount0 should not be zero")
			uassert.NotEqual(t, actualAmount1, "0", "amount1 should not be zero")

			uassert.Equal(t, dryAmount0, actualAmount0,
				"Amount0 mismatch between DrySwap and actual Swap")
			uassert.Equal(t, dryAmount1, actualAmount1,
				"Amount1 mismatch between DrySwap and actual Swap")
		})
	}
}

func TestSwapAndDrySwapComparison_amount_zero(t *testing.T) {
	addr := routerAddr

	tests := []struct {
		name        string
		setupFn     func(t *testing.T)
		action      func(t *testing.T)
		shouldPanic bool
		expected    string
	}{
		{
			name: "zero amount swap - zeroForOne = false",
			setupFn: func(t *testing.T) {
				InitialisePoolTest(t)
				TokenFaucet(t, gnsPath, addr)
				TokenApprove(t, gnsPath, addr, poolAddr, 1000)
			},
			action: func(t *testing.T) {
				dryAmount0, dryAmount1, drySuccess := DrySwap(
					cross,
					wugnotPath,
					gnsPath,
					fee3000,
					false,
					"0",
					maxSqrtPriceLimitX96,
				)
				uassert.Equal(t, "0", dryAmount0)
				uassert.Equal(t, "0", dryAmount1)
				uassert.Equal(t, false, drySuccess)
			},
			shouldPanic: false,
			expected:    "[GNOSWAP-POOL-014] invalid swap amount || amountSpecified == 0",
		},
		{
			name: "zero amount swap - zeroForOne = true",
			setupFn: func(t *testing.T) {
				InitialisePoolTest(t)
				TokenFaucet(t, gnsPath, addr)
				TokenApprove(t, gnsPath, addr, poolAddr, 1000)
			},
			action: func(t *testing.T) {
				testing.SetOriginCaller(routerAddr)
				actualAmount0, actualAmount1 := Swap(
					cross,
					wugnotPath,
					gnsPath,
					fee3000,
					addr,
					true,
					"0",
					maxSqrtPriceLimitX96,
					addr,
				)

				println("actualAmount0", actualAmount0)
				println("actualAmount1", actualAmount1)
			},
			shouldPanic: true,
			expected:    "[GNOSWAP-POOL-015] invalid swap amount || amountSpecified == 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetObject(t)
			burnTokens(t)

			if tt.setupFn != nil {
				tt.setupFn(t)
			}

			if tt.shouldPanic {
				uassert.AbortsWithMessage(t, tt.expected, func() { tt.action(t) })
			} else {
				tt.action(t)
			}
		})
	}
}

func TestSwap_amount_over_liquidity(t *testing.T) {
	addr := routerAddr

	tests := []struct {
		name        string
		setupFn     func(t *testing.T)
		action      func(t *testing.T) (string, string)
		shouldPanic bool
		expected    []string
	}{
		{
			name: "amount over liquidity - zeroForOne = false, token0:20000",
			setupFn: func(t *testing.T) {
				InitialisePoolTest(t)
				MintPositionAll(t, adminAddr)
				TokenFaucet(t, gnsPath, addr)
				TokenFaucet(t, wugnotPath, addr)
				TokenApprove(t, gnsPath, addr, poolAddr, maxApprove)
				TokenApprove(t, wugnotPath, addr, poolAddr, maxApprove)
			},
			action: func(t *testing.T) (string, string) {
				testing.SetOriginCaller(routerAddr)
				actualAmount0, actualAmount1 := Swap(
					cross,
					wugnotPath,
					gnsPath,
					fee3000,
					alice,
					false,
					"20000",
					common.TickMathGetSqrtRatioAtTick(887220).ToString(),
					addr,
				)
				return actualAmount0, actualAmount1
			},
			shouldPanic: false,
			expected:    []string{"-19936", "20000"},
		},
		{
			name: "amount over liquidity - zeroForOne = false, token1:-20000",
			setupFn: func(t *testing.T) {
				InitialisePoolTest(t)
				MintPositionAll(t, adminAddr)
				TokenFaucet(t, gnsPath, addr)
				TokenFaucet(t, wugnotPath, addr)
				TokenApprove(t, gnsPath, addr, poolAddr, maxApprove)
				TokenApprove(t, wugnotPath, addr, poolAddr, maxApprove)
			},
			action: func(t *testing.T) (string, string) {
				testing.SetOriginCaller(routerAddr)
				actualAmount0, actualAmount1 := Swap(
					cross,
					wugnotPath,
					gnsPath,
					fee3000,
					alice,
					false,
					"-20000",
					common.TickMathGetSqrtRatioAtTick(887220).ToString(),
					addr,
				)
				return actualAmount0, actualAmount1
			},
			shouldPanic: false,
			expected:    []string{"-20000", "20066"},
		},
		{
			name: "amount over liquidity - zeroForOne = true, token0:20000",
			setupFn: func(t *testing.T) {
				InitialisePoolTest(t)
				MintPositionAll(t, adminAddr)
				TokenFaucet(t, gnsPath, addr)
				TokenFaucet(t, wugnotPath, addr)
				TokenApprove(t, gnsPath, addr, poolAddr, maxApprove)
				TokenApprove(t, wugnotPath, addr, poolAddr, maxApprove)
			},
			action: func(t *testing.T) (string, string) {
				testing.SetOriginCaller(routerAddr)
				actualAmount0, actualAmount1 := Swap(
					cross,
					wugnotPath,
					gnsPath,
					fee3000,
					alice,
					true,
					"20000",
					consts.MIN_PRICE,
					addr,
				)
				return actualAmount0, actualAmount1
			},
			shouldPanic: false,
			expected:    []string{"20000", "-19936"},
		},
		{
			name: "amount over liquidity - zeroForOne = true, token1:-20000",
			setupFn: func(t *testing.T) {
				InitialisePoolTest(t)
				MintPositionAll(t, adminAddr)
				TokenFaucet(t, gnsPath, addr)
				TokenFaucet(t, wugnotPath, addr)
				TokenApprove(t, gnsPath, addr, poolAddr, maxApprove)
				TokenApprove(t, wugnotPath, addr, poolAddr, maxApprove)
			},
			action: func(t *testing.T) (string, string) {
				testing.SetOriginCaller(routerAddr)
				actualAmount0, actualAmount1 := Swap(
					cross,
					wugnotPath,
					gnsPath,
					fee3000,
					alice,
					true,
					"-20000",
					consts.MIN_PRICE,
					addr,
				)
				return actualAmount0, actualAmount1
			},
			shouldPanic: false,
			expected:    []string{"20066", "-20000"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetObject(t)
			burnTokens(t)

			if tt.setupFn != nil {
				tt.setupFn(t)
			}

			if tt.shouldPanic {
				tt.action(t)
			} else {
				amount0, amount1 := tt.action(t)
				uassert.Equal(t, tt.expected[0], amount0)
				uassert.Equal(t, tt.expected[1], amount1)
			}
		})
	}
}

func TestSwap_EXACTIN_OUT(t *testing.T) {
	addr := routerAddr

	tests := []struct {
		name        string
		setupFn     func(t *testing.T)
		action      func(t *testing.T) (string, string)
		shouldPanic bool
		expected    []string
	}{
		{
			name: "EXACT IN - zeroForOne = false, token1:200",
			setupFn: func(t *testing.T) {
				InitialisePoolTest(t)
				MintPositionAll(t, adminAddr)
				TokenFaucet(t, gnsPath, addr)
				TokenApprove(t, gnsPath, addr, poolAddr, maxApprove)
			},
			action: func(t *testing.T) (string, string) {
				testing.SetOriginCaller(routerAddr)
				actualAmount0, actualAmount1 := Swap(
					cross,
					wugnotPath,
					gnsPath,
					fee3000,
					alice,
					false,
					"200",
					common.TickMathGetSqrtRatioAtTick(887220).ToString(),
					addr,
				)
				return actualAmount0, actualAmount1
			},
			shouldPanic: false,
			expected:    []string{"-198", "200"},
		},
		{
			name: "EXACT OUT - zeroForOne = false, token0:-200",
			setupFn: func(t *testing.T) {
				InitialisePoolTest(t)
				MintPositionAll(t, adminAddr)
				TokenFaucet(t, gnsPath, addr)
				TokenApprove(t, gnsPath, addr, poolAddr, maxApprove)
			},
			action: func(t *testing.T) (string, string) {
				testing.SetOriginCaller(routerAddr)
				actualAmount0, actualAmount1 := Swap(
					cross,
					wugnotPath,
					gnsPath,
					fee3000,
					alice,
					false,
					"-200",
					common.TickMathGetSqrtRatioAtTick(887220).ToString(),
					addr,
				)
				return actualAmount0, actualAmount1
			},
			shouldPanic: false,
			expected:    []string{"-200", "202"},
		},
		{
			name: "EXACT IN - zeroForOne = true, token0:200",
			setupFn: func(t *testing.T) {
				InitialisePoolTest(t)
				MintPositionAll(t, adminAddr)
				TokenFaucet(t, gnsPath, addr)
				TokenFaucet(t, wugnotPath, addr)
				TokenApprove(t, gnsPath, addr, poolAddr, maxApprove)
				TokenApprove(t, wugnotPath, addr, poolAddr, maxApprove)
			},
			action: func(t *testing.T) (string, string) {
				testing.SetOriginCaller(routerAddr)
				actualAmount0, actualAmount1 := Swap(
					cross,
					wugnotPath,
					gnsPath,
					fee3000,
					alice,
					true,
					"200",
					consts.MIN_PRICE,
					addr,
				)
				return actualAmount0, actualAmount1
			},
			shouldPanic: false,
			expected:    []string{"200", "-198"},
		},
		{
			name: "EXACT OUT - zeroForOne = true, token1:-200",
			setupFn: func(t *testing.T) {
				InitialisePoolTest(t)
				MintPositionAll(t, adminAddr)
				TokenFaucet(t, gnsPath, addr)
				TokenFaucet(t, wugnotPath, addr)
				TokenApprove(t, gnsPath, addr, poolAddr, maxApprove)
				TokenApprove(t, wugnotPath, addr, poolAddr, maxApprove)
			},
			action: func(t *testing.T) (string, string) {
				testing.SetOriginCaller(routerAddr)
				actualAmount0, actualAmount1 := Swap(
					cross,
					wugnotPath,
					gnsPath,
					fee3000,
					alice,
					true,
					"-200",
					consts.MIN_PRICE,
					addr,
				)
				return actualAmount0, actualAmount1
			},
			shouldPanic: false,
			expected:    []string{"202", "-200"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetObject(t)
			burnTokens(t)

			if tt.setupFn != nil {
				tt.setupFn(t)
			}

			if tt.shouldPanic {
				tt.action(t)
			} else {
				amount0, amount1 := tt.action(t)
				uassert.Equal(t, tt.expected[0], amount0)
				uassert.Equal(t, tt.expected[1], amount1)
			}
		})
	}
}

func TestSwap_FeeTier100(t *testing.T) {
	const wugnotTokenPath = "gno.land/r/demo/wugnot"
	const gnsTokenPath = "gno.land/r/gnoswap/v1/gns"

	tests := []struct {
		name        string
		setupFn     func(t *testing.T)
		action      func(t *testing.T) (string, string)
		shouldPanic bool
		expected    []string
	}{
		{
			name: "EXACT IN - zeroForOne = true, feeTier=100",
			setupFn: func(t *testing.T) {
				testing.SetRealm(std.NewUserRealm(adminAddr))
				wugnot.Approve(cross, poolAddr, maxApprove)
				gns.Approve(cross, poolAddr, maxApprove)
				CreatePool(cross, wugnotTokenPath, gnsTokenPath, FeeTier100, "79228162514264337593543950336")
				wugnot.Approve(cross, routerAddr, maxApprove)
				gns.Approve(cross, routerAddr, maxApprove)

				faucetAddress := adminAddr
				testing.SetOriginCaller(faucetAddress)
				newCoins := std.Coins{{"ugnot", int64(10000000000)}}
				testing.IssueCoins(adminAddr, newCoins)
				testing.SetOriginSend(newCoins)
				banker := std.NewBanker(std.BankerTypeRealmSend)
				banker.SendCoins(adminAddr, consts.WUGNOT_ADDR, newCoins)
				wugnot.Deposit(cross)

				testing.SetRealm(posRealm)
				Mint(
					cross,
					wugnotTokenPath,
					gnsTokenPath,
					FeeTier100,
					adminAddr,
					-100,
					100,
					"10000000",
					adminAddr,
				)

				Mint(
					cross,
					wugnotTokenPath,
					gnsTokenPath,
					FeeTier100,
					adminAddr,
					-100,
					100,
					"10000000",
					adminAddr,
				)
				testing.SetRealm(std.NewUserRealm(adminAddr))
				gns.Transfer(cross, poolAddr, 20000000)
			},
			action: func(t *testing.T) (string, string) {
				testing.SetOriginCaller(adminAddr)
				wugnot.Transfer(cross, routerAddr, 20000000)
				gns.Transfer(cross, routerAddr, 20000000)

				testing.SetOriginCaller(routerAddr)
				wugnot.Approve(cross, poolAddr, maxApprove)

				actualAmount0, actualAmount1 := Swap(
					cross,
					wugnotTokenPath,
					gnsTokenPath,
					FeeTier100,
					adminAddr,
					true,
					"40012000",
					common.TickMathGetSqrtRatioAtTick(-887220).ToString(),
					adminAddr,
				)
				return actualAmount0, actualAmount1
			},
			shouldPanic: false,
			expected:    []string{"100257", "-99745"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFn != nil {
				tt.setupFn(t)
			}

			if tt.shouldPanic {
				tt.action(t)
			} else {
				amount0, amount1 := tt.action(t)
				uassert.Equal(t, tt.expected[0], amount0)
				uassert.Equal(t, tt.expected[1], amount1)
			}
		})
	}
}
