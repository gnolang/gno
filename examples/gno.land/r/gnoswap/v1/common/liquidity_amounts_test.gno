package common

import (
	"testing"

	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"
	"gno.land/p/gnoswap/consts"
	u256 "gno.land/p/gnoswap/uint256"
)

func TestToAscendingOrder(t *testing.T) {
	tests := []struct {
		name      string
		a         *u256.Uint
		b         *u256.Uint
		expectedA string
		expectedB string
	}{
		{
			name:      "Ascending order - a < b",
			a:         u256.MustFromDecimal("10"),
			b:         u256.MustFromDecimal("20"),
			expectedA: "10",
			expectedB: "20",
		},
		{
			name:      "Descending order - a > b",
			a:         u256.MustFromDecimal("50"),
			b:         u256.MustFromDecimal("30"),
			expectedA: "30",
			expectedB: "50",
		},
		{
			name:      "Equal values - a == b",
			a:         u256.MustFromDecimal("100"),
			b:         u256.MustFromDecimal("100"),
			expectedA: "100",
			expectedB: "100",
		},
		{
			name:      "Large numbers",
			a:         u256.MustFromDecimal("340282366920938463463374607431768211455"), // 2^128 - 1
			b:         u256.MustFromDecimal("170141183460469231731687303715884105727"), // 2^127 - 1
			expectedA: "170141183460469231731687303715884105727",
			expectedB: "340282366920938463463374607431768211455",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			min, max := toAscendingOrder(tt.a, tt.b)

			if min.ToString() != tt.expectedA {
				t.Errorf("Expected min to be %s, got %s", tt.expectedA, min.ToString())
			}

			if max.ToString() != tt.expectedB {
				t.Errorf("Expected max to be %s, got %s", tt.expectedB, max.ToString())
			}
		})
	}
}

func TestSafeConvertToUint128(t *testing.T) {
	tests := []struct {
		name        string
		input       *u256.Uint
		expected    *u256.Uint
		shouldPanic bool
	}{
		{
			name:        "Valid uint128 value",
			input:       u256.MustFromDecimal("340282366920938463463374607431768211455"), // MAX_UINT128
			expected:    u256.MustFromDecimal("340282366920938463463374607431768211455"),
			shouldPanic: false,
		},
		{
			name:        "Value exceeding uint128 range",
			input:       u256.MustFromDecimal("340282366920938463463374607431768211456"), // MAX_UINT128 + 1
			shouldPanic: true,
		},
		{
			name:        "Zero value",
			input:       u256.Zero(),
			expected:    u256.Zero(),
			shouldPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Expected panic but got none")
					}
				}()
				safeConvertToUint128(tt.input)
			} else {
				got := safeConvertToUint128(tt.input)
				if !got.Eq(tt.expected) {
					t.Errorf("Expected %s, got %s", tt.expected.ToString(), got.ToString())
				}
			}
		})
	}
}

func TestComputeLiquidityForAmount0(t *testing.T) {
	testCases := []struct {
		name        string
		sqrtRatioA  string
		sqrtRatioB  string
		amount0     string
		expected    string
		expectPanic bool
	}{
		{
			name:       "Basic liquidity calculation",
			sqrtRatioA: "79228162514264337593543950336",  // sqrt(1) << 96
			sqrtRatioB: "158456325028528675187087900672", // sqrt(4) << 96
			amount0:    "1000000",
			expected:   "2000000", // Expected liquidity
		},
		{
			name:       "No liquidity (zero amount)",
			sqrtRatioA: "79228162514264337593543950336",
			sqrtRatioB: "158456325028528675187087900672",
			amount0:    "0",
			expected:   "0",
		},
		{
			name:        "Liquidity overflow (exceeds uint128)",
			sqrtRatioA:  "158456325028528675187087900672",
			sqrtRatioB:  "316912650057057350374175801344",
			amount0:     "340282366920938463463374607431768211456", // Exceeds uint128
			expectPanic: true,
		},
		{
			name:        "Zero liquidity with equal ratios",
			sqrtRatioA:  "79228162514264337593543950336",
			sqrtRatioB:  "79228162514264337593543950336",
			amount0:     "1000000",
			expected:    "0",
			expectPanic: true,
		},
		{
			name:        "Panic with identical ticks",
			sqrtRatioA:  "79228162514264337593543950336",
			sqrtRatioB:  "79228162514264337593543950336",
			amount0:     "1000000",
			expectPanic: true,
		},
		{
			name:       "Large liquidity calculation",
			sqrtRatioA: "79228162514264337593543950336",
			sqrtRatioB: "158456325028528675187087900672",
			amount0:    "1000000000",
			expected:   "2000000000",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tc.expectPanic {
						t.Errorf("Unexpected panic for test case: %s", tc.name)
					}
				}
			}()

			sqrtRatioA := u256.MustFromDecimal(tc.sqrtRatioA)
			sqrtRatioB := u256.MustFromDecimal(tc.sqrtRatioB)
			amount0 := u256.MustFromDecimal(tc.amount0)

			result := computeLiquidityForAmount0(sqrtRatioA, sqrtRatioB, amount0)
			if !tc.expectPanic {
				if result.ToString() != tc.expected {
					t.Errorf("Expected %s but got %s", tc.expected, result.ToString())
				}
			}
		})
	}
}

func TestComputeLiquidityForAmount1(t *testing.T) {
	q96 := u256.MustFromDecimal(consts.Q96)
	amount1 := u256.MustFromDecimal("1000000")

	t.Run("Basic liquidity calculation", func(t *testing.T) {
		sqrtRatioAX96 := q96                                                // 2^96 (1 in Q96)
		sqrtRatioBX96 := new(u256.Uint).Mul(q96, u256.MustFromDecimal("4")) // 4^96 (4 in Q96)

		expected := u256.MustFromDecimal("333333") // Expected liquidity
		result := computeLiquidityForAmount1(sqrtRatioAX96, sqrtRatioBX96, amount1)

		if !result.Eq(expected) {
			t.Errorf("Expected %s but got %s", expected.ToString(), result.ToString())
		}
	})

	t.Run("Zero liquidity with equal ratios", func(t *testing.T) {
		sqrtRatioAX96 := q96 // 2^96 (1 in Q96)
		sqrtRatioBX96 := q96 // Same as lower tick

		uassert.PanicsWithMessage(t,
			"[GNOSWAP-COMMON-010] identical ticks || sqrtRatioAX96 (79228162514264337593543950336) and sqrtRatioBX96 (79228162514264337593543950336) are identical",
			func() {
				_ = computeLiquidityForAmount1(sqrtRatioAX96, sqrtRatioBX96, amount1)
			})
	})

	t.Run("Large liquidity calculation", func(t *testing.T) {
		sqrtRatioAX96 := q96                                       // 1x
		sqrtRatioBX96 := new(u256.Uint).Mul(q96, u256.NewUint(16)) // 16x
		largeAmount := u256.MustFromDecimal("1000000000")

		expected := u256.MustFromDecimal("66666666") // 1B / 16 = 62.5M
		result := computeLiquidityForAmount1(sqrtRatioAX96, sqrtRatioBX96, largeAmount)

		if !result.Eq(expected) {
			t.Errorf("Expected %s but got %s", expected.ToString(), result.ToString())
		}
	})
}

func TestGetLiquidityForAmounts(t *testing.T) {
	q96 := u256.MustFromDecimal(consts.Q96)

	tests := []struct {
		name          string
		sqrtRatioX96  string
		sqrtRatioAX96 string
		sqrtRatioBX96 string
		amount0       string
		amount1       string
		expected      string
	}{
		{
			name:          "Basic Liquidity Calculation - Token0 Dominant",
			sqrtRatioX96:  q96.ToString(),
			sqrtRatioAX96: (new(u256.Uint).Mul(q96, u256.NewUint(4))).ToString(),
			sqrtRatioBX96: (new(u256.Uint).Mul(q96, u256.NewUint(16))).ToString(),
			amount0:       "1000000",
			amount1:       "1000000",
			expected:      "5333333",
		},
		{
			name:          "Within Range - Both Token0 and Token1",
			sqrtRatioX96:  (new(u256.Uint).Mul(q96, u256.NewUint(2))).ToString(),
			sqrtRatioAX96: (new(u256.Uint).Mul(q96, u256.NewUint(4))).ToString(),
			sqrtRatioBX96: (new(u256.Uint).Mul(q96, u256.NewUint(16))).ToString(),
			amount0:       "2000000",
			amount1:       "3000000",
			expected:      "10666666",
		},
		{
			name:          "Token1 Dominant - Price Above Upper Bound",
			sqrtRatioX96:  (new(u256.Uint).Mul(q96, u256.NewUint(20))).ToString(),
			sqrtRatioAX96: (new(u256.Uint).Mul(q96, u256.NewUint(4))).ToString(),
			sqrtRatioBX96: (new(u256.Uint).Mul(q96, u256.NewUint(16))).ToString(),
			amount0:       "1500000",
			amount1:       "3000000",
			expected:      "250000",
		},
		{
			name:          "Edge Case - sqrtRatioX96 = Lower Bound",
			sqrtRatioX96:  (new(u256.Uint).Mul(q96, u256.NewUint(4))).ToString(),
			sqrtRatioAX96: (new(u256.Uint).Mul(q96, u256.NewUint(4))).ToString(),
			sqrtRatioBX96: (new(u256.Uint).Mul(q96, u256.NewUint(16))).ToString(),
			amount0:       "500000",
			amount1:       "500000",
			expected:      "2666666",
		},
		{
			name:          "Edge Case - sqrtRatioX96 = Upper Bound",
			sqrtRatioX96:  (new(u256.Uint).Mul(q96, u256.NewUint(16))).ToString(),
			sqrtRatioAX96: (new(u256.Uint).Mul(q96, u256.NewUint(4))).ToString(),
			sqrtRatioBX96: (new(u256.Uint).Mul(q96, u256.NewUint(16))).ToString(),
			amount0:       "1000000",
			amount1:       "1000000",
			expected:      "83333",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sqrtRatioX96 := u256.MustFromDecimal(tc.sqrtRatioX96)
			sqrtRatioAX96 := u256.MustFromDecimal(tc.sqrtRatioAX96)
			sqrtRatioBX96 := u256.MustFromDecimal(tc.sqrtRatioBX96)
			amount0 := u256.MustFromDecimal(tc.amount0)
			amount1 := u256.MustFromDecimal(tc.amount1)

			result := GetLiquidityForAmounts(sqrtRatioX96, sqrtRatioAX96, sqrtRatioBX96, amount0, amount1)
			expected := u256.MustFromDecimal(tc.expected)

			uassert.Equal(t, expected.ToString(), result.ToString())
		})
	}
}

func TestComputeAmount0ForLiquidity(t *testing.T) {
	q96 := u256.MustFromDecimal("79228162514264337593543950336") // 2^96

	tests := []struct {
		name          string
		sqrtRatioAX96 string
		sqrtRatioBX96 string
		liquidity     string
		expected      string
	}{
		{
			name:          "Basic Case - Small Range",
			sqrtRatioAX96: new(u256.Uint).Mul(q96, u256.NewUint(4)).ToString(),
			sqrtRatioBX96: new(u256.Uint).Mul(q96, u256.NewUint(8)).ToString(),
			liquidity:     "1000000",
			expected:      "125000",
		},
		{
			name:          "Large Liquidity - Wide Range",
			sqrtRatioAX96: new(u256.Uint).Mul(q96, u256.NewUint(2)).ToString(),
			sqrtRatioBX96: new(u256.Uint).Mul(q96, u256.NewUint(16)).ToString(),
			liquidity:     "5000000000",
			expected:      "2187500000",
		},
		{
			name:          "Edge Case - Equal Bounds",
			sqrtRatioAX96: new(u256.Uint).Mul(q96, u256.NewUint(8)).ToString(),
			sqrtRatioBX96: new(u256.Uint).Mul(q96, u256.NewUint(8)).ToString(),
			liquidity:     "1000000",
			expected:      "0",
		},
		{
			name:          "Minimum Liquidity",
			sqrtRatioAX96: new(u256.Uint).Mul(q96, u256.NewUint(5)).ToString(),
			sqrtRatioBX96: new(u256.Uint).Mul(q96, u256.NewUint(10)).ToString(),
			liquidity:     "1",
			expected:      "0",
		},
		{
			name:          "Max Liquidity",
			sqrtRatioAX96: new(u256.Uint).Mul(q96, u256.NewUint(1)).ToString(),
			sqrtRatioBX96: new(u256.Uint).Mul(q96, u256.NewUint(32)).ToString(),
			liquidity:     "1000000000000000000",
			expected:      "968750000000000000",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sqrtRatioAX96 := u256.MustFromDecimal(tc.sqrtRatioAX96)
			sqrtRatioBX96 := u256.MustFromDecimal(tc.sqrtRatioBX96)
			liquidity := u256.MustFromDecimal(tc.liquidity)

			result := computeAmount0ForLiquidity(sqrtRatioAX96, sqrtRatioBX96, liquidity)
			expected := u256.MustFromDecimal(tc.expected)

			if result.ToString() != expected.ToString() {
				t.Errorf("expected %s but got %s", expected.ToString(), result.ToString())
			}
		})
	}
}

func TestComputeAmount1ForLiquidity(t *testing.T) {
	q96 := u256.MustFromDecimal(consts.Q96) // 2^96 = 79228162514264337593543950336

	tests := []struct {
		name           string
		sqrtRatioAX96  *u256.Uint
		sqrtRatioBX96  *u256.Uint
		liquidity      *u256.Uint
		expectedAmount string
	}{
		{
			name:           "Basic Case - Small Liquidity",
			sqrtRatioAX96:  q96,
			sqrtRatioBX96:  new(u256.Uint).Mul(q96, u256.NewUint(4)), // sqrtRatioBX96 = 4 * Q96
			liquidity:      u256.NewUint(1000000000),
			expectedAmount: "3000000000", // (4-1)*liquidity = 3 * 10^9
		},
		{
			name:           "Edge Case - Equal Ratios",
			sqrtRatioAX96:  q96,
			sqrtRatioBX96:  q96,
			liquidity:      u256.NewUint(1000000),
			expectedAmount: "0",
		},
		{
			name:           "Zero Liquidity",
			sqrtRatioAX96:  q96,
			sqrtRatioBX96:  new(u256.Uint).Mul(q96, u256.NewUint(2)),
			liquidity:      u256.Zero(),
			expectedAmount: "0",
		},
		{
			name:           "Large Liquidity",
			sqrtRatioAX96:  q96,
			sqrtRatioBX96:  new(u256.Uint).Mul(q96, u256.NewUint(16)), // sqrtRatioBX96 = 16 * Q96
			liquidity:      u256.NewUint(1000000000000000000),         // 1e18 liquidity
			expectedAmount: "15000000000000000000",                    // (16-1) * 1e18 = 15 * 1e18
		},
		{
			name:           "Descending Ratios (Order Correction)",
			sqrtRatioAX96:  new(u256.Uint).Mul(q96, u256.NewUint(8)),
			sqrtRatioBX96:  q96,
			liquidity:      u256.NewUint(500000),
			expectedAmount: "3500000", // (8-1)*500000
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := computeAmount1ForLiquidity(tt.sqrtRatioAX96, tt.sqrtRatioBX96, tt.liquidity)
			uassert.Equal(t, tt.expectedAmount, result.ToString(), ufmt.Sprintf("expected %s but got %s", tt.expectedAmount, result.ToString()))
		})
	}
}

func TestGetAmountsForLiquidity(t *testing.T) {
	q96 := u256.MustFromDecimal(consts.Q96) // 2^96 = 79228162514264337593543950336

	tests := []struct {
		name            string
		sqrtRatioX96    *u256.Uint
		sqrtRatioAX96   *u256.Uint
		sqrtRatioBX96   *u256.Uint
		liquidity       *u256.Uint
		expectedAmount0 string
		expectedAmount1 string
	}{
		{
			name:            "Basic Case - Within Range",
			sqrtRatioX96:    new(u256.Uint).Mul(q96, u256.NewUint(2)), // Current price at 2 * Q96
			sqrtRatioAX96:   q96,                                      // Lower bound at 1 * Q96
			sqrtRatioBX96:   new(u256.Uint).Mul(q96, u256.NewUint(4)), // Upper bound at 4 * Q96
			liquidity:       u256.NewUint(1000000),
			expectedAmount0: "250000",
			expectedAmount1: "1000000",
		},
		{
			name:            "Edge Case - At Lower Bound (sqrtRatioX96 == sqrtRatioAX96)",
			sqrtRatioX96:    q96,
			sqrtRatioAX96:   q96,
			sqrtRatioBX96:   new(u256.Uint).Mul(q96, u256.NewUint(8)),
			liquidity:       u256.NewUint(1000000),
			expectedAmount0: "875000",
			expectedAmount1: "0",
		},
		{
			name:            "Edge Case - At Upper Bound (sqrtRatioX96 == sqrtRatioBX96)",
			sqrtRatioX96:    new(u256.Uint).Mul(q96, u256.NewUint(8)),
			sqrtRatioAX96:   q96,
			sqrtRatioBX96:   new(u256.Uint).Mul(q96, u256.NewUint(8)),
			liquidity:       u256.NewUint(1000000),
			expectedAmount0: "0",
			expectedAmount1: "7000000",
		},
		{
			name:            "Out of Range - Below Lower Bound",
			sqrtRatioX96:    new(u256.Uint).Div(q96, u256.NewUint(2)), // Current price at 0.5 * Q96
			sqrtRatioAX96:   q96,
			sqrtRatioBX96:   new(u256.Uint).Mul(q96, u256.NewUint(8)),
			liquidity:       u256.NewUint(500000),
			expectedAmount0: "437500",
			expectedAmount1: "0",
		},
		{
			name:            "Out of Range - Above Upper Bound",
			sqrtRatioX96:    new(u256.Uint).Mul(q96, u256.NewUint(10)), // Current price at 10 * Q96
			sqrtRatioAX96:   q96,
			sqrtRatioBX96:   new(u256.Uint).Mul(q96, u256.NewUint(8)),
			liquidity:       u256.NewUint(2000000),
			expectedAmount0: "0",
			expectedAmount1: "14000000",
		},
		{
			name:            "Zero Liquidity",
			sqrtRatioX96:    q96,
			sqrtRatioAX96:   q96,
			sqrtRatioBX96:   new(u256.Uint).Mul(q96, u256.NewUint(16)),
			liquidity:       u256.Zero(),
			expectedAmount0: "0",
			expectedAmount1: "0",
		},
		{
			name:            "Descending Ratios (Order Correction)",
			sqrtRatioX96:    q96,
			sqrtRatioAX96:   new(u256.Uint).Mul(q96, u256.NewUint(8)),
			sqrtRatioBX96:   q96,
			liquidity:       u256.NewUint(1000000),
			expectedAmount0: "875000",
			expectedAmount1: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amount0, amount1 := GetAmountsForLiquidity(tt.sqrtRatioX96, tt.sqrtRatioAX96, tt.sqrtRatioBX96, tt.liquidity)
			uassert.Equal(t, tt.expectedAmount0, amount0, ufmt.Sprintf("expected %s but got %s for amount0", tt.expectedAmount0, amount0))
			uassert.Equal(t, tt.expectedAmount1, amount1, ufmt.Sprintf("expected %s but got %s for amount1", tt.expectedAmount1, amount1))
		})
	}
}
