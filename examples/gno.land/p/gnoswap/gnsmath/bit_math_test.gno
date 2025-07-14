package gnsmath

import (
	"testing"

	"gno.land/p/demo/uassert"
	u256 "gno.land/p/gnoswap/uint256"
)

func TestBitMathMostSignificantBit(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    uint8
		shouldPanic bool
	}{
		{
			name:        "zero_panics",
			input:       "0",
			shouldPanic: true,
		},

		// Basic cases
		{
			name:     "one",
			input:    "1",
			expected: 0,
		},
		{
			name:     "two",
			input:    "2",
			expected: 1,
		},

		// Small numbers
		{
			name:     "three",
			input:    "3",
			expected: 1,
		},
		{
			name:     "four",
			input:    "4",
			expected: 2,
		},
		{
			name:     "five",
			input:    "5",
			expected: 2,
		},
		{
			name:     "255",
			input:    "255",
			expected: 7,
		},
		{
			name:     "256",
			input:    "256",
			expected: 8,
		},
		{
			name:     "257",
			input:    "257",
			expected: 8,
		},

		// Boundary values for each shift level
		{
			name:     "2^16-1",
			input:    "65535",
			expected: 15,
		},
		{
			name:     "2^16",
			input:    "65536",
			expected: 16,
		},
		{
			name:     "2^16+1",
			input:    "65537",
			expected: 16,
		},
		{
			name:     "2^32-1",
			input:    "4294967295",
			expected: 31,
		},
		{
			name:     "2^32",
			input:    "4294967296",
			expected: 32,
		},
		{
			name:     "2^32+1",
			input:    "4294967297",
			expected: 32,
		},
		{
			name:     "2^64-1",
			input:    "18446744073709551615",
			expected: 63,
		},
		{
			name:     "2^64",
			input:    "18446744073709551616",
			expected: 64,
		},
		{
			name:     "2^64+1",
			input:    "18446744073709551617",
			expected: 64,
		},
		{
			name:     "2^128-1",
			input:    "340282366920938463463374607431768211455",
			expected: 127,
		},
		{
			name:     "2^128",
			input:    "340282366920938463463374607431768211456",
			expected: 128,
		},
		{
			name:     "2^128+1",
			input:    "340282366920938463463374607431768211457",
			expected: 128,
		},

		// Large numbers
		{
			name:     "2^255",
			input:    "57896044618658097711785492504343953926634992332820282019728792003956564819968",
			expected: 255,
		},
		{
			name:     "max_uint256",
			input:    "115792089237316195423570985008687907853269984665640564039457584007913129639935",
			expected: 255,
		},

		// Special bit patterns
		{
			name:     "alternating_bits_0x5555",
			input:    "21845", // 0x5555
			expected: 14,
		},
		{
			name:     "alternating_bits_0xAAAA",
			input:    "43690", // 0xAAAA
			expected: 15,
		},
		{
			name:     "all_ones_32bit",
			input:    "4294967295", // 0xFFFFFFFF
			expected: 31,
		},
		{
			name:     "high_bit_only_64",
			input:    "9223372036854775808", // 0x8000000000000000
			expected: 63,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldPanic {
				uassert.PanicsWithMessage(t, errMSBZeroInput.Error(), func() {
					BitMathMostSignificantBit(u256.MustFromDecimal(tc.input))
				})
				return
			}

			x := u256.MustFromDecimal(tc.input)
			result := BitMathMostSignificantBit(x)
			uassert.Equal(t, result, tc.expected)
		})
	}
}

func TestBitMathMostSignificantBit_PowersOfTwo(t *testing.T) {
	// Test all powers of 2 from 2^0 to 2^255
	for i := 0; i < 256; i++ {
		num := new(u256.Uint).Lsh(u256.One(), uint(i))
		result := BitMathMostSignificantBit(num)
		uassert.Equal(t, result, uint8(i))
	}
}

func TestBitMathLeastSignificantBit(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    uint8
		shouldPanic bool
	}{
		{
			name:        "zero_panics",
			input:       "0",
			shouldPanic: true,
		},

		// Basic cases
		{
			name:     "one",
			input:    "1",
			expected: 0,
		},
		{
			name:     "two",
			input:    "2",
			expected: 1,
		},

		// Odd numbers (LSB = 0)
		{
			name:     "three",
			input:    "3",
			expected: 0,
		},
		{
			name:     "five",
			input:    "5",
			expected: 0,
		},
		{
			name:     "255",
			input:    "255",
			expected: 0,
		},
		{
			name:     "max_uint256_odd",
			input:    "115792089237316195423570985008687907853269984665640564039457584007913129639935",
			expected: 0,
		},

		// Even numbers
		{
			name:     "four",
			input:    "4",
			expected: 2,
		},
		{
			name:     "six",
			input:    "6",
			expected: 1,
		},
		{
			name:     "eight",
			input:    "8",
			expected: 3,
		},
		{
			name:     "256",
			input:    "256",
			expected: 8,
		},
		{
			name:     "1024",
			input:    "1024",
			expected: 10,
		},

		// Powers of 2 boundaries
		{
			name:     "2^16",
			input:    "65536",
			expected: 16,
		},
		{
			name:     "2^32",
			input:    "4294967296",
			expected: 32,
		},
		{
			name:     "2^64",
			input:    "18446744073709551616",
			expected: 64,
		},
		{
			name:     "2^128",
			input:    "340282366920938463463374607431768211456",
			expected: 128,
		},
		{
			name:     "2^255",
			input:    "57896044618658097711785492504343953926634992332820282019728792003956564819968",
			expected: 255,
		},

		// Special patterns
		{
			name:     "0xF0",
			input:    "240",
			expected: 4,
		},
		{
			name:     "0xFF00",
			input:    "65280",
			expected: 8,
		},
		{
			name:     "0xFFFF0000",
			input:    "4294901760",
			expected: 16,
		},

		// Multiple bits with different LSB
		{
			name:     "12", // 0b1100
			input:    "12",
			expected: 2,
		},
		{
			name:     "96", // 0b1100000
			input:    "96",
			expected: 5,
		},
		{
			name:     "192", // 0b11000000
			input:    "192",
			expected: 6,
		},

		// Large numbers with low LSB
		{
			name:     "2^255+2^10",
			input:    "57896044618658097711785492504343953926634992332820282019728792003956564820992",
			expected: 10,
		},
		{
			name:     "2^200+2^20",
			input:    "1606938044258990275541962092341162602522202993782792835302425600",
			expected: 10,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldPanic {
				uassert.PanicsWithMessage(t, errLSBZeroInput.Error(), func() {
					BitMathLeastSignificantBit(u256.MustFromDecimal(tc.input))
				})
				return
			}

			x := u256.MustFromDecimal(tc.input)
			result := BitMathLeastSignificantBit(x)
			uassert.Equal(t, result, tc.expected)
		})
	}
}

func TestBitMathLeastSignificantBit_PowersOfTwo(t *testing.T) {
	// Test all powers of 2 from 2^0 to 2^255
	for i := 0; i < 256; i++ {
		num := new(u256.Uint).Lsh(u256.One(), uint(i))
		result := BitMathLeastSignificantBit(num)
		uassert.Equal(t, result, uint8(i))
	}
}

func TestBitMath_Consistency(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"small_number", "42"},
		{"medium_number", "1234567890"},
		{"large_number", "123456789012345678901234567890123456789"},
		{"half_max", "57896044618658097711785492504343953926634992332820282019728792003956564819967"},
		{"near_max", "115792089237316195423570985008687907853269984665640564039457584007913129639934"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			x := u256.MustFromDecimal(tc.input)
			msb := BitMathMostSignificantBit(x)
			lsb := BitMathLeastSignificantBit(x)

			// MSB must be >= LSB
			if msb < lsb {
				t.Errorf("Invalid result: MSB(%d) < LSB(%d) for %s", msb, lsb, tc.input)
			}
		})
	}
}

func TestBitMath_SingleBitConsistency(t *testing.T) {
	// For single bit numbers, MSB == LSB
	for i := 0; i < 256; i++ {
		x := new(u256.Uint).Lsh(u256.One(), uint(i))
		msb := BitMathMostSignificantBit(x)
		lsb := BitMathLeastSignificantBit(x)

		if msb != lsb || msb != uint8(i) {
			t.Errorf("Single bit 2^%d: MSB=%d, LSB=%d, expected both=%d", i, msb, lsb, i)
		}
	}
}

// Additional test cases to add to the existing test file

func TestBitMathLeastSignificantBit_MaxUint256MinusOne(t *testing.T) {
	// MAX_UINT256 - 1 = ...639934 (even number, LSB should be 1)
	maxMinusOne := new(u256.Uint).Sub(
		u256.MustFromDecimal("115792089237316195423570985008687907853269984665640564039457584007913129639935"),
		u256.One(),
	)

	result := BitMathLeastSignificantBit(maxMinusOne)
	uassert.Equal(t, result, uint8(1))
}

func TestBitMath_PseudoRandomLargeNumbers(t *testing.T) {
	testCases := []struct {
		name   string
		input  string
		expMSB uint8
		expLSB uint8
	}{
		{
			name:   "large_random_1",
			input:  "98765432109876543210987654321098765432109876543210987654321098765432",
			expMSB: 225,
			expLSB: 3,
		},
		{
			name:   "large_random_2",
			input:  "11111111111111111111111111111111111111111111111111111111111111111111",
			expMSB: 222,
			expLSB: 0,
		},
		{
			name:   "large_random_3",
			input:  "99999999999999999999999999999999999999999999999999999999999999999999",
			expMSB: 225,
			expLSB: 0,
		},
		{
			name:   "large_prime_like",
			input:  "57896044618658097711785492504343953926634992332820282019728792003956564819973",
			expMSB: 255,
			expLSB: 0,
		},
		{
			name:   "fibonacci_large",
			input:  "354224848179261915075",
			expMSB: 68,
			expLSB: 0,
		},
		{
			name:   "near_max_even",
			input:  "115792089237316195423570985008687907853269984665640564039457584007913129639934",
			expMSB: 255,
			expLSB: 1,
		},
		{
			name:   "half_max",
			input:  "57896044618658097711785492504343953926634992332820282019728792003956564819967",
			expMSB: 254,
			expLSB: 0,
		},
		{
			name:   "quarter_max",
			input:  "28948022309329048855892746252171976963317496166410141009864396001978282409983",
			expMSB: 253,
			expLSB: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			x := u256.MustFromDecimal(tc.input)

			msb := BitMathMostSignificantBit(x)
			lsb := BitMathLeastSignificantBit(x)

			uassert.Equal(t, msb, tc.expMSB)
			uassert.Equal(t, lsb, tc.expLSB)

			// Verify MSB >= LSB
			if msb < lsb {
				t.Errorf("Invalid: MSB(%d) < LSB(%d)", msb, lsb)
			}
		})
	}
}

func TestBitMath_ActualValues(t *testing.T) {
	testValues := []struct {
		name  string
		input string
	}{
		{"2^200+2^20", "1606938044258990275541962092341162602522202993782792835302425600"},
		{"large1", "98765432109876543210987654321098765432109876543210987654321098765432"},
		{"large2", "11111111111111111111111111111111111111111111111111111111111111111111"},
		{"large3", "99999999999999999999999999999999999999999999999999999999999999999999"},
		{"half_max", "57896044618658097711785492504343953926634992332820282019728792003956564819967"},
		{"quarter_max", "28948022309329048855892746252171976963317496166410141009864396001978282409983"},
	}

	for _, tc := range testValues {
		t.Run(tc.name, func(t *testing.T) {
			x := u256.MustFromDecimal(tc.input)
			msb := BitMathMostSignificantBit(x)
			lsb := BitMathLeastSignificantBit(x)

			t.Logf("%s: MSB=%d, LSB=%d", tc.name, msb, lsb)

			temp := x.Clone()
			actualLSB := uint8(0)
			for i := uint8(0); i < 255; i++ {
				if new(u256.Uint).And(temp, u256.One()).Eq(u256.One()) {
					actualLSB = i
					break
				}
				temp.Rsh(temp, 1)
			}

			if actualLSB != lsb {
				t.Errorf("Calculated LSB %d doesn't match actual %d", lsb, actualLSB)
			}
		})
	}
}

func TestBitMath_VerifyCalculations(t *testing.T) {
	verifyTests := []struct {
		name        string
		input       string
		binaryForm  string
		expectedMSB uint8
		expectedLSB uint8
	}{
		{
			name:        "complex_number_with_msb_209",
			input:       "1606938044258990275541962092341162602522202993782792835302425600",
			binaryForm:  "complex number",
			expectedMSB: 209,
			expectedLSB: 10,
		},
		{
			name:        "verify_12",
			input:       "12",
			binaryForm:  "0b1100",
			expectedMSB: 3,
			expectedLSB: 2,
		},
		{
			name:        "verify_96",
			input:       "96",
			binaryForm:  "0b1100000",
			expectedMSB: 6,
			expectedLSB: 5,
		},
	}

	for _, tc := range verifyTests {
		t.Run(tc.name, func(t *testing.T) {
			x := u256.MustFromDecimal(tc.input)

			msb := BitMathMostSignificantBit(x)
			lsb := BitMathLeastSignificantBit(x)

			if msb != tc.expectedMSB {
				t.Errorf("MSB for %s: got %d, expected %d", tc.name, msb, tc.expectedMSB)
			}
			if lsb != tc.expectedLSB {
				t.Errorf("LSB for %s: got %d, expected %d", tc.name, lsb, tc.expectedLSB)
			}
		})
	}
}

func TestBitMath_ShiftTableCompleteness(t *testing.T) {
	// Verify msbShifts covers all necessary ranges
	t.Run("msb_shifts_coverage", func(t *testing.T) {
		expectedShifts := []uint{128, 64, 32, 16, 8, 4, 2, 1}

		if len(msbShifts) != len(expectedShifts) {
			t.Errorf("msbShifts has %d entries, expected %d", len(msbShifts), len(expectedShifts))
		}

		for i, expected := range expectedShifts {
			if i < len(msbShifts) && msbShifts[i].shift != expected {
				t.Errorf("msbShifts[%d].shift = %d, expected %d", i, msbShifts[i].shift, expected)
			}
		}

		// Verify bit patterns are correct powers of 2
		for i, s := range msbShifts {
			// Calculate expected value: 2^shift
			expected := new(u256.Uint).Lsh(u256.One(), s.shift)
			if !s.bitPattern.Eq(expected) {
				t.Errorf("msbShifts[%d].bitPattern incorrect for shift %d", i, s.shift)
			}
		}
	})

	// Verify lsbShifts covers all necessary ranges
	t.Run("lsb_shifts_coverage", func(t *testing.T) {
		expectedShifts := []uint{128, 64, 32, 16, 8, 4, 2, 1}

		if len(lsbShifts) != len(expectedShifts) {
			t.Errorf("lsbShifts has %d entries, expected %d", len(lsbShifts), len(expectedShifts))
		}

		for i, expected := range expectedShifts {
			if i < len(lsbShifts) && lsbShifts[i].shift != expected {
				t.Errorf("lsbShifts[%d].shift = %d, expected %d", i, lsbShifts[i].shift, expected)
			}
		}

		// Verify bit patterns are correct (2^shift - 1)
		for i, s := range lsbShifts {
			// Calculate expected value: 2^shift - 1
			powerOfTwo := new(u256.Uint).Lsh(u256.One(), s.shift)
			expected := new(u256.Uint).Sub(powerOfTwo, u256.One())

			if !s.bitPattern.Eq(expected) {
				t.Errorf("lsbShifts[%d].bitPattern incorrect for shift %d", i, s.shift)
			}
		}
	})
}

// TestBitMathInputPreservation verifies that bit math functions do not mutate their input.
func TestBitMathInputPreservation(t *testing.T) {
	cases := []struct {
		name string
		val  string
	}{
		{"zero_plus_one", "1"},
		{"small", "12345"},
		{"power_of_two_small", "256"},         // 2^8
		{"power_of_two_medium", "4294967296"}, // 2^32
		{"power_of_two_large", "340282366920938463463374607431768211456"},                                 // 2^128
		{"max_uint256", "115792089237316195423570985008687907853269984665640564039457584007913129639935"}, // 2^256 - 1
		{"random_large", "98765432109876543210987654321098765432109876543210987654321098765432"},
		{"all_bits_set_lower", "4294967295"},         // 0xFFFFFFFF
		{"alternating_bits", "12297829382473034410"}, // 0xAAAAAAAAAAAAAAAA
	}

	for _, tc := range cases {
		t.Run(tc.name+"/MSB", func(t *testing.T) {
			x := u256.MustFromDecimal(tc.val)
			before := x.Clone()
			_ = BitMathMostSignificantBit(x)

			// Verify input was not mutated
			if !x.Eq(before) {
				t.Errorf("MSB mutated input: before=%s, after=%s",
					before.ToString(), x.ToString())
			}
		})

		t.Run(tc.name+"/LSB", func(t *testing.T) {
			x := u256.MustFromDecimal(tc.val)
			before := x.Clone()
			_ = BitMathLeastSignificantBit(x)

			// Verify input was not mutated
			if !x.Eq(before) {
				t.Errorf("LSB mutated input: before=%s, after=%s",
					before.ToString(), x.ToString())
			}
		})
	}
}

// TestBitMathIdempotency ensures that multiple calls with the same input produce consistent results
func TestBitMathIdempotency(t *testing.T) {
	values := []string{
		"42",
		"65535",
		"340282366920938463463374607431768211456",
		"98765432109876543210987654321098765432109876543210987654321098765432",
	}

	for _, v := range values {
		x := u256.MustFromDecimal(v)

		// Call MSB multiple times and verify consistency
		msb1 := BitMathMostSignificantBit(x)
		msb2 := BitMathMostSignificantBit(x)
		msb3 := BitMathMostSignificantBit(x)

		if msb1 != msb2 || msb2 != msb3 {
			t.Errorf("MSB inconsistent for %s: %d, %d, %d", v, msb1, msb2, msb3)
		}

		// Call LSB multiple times and verify consistency
		lsb1 := BitMathLeastSignificantBit(x)
		lsb2 := BitMathLeastSignificantBit(x)
		lsb3 := BitMathLeastSignificantBit(x)

		if lsb1 != lsb2 || lsb2 != lsb3 {
			t.Errorf("LSB inconsistent for %s: %d, %d, %d", v, lsb1, lsb2, lsb3)
		}

		// Final verification that input remains unchanged after multiple calls
		if !x.Eq(u256.MustFromDecimal(v)) {
			t.Errorf("Value mutated after multiple calls: %s", v)
		}
	}
}

// TestBitMathMemoryReuse specifically tests temp.Clone() is used instead of mutating the input parameter.
func TestBitMathMemoryReuse(t *testing.T) {
	// Test with 2^128 - a value that triggers multiple shift operations
	v := "340282366920938463463374607431768211456"

	t.Run("MSB_no_reallocation", func(t *testing.T) {
		x := u256.MustFromDecimal(v)

		// Store the original pointer value (*Uint itself, not address of x)
		originalPtr := x

		// Execute the function that previously had mutation issues
		msb := BitMathMostSignificantBit(x)

		// Verify the pointer hasn't changed (no reallocation)
		if x != originalPtr {
			t.Error("Pointer value changed - unexpected reallocation occurred")
		}

		// Verify the value hasn't been mutated
		if !x.Eq(u256.MustFromDecimal(v)) {
			t.Error("Value was mutated")
		}

		// Verify the result is correct
		if msb != 128 {
			t.Errorf("Incorrect MSB: got %d, expected 128", msb)
		}
	})

	t.Run("LSB_no_reallocation", func(t *testing.T) {
		x := u256.MustFromDecimal(v)

		// Store the original pointer value (*Uint itself, not address of x)
		originalPtr := x

		// Execute the function
		lsb := BitMathLeastSignificantBit(x)

		// Verify the pointer hasn't changed (no reallocation)
		if x != originalPtr {
			t.Error("Pointer value changed - unexpected reallocation occurred")
		}

		// Verify the value hasn't been mutated
		if !x.Eq(u256.MustFromDecimal(v)) {
			t.Error("Value was mutated")
		}

		// Verify the result is correct (2^128 has LSB at position 128)
		if lsb != 128 {
			t.Errorf("Incorrect LSB: got %d, expected 128", lsb)
		}
	})
}

func TestBitMath_SpecialPatterns(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expMSB uint8
		expLSB uint8
		desc   string
	}{
		{
			name:   "all_even_bits",
			input:  "6148914691236517205", // 0x5555555555555555
			expMSB: 62,
			expLSB: 0,
			desc:   "alternating 01 pattern",
		},
		{
			name:   "all_odd_bits",
			input:  "12297829382473034410", // 0xAAAAAAAAAAAAAAAA
			expMSB: 63,
			expLSB: 1,
			desc:   "alternating 10 pattern",
		},
		{
			name:   "one_bit_per_byte",
			input:  "72340172838076673", // 0x0101010101010101
			expMSB: 56,
			expLSB: 0,
			desc:   "one bit set per byte",
		},
		{
			name:   "high_byte_only",
			input:  "18374686479671623680", // 0xFF00000000000000
			expMSB: 63,
			expLSB: 56,
			desc:   "only highest byte set",
		},
		{
			name:   "mersenne_127",
			input:  "170141183460469231731687303715884105727", // 2^127 - 1
			expMSB: 126,
			expLSB: 0,
			desc:   "Mersenne prime 2^127-1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			x := u256.MustFromDecimal(tc.input)

			msb := BitMathMostSignificantBit(x)
			lsb := BitMathLeastSignificantBit(x)

			uassert.Equal(t, msb, tc.expMSB)
			uassert.Equal(t, lsb, tc.expLSB)
		})
	}
}

// Benchmark-style exhaustive test for small ranges
func TestBitMath_ExhaustiveSmallRange(t *testing.T) {
	// Test first 1000 numbers exhaustively
	for i := uint64(1); i <= 1000; i++ {
		x := u256.NewUint(i)
		msb := BitMathMostSignificantBit(x)
		lsb := BitMathLeastSignificantBit(x)

		// Manual calculation for verification
		expectedMSB := uint8(0)
		// Start from bit 63 and work down to find the highest set bit
		for bit := 63; bit >= 0; bit-- {
			if i&(1<<uint(bit)) != 0 {
				expectedMSB = uint8(bit)
				break
			}
		}

		expectedLSB := uint8(0)
		for bit := uint8(0); bit < 64; bit++ {
			if i&(1<<bit) != 0 {
				expectedLSB = bit
				break
			}
		}

		if msb != expectedMSB {
			t.Errorf("MSB mismatch for %d: got %d, expected %d", i, msb, expectedMSB)
		}
		if lsb != expectedLSB {
			t.Errorf("LSB mismatch for %d: got %d, expected %d", i, lsb, expectedLSB)
		}
	}
}
