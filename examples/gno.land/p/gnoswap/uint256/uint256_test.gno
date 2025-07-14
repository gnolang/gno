package uint256

import (
	"testing"

	"gno.land/p/demo/uassert"
)

func TestFromDecimal(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		shouldPanic bool
		panicMsg    string
	}{
		// Basic cases
		{
			name:     "zero",
			input:    "0",
			expected: "0",
		},
		{
			name:     "one",
			input:    "1",
			expected: "1",
		},
		{
			name:     "max_uint64",
			input:    "18446744073709551615",
			expected: "18446744073709551615",
		},
		{
			name:     "max_uint128",
			input:    "340282366920938463463374607431768211455",
			expected: "340282366920938463463374607431768211455",
		},
		{
			name:     "max_uint256",
			input:    MAX_UINT256,
			expected: MAX_UINT256,
		},

		// Format cases
		{
			name:     "leading_zeros",
			input:    "00000000000000000000000001234567890",
			expected: "1234567890",
		},
		{
			name:     "plus_sign",
			input:    "+12345",
			expected: "12345",
		},

		// Error cases
		{
			name:        "max_uint256_plus_one",
			input:       "115792089237316195423570985008687907853269984665640564039457584007913129639936",
			shouldPanic: true,
			panicMsg:    "hex number > 256 bits",
		},
		{
			name:        "multiple_plus_signs",
			input:       "++12345",
			shouldPanic: true,
		},
		{
			name:        "negative_number",
			input:       "-12345",
			shouldPanic: true,
		},
		{
			name:        "empty_string",
			input:       "",
			shouldPanic: true,
			panicMsg:    "EOF",
		},
		{
			name:        "invalid_characters",
			input:       "123abc456",
			shouldPanic: true,
		},
		{
			name:        "spaces_in_number",
			input:       "123 456",
			shouldPanic: true,
		},
		{
			name:        "decimal_point",
			input:       "123.456",
			shouldPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				if tt.panicMsg != "" {
					uassert.PanicsWithMessage(t, tt.panicMsg, func() {
						MustFromDecimal(tt.input)
					})
				} else {
					defer func() {
						if r := recover(); r == nil {
							t.Errorf("Expected panic but got none")
						}
					}()
					MustFromDecimal(tt.input)
				}
			} else {
				result := MustFromDecimal(tt.input)
				uassert.Equal(t, tt.expected, result.ToString())
			}
		})
	}
}

func TestFromHex(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		shouldPanic bool
		panicMsg    string
	}{
		// Basic cases
		{
			name:     "zero",
			input:    "0x0",
			expected: "0",
		},
		{
			name:     "one",
			input:    "0x1",
			expected: "1",
		},
		{
			name:     "max_uint256",
			input:    "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
			expected: MAX_UINT256,
		},

		// Format cases
		{
			name:     "uppercase_0X",
			input:    "0Xff",
			expected: "255",
		},
		{
			name:     "mixed_case",
			input:    "0xAbCdEf",
			expected: "11259375",
		},

		// Error cases
		{
			name:        "hex_overflow_67_chars",
			input:       "0x10000000000000000000000000000000000000000000000000000000000000000",
			shouldPanic: true,
			panicMsg:    "hex number > 256 bits",
		},
		{
			name:        "no_0x_prefix",
			input:       "ffffffff",
			shouldPanic: true,
			panicMsg:    "UnmarshalText: ffffffff: hex string without 0x prefix",
		},
		{
			name:        "empty_string",
			input:       "",
			shouldPanic: true,
			panicMsg:    "UnmarshalText: : empty hex string",
		},
		{
			name:        "only_0x",
			input:       "0x",
			shouldPanic: true,
			panicMsg:    "UnmarshalText: 0x: hex string \"0x\"",
		},
		{
			name:        "invalid_hex_chars",
			input:       "0xgg",
			shouldPanic: true,
			panicMsg:    "invalid hex string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				if tt.panicMsg != "" {
					uassert.PanicsWithMessage(t, tt.panicMsg, func() {
						MustFromHex(tt.input)
					})
				} else {
					defer func() {
						if r := recover(); r == nil {
							t.Errorf("Expected panic but got none")
						}
					}()
					MustFromHex(tt.input)
				}
			} else {
				result := MustFromHex(tt.input)
				uassert.Equal(t, tt.expected, result.ToString())
			}
		})
	}
}

func TestComparisonOperations(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		x         string
		y         string
		expected  bool
	}{
		// Equality tests
		{
			name:      "eq_true",
			operation: "eq",
			x:         "12345",
			y:         "12345",
			expected:  true,
		},
		{
			name:      "eq_false",
			operation: "eq",
			x:         "12345",
			y:         "12346",
			expected:  false,
		},

		// Less than tests
		{
			name:      "lt_true",
			operation: "lt",
			x:         "12345",
			y:         "12346",
			expected:  true,
		},
		{
			name:      "lt_false_greater",
			operation: "lt",
			x:         "12346",
			y:         "12345",
			expected:  false,
		},
		{
			name:      "lt_false_equal",
			operation: "lt",
			x:         "12345",
			y:         "12345",
			expected:  false,
		},

		// Greater than tests
		{
			name:      "gt_true",
			operation: "gt",
			x:         "12346",
			y:         "12345",
			expected:  true,
		},
		{
			name:      "gt_false_smaller",
			operation: "gt",
			x:         "12345",
			y:         "12346",
			expected:  false,
		},
		{
			name:      "gt_false_equal",
			operation: "gt",
			x:         "12345",
			y:         "12345",
			expected:  false,
		},

		// Less than or equal tests
		{
			name:      "lte_true_less",
			operation: "lte",
			x:         "12345",
			y:         "12346",
			expected:  true,
		},
		{
			name:      "lte_true_equal",
			operation: "lte",
			x:         "12345",
			y:         "12345",
			expected:  true,
		},
		{
			name:      "lte_false",
			operation: "lte",
			x:         "12346",
			y:         "12345",
			expected:  false,
		},

		// Greater than or equal tests
		{
			name:      "gte_true_greater",
			operation: "gte",
			x:         "12346",
			y:         "12345",
			expected:  true,
		},
		{
			name:      "gte_true_equal",
			operation: "gte",
			x:         "12345",
			y:         "12345",
			expected:  true,
		},
		{
			name:      "gte_false",
			operation: "gte",
			x:         "12345",
			y:         "12346",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x := MustFromDecimal(tt.x)
			y := MustFromDecimal(tt.y)

			var result bool
			switch tt.operation {
			case "eq":
				result = x.Eq(y)
			case "lt":
				result = x.Lt(y)
			case "gt":
				result = x.Gt(y)
			case "lte":
				result = x.Lte(y)
			case "gte":
				result = x.Gte(y)
			}

			uassert.Equal(t, tt.expected, result)
		})
	}
}

func TestBitOperations(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		x         string
		n         uint
		expected  string
	}{
		// Left shift tests
		{
			name:      "lsh_zero",
			operation: "lsh",
			x:         "0",
			n:         100,
			expected:  "0",
		},
		{
			name:      "lsh_one_by_zero",
			operation: "lsh",
			x:         "1",
			n:         0,
			expected:  "1",
		},
		{
			name:      "lsh_one_by_one",
			operation: "lsh",
			x:         "1",
			n:         1,
			expected:  "2",
		},
		{
			name:      "lsh_one_by_64",
			operation: "lsh",
			x:         "1",
			n:         64,
			expected:  "18446744073709551616",
		},
		{
			name:      "lsh_one_by_255",
			operation: "lsh",
			x:         "1",
			n:         255,
			expected:  "57896044618658097711785492504343953926634992332820282019728792003956564819968",
		},
		{
			name:      "lsh_one_by_256",
			operation: "lsh",
			x:         "1",
			n:         256,
			expected:  "0", // Shifts out of range
		},

		// Right shift tests
		{
			name:      "rsh_zero",
			operation: "rsh",
			x:         "0",
			n:         100,
			expected:  "0",
		},
		{
			name:      "rsh_by_zero",
			operation: "rsh",
			x:         "123456",
			n:         0,
			expected:  "123456",
		},
		{
			name:      "rsh_by_one",
			operation: "rsh",
			x:         "123456",
			n:         1,
			expected:  "61728",
		},
		{
			name:      "rsh_by_64",
			operation: "rsh",
			x:         "340282366920938463463374607431768211456", // 2^128
			n:         64,
			expected:  "18446744073709551616", // 2^64
		},
		{
			name:      "rsh_max_by_255",
			operation: "rsh",
			x:         MAX_UINT256,
			n:         255,
			expected:  "1",
		},
		{
			name:      "rsh_max_by_256",
			operation: "rsh",
			x:         MAX_UINT256,
			n:         256,
			expected:  "0", // Shifts out of range
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x := MustFromDecimal(tt.x)
			var result *Uint

			switch tt.operation {
			case "lsh":
				result = new(Uint).Lsh(x, tt.n)
			case "rsh":
				result = new(Uint).Rsh(x, tt.n)
			}

			uassert.Equal(t, tt.expected, result.ToString())
		})
	}
}

func TestByteOperation(t *testing.T) {
	tests := []struct {
		name     string
		x        string
		n        uint64
		expected string
	}{
		{
			name:     "byte_0_of_max",
			x:        MAX_UINT256,
			n:        0,
			expected: "255",
		},
		{
			name:     "byte_31_of_max",
			x:        MAX_UINT256,
			n:        31,
			expected: "255",
		},
		{
			name:     "byte_32_out_of_range",
			x:        MAX_UINT256,
			n:        32,
			expected: "0",
		},
		{
			name:     "byte_0_of_256",
			x:        "256",
			n:        30,
			expected: "1",
		},
		{
			name:     "byte_31_of_single_byte",
			x:        "255",
			n:        31,
			expected: "255",
		},
		{
			name:     "byte_at_boundary",
			x:        "18446744073709551616", // 2^64
			n:        23,
			expected: "1",
		},
		{
			name:     "byte_in_middle",
			x:        "0xff00ff00ff00ff00ff00", // Hexadecimal format without leading zeros
			n:        15,
			expected: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var x *Uint
			if len(tt.x) >= 2 && tt.x[:2] == "0x" {
				x = MustFromHex(tt.x)
			} else {
				x = MustFromDecimal(tt.x)
			}

			n := NewUint(tt.n)
			result := x.Byte(n)

			uassert.Equal(t, tt.expected, result.ToString())
		})
	}
}

func TestMulDiv(t *testing.T) {
	tests := []struct {
		name        string
		x           string
		y           string
		denominator string
		expected    string
		shouldPanic bool
		panicMsg    string
	}{
		{
			name:        "reverts_if_denominator_is_0",
			x:           "340282366920938463463374607431768211456", // Q128
			y:           "5",
			denominator: "0",
			shouldPanic: true,
			panicMsg:    "denominator must be greater than 0. got 0",
		},
		{
			name:        "reverts_if_output_overflows_uint256",
			x:           "340282366920938463463374607431768211456", // Q128
			y:           "340282366920938463463374607431768211456", // Q128
			denominator: "1",
			shouldPanic: true,
			panicMsg:    "overflow: denominator(1) must be greater than hi(1)",
		},
		{
			name:        "all_max_inputs",
			x:           MAX_UINT256,
			y:           MAX_UINT256,
			denominator: MAX_UINT256,
			expected:    MAX_UINT256,
		},
		{
			name:        "simple_case_no_remainder",
			x:           "1000000",
			y:           "1000000",
			denominator: "1000",
			expected:    "1000000000",
		},
		{
			name:        "accurate_without_phantom_overflow",
			x:           "340282366920938463463374607431768211456", // Q128
			y:           "170141183460469231731687303715884105728", // 50 * Q128 / 100
			denominator: "510423550381407695195061911147652317184", // 150 * Q128 / 100
			expected:    "113427455640312821154458202477256070485", // Q128 / 3
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x := MustFromDecimal(tt.x)
			y := MustFromDecimal(tt.y)
			denominator := MustFromDecimal(tt.denominator)

			if tt.shouldPanic {
				uassert.PanicsWithMessage(t, tt.panicMsg, func() {
					MulDiv(x, y, denominator)
				})
			} else {
				result := MulDiv(x, y, denominator)
				uassert.Equal(t, tt.expected, result.ToString())
			}
		})
	}
}

func TestMulDivRoundingUpSimple(t *testing.T) {
	tests := []struct {
		name        string
		x           string
		y           string
		denominator string
		expected    string
		shouldPanic bool
		panicMsg    string
	}{
		{
			name:        "reverts_if_denominator_is_0",
			x:           "340282366920938463463374607431768211456", // Q128
			y:           "5",
			denominator: "0",
			shouldPanic: true,
			panicMsg:    "denominator must be greater than 0. got 0",
		},
		{
			name:        "rounds_up_when_remainder",
			x:           "5",
			y:           "2",
			denominator: "3",
			expected:    "4", // (5*2)/3 = 3.33, rounded up to 4
		},
		{
			name:        "no_rounding_when_exact",
			x:           "6",
			y:           "3",
			denominator: "2",
			expected:    "9", // (6*3)/2 = 9, no rounding needed
		},
		{
			name:        "accurate_with_rounding",
			x:           "340282366920938463463374607431768211456",   // Q128
			y:           "17014118346046923173168730371588410572800", // 50 * Q128
			denominator: "51042355038140769519506191114765231718400", // 150 * Q128
			expected:    "113427455640312821154458202477256070486",   // Q128/3 + 1
		},
		{
			name:        "max_result_no_remainder",
			x:           MAX_UINT256,
			y:           "1",
			denominator: "1",
			expected:    MAX_UINT256,
		},
		{
			name:        "rounding_overflow",
			x:           MAX_UINT256,
			y:           "2",
			denominator: "1",
			shouldPanic: true,
			panicMsg:    "overflow: denominator(1) must be greater than hi(1)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x := MustFromDecimal(tt.x)
			y := MustFromDecimal(tt.y)
			denominator := MustFromDecimal(tt.denominator)

			if tt.shouldPanic {
				uassert.PanicsWithMessage(t, tt.panicMsg, func() {
					MulDivRoundingUp(x, y, denominator)
				})
			} else {
				result := MulDivRoundingUp(x, y, denominator)
				uassert.Equal(t, tt.expected, result.ToString())
			}
		})
	}
}

func TestBitLenAndByteLen(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedBit  int
		expectedByte int
	}{
		{
			name:         "zero",
			input:        "0",
			expectedBit:  0,
			expectedByte: 0,
		},
		{
			name:         "one",
			input:        "1",
			expectedBit:  1,
			expectedByte: 1,
		},
		{
			name:         "byte_boundary",
			input:        "255", // 2^8 - 1
			expectedBit:  8,
			expectedByte: 1,
		},
		{
			name:         "word_boundary",
			input:        "18446744073709551615", // 2^64 - 1
			expectedBit:  64,
			expectedByte: 8,
		},
		{
			name:         "max_uint256",
			input:        MAX_UINT256,
			expectedBit:  256,
			expectedByte: 32,
		},
		{
			name:         "half_max",
			input:        "57896044618658097711785492504343953926634992332820282019728792003956564819968", // 2^255
			expectedBit:  256,
			expectedByte: 32,
		},
		{
			name:         "uneven_bits",
			input:        "123456789",
			expectedBit:  27,
			expectedByte: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x := MustFromDecimal(tt.input)
			bitLen := x.BitLen()
			byteLen := x.ByteLen()

			uassert.Equal(t, tt.expectedBit, bitLen)
			uassert.Equal(t, tt.expectedByte, byteLen)
		})
	}
}

func TestInPlaceSafety(t *testing.T) {
	// Testing that operations don't mutate their inputs
	t.Run("original_values_preserved", func(t *testing.T) {
		a := MustFromDecimal("1234567890")
		b := MustFromDecimal("9876543210")

		original_a := a.Clone()
		original_b := b.Clone()

		_ = new(Uint).Add(a, b)

		uassert.True(t, a.Eq(original_a))
		uassert.True(t, b.Eq(original_b))
	})

	// Testing that chained operations work correctly
	t.Run("chained_operations", func(t *testing.T) {
		a := MustFromDecimal("1000000")
		b := MustFromDecimal("2000000")
		c := MustFromDecimal("3000000")

		// (a + b) * c
		temp := new(Uint).Add(a, b)
		result := new(Uint).Mul(temp, c)

		expected := MustFromDecimal("9000000000000")
		uassert.True(t, result.Eq(expected))
	})
}

func TestOperationConsistency(t *testing.T) {
	// Testing mathematical properties
	t.Run("addition_associativity", func(t *testing.T) {
		a := NewUint(12345)
		b := NewUint(67890)
		c := NewUint(11111)

		// (a + b) + c
		path1 := new(Uint).Add(a, b)
		path1 = new(Uint).Add(path1, c)

		// a + (b + c)
		path2 := new(Uint).Add(b, c)
		path2 = new(Uint).Add(a, path2)

		uassert.True(t, path1.Eq(path2))
	})

	t.Run("distributive_property", func(t *testing.T) {
		a := NewUint(12345)
		b := NewUint(67)
		c := NewUint(89)

		// a * (b + c)
		sum := new(Uint).Add(b, c)
		dist1 := new(Uint).Mul(a, sum)

		// (a * b) + (a * c)
		prod1 := new(Uint).Mul(a, b)
		prod2 := new(Uint).Mul(a, c)
		dist2 := new(Uint).Add(prod1, prod2)

		uassert.True(t, dist1.Eq(dist2))
	})

	t.Run("inverse_operations", func(t *testing.T) {
		x := MustFromDecimal("123456789012345678901234567890")

		// x + y - y = x
		y := MustFromDecimal("999999999999999999999999999999")
		result := new(Uint).Add(x, y)
		result = new(Uint).Sub(result, y)
		uassert.True(t, result.Eq(x))

		// x * y / y = x (when no remainder)
		y = NewUint(12345)
		result = new(Uint).Mul(x, y)
		result = new(Uint).Div(result, y)
		uassert.True(t, result.Eq(x))

		// x << n >> n = x (when n < bitlen(x))
		n := uint(10)
		result = new(Uint).Lsh(x, n)
		result = new(Uint).Rsh(result, n)
		uassert.True(t, result.Eq(x))
	})
}
