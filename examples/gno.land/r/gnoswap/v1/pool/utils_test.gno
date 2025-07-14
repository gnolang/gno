package pool

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"
	i256 "gno.land/p/gnoswap/int256"
	u256 "gno.land/p/gnoswap/uint256"

	"gno.land/p/gnoswap/consts"
)

func TestU256Min(t *testing.T) {
	tests := []struct {
		name     string
		num1     string
		num2     string
		expected string
	}{
		// Basic cases
		{
			name:     "num1 is less than num2",
			num1:     "1",
			num2:     "2",
			expected: "1",
		},
		{
			name:     "num1 is greater than num2",
			num1:     "2",
			num2:     "1",
			expected: "1",
		},
		{
			name:     "num1 is equal to num2",
			num1:     "1",
			num2:     "1",
			expected: "1",
		},

		// Edge cases with zero
		{
			name:     "compare max u256 with zero",
			num1:     consts.MAX_UINT256,
			num2:     "0",
			expected: "0",
		},
		{
			name:     "compare zero with non-zero",
			num1:     "0",
			num2:     "1",
			expected: "0",
		},

		// Close values
		{
			name:     "very close values (num1 smaller)",
			num1:     "999999999999999999",
			num2:     "1000000000000000000",
			expected: "999999999999999999",
		},
		{
			name:     "very close values (num2 smaller)",
			num1:     "1000000000000000000",
			num2:     "999999999999999999",
			expected: "999999999999999999",
		},

		// Large values
		{
			name:     "large values (num1 smaller)",
			num1:     "115792089237316195423570985008687907853269984665640564039457584007913129639934",
			num2:     consts.MAX_UINT256,
			expected: "115792089237316195423570985008687907853269984665640564039457584007913129639934",
		},
		{
			name:     "large values (num2 smaller)",
			num1:     consts.MAX_UINT256,
			num2:     "115792089237316195423570985008687907853269984665640564039457584007913129639934",
			expected: "115792089237316195423570985008687907853269984665640564039457584007913129639934",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			num1 := u256.MustFromDecimal(tt.num1)
			num2 := u256.MustFromDecimal(tt.num2)
			expected := u256.MustFromDecimal(tt.expected)

			result := u256Min(num1, num2)
			uassert.Equal(t, expected.ToString(), result.ToString())
		})
	}
}

func TestU256Min_AdditionalCases(t *testing.T) {
	tests := []struct {
		name     string
		num1     string
		num2     string
		expected string
	}{
		// Q values comparison
		{
			name:     "Q64 vs Q96",
			num1:     consts.Q64,
			num2:     consts.Q96,
			expected: consts.Q64,
		},
		{
			name:     "Q96 vs Q128",
			num1:     consts.Q96,
			num2:     consts.Q128,
			expected: consts.Q96,
		},
		// Price boundaries
		{
			name:     "MIN_SQRT_RATIO vs MAX_SQRT_RATIO",
			num1:     consts.MIN_SQRT_RATIO,
			num2:     consts.MAX_SQRT_RATIO,
			expected: consts.MIN_SQRT_RATIO,
		},
		// Different magnitudes
		{
			name:     "very small vs very large",
			num1:     "1",
			num2:     consts.MAX_UINT256,
			expected: "1",
		},
		{
			name:     "consecutive values",
			num1:     "999999999999999999999999999999999999999",
			num2:     "1000000000000000000000000000000000000000",
			expected: "999999999999999999999999999999999999999",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			num1 := u256.MustFromDecimal(tt.num1)
			num2 := u256.MustFromDecimal(tt.num2)
			expected := u256.MustFromDecimal(tt.expected)
			
			result := u256Min(num1, num2)
			uassert.Equal(t, expected.ToString(), result.ToString())
			
			// Test commutativity
			resultReversed := u256Min(num2, num1)
			uassert.Equal(t, expected.ToString(), resultReversed.ToString())
		})
	}
}

func TestGetPrevAsString(t *testing.T) {
	tests := []struct {
		name            string
		_realm          std.Realm
		expectedAddr    string
		expectedPkgPath string
	}{
		{
			name:            "user call",
			_realm:          std.NewUserRealm(std.Address("user")),
			expectedAddr:    "user",
			expectedPkgPath: "",
		},
		{
			name:            "code call",
			_realm:          std.NewCodeRealm("gno.land/r/demo/realm"),
			expectedAddr:    std.DerivePkgAddr("gno.land/r/demo/realm").String(),
			expectedPkgPath: "gno.land/r/demo/realm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testing.SetRealm(tt._realm)
			func() {
				testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/tests"))
				addr, pkgPath := getPrevAsString()
				uassert.Equal(t, tt.expectedAddr, addr)
				uassert.Equal(t, tt.expectedPkgPath, pkgPath)
			}()
		})
	}
}

func TestSafeConvertToUint64(t *testing.T) {
	tests := []struct {
		name      string
		value     *u256.Uint
		wantRes   uint64
		wantPanic bool
		errorMsg  string
	}{
		// Normal cases
		{
			name:      "normal conversion - small",
			value:     u256.NewUint(123),
			wantRes:   123,
			wantPanic: false,
			errorMsg:  "",
		},
		{
			name:      "normal conversion - medium",
			value:     u256.NewUint(1000000),
			wantRes:   1000000,
			wantPanic: false,
			errorMsg:  "",
		},
		{
			name:      "zero",
			value:     u256.NewUint(0),
			wantRes:   0,
			wantPanic: false,
			errorMsg:  "",
		},

		// Boundary cases
		{
			name:      "max uint64 - 2",
			value:     u256.NewUint(18446744073709551613),
			wantRes:   18446744073709551613,
			wantPanic: false,
			errorMsg:  "",
		},
		{
			name:      "max uint64 - 1",
			value:     u256.NewUint(18446744073709551614),
			wantRes:   18446744073709551614,
			wantPanic: false,
			errorMsg:  "",
		},
		{
			name:      "max uint64",
			value:     u256.NewUint(1<<64 - 1),
			wantRes:   1<<64 - 1,
			wantPanic: false,
			errorMsg:  "",
		},

		// Overflow cases
		{
			name:      "overflow - max uint64 + 1",
			value:     u256.MustFromDecimal("18446744073709551616"),
			wantRes:   0,
			wantPanic: true,
			errorMsg:  "[GNOSWAP-POOL-005] out of range for numeric value: amount(18446744073709551616) overflows uint64 range (max 18446744073709551615)",
		},
		{
			name:      "overflow - large value",
			value:     u256.MustFromDecimal(consts.MAX_UINT128),
			wantRes:   0,
			wantPanic: true,
			errorMsg:  "[GNOSWAP-POOL-005] out of range for numeric value: amount(340282366920938463463374607431768211455) overflows uint64 range (max 18446744073709551615)",
		},
		{
			name:      "overflow - very large value",
			value:     u256.MustFromDecimal(consts.MAX_UINT256),
			wantRes:   0,
			wantPanic: true,
			errorMsg:  "[GNOSWAP-POOL-005] out of range for numeric value: amount(115792089237316195423570985008687907853269984665640564039457584007913129639935) overflows uint64 range (max 18446744073709551615)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				uassert.PanicsWithMessage(t, tt.errorMsg, func() {
					safeConvertToUint64(tt.value)
				})
			} else {
				res := safeConvertToUint64(tt.value)
				uassert.Equal(t, tt.wantRes, res)
			}
		})
	}
}

func TestSafeConvertToInt64(t *testing.T) {
	tests := []struct {
		name      string
		value     *u256.Uint
		wantRes   int64
		wantPanic bool
		errorMsg  string
	}{
		// Normal cases
		{"small positive number", u256.NewUint(123), 123, false, ""},
		{"medium positive number", u256.NewUint(1000000), 1000000, false, ""},
		{"zero", u256.NewUint(0), 0, false, ""},

		// Boundary cases for int64
		{"max int64 - 1", u256.NewUint(9223372036854775806), 9223372036854775806, false, ""},
		{"max int64", u256.NewUint(9223372036854775807), 9223372036854775807, false, ""},
		{"max int64 + 1 (overflow)", u256.NewUint(9223372036854775808), 0, true, "[GNOSWAP-POOL-005] out of range for numeric value: amount(9223372036854775808) overflows int64 range (max 9223372036854775807)"},

		// Mid-range boundaries
		{"2^32", u256.NewUint(4294967296), 4294967296, false, ""},
		{"2^48", u256.NewUint(281474976710656), 281474976710656, false, ""},

		// Values between max int64 and max uint64
		{"value between max int64 and max uint64 (1)", u256.NewUint(10000000000000000000), 0, true, "[GNOSWAP-POOL-005] out of range for numeric value: amount(10000000000000000000) overflows int64 range (max 9223372036854775807)"},
		{"value between max int64 and max uint64 (2)", u256.NewUint(15000000000000000000), 0, true, "[GNOSWAP-POOL-005] out of range for numeric value: amount(15000000000000000000) overflows int64 range (max 9223372036854775807)"},

		// Extreme cases - these will fail on uint64 conversion first
		{"max uint64 + 1 (overflow)", u256.MustFromDecimal("18446744073709551616"), 0, true, "[GNOSWAP-POOL-005] out of range for numeric value: amount(18446744073709551616) overflows uint64 range (max 18446744073709551615)"},
		{"very large number (MAX_UINT128)", u256.MustFromDecimal(consts.MAX_UINT128), 0, true, "[GNOSWAP-POOL-005] out of range for numeric value: amount(340282366920938463463374607431768211455) overflows uint64 range (max 18446744073709551615)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.wantPanic {
						t.Errorf("unexpected panic: %v", r)
					}
					return
				}
				if tt.wantPanic {
					t.Errorf("expected panic, but none occurred")
				}
			}()

			res := safeConvertToInt64(tt.value)
			if res != tt.wantRes {
				t.Errorf("safeConvertToInt64() = %v, want %v", res, tt.wantRes)
			}
		})
	}
}

func TestSafeConvertToInt128(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantRes   string
		wantPanic bool
		errorMsg  string
	}{
		// Normal cases
		{"zero", "0", "0", false, ""},
		{"small positive value", "123", "123", false, ""},
		{"medium positive value", "1000000", "1000000", false, ""},

		// Power of 2 boundaries
		{"2^64", consts.Q64, consts.Q64, false, ""},
		{"2^96", consts.Q96, consts.Q96, false, ""},

		// Boundary cases
		{"close to MAX_INT128", "170141183460469231731687303715884105726", "170141183460469231731687303715884105726", false, ""},
		{"MAX_INT128", consts.MAX_INT128, consts.MAX_INT128, false, ""},

		// Overflow cases
		{"just over MAX_INT128", "170141183460469231731687303715884105728", "", true, "[GNOSWAP-POOL-026] overflow: amount(170141183460469231731687303715884105728) overflows int128 range"},
		{"large overflow", consts.MAX_UINT128, "", true, "[GNOSWAP-POOL-026] overflow: amount(340282366920938463463374607431768211455) overflows int128 range"},
		{"extreme overflow", consts.MAX_UINT256, "", true, "[GNOSWAP-POOL-026] overflow: amount(115792089237316195423570985008687907853269984665640564039457584007913129639935) overflows int128 range"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.wantPanic {
						t.Errorf("unexpected panic: %v", r)
					}
					return
				}
				if tt.wantPanic {
					t.Errorf("expected panic, but none occurred")
				}
			}()

			res := safeConvertToInt128(u256.MustFromDecimal(tt.value))
			if res.ToString() != tt.wantRes {
				t.Errorf("safeConvertToInt128() = %v, want %v", res.ToString(), tt.wantRes)
			}
		})
	}
}

func TestToUint128(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "within range - no change",
			input:    "12345",
			expected: "12345",
		},
		{
			name:     "zero value",
			input:    "0",
			expected: "0",
		},
		{
			name:     "max uint128 - no change",
			input:    consts.MAX_UINT128,
			expected: consts.MAX_UINT128,
		},
		{
			name:     "overflow - should mask to max uint128",
			input:    consts.MAX_UINT256,
			expected: consts.MAX_UINT128,
		},
		{
			name:     "overflow - just over max uint128",
			input:    "340282366920938463463374607431768211456", // MAX_UINT128 + 1
			expected: "0",                                       // Should wrap to 0 due to masking
		},
		// Additional masking tests
		{
			name:     "exactly 2^128 wraps to 0",
			input:    "340282366920938463463374607431768211456", // 2^128
			expected: "0",
		},
		{
			name:     "2^128 + 1 wraps to 1",
			input:    "340282366920938463463374607431768211457",
			expected: "1",
		},
		{
			name:     "2^129 wraps to 0",
			input:    "680564733841876926926749214863536422912", // 2^129
			expected: "0",
		},
		{
			name:     "q96 value",
			input:    consts.Q96,
			expected: consts.Q96,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := u256.MustFromDecimal(tt.input)
			originalStr := original.ToString()

			result := toUint128(original)

			// Verify result
			uassert.Equal(t, tt.expected, result.ToString())

			// Verify input wasn't mutated
			uassert.Equal(t, originalStr, original.ToString())
		})
	}
}

func TestToUint128InputPreservation(t *testing.T) {
	tests := []struct {
		name                      string
		input                     string
		expectedResult            string
		expectedOriginal          string
		shouldBeDifferentInstance bool
	}{
		{
			name:                      "overflow case - input preservation",
			input:                     consts.MAX_UINT256,
			expectedResult:            consts.MAX_UINT128,
			expectedOriginal:          consts.MAX_UINT256,
			shouldBeDifferentInstance: true,
		},
		{
			name:                      "within range case - input preservation",
			input:                     "1000",
			expectedResult:            "1000",
			expectedOriginal:          "1000",
			shouldBeDifferentInstance: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := u256.MustFromDecimal(tt.input)
			result := toUint128(original)

			// Original should be unchanged
			uassert.Equal(t, tt.expectedOriginal, original.ToString())
			// Result should have expected value
			uassert.Equal(t, tt.expectedResult, result.ToString())

			// Check instance relationship
			if tt.shouldBeDifferentInstance {
				uassert.True(t, result != original, "result should be a different instance")
			} else {
				uassert.True(t, result == original, "result should be the same instance")
			}
		})
	}
}

func TestToUint128_PowerOfTwoBoundaries(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Q64 (2^64)",
			input:    consts.Q64,
			expected: consts.Q64,
		},
		{
			name:     "Q96 (2^96)",
			input:    consts.Q96,
			expected: consts.Q96,
		},
		{
			name:     "2^127",
			input:    "170141183460469231731687303715884105728",
			expected: "170141183460469231731687303715884105728",
		},
		{
			name:     "Q128 (2^128) wraps to 0",
			input:    consts.Q128,
			expected: "0",
		},
		{
			name:     "2^128 + 1 wraps to 1",
			input:    "340282366920938463463374607431768211457",
			expected: "1",
		},
		{
			name:     "2^129 wraps to 0",
			input:    "680564733841876926926749214863536422912",
			expected: "0",
		},
		{
			name:     "MAX_UINT128 * 2 + 10",
			input:    "680564733841876926926749214863536422920",
			expected: "8",
		},
		// Price boundaries
		{
			name:     "MIN_SQRT_RATIO",
			input:    consts.MIN_SQRT_RATIO,
			expected: consts.MIN_SQRT_RATIO,
		},
		{
			name:     "MAX_SQRT_RATIO overflows uint128",
			input:    consts.MAX_SQRT_RATIO,
			expected: "318775800626314356294205765087544249638",
			// MAX_SQRT_RATIO = 1461446703485210103287273052203988822378723970342
			// This value is used as the maximum price square root in Uniswap V3
			// When masked to uint128 (keeping only lower 128 bits):
			// Binary: ...10101010101010101010101010100110 (lower 128 bits)
			// Decimal: 318775800626314356294205765087544249638
		},
		{
			name:     "MIN_PRICE",
			input:    consts.MIN_PRICE,
			expected: consts.MIN_PRICE,
		},
		{
			name:     "MAX_PRICE overflows uint128",
			input:    consts.MAX_PRICE,
			expected: "318775800626314356294205765087544249637",
			// MAX_PRICE = MAX_SQRT_RATIO - 1 = 1461446703485210103287273052203988822378723970341
			// When masked to uint128:
			// Result: 318775800626314356294205765087544249637 (masked MAX_SQRT_RATIO - 1)
			// This masking ensures price calculations remain within bounds
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := u256.MustFromDecimal(tt.input)
			originalStr := original.ToString()
			
			result := toUint128(original)
			
			uassert.Equal(t, tt.expected, result.ToString())
			uassert.Equal(t, originalStr, original.ToString())
		})
	}
}

func TestVerifyToUint128MaskingLogic(t *testing.T) {
	// 1. Verify toUint128 function logic
	// toUint128 keeps only lower 128 bits when value > MAX_UINT128
	
	// Verify MAX_UINT128
	maxUint128 := u256.MustFromDecimal(consts.MAX_UINT128)
	uassert.Equal(t, "340282366920938463463374607431768211455", maxUint128.ToString())
	uassert.Equal(t, consts.MAX_UINT128, maxUint128.ToString())
	
	// 2. Verify MAX_SQRT_RATIO
	maxSqrtRatio := u256.MustFromDecimal(consts.MAX_SQRT_RATIO)
	uassert.Equal(t, consts.MAX_SQRT_RATIO, "1461446703485210103287273052203988822378723970342")
	uassert.True(t, maxSqrtRatio.Gt(maxUint128), "MAX_SQRT_RATIO should be > MAX_UINT128")
	
	// 3. Verify masking logic
	// mask = (1 << 128) - 1 = MAX_UINT128
	mask := new(u256.Uint).Lsh(u256.One(), 128)
	mask = new(u256.Uint).Sub(mask, u256.One())
	uassert.Equal(t, mask.ToString(), maxUint128.ToString())
	
	// 4. Verify actual masking
	maskedMaxSqrt := new(u256.Uint).And(maxSqrtRatio, mask)
	uassert.Equal(t, "318775800626314356294205765087544249638", maskedMaxSqrt.ToString())
	
	// 5. Verify toUint128 function result matches manual masking
	toUint128Result := toUint128(maxSqrtRatio)
	uassert.Equal(t, maskedMaxSqrt.ToString(), toUint128Result.ToString())
	
	// 6. Verify MAX_PRICE (MAX_SQRT_RATIO - 1)
	maxPrice := u256.MustFromDecimal(consts.MAX_PRICE)
	uassert.Equal(t, consts.MAX_PRICE, "1461446703485210103287273052203988822378723970341")
	
	// Verify MAX_PRICE = MAX_SQRT_RATIO - 1
	maxSqrtMinusOne := new(u256.Uint).Sub(maxSqrtRatio, u256.One())
	uassert.True(t, maxPrice.Eq(maxSqrtMinusOne), "MAX_PRICE should equal MAX_SQRT_RATIO - 1")
	
	// 7. Verify MAX_PRICE masking
	maskedMaxPrice := new(u256.Uint).And(maxPrice, mask)
	uassert.Equal(t, "318775800626314356294205765087544249637", maskedMaxPrice.ToString())
	
	toUint128PriceResult := toUint128(maxPrice)
	uassert.Equal(t, maskedMaxPrice.ToString(), toUint128PriceResult.ToString())
	
	// 8. Verify difference between masked values is 1
	diff := new(u256.Uint).Sub(maskedMaxSqrt, maskedMaxPrice)
	uassert.Equal(t, "1", diff.ToString())
	
	// Final verification
	uassert.Equal(t, "318775800626314356294205765087544249638", toUint128Result.ToString())
	uassert.Equal(t, "318775800626314356294205765087544249637", toUint128PriceResult.ToString())
}

// Additional verification: Ensure masked values are within uint128 range
func TestToUint128MaskedValuesRange(t *testing.T) {
	maxSqrtRatio := u256.MustFromDecimal(consts.MAX_SQRT_RATIO)
	maxUint128 := u256.MustFromDecimal(consts.MAX_UINT128)
	
	// toUint128 results
	resultMaxSqrt := toUint128(maxSqrtRatio)
	resultMaxPrice := toUint128(u256.MustFromDecimal(consts.MAX_PRICE))
	
	// Verify results are <= MAX_UINT128
	uassert.True(t, resultMaxSqrt.Lte(maxUint128), "Masked MAX_SQRT_RATIO should be <= MAX_UINT128")
	uassert.True(t, resultMaxPrice.Lte(maxUint128), "Masked MAX_PRICE should be <= MAX_UINT128")
	
	// Verify specific values
	uassert.Equal(t, "318775800626314356294205765087544249638", resultMaxSqrt.ToString())
	uassert.Equal(t, "318775800626314356294205765087544249637", resultMaxPrice.ToString())
}

//assertOnlyNotNil
func TestAssertOnlyInitializedUint256(t *testing.T) {
	tests := []struct {
		name        string
		input       *u256.Uint
		shouldPanic bool
		errorMsg    string
	}{
		{
			name:        "nil input",
			input:       nil,
			shouldPanic: true,
			errorMsg:    "[GNOSWAP-POOL-006] invalid input data || value is nil",
		},
		{
			name:        "valid zero",
			input:       u256.NewUint(0),
			shouldPanic: false,
		},
		{
			name:        "valid non-zero",
			input:       u256.NewUint(123),
			shouldPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				uassert.PanicsWithMessage(t, tt.errorMsg, func() {
					assertOnlyInitializedUint256(tt.input)
				})
			} else {
				uassert.NotPanics(t, func() {
					assertOnlyInitializedUint256(tt.input)
				})
			}
		})
	}
}

func TestDerivePkgAddr(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Position package path",
			input:    "gno.land/r/gnoswap/v1/position",
			expected: "g1q646ctzhvn60v492x8ucvyqnrj2w30cwh6efk5",
		},
		{
			name:     "Pool package path",
			input:    "gno.land/r/gnoswap/v1/pool",
			expected: "g148tjamj80yyrm309z7rk690an22thd2l3z8ank",
		},
		{
			name:     "Router package path",
			input:    "gno.land/r/gnoswap/v1/router",
			expected: "g1lm2l7tf49h3mykesct7rhfml30yx8dw5xrval7",
		},
		{
			name:     "GNS package path",
			input:    "gno.land/r/gnoswap/v1/gns",
			expected: "g1jgqwaa2le3yr63d533fj785qkjspumzv22ys5m",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := derivePkgAddr(tc.input)
			uassert.Equal(t, tc.expected, got.String())
		})
	}
}

func TestCheckOverFlowInt128(t *testing.T) {
	tests := []struct {
		name        string
		input       *i256.Int
		shouldPanic bool
		errorMsg    string
	}{
		{
			name:        "Valid value within int128 range",
			input:       i256.MustFromDecimal("1"),
			shouldPanic: false,
		},
		{
			name:        "Edge case - MAX_INT128",
			input:       i256.MustFromDecimal(consts.MAX_INT128),
			shouldPanic: false,
		},
		{
			name:        "Overflow case - exceeds MAX_INT128",
			input:       i256.MustFromDecimal("170141183460469231731687303715884105728"), // MAX_INT128 + 1
			shouldPanic: true,
			errorMsg:    "[GNOSWAP-POOL-026] overflow: amount(170141183460469231731687303715884105728) overflows int128 range",
		},
		// Test negative values - they should NOT panic since they're within int128 range
		{
			name:        "negative small value",
			input:       i256.MustFromDecimal("-1"),
			shouldPanic: false,
		},
		{
			name:        "negative large value",
			input:       i256.MustFromDecimal("-170141183460469231731687303715884105728"), // -2^127
			shouldPanic: false,
		},
		{
			name:        "very large positive (MAX_INT256)",
			input:       i256.MustFromDecimal(consts.MAX_INT256),
			shouldPanic: true,
			errorMsg:    "[GNOSWAP-POOL-026] overflow: amount(57896044618658097711785492504343953926634992332820282019728792003956564819967) overflows int128 range",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.shouldPanic {
						t.Errorf("Unexpected panic: %v", r)
					}
				} else if tt.shouldPanic {
					t.Errorf("Expected panic but none occurred")
				}
			}()
			checkOverFlowInt128(tt.input)
		})
	}
}

func TestCheckOverFlowInt128_BoundaryValues(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		shouldPanic bool
		errorMsg    string
	}{
		// Exact boundaries
		{
			name:        "MAX_INT128 exact",
			input:       consts.MAX_INT128,
			shouldPanic: false,
		},
		{
			name:        "MAX_INT128 + 1",
			input:       "170141183460469231731687303715884105728",
			shouldPanic: true,
			errorMsg:    "[GNOSWAP-POOL-026] overflow: amount(170141183460469231731687303715884105728) overflows int128 range",
		},
		// Negative values
		{
			name:        "negative MAX_INT128",
			input:       "-170141183460469231731687303715884105727",
			shouldPanic: false,
		},
		{
			name:        "negative MAX_INT128 - 1",
			input:       "-170141183460469231731687303715884105728",
			shouldPanic: false,
		},
		// Q values
		{
			name:        "Q96 within range",
			input:       consts.Q96,
			shouldPanic: false,
		},
		{
			name:        "Q128 exceeds range",
			input:       consts.Q128,
			shouldPanic: true,
			errorMsg:    "[GNOSWAP-POOL-026] overflow: amount(340282366920938463463374607431768211456) overflows int128 range",
		},
		// Special values
		{
			name:        "zero",
			input:       "0",
			shouldPanic: false,
		},
		{
			name:        "one",
			input:       "1",
			shouldPanic: false,
		},
		{
			name:        "negative one",
			input:       "-1",
			shouldPanic: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := i256.MustFromDecimal(tt.input)
			if tt.shouldPanic {
				uassert.PanicsWithMessage(t, tt.errorMsg, func() {
					checkOverFlowInt128(value)
				})
			} else {
				uassert.NotPanics(t, func() {
					checkOverFlowInt128(value)
				})
			}
		})
	}
}

func TestCheckTickSpacing(t *testing.T) {
	tests := []struct {
		name        string
		tick        int32
		tickSpacing int32
		shouldPanic bool
		errorMsg    string
	}{
		{
			name:        "Valid tick - divisible by tickSpacing",
			tick:        120,
			tickSpacing: 60,
			shouldPanic: false,
		},
		{
			name:        "Valid tick - zero tick",
			tick:        0,
			tickSpacing: 10,
			shouldPanic: false,
		},
		{
			name:        "Invalid tick - not divisible",
			tick:        15,
			tickSpacing: 10,
			shouldPanic: true,
			errorMsg:    "[GNOSWAP-POOL-022] invalid tick and tick spacing requested || tick(15) MOD tickSpacing(10) != 0(5)",
		},
		{
			name:        "Invalid tick - negative tick",
			tick:        -35,
			tickSpacing: 20,
			shouldPanic: true,
			errorMsg:    "[GNOSWAP-POOL-022] invalid tick and tick spacing requested || tick(-35) MOD tickSpacing(20) != 0(-15)",
		},
		// Edge cases with extreme tick values
		{
			name:        "min tick boundary",
			tick:        consts.MIN_TICK, // -887272
			tickSpacing: 1,
			shouldPanic: false,
		},
		{
			name:        "max tick boundary",
			tick:        consts.MAX_TICK, // 887272
			tickSpacing: 1,
			shouldPanic: false,
		},
		{
			name:        "large tick spacing",
			tick:        600,
			tickSpacing: 200,
			shouldPanic: false,
		},
		{
			name:        "negative tick valid spacing",
			tick:        -600,
			tickSpacing: 200,
			shouldPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.shouldPanic {
						t.Errorf("Unexpected panic: %v", r)
					}
				} else if tt.shouldPanic {
					t.Errorf("Expected panic but none occurred")
				}
			}()
			checkTickSpacing(tt.tick, tt.tickSpacing)
		})
	}
}

func TestCheckTickSpacing_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		tick        int32
		tickSpacing int32
		shouldPanic bool
		errorMsg    string
	}{
		// Boundary values with tick spacing 1
		{
			name:        "MIN_TICK with spacing 1",
			tick:        consts.MIN_TICK,
			tickSpacing: 1,
			shouldPanic: false,
		},
		{
			name:        "MAX_TICK with spacing 1",
			tick:        consts.MAX_TICK,
			tickSpacing: 1,
			shouldPanic: false,
		},
		// Fee tier tick spacings
		{
			name:        "fee tier 100 - valid tick",
			tick:        12345,
			tickSpacing: 1,
			shouldPanic: false,
		},
		{
			name:        "fee tier 500 - valid tick",
			tick:        12340,
			tickSpacing: 10,
			shouldPanic: false,
		},
		{
			name:        "fee tier 500 - invalid tick",
			tick:        12345,
			tickSpacing: 10,
			shouldPanic: true,
			errorMsg:    "[GNOSWAP-POOL-022] invalid tick and tick spacing requested || tick(12345) MOD tickSpacing(10) != 0(5)",
		},
		{
			name:        "fee tier 3000 - valid tick",
			tick:        -12360,
			tickSpacing: 60,
			shouldPanic: false,
		},
		{
			name:        "fee tier 3000 - invalid tick",
			tick:        -12359,
			tickSpacing: 60,
			shouldPanic: true,
			errorMsg:    "[GNOSWAP-POOL-022] invalid tick and tick spacing requested || tick(-12359) MOD tickSpacing(60) != 0(-59)",
		},
		{
			name:        "fee tier 10000 - valid tick",
			tick:        -12400,
			tickSpacing: 200,
			shouldPanic: false,
		},
		{
			name:        "fee tier 10000 - invalid tick",
			tick:        -12399,
			tickSpacing: 200,
			shouldPanic: true,
			errorMsg:    "[GNOSWAP-POOL-022] invalid tick and tick spacing requested || tick(-12399) MOD tickSpacing(200) != 0(-199)",
		},
		// Extreme boundary cases
		{
			name:        "near MIN_TICK aligned",
			tick:        -887200,
			tickSpacing: 100,
			shouldPanic: false,
		},
		{
			name:        "near MIN_TICK not aligned",
			tick:        -887271,
			tickSpacing: 100,
			shouldPanic: true,
			errorMsg:    "[GNOSWAP-POOL-022] invalid tick and tick spacing requested || tick(-887271) MOD tickSpacing(100) != 0(-71)",
		},
		{
			name:        "maximum tick spacing",
			tick:        0,
			tickSpacing: 887272,
			shouldPanic: false,
		},
		{
			name:        "maximum tick spacing not aligned",
			tick:        1,
			tickSpacing: 887272,
			shouldPanic: true,
			errorMsg:    "[GNOSWAP-POOL-022] invalid tick and tick spacing requested || tick(1) MOD tickSpacing(887272) != 0(1)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				uassert.PanicsWithMessage(t, tt.errorMsg, func() {
					checkTickSpacing(tt.tick, tt.tickSpacing)
				})
			} else {
				uassert.NotPanics(t, func() {
					checkTickSpacing(tt.tick, tt.tickSpacing)
				})
			}
		})
	}
}

func TestAssertOnlyValidAddress(t *testing.T) {
	tests := []struct {
		name     string
		addr     std.Address
		expected bool
		errorMsg string
	}{
		{
			name:     "Success - valid address",
			addr:     "g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d", // ADMIN address from consts
			expected: true,
		},
		{
			name:     "Failure - invalid address",
			addr:     "g1lmvrrrr4er2us84h2732sru76c9zl2nvknha8", // invalid length
			expected: false,
			errorMsg: "[GNOSWAP-POOL-023] invalid address || (g1lmvrrrr4er2us84h2732sru76c9zl2nvknha8)",
		},
		{
			name:     "Failure - empty address",
			addr:     "",
			expected: false,
			errorMsg: "[GNOSWAP-POOL-023] invalid address || ()",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expected {
				uassert.NotPanics(t, func() {
					assertOnlyValidAddress(tc.addr)
				})
			} else {
				uassert.PanicsWithMessage(t, tc.errorMsg, func() {
					assertOnlyValidAddress(tc.addr)
				})
			}
		})
	}
}

func TestAssertOnlyValidAddress_SpecialCases(t *testing.T) {
	tests := []struct {
		name        string
		addr        std.Address
		shouldPanic bool
		errorMsg    string
	}{
		{
			name:        "zero address is valid",
			addr:        consts.ZERO_ADDRESS,
			shouldPanic: false,
		},
		{
			name:        "admin address",
			addr:        consts.ADMIN,
			shouldPanic: false,
		},
		{
			name:        "pool address",
			addr:        consts.POOL_ADDR,
			shouldPanic: false,
		},
		{
			name:        "empty string",
			addr:        "",
			shouldPanic: true,
			errorMsg:    "[GNOSWAP-POOL-023] invalid address || ()",
		},
		{
			name:        "invalid format - wrong length",
			addr:        "g1lmvrrrr4er2us84h2732sru76c9zl2nvknha", // 39 chars instead of 40
			shouldPanic: true,
			errorMsg:    "[GNOSWAP-POOL-023] invalid address || (g1lmvrrrr4er2us84h2732sru76c9zl2nvknha)",
		},
		{
			name:        "too short",
			addr:        "g1lmvrrrr4er2us84h2732",
			shouldPanic: true,
			errorMsg:    "[GNOSWAP-POOL-023] invalid address || (g1lmvrrrr4er2us84h2732)",
		},
		{
			name:        "wrong prefix",
			addr:        "x17290cwvmrapvp869xfnhhawa8sm9edpufzat7d",
			shouldPanic: true,
			errorMsg:    "[GNOSWAP-POOL-023] invalid address || (x17290cwvmrapvp869xfnhhawa8sm9edpufzat7d)",
		},
		{
			name:        "invalid characters",
			addr:        "g1!@#$%^&*()_+{}[]|\\:;\"'<>,.?/~`12345678",
			shouldPanic: true,
			errorMsg:    "[GNOSWAP-POOL-023] invalid address || (g1!@#$%^&*()_+{}[]|\\:;\"'<>,.?/~`12345678)",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				uassert.PanicsWithMessage(t, tt.errorMsg, func() {
					assertOnlyValidAddress(tt.addr)
				})
			} else {
				uassert.NotPanics(t, func() {
					assertOnlyValidAddress(tt.addr)
				})
			}
		})
	}
}

func TestPermissionFunctions(t *testing.T) {
	t.Run("user realm detection", func(t *testing.T) {
		testing.SetRealm(std.NewUserRealm(std.Address("user")))
		func() {
			testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/tests"))
			uassert.True(t, isUserCall())
		}()
	})

	t.Run("code realm detection", func(t *testing.T) {
		testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/realm"))
		func() {
			testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/tests"))
			uassert.False(t, isUserCall())
		}()
	})
}

// Additional edge case tests for better coverage
func TestConversionBoundaryValues(t *testing.T) {
	t.Run("uint64 power of 2 boundaries", func(t *testing.T) {
		// Test various power of 2 values
		pow32 := u256.NewUint(1 << 32) // 4,294,967,296
		result := safeConvertToUint64(pow32)
		uassert.Equal(t, uint64(1<<32), result)

		pow48 := u256.NewUint(1 << 48)
		result = safeConvertToUint64(pow48)
		uassert.Equal(t, uint64(1<<48), result)

		pow63 := u256.NewUint(1 << 63) // 9,223,372,036,854,775,808
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic for 2^63, but none occurred")
			}
		}()
		safeConvertToInt64(pow63)
	})

	t.Run("edge case q values", func(t *testing.T) {
		// Q96 should work with int128
		q96 := u256.MustFromDecimal(consts.Q96)
		result := safeConvertToInt128(q96)
		uassert.Equal(t, consts.Q96, result.ToString())

		// Q128 is actually 2^128 which is larger than MAX_INT128 (2^127 - 1)
		// So it should panic
		q128Val := u256.MustFromDecimal(consts.Q128)
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic for Q128 value, but none occurred")
			}
		}()
		safeConvertToInt128(q128Val)
	})
}

func TestSafeConversions_ExactBoundaries(t *testing.T) {
	t.Run("uint64 exact boundary", func(t *testing.T) {
		// MAX_UINT64
		maxValue := u256.MustFromDecimal(consts.MAX_UINT64)
		result := safeConvertToUint64(maxValue)
		uassert.Equal(t, consts.UINT64_MAX, result)
		
		// MAX_UINT64 + 1
		overValue := new(u256.Uint).Add(maxValue, u256.One())
		uassert.PanicsWithMessage(t, 
			"[GNOSWAP-POOL-005] out of range for numeric value: amount(18446744073709551616) overflows uint64 range (max 18446744073709551615)",
			func() { safeConvertToUint64(overValue) })
	})
	
	t.Run("int64 exact boundary", func(t *testing.T) {
		// MAX_INT64
		maxValue := u256.MustFromDecimal(consts.MAX_INT64)
		result := safeConvertToInt64(maxValue)
		uassert.Equal(t, consts.INT64_MAX, result)
		
		// MAX_INT64 + 1
		overValue := new(u256.Uint).Add(maxValue, u256.One())
		uassert.PanicsWithMessage(t,
			"[GNOSWAP-POOL-005] out of range for numeric value: amount(9223372036854775808) overflows int64 range (max 9223372036854775807)",
			func() { safeConvertToInt64(overValue) })
	})
	
	t.Run("int128 exact boundary", func(t *testing.T) {
		// MAX_INT128
		maxValue := u256.MustFromDecimal(consts.MAX_INT128)
		result := safeConvertToInt128(maxValue)
		uassert.Equal(t, consts.MAX_INT128, result.ToString())
		
		// MAX_INT128 + 1
		overValue := new(u256.Uint).Add(maxValue, u256.One())
		uassert.PanicsWithMessage(t,
			"[GNOSWAP-POOL-026] overflow: amount(170141183460469231731687303715884105728) overflows int128 range",
			func() { safeConvertToInt128(overValue) })
	})
}

func TestFormatFunctions(t *testing.T) {
	t.Run("formatUint", func(t *testing.T) {
		tests := []struct {
			name     string
			input    any
			expected string
			panics   bool
			errorMsg string
		}{
			{
				name:     "uint8",
				input:    uint8(255),
				expected: "255",
				panics:   false,
				errorMsg: "",
			},
			{
				name:     "uint32",
				input:    uint32(4294967295),
				expected: "4294967295",
				panics:   false,
				errorMsg: "",
			},
			{
				name:     "uint64",
				input:    uint64(18446744073709551615),
				expected: "18446744073709551615",
				panics:   false,
				errorMsg: "",
			},
			{
				name:     "uint8 zero",
				input:    uint8(0),
				expected: "0",
				panics:   false,
				errorMsg: "",
			},
			{
				name:     "invalid type string",
				input:    "string",
				expected: "",
				panics:   true,
				errorMsg: "invalid type: string",
			},
			{
				name:     "invalid type int",
				input:    int(42),
				expected: "",
				panics:   true,
				errorMsg: "invalid type: int",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if tt.panics {
					uassert.PanicsWithMessage(t, tt.errorMsg, func() {
						formatUint(tt.input)
					})
				} else {
					result := formatUint(tt.input)
					uassert.Equal(t, tt.expected, result)
				}
			})
		}
	})

	t.Run("formatInt", func(t *testing.T) {
		tests := []struct {
			name     string
			input    any
			expected string
			panics   bool
			errorMsg string
		}{
			{
				name:     "int32",
				input:    int32(-2147483648),
				expected: "-2147483648",
				panics:   false,
				errorMsg: "",
			},
			{
				name:     "int64",
				input:    int64(9223372036854775807),
				expected: "9223372036854775807",
				panics:   false,
				errorMsg: "",
			},
			{
				name:     "int",
				input:    int(42),
				expected: "42",
				panics:   false,
				errorMsg: "",
			},
			{
				name:     "int negative",
				input:    int(-42),
				expected: "-42",
				panics:   false,
				errorMsg: "",
			},
			{
				name:     "int32 zero",
				input:    int32(0),
				expected: "0",
				panics:   false,
				errorMsg: "",
			},
			{
				name:     "invalid type string",
				input:    "string",
				expected: "",
				panics:   true,
				errorMsg: "invalid type: string",
			},
			{
				name:     "invalid type uint",
				input:    uint(42),
				expected: "",
				panics:   true,
				errorMsg: "invalid type: uint",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if tt.panics {
					uassert.PanicsWithMessage(t, tt.errorMsg, func() {
						formatInt(tt.input)
					})
				} else {
					result := formatInt(tt.input)
					uassert.Equal(t, tt.expected, result)
				}
			})
		}
	})

	t.Run("formatBool", func(t *testing.T) {
		tests := []struct {
			name     string
			input    bool
			expected string
		}{
			{
				name:     "true",
				input:    true,
				expected: "true",
			},
			{
				name:     "false",
				input:    false,
				expected: "false",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := formatBool(tt.input)
				uassert.Equal(t, tt.expected, result)
			})
		}
	})
}

func TestRealm(t *testing.T) {
	t.Run("getPrevRealm", func(t *testing.T) {
		testing.SetRealm(std.NewUserRealm(std.Address("test_user")))
		func() {
			testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/tests"))
			r := getPrevRealm()
			uassert.True(t, r.IsUser())
			uassert.Equal(t, "test_user", r.Address().String())
		}()
	})

	t.Run("getPrevAddr", func(t *testing.T) {
		testing.SetRealm(std.NewUserRealm(std.Address("test_user2")))
		func() {
			testing.SetRealm(std.NewCodeRealm("gno.land/r/demo/tests"))
			addr := getPrevAddr()
			uassert.Equal(t, "test_user2", addr.String())
		}()
	})
}