package gnsmath

import (
	"testing"

	"gno.land/p/demo/uassert"

	"gno.land/p/gnoswap/consts"
	i256 "gno.land/p/gnoswap/int256"
	u256 "gno.land/p/gnoswap/uint256"
)

func TestSqrtPriceMathGetNextSqrtPriceFromAmount0RoundingUp(t *testing.T) {
	t.Run("zero amount returns same price", func(t *testing.T) {
		sqrtPX96 := u256.MustFromDecimal("1000000")
		liquidity := u256.MustFromDecimal("2000000")
		amount := u256.Zero()

		result := getNextSqrtPriceFromAmount0RoundingUp(
			sqrtPX96,
			liquidity,
			amount,
			true,
		)

		if !result.Eq(sqrtPX96) {
			t.Errorf("Expected %s, got %s", sqrtPX96.ToString(), result.ToString())
		}
	})

	t.Run("remove token0", func(t *testing.T) {
		sqrtPX96 := u256.MustFromDecimal("1000000")
		liquidity := u256.MustFromDecimal("2000000")
		amount := u256.MustFromDecimal("500000")

		result := getNextSqrtPriceFromAmount0RoundingUp(
			sqrtPX96,
			liquidity,
			amount,
			false,
		)

		if result.Lte(sqrtPX96) {
			t.Error("Price should increase when removing token0")
		}
	})
}

func TestSqrtPriceMathGetNextSqrtPriceFromAmount1RoundingDown(t *testing.T) {
	t.Run("add token1 small amount", func(t *testing.T) {
		sqrtPX96 := u256.MustFromDecimal("1000000")
		liquidity := u256.MustFromDecimal("2000000")
		amount := u256.MustFromDecimal("100000")

		result := getNextSqrtPriceFromAmount1RoundingDown(
			sqrtPX96,
			liquidity,
			amount,
			true,
		)

		if result.Lte(sqrtPX96) {
			t.Error("Price should increase when adding token1")
		}
	})
}

func TestSqrtPriceMathGetAmount0DeltaStr(t *testing.T) {
	t.Run("positive liquidity", func(t *testing.T) {
		ratioA := u256.MustFromDecimal("1000000")
		ratioB := u256.MustFromDecimal("2000000")
		liquidity := i256.FromUint256(u256.MustFromDecimal("5000000"))

		result := GetAmount0DeltaStr(ratioA, ratioB, liquidity)

		if result[0] == '-' {
			t.Error("Result should be positive for positive liquidity")
		}
	})

	t.Run("negative liquidity", func(t *testing.T) {
		ratioA := u256.MustFromDecimal("1000000")
		ratioB := u256.MustFromDecimal("2000000")
		liquidity := i256.Zero().Neg(i256.FromUint256(u256.MustFromDecimal("5000000")))

		result := GetAmount0DeltaStr(ratioA, ratioB, liquidity)

		if result[0] != '-' {
			t.Error("Result should be negative for negative liquidity")
		}
	})

	t.Run("panic overflow when getting amount0 with positive liquidity", func(t *testing.T) {
		// Inputs to trigger panic
		sqrtRatioAX96 := u256.MustFromDecimal("1")                                       // very low value
		sqrtRatioBX96 := u256.MustFromDecimal("340282366920938463463374607431768211455") // very high value(2^128-1)
		liquidity := i256.FromUint256(u256.MustFromDecimal("115792089237316195423570985008687907853269984665640564039457584007913129639935"))

		uassert.PanicsWithMessage(t, errAmount0DeltaOverflow.Error(), func() {
			GetAmount0DeltaStr(sqrtRatioAX96, sqrtRatioBX96, liquidity)
		})
	})

	t.Run("panic overflow when getting amount0 with negative liquidity", func(t *testing.T) {
		// Inputs to trigger panic
		sqrtRatioAX96 := u256.MustFromDecimal("1")                                       // very low value
		sqrtRatioBX96 := u256.MustFromDecimal("340282366920938463463374607431768211455") // very high value(2^128-1)
		liquidity := i256.FromUint256(u256.MustFromDecimal("115792089237316195423570985008687907853269984665640564039457584007913129639935"))
		liquidity = liquidity.Neg(liquidity) // Make liquidity negative

		uassert.PanicsWithMessage(t, errAmount0DeltaOverflow.Error(), func() {
			GetAmount0DeltaStr(sqrtRatioAX96, sqrtRatioBX96, liquidity)
		})
	})
}

func TestSqrtPriceMathGetAmount1DeltaStr(t *testing.T) {
	t.Run("positive liquidity", func(t *testing.T) {
		ratioA := u256.MustFromDecimal("1000000")
		ratioB := u256.MustFromDecimal("2000000")
		liquidity := i256.FromUint256(u256.MustFromDecimal("5000000"))

		result := GetAmount1DeltaStr(ratioA, ratioB, liquidity)

		if result[0] == '-' {
			t.Error("Result should be positive for positive liquidity")
		}
	})

	t.Run("negative liquidity", func(t *testing.T) {
		ratioA := u256.MustFromDecimal("1000000")
		ratioB := u256.MustFromDecimal("2000000")
		liquidity := i256.Zero().Neg(i256.FromUint256(u256.MustFromDecimal("5000000")))

		result := GetAmount0DeltaStr(ratioA, ratioB, liquidity)

		if result[0] != '-' {
			t.Error("Result should be negative for negative liquidity")
		}
	})

	t.Run("panic overflow when getting amount1 with positive liquidity", func(t *testing.T) {
		// Inputs to trigger panic
		sqrtRatioAX96 := u256.MustFromDecimal("1")                             // very low value
		sqrtRatioBX96 := u256.MustFromDecimal("79228162514264337593543950335") // slightly below Q96
		liquidity := i256.FromUint256(u256.MustFromDecimal("115792089237316195423570985008687907853269984665640564039457584007913129639935"))

		uassert.PanicsWithMessage(t, errAmount1DeltaOverflow.Error(), func() {
			GetAmount1DeltaStr(sqrtRatioAX96, sqrtRatioBX96, liquidity)
		})
	})

	t.Run("panic overflow when getting amount1 with negative liquidity", func(t *testing.T) {
		// Inputs to trigger panic
		sqrtRatioAX96 := u256.MustFromDecimal("1")                             // very low value
		sqrtRatioBX96 := u256.MustFromDecimal("79228162514264337593543950335") // slightly below Q96
		liquidity := i256.FromUint256(u256.MustFromDecimal("115792089237316195423570985008687907853269984665640564039457584007913129639935"))
		liquidity = liquidity.Neg(liquidity) // Make liquidity negative

		uassert.PanicsWithMessage(t, errAmount1DeltaOverflow.Error(), func() {
			GetAmount1DeltaStr(sqrtRatioAX96, sqrtRatioBX96, liquidity)
		})
	})
}

func TestSqrtPriceMathGetNextSqrtPriceFromInput(t *testing.T) {
	tests := []struct {
		name                 string
		sqrtPriceX96         *u256.Uint
		liquidity            *u256.Uint
		amountIn             *u256.Uint
		zeroForOne           bool
		shouldPanic          bool
		panicMsg             string
		expectedSqrtPriceX96 string
	}{
		{
			name:         "fails if price is zero",
			sqrtPriceX96: u256.Zero(),
			liquidity:    u256.Zero(),
			amountIn:     u256.MustFromDecimal("100000000000000000"),
			zeroForOne:   false,
			shouldPanic:  true,
			panicMsg:     "sqrtPX96 should not be zero",
		},
		{
			name:         "fails if liquidity is zero",
			sqrtPriceX96: u256.One(),
			liquidity:    u256.Zero(),
			amountIn:     u256.MustFromDecimal("100000000000000000"),
			zeroForOne:   true,
			shouldPanic:  true,
			panicMsg:     "liquidity should not be zero",
		},
		{
			name:         "fails if input amount overflows the price",
			sqrtPriceX96: u256.MustFromDecimal("1461501637330902918203684832716283019655932542975"), // 2^160 - 1
			liquidity:    u256.MustFromDecimal("1024"),
			amountIn:     u256.MustFromDecimal("1024"),
			zeroForOne:   false,
			shouldPanic:  true,
			panicMsg:     "GetNextSqrtPriceFromAmount1RoundingDown sqrtPx96 + quotient overflow uint160",
		},
		{
			name:                 "any input amount cannot underflow the price",
			sqrtPriceX96:         u256.MustFromDecimal("1"),
			liquidity:            u256.MustFromDecimal("1"),
			amountIn:             u256.MustFromDecimal("57896044618658097711785492504343953926634992332820282019728792003956564819968"), // 2^255
			zeroForOne:           true,
			expectedSqrtPriceX96: "1",
		},
		{
			name:                 "returns input price if amount in is zero and zeroForOne = true",
			sqrtPriceX96:         u256.MustFromDecimal("79228162514264337593543950336"),
			liquidity:            u256.MustFromDecimal("100000000000000000"),
			amountIn:             u256.Zero(),
			zeroForOne:           true,
			expectedSqrtPriceX96: "79228162514264337593543950336",
		},
		{
			name:                 "returns input price if amount in is zero and zeroForOne = false",
			sqrtPriceX96:         u256.MustFromDecimal("79228162514264337593543950336"),
			liquidity:            u256.MustFromDecimal("100000000000000000"),
			amountIn:             u256.Zero(),
			zeroForOne:           false,
			expectedSqrtPriceX96: "79228162514264337593543950336",
		},
		{
			name:                 "returns the minimum price for max inputs",
			sqrtPriceX96:         u256.MustFromDecimal("1461501637330902918203684832716283019655932542975"), // 2^160 - 1
			liquidity:            u256.MustFromDecimal("340282366920938463463374607431768211455"),           // 2^128 - 1
			amountIn:             u256.MustFromDecimal("115792089237316195423570985008687907853269984665640564039439137263839420088320"),
			zeroForOne:           true,
			expectedSqrtPriceX96: "1",
		},
		{
			name:                 "input amount of 0.1 token1",
			sqrtPriceX96:         encodePriceSqrt("1", "1"),
			liquidity:            u256.MustFromDecimal("1000000000000000000"),
			amountIn:             u256.MustFromDecimal("100000000000000000"),
			zeroForOne:           false,
			expectedSqrtPriceX96: "87150978765690771352898345369",
		},
		{
			name:                 "input amount of 0.1 token0",
			sqrtPriceX96:         encodePriceSqrt("1", "1"),
			liquidity:            u256.MustFromDecimal("1000000000000000000"),
			amountIn:             u256.MustFromDecimal("100000000000000000"),
			zeroForOne:           true,
			expectedSqrtPriceX96: "72025602285694852357767227579",
		},
		{
			name:                 "amountIn > type(uint96).max and zeroForOne = true",
			sqrtPriceX96:         encodePriceSqrt("1", "1"),
			liquidity:            u256.MustFromDecimal("10000000000000000000"),
			amountIn:             u256.MustFromDecimal("1267650600228229401496703205376"), // 2^128 - 1
			zeroForOne:           true,
			expectedSqrtPriceX96: "624999999995069620",
		},
		{
			name:                 "can return 1 with enough amountIn and zeroForOne = true",
			sqrtPriceX96:         encodePriceSqrt("1", "1"),
			liquidity:            u256.One(),
			amountIn:             u256.MustFromDecimal("57896044618658097711785492504343953926634992332820282019728792003956564819967"),
			zeroForOne:           true,
			expectedSqrtPriceX96: "1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				uassert.PanicsWithMessage(t, tt.panicMsg, func() {
					getNextSqrtPriceFromInput(tt.sqrtPriceX96, tt.liquidity, tt.amountIn, tt.zeroForOne)
				})
			} else {
				actual := getNextSqrtPriceFromInput(tt.sqrtPriceX96, tt.liquidity, tt.amountIn, tt.zeroForOne)
				uassert.Equal(t, tt.expectedSqrtPriceX96, actual.ToString())
			}
		})
	}
}

func TestSqrtPriceMathGetNextSqrtPriceFromInput2(t *testing.T) {
	t.Run("zero sqrtPX96 should panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for zero sqrtPX96")
			}
		}()

		getNextSqrtPriceFromInput(
			u256.Zero(),
			u256.MustFromDecimal("1000000"),
			u256.MustFromDecimal("500000"),
			true,
		)
	})

	t.Run("zero liquidity should panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for zero liquidity")
			}
		}()

		getNextSqrtPriceFromInput(
			u256.MustFromDecimal("1000000"),
			u256.Zero(),
			u256.MustFromDecimal("500000"),
			true,
		)
	})
}

func TestSqrtPriceMathGetNextSqrtPriceFromOutput(t *testing.T) {
	tests := []struct {
		name                 string
		sqrtPriceX96         *u256.Uint
		liquidity            *u256.Uint
		amountOut            *u256.Uint
		zeroForOne           bool
		shouldPanic          bool
		expectedSqrtPriceX96 string
	}{
		{
			name:         "fails if price is zero",
			sqrtPriceX96: u256.Zero(),
			liquidity:    u256.Zero(),
			amountOut:    u256.MustFromDecimal("100000000000000000"),
			zeroForOne:   false,
			shouldPanic:  true,
		},
		{
			name:         "fails if liquidity is zero",
			sqrtPriceX96: u256.One(),
			liquidity:    u256.Zero(),
			amountOut:    u256.MustFromDecimal("100000000000000000"),
			zeroForOne:   true,
			shouldPanic:  true,
		},
		{
			name:         "fails if output amount is exactly the virtual reserves of token0",
			sqrtPriceX96: u256.MustFromDecimal("20282409603651670423947251286016"),
			liquidity:    u256.MustFromDecimal("1024"),
			amountOut:    u256.NewUint(4),
			zeroForOne:   false,
			shouldPanic:  true,
		},
		{
			name:         "fails if output amount is greater than virtual reserves of token0",
			sqrtPriceX96: u256.MustFromDecimal("20282409603651670423947251286016"),
			liquidity:    u256.MustFromDecimal("1024"),
			amountOut:    u256.NewUint(5),
			zeroForOne:   false,
			shouldPanic:  true,
		},
		{
			name:         "fails if output amount is greater than virtual reserves of token1",
			sqrtPriceX96: u256.MustFromDecimal("20282409603651670423947251286016"),
			liquidity:    u256.MustFromDecimal("1024"),
			amountOut:    u256.NewUint(262145),
			zeroForOne:   true,
			shouldPanic:  true,
		},
		{
			name:         "fails if output amount is exactly the virtual reserves of token1",
			sqrtPriceX96: u256.MustFromDecimal("20282409603651670423947251286016"),
			liquidity:    u256.MustFromDecimal("1024"),
			amountOut:    u256.NewUint(262144),
			zeroForOne:   true,
			shouldPanic:  true,
		},
		{
			name:                 "succeeds if output amount is just less than the virtual reserves of token1",
			sqrtPriceX96:         u256.MustFromDecimal("20282409603651670423947251286016"),
			liquidity:            u256.MustFromDecimal("1024"),
			amountOut:            u256.NewUint(262143),
			zeroForOne:           true,
			expectedSqrtPriceX96: "77371252455336267181195264",
		},
		{
			name:         "puzzling echidna test",
			sqrtPriceX96: u256.MustFromDecimal("20282409603651670423947251286016"),
			liquidity:    u256.MustFromDecimal("1024"),
			amountOut:    u256.NewUint(4),
			zeroForOne:   false,
			shouldPanic:  true,
		},
		{
			name:                 "returns input price if amount in is zero and zeroForOne = true",
			sqrtPriceX96:         encodePriceSqrt("1", "1"),
			liquidity:            u256.MustFromDecimal("100000000000000000"),
			amountOut:            u256.Zero(),
			zeroForOne:           true,
			expectedSqrtPriceX96: encodePriceSqrt("1", "1").ToString(),
		},
		{
			name:                 "returns input price if amount in is zero and zeroForOne = false",
			sqrtPriceX96:         encodePriceSqrt("1", "1"),
			liquidity:            u256.MustFromDecimal("100000000000000000"),
			amountOut:            u256.Zero(),
			zeroForOne:           false,
			expectedSqrtPriceX96: encodePriceSqrt("1", "1").ToString(),
		},
		{
			name:                 "output amount of 0.1 token1, zeroForOne = false",
			sqrtPriceX96:         encodePriceSqrt("1", "1"),
			liquidity:            u256.MustFromDecimal("1000000000000000000"),
			amountOut:            u256.MustFromDecimal("100000000000000000"),
			zeroForOne:           false,
			expectedSqrtPriceX96: "88031291682515930659493278152",
		},
		{
			name:                 "output amount of 0.1 token1, zeroForOne = true",
			sqrtPriceX96:         encodePriceSqrt("1", "1"),
			liquidity:            u256.MustFromDecimal("1000000000000000000"),
			amountOut:            u256.MustFromDecimal("100000000000000000"),
			zeroForOne:           true,
			expectedSqrtPriceX96: "71305346262837903834189555302",
		},
		{
			name:         "reverts if amountOut is impossible in zero for one direction",
			sqrtPriceX96: encodePriceSqrt("1", "1"),
			liquidity:    u256.NewUint(1),
			amountOut:    u256.MustFromDecimal(consts.MAX_UINT256),
			zeroForOne:   true,
			shouldPanic:  true,
		},
		{
			name:         "reverts if amountOut is impossible in one for zero direction",
			sqrtPriceX96: encodePriceSqrt("1", "1"),
			liquidity:    u256.NewUint(1),
			amountOut:    u256.MustFromDecimal(consts.MAX_UINT256),
			zeroForOne:   false,
			shouldPanic:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected panic for %s", tt.name)
					}
				}()
				getNextSqrtPriceFromOutput(tt.sqrtPriceX96, tt.liquidity, tt.amountOut, tt.zeroForOne)

			} else {
				actual := getNextSqrtPriceFromOutput(tt.sqrtPriceX96, tt.liquidity, tt.amountOut, tt.zeroForOne)
				uassert.Equal(t, tt.expectedSqrtPriceX96, actual.ToString())
			}
		})
	}
}

func TestSqrtPriceMathGetAmount0DeltaHelper(t *testing.T) {
	tests := []struct {
		name                                    string
		sqrtRatioAX96, sqrtRatioBX96, liquidity *u256.Uint
		roundUp                                 bool
		expectedAmount0Delta                    string
	}{
		{
			name:                 "returns 0 if liquidity is 0",
			sqrtRatioAX96:        encodePriceSqrt("1", "1"),
			sqrtRatioBX96:        encodePriceSqrt("2", "1"),
			liquidity:            u256.Zero(),
			roundUp:              true,
			expectedAmount0Delta: "0",
		},
		{
			name:                 "returns 0 if prices are equal",
			sqrtRatioAX96:        encodePriceSqrt("1", "1"),
			sqrtRatioBX96:        encodePriceSqrt("1", "1"),
			liquidity:            u256.Zero(),
			roundUp:              true,
			expectedAmount0Delta: "0",
		},
		{
			name:                 "returns 0.1 amount1 for price of 1 to 1.21, roundUp = true",
			sqrtRatioAX96:        encodePriceSqrt("1", "1"),
			sqrtRatioBX96:        encodePriceSqrt("121", "100"),
			liquidity:            u256.MustFromDecimal("1000000000000000000"),
			roundUp:              true,
			expectedAmount0Delta: "90909090909090910",
		},
		{
			name:                 "returns 0.1 amount1 for price of 1 to 1.21, roundUp = false",
			sqrtRatioAX96:        encodePriceSqrt("1", "1"),
			sqrtRatioBX96:        encodePriceSqrt("121", "100"),
			liquidity:            u256.MustFromDecimal("1000000000000000000"),
			roundUp:              false,
			expectedAmount0Delta: "90909090909090909",
		},
		{
			name:                 "works for prices that overflow, roundUp = true",
			sqrtRatioAX96:        u256.MustFromDecimal("43556142965880123323311949751266331066368"),
			sqrtRatioBX96:        u256.MustFromDecimal("22300745198530623141535718272648361505980416"),
			liquidity:            u256.MustFromDecimal("1000000000000000000"),
			roundUp:              true,
			expectedAmount0Delta: "1815437",
		},
		{
			name:                 "works for prices that overflow, roundUp = false",
			sqrtRatioAX96:        u256.MustFromDecimal("43556142965880123323311949751266331066368"),
			sqrtRatioBX96:        u256.MustFromDecimal("22300745198530623141535718272648361505980416"),
			liquidity:            u256.MustFromDecimal("1000000000000000000"),
			roundUp:              false,
			expectedAmount0Delta: "1815436",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := getAmount0DeltaHelper(tt.sqrtRatioAX96, tt.sqrtRatioBX96, tt.liquidity, tt.roundUp)
			uassert.Equal(t, tt.expectedAmount0Delta, actual.ToString())
		})
	}
}

func TestSqrtPriceMathGetAmount1DeltaHelper(t *testing.T) {
	tests := []struct {
		name                                    string
		sqrtRatioAX96, sqrtRatioBX96, liquidity *u256.Uint
		roundUp                                 bool
		expectedAmount1Delta                    string
	}{
		{
			name:                 "returns 0 if liquidity is 0",
			sqrtRatioAX96:        encodePriceSqrt("1", "1"),
			sqrtRatioBX96:        encodePriceSqrt("2", "1"),
			liquidity:            u256.Zero(),
			roundUp:              true,
			expectedAmount1Delta: "0",
		},
		{
			name:                 "returns 0 if prices are equal",
			sqrtRatioAX96:        encodePriceSqrt("1", "1"),
			sqrtRatioBX96:        encodePriceSqrt("1", "1"),
			liquidity:            u256.Zero(),
			roundUp:              true,
			expectedAmount1Delta: "0",
		},
		{
			name:                 "returns 0.1 amount1 for price of 1 to 1.21, roundUp = true",
			sqrtRatioAX96:        encodePriceSqrt("1", "1"),
			sqrtRatioBX96:        encodePriceSqrt("121", "100"),
			liquidity:            u256.MustFromDecimal("1000000000000000000"),
			roundUp:              true,
			expectedAmount1Delta: "100000000000000000",
		},
		{
			name:                 "returns 0.1 amount1 for price of 1 to 1.21, roundUp = false",
			sqrtRatioAX96:        encodePriceSqrt("1", "1"),
			sqrtRatioBX96:        encodePriceSqrt("121", "100"),
			liquidity:            u256.MustFromDecimal("1000000000000000000"),
			roundUp:              false,
			expectedAmount1Delta: "99999999999999999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := getAmount1DeltaHelper(tt.sqrtRatioAX96, tt.sqrtRatioBX96, tt.liquidity, tt.roundUp)
			uassert.Equal(t, tt.expectedAmount1Delta, actual.ToString())
		})
	}
}

func TestSwapComputation_SqrtP_SqrtQ_Mul_Overflow(t *testing.T) {
	sqrtP := u256.MustFromDecimal("1025574284609383690408304870162715216695788925244")
	liquidity := u256.MustFromDecimal("50015962439936049619261659728067971248")
	amountIn := u256.MustFromDecimal("406")
	zeroForOne := true

	sqrtQ := getNextSqrtPriceFromInput(sqrtP, liquidity, amountIn, zeroForOne)
	uassert.Equal(t, "1025574284609383582644711336373707553698163132913", sqrtQ.ToString())

	amount0Delta := getAmount0DeltaHelper(sqrtQ, sqrtP, liquidity, true)
	uassert.Equal(t, "406", amount0Delta.ToString())
}

// encodePriceSqrt calculates the sqrt((reserve1 << 192) / reserve0)
func encodePriceSqrt(reserve1, reserve0 string) *u256.Uint {
	reserve1Uint := u256.MustFromDecimal(reserve1)
	reserve0Uint := u256.MustFromDecimal(reserve0)

	if reserve0Uint.IsZero() {
		panic("division by zero")
	}

	// numerator = reserve1 * (2^192)
	two192 := new(u256.Uint).Lsh(u256.NewUint(1), 192)
	numerator := new(u256.Uint).Mul(reserve1Uint, two192)

	// ratioX192 = numerator / reserve0
	ratioX192 := new(u256.Uint).Div(numerator, reserve0Uint)

	// Return sqrt(ratioX192)
	return sqrt(ratioX192)
}

// sqrt computes the integer square root of a u256.Uint
func sqrt(x *u256.Uint) *u256.Uint {
	if x.IsZero() {
		return u256.NewUint(0)
	}

	z := new(u256.Uint).Set(x)
	y := new(u256.Uint).Rsh(z, 1) // Initial guess is x / 2

	for y.Cmp(z) < 0 {
		z.Set(y)
		temp := new(u256.Uint).Div(x, z)
		y.Add(z, temp).Rsh(y, 1)
	}
	return z
}
