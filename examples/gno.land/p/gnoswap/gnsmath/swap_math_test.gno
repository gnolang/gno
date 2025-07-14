package gnsmath

import (
	"testing"

	"gno.land/p/demo/uassert"

	i256 "gno.land/p/gnoswap/int256"
	u256 "gno.land/p/gnoswap/uint256"
)

func TestSwapMathComputeSwapStepStr(t *testing.T) {
	tests := []struct {
		name                           string
		currentX96, targetX96          *u256.Uint
		liquidity                      *u256.Uint
		amountRemaining                *i256.Int
		feePips                        uint64
		sqrtNextX96                    *u256.Uint
		chkSqrtNextX96                 func(sqrtRatioNextX96, priceTarget *u256.Uint)
		amountIn, amountOut, feeAmount string
	}{
		{
			name:            "exact amount in that gets capped at price target in one for zero",
			currentX96:      encodePriceSqrtTest(t, "1", "1"),
			targetX96:       encodePriceSqrtTest(t, "101", "100"),
			liquidity:       u256.MustFromDecimal("2000000000000000000"),
			amountRemaining: i256.MustFromDecimal("1000000000000000000"),
			feePips:         600,
			sqrtNextX96:     encodePriceSqrtTest(t, "101", "100"),
			chkSqrtNextX96: func(sqrtRatioNextX96, priceTarget *u256.Uint) {
				uassert.True(t, sqrtRatioNextX96.Eq(priceTarget))
			},
			amountIn:  "9975124224178055",
			amountOut: "9925619580021728",
			feeAmount: "5988667735148",
		},
		{
			name:            "exact amount out that gets capped at price target in one for zero",
			currentX96:      encodePriceSqrtTest(t, "1", "1"),
			targetX96:       encodePriceSqrtTest(t, "101", "100"),
			liquidity:       u256.MustFromDecimal("2000000000000000000"),
			amountRemaining: i256.MustFromDecimal("-1000000000000000000"),
			feePips:         600,
			sqrtNextX96:     encodePriceSqrtTest(t, "101", "100"),
			chkSqrtNextX96: func(sqrtRatioNextX96, priceTarget *u256.Uint) {
				uassert.True(t, sqrtRatioNextX96.Eq(priceTarget))
			},
			amountIn:  "9975124224178055",
			amountOut: "9925619580021728",
			feeAmount: "5988667735148",
		},
		{
			name:            "exact amount in that is fully spent in one for zero",
			currentX96:      encodePriceSqrtTest(t, "1", "1"),
			targetX96:       encodePriceSqrtTest(t, "1000", "100"),
			liquidity:       u256.MustFromDecimal("2000000000000000000"),
			amountRemaining: i256.MustFromDecimal("1000000000000000000"),
			sqrtNextX96:     encodePriceSqrtTest(t, "1000", "100"),
			feePips:         600,
			chkSqrtNextX96: func(sqrtRatioNextX96, priceTarget *u256.Uint) {
				uassert.True(t, sqrtRatioNextX96.Lte(priceTarget))
			},
			amountIn:  "999400000000000000",
			amountOut: "666399946655997866",
			feeAmount: "600000000000000",
		},
		{
			name:            "exact amount out that is fully received in one for zero",
			currentX96:      encodePriceSqrtTest(t, "1", "1"),
			targetX96:       encodePriceSqrtTest(t, "1000", "100"),
			liquidity:       u256.MustFromDecimal("2000000000000000000"),
			amountRemaining: i256.MustFromDecimal("-1000000000000000000"),
			feePips:         600,
			sqrtNextX96:     encodePriceSqrtTest(t, "1000", "100"),
			chkSqrtNextX96: func(sqrtRatioNextX96, priceTarget *u256.Uint) {
				uassert.True(t, sqrtRatioNextX96.Lt(priceTarget))
			},
			amountIn:  "2000000000000000000",
			amountOut: "1000000000000000000",
			feeAmount: "1200720432259356",
		},
		{
			name:            "amount out is capped at the desired amount out",
			currentX96:      u256.MustFromDecimal("417332158212080721273783715441582"),
			targetX96:       u256.MustFromDecimal("1452870262520218020823638996"),
			liquidity:       u256.MustFromDecimal("159344665391607089467575320103"),
			amountRemaining: i256.MustFromDecimal("-1"),
			feePips:         1,
			sqrtNextX96:     u256.MustFromDecimal("417332158212080721273783715441581"),
			chkSqrtNextX96: func(sqrtRatioNextX96, priceTarget *u256.Uint) {
				uassert.True(t, sqrtRatioNextX96.Eq(priceTarget))
			},
			amountIn:  "1",
			amountOut: "1",
			feeAmount: "1",
		},
		{
			name:            "target price of 1 uses partial input amount",
			currentX96:      u256.MustFromDecimal("2"),
			targetX96:       u256.MustFromDecimal("1"),
			liquidity:       u256.MustFromDecimal("1"),
			amountRemaining: i256.MustFromDecimal("3915081100057732413702495386755767"),
			feePips:         1,
			sqrtNextX96:     u256.MustFromDecimal("1"),
			chkSqrtNextX96: func(sqrtRatioNextX96, priceTarget *u256.Uint) {
				uassert.True(t, sqrtRatioNextX96.Eq(priceTarget))
			},
			amountIn:  "39614081257132168796771975168",
			amountOut: "0",
			feeAmount: "39614120871253040049813",
		},
		{
			name:            "entire input amount taken as fee",
			currentX96:      u256.MustFromDecimal("4295128739"),
			targetX96:       u256.MustFromDecimal("79887613182836312"),
			liquidity:       u256.MustFromDecimal("1985041575832132834610021537970"),
			amountRemaining: i256.MustFromDecimal("10"),
			feePips:         1872,
			sqrtNextX96:     u256.MustFromDecimal("4295128739"),
			chkSqrtNextX96: func(sqrtRatioNextX96, priceTarget *u256.Uint) {
				uassert.True(t, sqrtRatioNextX96.Eq(priceTarget))
			},
			amountIn:  "0",
			amountOut: "0",
			feeAmount: "10",
		},
		{
			name:            "handles intermediate insufficient liquidity in zero for one exact output case",
			currentX96:      u256.MustFromDecimal("20282409603651670423947251286016"),
			targetX96:       u256.MustFromDecimal("22310650564016837466341976414617"),
			liquidity:       u256.MustFromDecimal("1024"),
			amountRemaining: i256.MustFromDecimal("-4"),
			feePips:         3000,
			sqrtNextX96:     u256.MustFromDecimal("22310650564016837466341976414617"),
			chkSqrtNextX96: func(sqrtRatioNextX96, priceTarget *u256.Uint) {
				uassert.True(t, sqrtRatioNextX96.Eq(priceTarget))
			},
			amountIn:  "26215",
			amountOut: "0",
			feeAmount: "79",
		},
		{
			name:            "handles intermediate insufficient liquidity in one for zero exact output case",
			currentX96:      u256.MustFromDecimal("20282409603651670423947251286016"),
			targetX96:       u256.MustFromDecimal("18254168643286503381552526157414"),
			liquidity:       u256.MustFromDecimal("1024"),
			amountRemaining: i256.MustFromDecimal("-263000"),
			feePips:         3000,
			sqrtNextX96:     u256.MustFromDecimal("18254168643286503381552526157414"),
			chkSqrtNextX96: func(sqrtRatioNextX96, priceTarget *u256.Uint) {
				uassert.True(t, sqrtRatioNextX96.Eq(priceTarget))
			},
			amountIn:  "1",
			amountOut: "26214",
			feeAmount: "1",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sqrtRatioNextX96, amountIn, amountOut, feeAmount := SwapMathComputeSwapStepStr(test.currentX96, test.targetX96, test.liquidity, test.amountRemaining, test.feePips)
			test.chkSqrtNextX96(u256.MustFromDecimal(sqrtRatioNextX96), test.sqrtNextX96)
			uassert.Equal(t, amountIn, test.amountIn)
			uassert.Equal(t, amountOut, test.amountOut)
			uassert.Equal(t, feeAmount, test.feeAmount)
		})
	}
}

func TestSwapMathComputeSwapStepStrFail(t *testing.T) {
	tests := []struct {
		name                           string
		currentX96, targetX96          *u256.Uint
		liquidity                      *u256.Uint
		amountRemaining                *i256.Int
		feePips                        uint64
		sqrtNextX96                    *u256.Uint
		chkSqrtNextX96                 func(sqrtRatioNextX96, priceTarget *u256.Uint)
		amountIn, amountOut, feeAmount string
		shouldPanic                    bool
		expectedMessage                string
	}{
		{
			name:            "input parameter is nil",
			currentX96:      nil,
			targetX96:       nil,
			liquidity:       nil,
			amountRemaining: nil,
			feePips:         600,
			sqrtNextX96:     encodePriceSqrtTest(t, "101", "100"),
			chkSqrtNextX96: func(sqrtRatioNextX96, priceTarget *u256.Uint) {
				uassert.True(t, sqrtRatioNextX96.Eq(priceTarget))
			},
			amountIn:        "9975124224178055",
			amountOut:       "9925619580021728",
			feeAmount:       "5988667735148",
			shouldPanic:     true,
			expectedMessage: "SwapMathComputeSwapStepStr: invalid input",
		},
		{
			name:            "feePips exceeds maximum allowed value",
			currentX96:      encodePriceSqrtTest(t, "1", "1"),
			targetX96:       encodePriceSqrtTest(t, "101", "100"),
			liquidity:       u256.MustFromDecimal("2000000000000000000"),
			amountRemaining: i256.MustFromDecimal("1000000000000000000"),
			feePips:         1000001,
			sqrtNextX96:     encodePriceSqrtTest(t, "101", "100"),
			chkSqrtNextX96: func(sqrtRatioNextX96, priceTarget *u256.Uint) {
				uassert.True(t, sqrtRatioNextX96.Eq(priceTarget))
			},
			amountIn:        "9975124224178055",
			amountOut:       "9925619580021728",
			feeAmount:       "5988667735148",
			shouldPanic:     true,
			expectedMessage: "SwapMathComputeSwapStepStr: feePips must be less than or equal to 1000000",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if test.shouldPanic {
						if errMsg, ok := r.(string); ok {
							uassert.Equal(t, test.expectedMessage, errMsg)
						} else {
							t.Errorf("expected a panic with message, got: %v", r)
						}
					} else {
						t.Errorf("unexpected panic: %v", r)
					}
				} else {
					if test.shouldPanic {
						t.Errorf("expected a panic, but none occurred")
					}
				}
			}()

			SwapMathComputeSwapStepStr(
				test.currentX96,
				test.targetX96,
				test.liquidity,
				test.amountRemaining,
				test.feePips)
		})
	}
}

// encodePriceSqrtTest calculates the sqrt((reserve1 << 192) / reserve0)
func encodePriceSqrtTest(t *testing.T, reserve1, reserve0 string) *u256.Uint {
	t.Helper()

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
	return sqrtTest(t, ratioX192)
}

// sqrt computes the integer square root of a u256.Uint
func sqrtTest(t *testing.T, x *u256.Uint) *u256.Uint {
	t.Helper()

	if x.IsZero() {
		return u256.NewUint(0)
	}

	z := new(u256.Uint).Set(x)
	y := new(u256.Uint).Rsh(z, 1) // Initial guess is x / 2

	temp := new(u256.Uint)
	for y.Cmp(z) < 0 {
		z.Set(y)
		temp.Div(x, z)
		y.Add(z, temp).Rsh(y, 1)
	}
	return z
}
