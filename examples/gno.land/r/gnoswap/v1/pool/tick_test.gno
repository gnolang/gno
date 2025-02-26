package pool

import (
	"testing"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/uassert"

	i256 "gno.land/p/gnoswap/int256"
	u256 "gno.land/p/gnoswap/uint256"
)

func TestcalculateMaxLiquidityPerTick(t *testing.T) {
	tests := []struct {
		name        string
		tickSpacing int32
		want        string // expected result in string format
	}{
		{
			name:        "tick spacing 1 (for 0.01% pool)",
			tickSpacing: 1,
			want:        "191757530477355301479181766273477",
		},
		{
			name:        "tick spacing 10 (for 0.05% pool)",
			tickSpacing: 10,
			want:        "1917569901783203986719870431555990",
		},
		{
			name:        "tick spacing 60 (for 0.3% pool)",
			tickSpacing: 60,
			want:        "11505743598341114571880798222544994",
		},
		{
			name:        "tick spacing 200 (for 1% pool)",
			tickSpacing: 200,
			want:        "38350317471085141830651933667504588",
		},
		{
			name:        "entire range",
			tickSpacing: 887272,
			want:        "113427455640312821154458202477256070485",
		},
		{
			name:        "custom tick spacing",
			tickSpacing: 2302,
			want:        "441351967472034323558203122479595605",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateMaxLiquidityPerTick(tt.tickSpacing)
			if got.ToString() != tt.want {
				t.Errorf("calculateMaxLiquidityPerTick() = %v, want %v", got.ToString(), tt.want)
			}
		})
	}
}

func TestCalculateFeeGrowthInside(t *testing.T) {
	// Create a mock pool
	pool := &Pool{
		ticks: avl.NewTree(),
	}

	// Setup test ticks
	pool.ticks.Set("0", TickInfo{
		liquidityGross:        u256.NewUint(1000),
		liquidityNet:          i256.NewInt(100),
		feeGrowthOutside0X128: u256.NewUint(5),
		feeGrowthOutside1X128: u256.NewUint(7),
		initialized:           true,
	})
	pool.ticks.Set("100", TickInfo{
		liquidityGross:        u256.NewUint(2000),
		liquidityNet:          i256.NewInt(-100),
		feeGrowthOutside0X128: u256.NewUint(10),
		feeGrowthOutside1X128: u256.NewUint(15),
		initialized:           true,
	})

	tests := []struct {
		name                 string
		tickLower            int32
		tickUpper            int32
		tickCurrent          int32
		feeGrowthGlobal0X128 *u256.Uint
		feeGrowthGlobal1X128 *u256.Uint
		want0                string
		want1                string
		preconditions        func()
	}{
		{
			name:                 "returns all for two uninitialized ticks if tick is inside",
			tickLower:            -2,
			tickUpper:            2,
			tickCurrent:          0,
			feeGrowthGlobal0X128: u256.NewUint(15),
			feeGrowthGlobal1X128: u256.NewUint(15),
			want0:                "15",
			want1:                "15",
		},
		{
			name:                 "returns 0 for two uninitialized ticks if tick is above",
			tickLower:            -2,
			tickUpper:            2,
			tickCurrent:          4,
			feeGrowthGlobal0X128: u256.NewUint(15),
			feeGrowthGlobal1X128: u256.NewUint(15),
			want0:                "0",
			want1:                "0",
		},
		{
			name:                 "returns 0 for two uninitialized ticks if tick is below",
			tickLower:            -2,
			tickUpper:            2,
			tickCurrent:          4,
			feeGrowthGlobal0X128: u256.NewUint(15),
			feeGrowthGlobal1X128: u256.NewUint(15),
			want0:                "0",
			want1:                "0",
		},
		{
			name:                 "subtracts upper tick if below",
			tickLower:            -2,
			tickUpper:            2,
			tickCurrent:          0,
			feeGrowthGlobal0X128: u256.NewUint(15),
			feeGrowthGlobal1X128: u256.NewUint(15),
			want0:                "13",
			want1:                "12",
			preconditions: func() {
				setTick(
					t,
					pool,
					2,
					u256.NewUint(2),
					u256.NewUint(3),
					u256.NewUint(0),
					i256.NewInt(0),
					u256.NewUint(0),
					0,
					0,
					true,
				)
			},
		},
		{
			name:                 "subtracts lower tick if below",
			tickLower:            -2,
			tickUpper:            2,
			tickCurrent:          0,
			feeGrowthGlobal0X128: u256.NewUint(15),
			feeGrowthGlobal1X128: u256.NewUint(15),
			want0:                "13",
			want1:                "12",
			preconditions: func() {
				deleteTick(t, pool, 2) // delete tick from previous test
				setTick(
					t,
					pool,
					-2,
					u256.NewUint(2),
					u256.NewUint(3),
					u256.NewUint(0),
					i256.NewInt(0),
					u256.NewUint(0),
					0,
					0,
					true,
				)
			},
		},
		{
			name:                 "subtracts upper and lower tick if inside",
			tickLower:            -2,
			tickUpper:            2,
			tickCurrent:          0,
			feeGrowthGlobal0X128: u256.NewUint(15),
			feeGrowthGlobal1X128: u256.NewUint(15),
			want0:                "9",
			want1:                "11",
			preconditions: func() {
				// we already have tick -2
				setTick(
					t,
					pool,
					2,
					u256.NewUint(4),
					u256.NewUint(1),
					u256.NewUint(0),
					i256.NewInt(0),
					u256.NewUint(0),
					0,
					0,
					true,
				)
			},
		},
		{
			name:                 "works correctly with overflow on inside tick",
			tickLower:            -2,
			tickUpper:            2,
			tickCurrent:          0,
			feeGrowthGlobal0X128: u256.NewUint(15),
			feeGrowthGlobal1X128: u256.NewUint(15),
			want0:                "16",
			want1:                "13",
			preconditions: func() {
				deleteTick(t, pool, 2)
				deleteTick(t, pool, -2)
				setTick(
					t,
					pool,
					-2,
					u256.MustFromDecimal("115792089237316195423570985008687907853269984665640564039457584007913129639932"), // max uint256 - 3
					u256.MustFromDecimal("115792089237316195423570985008687907853269984665640564039457584007913129639933"), // max uint256 - 2
					u256.NewUint(0),
					i256.NewInt(0),
					u256.NewUint(0),
					0,
					0,
					true,
				)
				setTick(
					t,
					pool,
					2,
					u256.NewUint(3),
					u256.NewUint(5),
					u256.NewUint(0),
					i256.NewInt(0),
					u256.NewUint(0),
					0,
					0,
					true,
				)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.preconditions != nil {
				tt.preconditions()
			}
			got0, got1 := pool.getFeeGrowthInside(
				tt.tickLower,
				tt.tickUpper,
				tt.tickCurrent,
				tt.feeGrowthGlobal0X128,
				tt.feeGrowthGlobal1X128,
			)
			if got0.ToString() != tt.want0 || got1.ToString() != tt.want1 {
				t.Errorf("getFeeGrowthInside() = (%v, %v), want (%v, %v)",
					got0.ToString(), got1.ToString(), tt.want0, tt.want1)
			}
		})
	}
}

func TestTickUpdate(t *testing.T) {
	pool := &Pool{
		ticks: avl.NewTree(),
	}

	tests := []struct {
		name                 string
		preconditions        func()
		tick                 int32
		tickCurrent          int32
		liquidityDelta       *i256.Int
		feeGrowthGlobal0X128 *u256.Uint
		feeGrowthGlobal1X128 *u256.Uint
		upper                bool
		maxLiquidity         *u256.Uint
		wantFlipped          bool
		shouldPanic          bool
		verify               func()
	}{
		{
			name:                 "flips from zero to non zero",
			tick:                 0,
			tickCurrent:          0,
			liquidityDelta:       i256.One(),
			feeGrowthGlobal0X128: u256.Zero(),
			feeGrowthGlobal1X128: u256.Zero(),
			upper:                false,
			maxLiquidity:         u256.NewUint(3),
			wantFlipped:          true,
		},
		{
			name: "does not flip from nonzero to greater nonzero",
			preconditions: func() {
				deleteTick(t, pool, 0)
				pool.tickUpdate(0, 0, i256.One(), u256.Zero(), u256.Zero(), false, u256.NewUint(3))
			},
			tick:                 0,
			tickCurrent:          0,
			liquidityDelta:       i256.One(),
			feeGrowthGlobal0X128: u256.Zero(),
			feeGrowthGlobal1X128: u256.Zero(),
			upper:                false,
			maxLiquidity:         u256.NewUint(3),
			wantFlipped:          false,
		},
		{
			name: "flips from nonzero to zero",
			preconditions: func() {
				deleteTick(t, pool, 0)
				pool.tickUpdate(0, 0, i256.One(), u256.Zero(), u256.Zero(), false, u256.NewUint(3))
			},
			tick:                 0,
			tickCurrent:          0,
			liquidityDelta:       i256.NewInt(-1),
			feeGrowthGlobal0X128: u256.Zero(),
			feeGrowthGlobal1X128: u256.Zero(),
			upper:                false,
			maxLiquidity:         u256.NewUint(3),
			wantFlipped:          true,
		},
		{
			name: "does not flip from nonzero to lesser nonzero",
			preconditions: func() {
				deleteTick(t, pool, 0)
				pool.tickUpdate(0, 0, i256.NewInt(2), u256.Zero(), u256.Zero(), false, u256.NewUint(3))
			},
			tick:                 0,
			tickCurrent:          0,
			liquidityDelta:       i256.NewInt(-1),
			feeGrowthGlobal0X128: u256.Zero(),
			feeGrowthGlobal1X128: u256.Zero(),
			upper:                false,
			maxLiquidity:         u256.NewUint(3),
			wantFlipped:          false,
		},
		{
			name: "reverts if total liquidity gross is greater than max",
			preconditions: func() {
				deleteTick(t, pool, 0)
				pool.tickUpdate(0, 0, i256.NewInt(2), u256.Zero(), u256.Zero(), false, u256.NewUint(3))
				pool.tickUpdate(0, 0, i256.One(), u256.Zero(), u256.Zero(), true, u256.NewUint(3))
			},
			tick:                 0,
			tickCurrent:          0,
			liquidityDelta:       i256.One(),
			feeGrowthGlobal0X128: u256.Zero(),
			feeGrowthGlobal1X128: u256.Zero(),
			upper:                false,
			maxLiquidity:         u256.NewUint(3),
			wantFlipped:          false,
			shouldPanic:          true,
		},
		{
			name: "nets the liquidity based on upper flag",
			preconditions: func() {
				pool.tickUpdate(0, 0, i256.NewInt(2), u256.Zero(), u256.Zero(), false, u256.NewUint(10))
				pool.tickUpdate(0, 0, i256.One(), u256.Zero(), u256.Zero(), true, u256.NewUint(10))
				pool.tickUpdate(0, 0, i256.NewInt(3), u256.Zero(), u256.Zero(), true, u256.NewUint(10))
				pool.tickUpdate(0, 0, i256.One(), u256.Zero(), u256.Zero(), false, u256.NewUint(10))
			},
			tick:                 0,
			tickCurrent:          0,
			liquidityDelta:       i256.One(),
			feeGrowthGlobal0X128: u256.Zero(),
			feeGrowthGlobal1X128: u256.Zero(),
			upper:                false,
			maxLiquidity:         u256.NewUint(3),
			wantFlipped:          false,
			shouldPanic:          true,
		},
		{
			name: "reverts on overflow liquidity gross",
			preconditions: func() {
				pool.tickUpdate(0, 0, i256.MustFromDecimal("170141183460469231731687303715884105726"), u256.Zero(), u256.Zero(), false, u256.MustFromDecimal("340282366920938463463374607431768211455"))
			},
			tick:                 0,
			tickCurrent:          0,
			liquidityDelta:       i256.MustFromDecimal("170141183460469231731687303715884105726"), // (maxUint128 / 2) + 1
			feeGrowthGlobal0X128: u256.Zero(),
			feeGrowthGlobal1X128: u256.Zero(),
			upper:                false,
			maxLiquidity:         u256.MustFromDecimal("340282366920938463463374607431768211455"), // maxUint128
			wantFlipped:          false,
			shouldPanic:          true,
		},
		{
			name: "assumes all growth happens below ticks lte current tick",
			preconditions: func() {
				deleteTick(t, pool, 0)
				deleteTick(t, pool, 1)
				pool.tickUpdate(1, 1, i256.One(), u256.One(), u256.NewUint(2), false, u256.MustFromDecimal("340282366920938463463374607431768211455"))
			},
			tick:                 0,
			tickCurrent:          0,
			liquidityDelta:       i256.MustFromDecimal("170141183460469231731687303715884105726"),
			feeGrowthGlobal0X128: u256.Zero(),
			feeGrowthGlobal1X128: u256.Zero(),
			upper:                false,
			maxLiquidity:         u256.MustFromDecimal("340282366920938463463374607431768211455"),
			wantFlipped:          false,
			shouldPanic:          false,
			verify: func() {
				info := pool.mustGetTick(1)
				uassert.Equal(t, info.feeGrowthOutside0X128.ToString(), "1")
				uassert.Equal(t, info.feeGrowthOutside1X128.ToString(), "2")
			},
		},
		{
			name: "does not set any growth fields if tick is already initialized",
			preconditions: func() {
				pool.tickUpdate(1, 1, i256.One(), u256.One(), u256.NewUint(2), false, u256.MustFromDecimal("340282366920938463463374607431768211455"))
				pool.tickUpdate(1, 1, i256.One(), u256.NewUint(6), u256.NewUint(7), false, u256.MustFromDecimal("340282366920938463463374607431768211455"))
			},
			tick:                 1,
			tickCurrent:          1,
			liquidityDelta:       i256.One(),
			feeGrowthGlobal0X128: u256.NewUint(6),
			feeGrowthGlobal1X128: u256.NewUint(7),
			upper:                false,
			maxLiquidity:         u256.MustFromDecimal("340282366920938463463374607431768211455"),
			wantFlipped:          false,
			shouldPanic:          false,
			verify: func() {
				info := pool.mustGetTick(1)
				uassert.Equal(t, info.feeGrowthOutside0X128.ToString(), "1")
				uassert.Equal(t, info.feeGrowthOutside1X128.ToString(), "2")
			},
		},
		{
			name:                 "does not set any growth fields for ticks gt current tick",
			tick:                 2,
			tickCurrent:          1,
			liquidityDelta:       i256.One(),
			feeGrowthGlobal0X128: u256.NewUint(1),
			feeGrowthGlobal1X128: u256.NewUint(2),
			upper:                false,
			maxLiquidity:         u256.MustFromDecimal("340282366920938463463374607431768211455"),
			wantFlipped:          false,
			shouldPanic:          false,
			verify: func() {
				info := pool.getTick(2)
				uassert.Equal(t, info.feeGrowthOutside0X128.ToString(), "0")
				uassert.Equal(t, info.feeGrowthOutside1X128.ToString(), "0")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.preconditions != nil {
				tt.preconditions()
			}

			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("expected panic, but got nil")
					}
				}()

				pool.tickUpdate(
					tt.tick,
					tt.tickCurrent,
					tt.liquidityDelta,
					tt.feeGrowthGlobal0X128,
					tt.feeGrowthGlobal1X128,
					tt.upper,
					tt.maxLiquidity,
				)
			} else {
				if tt.verify != nil {
					tt.verify()
				} else {
					gotFlipped := pool.tickUpdate(
						tt.tick,
						tt.tickCurrent,
						tt.liquidityDelta,
						tt.feeGrowthGlobal0X128,
						tt.feeGrowthGlobal1X128,
						tt.upper,
						tt.maxLiquidity,
					)
					if gotFlipped != tt.wantFlipped {
						t.Errorf("tickUpdate() flipped = %v, want %v", gotFlipped, tt.wantFlipped)
					}
				}
			}
		})
	}
}

func TestTickCross(t *testing.T) {
	pool := &Pool{
		ticks: avl.NewTree(),
	}

	// Setup initial tick state
	pool.ticks.Set("100", TickInfo{
		liquidityGross:        u256.NewUint(1000),
		liquidityNet:          i256.NewInt(500),
		feeGrowthOutside0X128: u256.NewUint(10),
		feeGrowthOutside1X128: u256.NewUint(15),
		initialized:           true,
	})

	tests := []struct {
		name                 string
		tick                 int32
		feeGrowthGlobal0X128 *u256.Uint
		feeGrowthGlobal1X128 *u256.Uint
		wantLiquidityNet     string
	}{
		{
			name:                 "cross tick upwards",
			tick:                 100,
			feeGrowthGlobal0X128: u256.NewUint(20),
			feeGrowthGlobal1X128: u256.NewUint(25),
			wantLiquidityNet:     "500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLiquidityNet := pool.tickCross(
				tt.tick,
				tt.feeGrowthGlobal0X128,
				tt.feeGrowthGlobal1X128,
			)
			if gotLiquidityNet.ToString() != tt.wantLiquidityNet {
				t.Errorf("tickCross() liquidityNet = %v, want %v",
					gotLiquidityNet.ToString(), tt.wantLiquidityNet)
			}
		})
	}
}

func TestGetTick(t *testing.T) {
	pool := &Pool{
		ticks: avl.NewTree(),
	}

	// Setup a tick
	expectedTick := TickInfo{
		liquidityGross:        u256.NewUint(1000),
		liquidityNet:          i256.NewInt(500),
		feeGrowthOutside0X128: u256.NewUint(10),
		feeGrowthOutside1X128: u256.NewUint(15),
		initialized:           true,
	}
	pool.setTick(50, expectedTick)

	tests := []struct {
		name     string
		tick     int32
		wantInit bool
	}{
		{
			name:     "get existing tick",
			tick:     50,
			wantInit: true,
		},
		{
			name:     "get non-existing tick",
			tick:     100,
			wantInit: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pool.getTick(tt.tick)
			if got.initialized != tt.wantInit {
				t.Errorf("getTick() initialized = %v, want %v", got.initialized, tt.wantInit)
			}
		})
	}
}

func TestGetFeeGrowthBelowX128(t *testing.T) {
	// Setup test data
	globalFeeGrowth0 := u256.NewUint(1000) // Global fee growth for token 0
	globalFeeGrowth1 := u256.NewUint(2000) // Global fee growth for token 1

	lowerTick := TickInfo{
		feeGrowthOutside0X128: u256.NewUint(300), // fee growth outside for token 0
		feeGrowthOutside1X128: u256.NewUint(500), // fee growth outside for token 1
	}

	tests := []struct {
		name               string
		tickLower          int32
		tickCurrent        int32
		expectedFeeGrowth0 *u256.Uint
		expectedFeeGrowth1 *u256.Uint
	}{
		{
			name:               "tickCurrent >= tickLower - Return feeGrowthOutside directly",
			tickLower:          100,
			tickCurrent:        100,
			expectedFeeGrowth0: lowerTick.feeGrowthOutside0X128,
			expectedFeeGrowth1: lowerTick.feeGrowthOutside1X128,
		},
		{
			name:               "tickCurrent > tickLower - Return feeGrowthOutside directly",
			tickLower:          50,
			tickCurrent:        100,
			expectedFeeGrowth0: lowerTick.feeGrowthOutside0X128,
			expectedFeeGrowth1: lowerTick.feeGrowthOutside1X128,
		},
		{
			name:               "tickCurrent < tickLower - Subtract feeGrowthOutside from global",
			tickLower:          100,
			tickCurrent:        50,
			expectedFeeGrowth0: u256.NewUint(700),  // 1000 - 300
			expectedFeeGrowth1: u256.NewUint(1500), // 2000 - 500
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the function
			feeGrowth0, feeGrowth1 := getFeeGrowthBelowX128(
				tt.tickLower, tt.tickCurrent,
				globalFeeGrowth0, globalFeeGrowth1,
				lowerTick,
			)

			// Assertions
			uassert.True(t, feeGrowth0.Eq(tt.expectedFeeGrowth0),
				"Expected feeGrowth0: %s, got: %s", tt.expectedFeeGrowth0.ToString(), feeGrowth0.ToString())
			uassert.True(t, feeGrowth1.Eq(tt.expectedFeeGrowth1),
				"Expected feeGrowth1: %s, got: %s", tt.expectedFeeGrowth1.ToString(), feeGrowth1.ToString())
		})
	}
}

func TestGetFeeGrowthAboveX128(t *testing.T) {
	// Setup test data
	globalFeeGrowth0 := u256.NewUint(1000) // Global fee growth for token 0
	globalFeeGrowth1 := u256.NewUint(2000) // Global fee growth for token 1

	upperTick := TickInfo{
		feeGrowthOutside0X128: u256.NewUint(300), // Fee growth outside for token 0
		feeGrowthOutside1X128: u256.NewUint(500), // Fee growth outside for token 1
	}

	tests := []struct {
		name               string
		tickUpper          int32
		tickCurrent        int32
		expectedFeeGrowth0 *u256.Uint
		expectedFeeGrowth1 *u256.Uint
	}{
		{
			name:               "tickCurrent < tickUpper - Return feeGrowthOutside directly",
			tickUpper:          100,
			tickCurrent:        50,
			expectedFeeGrowth0: upperTick.feeGrowthOutside0X128, // 300
			expectedFeeGrowth1: upperTick.feeGrowthOutside1X128, // 500
		},
		{
			name:               "tickCurrent >= tickUpper - Subtract feeGrowthOutside from global",
			tickUpper:          100,
			tickCurrent:        150,
			expectedFeeGrowth0: u256.NewUint(700),  // 1000 - 300
			expectedFeeGrowth1: u256.NewUint(1500), // 2000 - 500
		},
		{
			name:               "tickCurrent == tickUpper - Subtract feeGrowthOutside from global",
			tickUpper:          100,
			tickCurrent:        100,
			expectedFeeGrowth0: u256.NewUint(700),  // 1000 - 300
			expectedFeeGrowth1: u256.NewUint(1500), // 2000 - 500
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the function
			feeGrowth0, feeGrowth1 := getFeeGrowthAboveX128(
				tt.tickUpper, tt.tickCurrent,
				globalFeeGrowth0, globalFeeGrowth1,
				upperTick,
			)

			// Assertions
			uassert.True(t, feeGrowth0.Eq(tt.expectedFeeGrowth0),
				"Expected feeGrowth0: %s, got: %s", tt.expectedFeeGrowth0.ToString(), feeGrowth0.ToString())
			uassert.True(t, feeGrowth1.Eq(tt.expectedFeeGrowth1),
				"Expected feeGrowth1: %s, got: %s", tt.expectedFeeGrowth1.ToString(), feeGrowth1.ToString())
		})
	}
}

func setTick(
	t *testing.T,
	pool *Pool,
	tick int32,
	feeGrowthOutside0X128 *u256.Uint,
	feeGrowthOutside1X128 *u256.Uint,
	liquidityGross *u256.Uint,
	liquidityNet *i256.Int,
	secondsPerLiquidityOutsideX128 *u256.Uint,
	tickCumulativeOutside int64,
	secondsOutside uint32,
	initialized bool,
) {
	t.Helper()

	info := pool.getTick(tick)
	info.feeGrowthOutside0X128 = feeGrowthOutside0X128
	info.feeGrowthOutside1X128 = feeGrowthOutside1X128
	info.liquidityGross = liquidityGross
	info.liquidityNet = liquidityNet
	info.secondsPerLiquidityOutsideX128 = secondsPerLiquidityOutsideX128
	info.tickCumulativeOutside = tickCumulativeOutside
	info.secondsOutside = secondsOutside
	info.initialized = initialized

	pool.setTick(tick, info)
}

func deleteTick(t *testing.T, pool *Pool, tick int32) {
	t.Helper()
	pool.deleteTick(tick)
}
