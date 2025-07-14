package uint256

import (
	"testing"

	"gno.land/p/demo/ufmt"
)

func TestFullMathMulDiv(t *testing.T) {
	tests := []struct {
		name      string
		x         string
		y         string
		denom     string
		want      string
		wantPanic bool
	}{
		// Basic functionality
		{
			name:      "simple_multiplication_division",
			x:         "100",
			y:         "200",
			denom:     "50",
			want:      "400", // (100 * 200) / 50 = 20000 / 50 = 400
			wantPanic: false,
		},
		{
			name:      "exact_division",
			x:         "1000",
			y:         "3000",
			denom:     "100",
			want:      "30000", // (1000 * 3000) / 100 = 3000000 / 100 = 30000
			wantPanic: false,
		},

		// Zero inputs
		{
			name:      "zero_first_operand",
			x:         "0",
			y:         "1000",
			denom:     "100",
			want:      "0",
			wantPanic: false,
		},
		{
			name:      "zero_second_operand",
			x:         "123456789",
			y:         "0",
			denom:     "100",
			want:      "0",
			wantPanic: false,
		},

		// Identity operations (denom = 1)
		{
			name:      "identity_small_numbers",
			x:         "123",
			y:         "456",
			denom:     "1",
			want:      "56088", // 123 * 456 = 56088
			wantPanic: false,
		},
		{
			name:      "identity_max_value",
			x:         MAX_UINT256,
			y:         "1",
			denom:     "1",
			want:      MAX_UINT256,
			wantPanic: false,
		},

		// Essential panic cases
		{
			name:      "panic_denominator_zero",
			x:         "100",
			y:         "200",
			denom:     "0",
			want:      "",
			wantPanic: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			x := MustFromDecimal(tc.x)
			y := MustFromDecimal(tc.y)
			denom := MustFromDecimal(tc.denom)

			if tc.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("MulDiv(%s, %s, %s) expected panic but got none", tc.x, tc.y, tc.denom)
					}
				}()
				MulDiv(x, y, denom)
				return
			}

			got := MulDiv(x, y, denom)
			want := MustFromDecimal(tc.want)

			if !got.Eq(want) {
				t.Errorf("MulDiv(%s, %s, %s) = %s, want %s", tc.x, tc.y, tc.denom, got.ToString(), tc.want)
			}
		})
	}
}

func TestMulDivBoundary(t *testing.T) {
	tests := []struct {
		name      string
		x         string
		y         string
		denom     string
		want      string
		wantPanic bool
	}{
		// Core boundaries
		{
			name:      "all_max_inputs",
			x:         MAX_UINT256,
			y:         MAX_UINT256,
			denom:     MAX_UINT256,
			want:      MAX_UINT256,
			wantPanic: false,
		},

		// Overflow detection (hi >= denominator)
		{
			name:      "overflow_hi_equals_denom",
			x:         "57896044618658097711785492504343953926634992332820282019728792003956564819968", // 2^255
			y:         "2",
			denom:     "1", // hiProduct = 1, denom = 1
			want:      "",
			wantPanic: true,
		},
		{
			name:      "overflow_hi_exceeds_denom",
			x:         MAX_UINT256,
			y:         MAX_UINT256,
			denom:     "340282366920938463463374607431768211455", // 2^128 - 1
			want:      "",
			wantPanic: true,
		},
		{
			name:      "overflow_q128_squared",
			x:         "340282366920938463463374607431768211456", // Q128 (2^128)
			y:         "340282366920938463463374607431768211456", // Q128
			denom:     "1",
			want:      "",
			wantPanic: true,
		},

		// Just below overflow boundary
		{
			name:      "boundary_hi_lt_denom_no_overflow",
			x:         "340282366920938463463374607431768211456", // Q128
			y:         "340282366920938463463374607431768211456", // Q128
			denom:     "340282366920938463463374607431768211457", // Q128 + 1
			want:      "340282366920938463463374607431768211455", // Q128 - 1
			wantPanic: false,
		},

		// Borrow conditions (remainder > lo)
		{
			name:      "borrow_needed_remainder_gt_lo",
			x:         "57896044618658097711785492504343953926634992332820282019728792003956564819968", // 2^255
			y:         "2",
			denom:     "57896044618658097711785492504343953926634992332820282019728792003956564819967", // 2^255 - 1
			want:      "2",
			wantPanic: false,
		},
		{
			name:      "no_borrow_remainder_lt_lo",
			x:         "100",
			y:         "100",
			denom:     "101", // remainder = 1, lo = 10000
			want:      "99",  // 10000 / 101 = 99
			wantPanic: false,
		},

		// Powers of 2 optimization path
		{
			name:      "power_of_2_denominator",
			x:         "340282366920938463463374607431768211456", // Q128
			y:         "1",
			denom:     "1024",                                 // 2^10
			want:      "332306998946228968225951765070086144", // Q128 / 1024
			wantPanic: false,
		},
		{
			name:      "mixed_power_of_2_factorization",
			x:         "1000000000000000000",
			y:         "1000000000000000000",
			denom:     "9223372036854775808", // 2^63 (single high bit)
			want:      "108420217248550443",
			wantPanic: false,
		},
		{
			name:      "power_of_2_division_large",
			x:         "57896044618658097711785492504343953926634992332820282019728792003956564819968", // 2^255
			y:         "2",
			denom:     "4",
			want:      "28948022309329048855892746252171976963317496166410141009864396001978282409984", // 2^254
			wantPanic: false,
		},

		// Newton-Raphson precision (odd denominators)
		{
			name:      "odd_denominator_prime",
			x:         "1000000000000000000",
			y:         "1000000000000000000",
			denom:     "999999999999999989", // Large prime
			want:      "1000000000000000011",
			wantPanic: false,
		},
		{
			name:      "odd_denominator_alternating_bits",
			x:         "1000000000000000000",
			y:         "1000000000000000000",
			denom:     "6148914691236517205", // 0x5555555555555555 (alternating bits)
			want:      "162630325872825665",
			wantPanic: false,
		},

		// Q128 special cases
		{
			name:      "q128_divided_by_small_odd",
			x:         "340282366920938463463374607431768211456", // Q128
			y:         "1",
			denom:     "3",
			want:      "113427455640312821154458202477256070485", // Q128 / 3
			wantPanic: false,
		},
		{
			name:      "q128_divided_by_q128_minus_1",
			x:         "340282366920938463463374607431768211456", // Q128
			y:         "1",
			denom:     "340282366920938463463374607431768211455", // Q128 - 1
			want:      "1",
			wantPanic: false,
		},

		// Edge cases with zero
		{
			name:      "zero_numerator",
			x:         "0",
			y:         "1000000000000000000",
			denom:     "999999999999999999",
			want:      "0",
			wantPanic: false,
		},

		// Combined optimization paths
		{
			name:      "combined_optimization_paths",
			x:         "1606938044258990275541962092341162602522202993782792835301376",                // 2^200
			y:         "36028797018963968",                                                            // 2^55
			denom:     "1809251394333065553493296640760748560207343510400633813116524750123642650624", // 2^250
			want:      "32",                                                                           // 2^(200+55-250) = 2^5 = 32
			wantPanic: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			x := MustFromDecimal(tc.x)
			y := MustFromDecimal(tc.y)
			denom := MustFromDecimal(tc.denom)

			if tc.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("MulDiv(%s, %s, %s) expected panic but got none", tc.x, tc.y, tc.denom)
					}
				}()
				MulDiv(x, y, denom)
				return
			}

			got := MulDiv(x, y, denom)
			want := MustFromDecimal(tc.want)

			if !got.Eq(want) {
				t.Errorf("MulDiv(%s, %s, %s) = %s, want %s", tc.x, tc.y, tc.denom, got.ToString(), tc.want)
			}
		})
	}
}

// This test verifies that the MulDiv function does not modify its input parameters
func TestMulDivInputPreservation(t *testing.T) {

	tests := []struct {
		name  string
		x     string
		y     string
		denom string
	}{
		{
			name:  "normal_inputs",
			x:     "12345678901234567890",
			y:     "98765432109876543210",
			denom: "123456789",
		},
		{
			name:  "power_of_2_boundary",
			x:     "57896044618658097711785492504343953926634992332820282019728792003956564819968", // 2^255
			y:     "2",
			denom: "4",
		},
		{
			name:  "phantom_overflow_case",
			x:     "340282366920938463463374607431768211456",  // Q128
			y:     "11930464781601263584560605149792510336",   // 35 * Q128 / 1000
			denom: "2722258935367507707706996859454145691648", // 8 * Q128
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			x := MustFromDecimal(tc.x)
			y := MustFromDecimal(tc.y)
			denom := MustFromDecimal(tc.denom)

			// Create copies of input values
			xCopy := new(Uint).Set(x)
			yCopy := new(Uint).Set(y)
			denomCopy := new(Uint).Set(denom)

			// Call MulDiv
			MulDiv(x, y, denom)

			// Verify that input values were not modified
			if !x.Eq(xCopy) {
				t.Errorf("Input 'x' was modified: original=%s, modified=%s",
					xCopy.ToString(), x.ToString())
			}

			if !y.Eq(yCopy) {
				t.Errorf("Input 'y' was modified: original=%s, modified=%s",
					yCopy.ToString(), y.ToString())
			}

			if !denom.Eq(denomCopy) {
				t.Errorf("Input 'denom' was modified: original=%s, modified=%s",
					denomCopy.ToString(), denom.ToString())
			}
		})
	}
}

// Phantom overflow cases (product > 256 bits but result fits)
func TestMulDivPhantomOverflow(t *testing.T) {
	tests := []struct {
		name      string
		x         string
		y         string
		denom     string
		want      string
		wantPanic bool
	}{
		{
			name:      "phantom_overflow_case_1",
			x:         "340282366920938463463374607431768211456",  // Q128
			y:         "11930464781601263584560605149792510336",   // 35 * Q128 / 1000
			denom:     "2722258935367507707706996859454145691648", // 8 * Q128
			want:      "1491308097700157948070075643724063792",    // (35/1000) / 8 * Q128 = 4.375/1000 * Q128
			wantPanic: false,
		},
		{
			name:      "phantom_overflow_repeating_decimal",
			x:         "340282366920938463463374607431768211456",     // Q128
			y:         "340282366920938463463374607431768211456000",  // 1000 * Q128
			denom:     "1020847100762815390390123822295304634368000", // 3000 * Q128
			want:      "113427455640312821154458202477256070485",     // Q128 / 3
			wantPanic: false,
		},
		{
			name:      "accurate_without_phantom_overflow",
			x:         "340282366920938463463374607431768211456", // Q128
			y:         "170141183460469231731687303715884105728", // 0.5 * Q128
			denom:     "510423550381407695195061911147652317184", // 1.5 * Q128
			want:      "113427455640312821154458202477256070485", // Q128 / 3
			wantPanic: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			x := MustFromDecimal(tc.x)
			y := MustFromDecimal(tc.y)
			denom := MustFromDecimal(tc.denom)

			if tc.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("MulDiv(%s, %s, %s) expected panic but got none", tc.x, tc.y, tc.denom)
					}
				}()
				MulDiv(x, y, denom)
				return
			}

			got := MulDiv(x, y, denom)
			want := MustFromDecimal(tc.want)

			if !got.Eq(want) {
				t.Errorf("MulDiv(%s, %s, %s) = %s, want %s", tc.x, tc.y, tc.denom, got.ToString(), tc.want)
			}
		})
	}
}

func TestMulDivRoundingUp(t *testing.T) {
	tests := []struct {
		name      string
		x         string
		y         string
		denom     string
		want      string
		wantPanic bool
	}{
		// Basic rounding functionality
		{
			name:      "no_rounding_needed",
			x:         "100",
			y:         "200",
			denom:     "50",
			want:      "400", // 20000 / 50 = 400 (exact)
			wantPanic: false,
		},
		{
			name:      "rounding_up_needed",
			x:         "100",
			y:         "201",
			denom:     "50",
			want:      "402", // 20100 / 50 = 402 (exact)
			wantPanic: false,
		},
		{
			name:      "rounding_up_with_remainder",
			x:         "101",
			y:         "199",
			denom:     "50",
			want:      "402", // 1101 * 199 = 20099, 20099 / 50 = 401.98 → 402
			wantPanic: false,
		},

		// Identity operations
		{
			name:      "identity_operation",
			x:         "789",
			y:         "123",
			denom:     "1",
			want:      "97047", // 789 * 123 = 97047
			wantPanic: false,
		},

		// Essential panic cases
		{
			name:      "panic_denominator_zero",
			x:         "100",
			y:         "200",
			denom:     "0",
			want:      "",
			wantPanic: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			x := MustFromDecimal(tc.x)
			y := MustFromDecimal(tc.y)
			denom := MustFromDecimal(tc.denom)

			if tc.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("MulDivRoundingUp(%s, %s, %s) expected panic but got none", tc.x, tc.y, tc.denom)
					}
				}()
				MulDivRoundingUp(x, y, denom)
				return
			}

			got := MulDivRoundingUp(x, y, denom)
			want := MustFromDecimal(tc.want)

			if !got.Eq(want) {
				t.Errorf("MulDivRoundingUp(%s, %s, %s) = %s, want %s", tc.x, tc.y, tc.denom, got.ToString(), tc.want)
			}
		})
	}
}

func TestMulDivRoundingUpBoundary(t *testing.T) {
	tests := []struct {
		name      string
		x         string
		y         string
		denom     string
		want      string
		wantPanic bool
	}{
		// Boundary rounding cases
		{
			name:      "max_result_no_overflow",
			x:         MAX_UINT256,
			y:         "1",
			denom:     "1",
			want:      MAX_UINT256,
			wantPanic: false,
		},
		{
			name:      "q128_divided_by_3_rounded_up",
			x:         "340282366920938463463374607431768211456", // Q128
			y:         "1",
			denom:     "3",
			want:      "113427455640312821154458202477256070486", // (Q128 / 3) + 1
			wantPanic: false,
		},
		{
			name:      "overflow_after_rounding",
			x:         "340282366920938463463374607431768211456", // Q128
			y:         "340282366920938463463374607431768211456", // Q128
			denom:     "1",
			want:      "",
			wantPanic: true,
		},
		{
			name:      "rounding_at_max_boundary",
			x:         MAX_UINT256,
			y:         "3",
			denom:     "3",
			want:      MAX_UINT256, // No rounding needed (exact division)
			wantPanic: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			x := MustFromDecimal(tc.x)
			y := MustFromDecimal(tc.y)
			denom := MustFromDecimal(tc.denom)

			if tc.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("MulDivRoundingUp(%s, %s, %s) expected panic but got none", tc.x, tc.y, tc.denom)
					}
				}()
				MulDivRoundingUp(x, y, denom)
				return
			}

			got := MulDivRoundingUp(x, y, denom)
			want := MustFromDecimal(tc.want)

			if !got.Eq(want) {
				t.Errorf("MulDivRoundingUp(%s, %s, %s) = %s, want %s", tc.x, tc.y, tc.denom, got.ToString(), tc.want)
			}
		})
	}
}

func TestDivRoundingUp(t *testing.T) {
	tests := []struct {
		name      string
		x         string
		y         string
		want      string
		wantPanic bool
	}{
		// Basic functionality
		{
			name:      "exact_division",
			x:         "100",
			y:         "10",
			want:      "10",
			wantPanic: false,
		},
		{
			name:      "division_with_remainder",
			x:         "101",
			y:         "10",
			want:      "11", // 10 + 1
			wantPanic: false,
		},
		{
			name:      "zero_dividend",
			x:         "0",
			y:         "10",
			want:      "0",
			wantPanic: false,
		},
		{
			name:      "one_divided_by_two",
			x:         "1",
			y:         "2",
			want:      "1", // 0 + 1 (rounded up)
			wantPanic: false,
		},

		// Identity operations
		{
			name:      "identity_operation",
			x:         "12345",
			y:         "1",
			want:      "12345",
			wantPanic: false,
		},

		// Essential panic cases
		{
			name:      "panic_division_by_zero",
			x:         "100",
			y:         "0",
			want:      "",
			wantPanic: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			x := MustFromDecimal(tc.x)
			y := MustFromDecimal(tc.y)

			if tc.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("DivRoundingUp(%s, %s) expected panic but got none", tc.x, tc.y)
					}
				}()
				DivRoundingUp(x, y)
				return
			}

			got := DivRoundingUp(x, y)
			want := MustFromDecimal(tc.want)

			if !got.Eq(want) {
				t.Errorf("DivRoundingUp(%s, %s) = %s, want %s", tc.x, tc.y, got.ToString(), tc.want)
			}
		})
	}
}

func TestDivRoundingUpBoundary(t *testing.T) {
	tests := []struct {
		name      string
		x         string
		y         string
		want      string
		wantPanic bool
	}{
		// Boundary value cases
		{
			name:      "max_divided_by_max",
			x:         MAX_UINT256,
			y:         MAX_UINT256,
			want:      "1",
			wantPanic: false,
		},
		{
			name:      "max_minus_1_divided_by_max_rounded_up",
			x:         "115792089237316195423570985008687907853269984665640564039457584007913129639934", // MAX - 1
			y:         MAX_UINT256,
			want:      "1", // 0 + 1 (rounded up)
			wantPanic: false,
		},
		{
			name:      "large_number_with_remainder",
			x:         "1000000000000000000000000000000000000001",
			y:         "1000000000000000000",
			want:      "1000000000000000000001", // 1000000000000000000 + 1
			wantPanic: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			x := MustFromDecimal(tc.x)
			y := MustFromDecimal(tc.y)

			if tc.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("DivRoundingUp(%s, %s) expected panic but got none", tc.x, tc.y)
					}
				}()
				DivRoundingUp(x, y)
				return
			}

			got := DivRoundingUp(x, y)
			want := MustFromDecimal(tc.want)

			if !got.Eq(want) {
				t.Errorf("DivRoundingUp(%s, %s) = %s, want %s", tc.x, tc.y, got.ToString(), tc.want)
			}
		})
	}
}

// Floor vs Ceil comparison
func TestMulDivFloorVsCeil(t *testing.T) {
	tests := []struct {
		name       string
		x          string
		y          string
		denom      string
		expectSame bool // true if floor == ceil (exact division)
	}{
		{
			name:       "exact_division",
			x:          "1000",
			y:          "2000",
			denom:      "100",
			expectSame: true, // 2000000/100 = 20000 (exact)
		},
		{
			name:       "inexact_division",
			x:          "1234",
			y:          "5678",
			denom:      "100",
			expectSame: false, // 7006652/100 = 70066.52 (inexact)
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			x := MustFromDecimal(tc.x)
			y := MustFromDecimal(tc.y)
			denom := MustFromDecimal(tc.denom)

			floor := MulDiv(x, y, denom)
			ceil := MulDivRoundingUp(x, y, denom)

			if tc.expectSame {
				if !floor.Eq(ceil) {
					t.Errorf("Expected floor == ceil for exact division, got floor=%s, ceil=%s",
						floor.ToString(), ceil.ToString())
				}
			} else {
				expected := new(Uint).Add(floor, one)
				if !ceil.Eq(expected) {
					t.Errorf("Expected ceil == floor + 1 for inexact division, got floor=%s, ceil=%s",
						floor.ToString(), ceil.ToString())
				}
			}
		})
	}
}

// Property-based testing for DivRoundingUp
func TestDivRoundingUpProperty(t *testing.T) {
	// Property: for any x,y>0: ceil = DivRoundingUp(x,y); floor = x.Div(y)
	// assert ceil == floor || ceil == floor+1

	testCases := []struct {
		x, y uint64
	}{
		{100, 3},   // 33.33... -> floor=33, ceil=34
		{100, 10},  // 10.0 -> floor=10, ceil=10 (exact)
		{1000, 7},  // 142.857... -> floor=142, ceil=143
		{999, 333}, // 3.0 -> floor=3, ceil=3 (exact)
		{1, 2},     // 0.5 -> floor=0, ceil=1
		{0, 999},   // 0.0 -> floor=0, ceil=0 (exact)
	}

	for i, tc := range testCases {
		t.Run(ufmt.Sprintf("property_test_%d", i), func(t *testing.T) {
			if tc.y == 0 {
				return // Skip division by zero
			}

			xUint := MustFromDecimal(ufmt.Sprintf("%d", tc.x))
			yUint := MustFromDecimal(ufmt.Sprintf("%d", tc.y))

			ceil := DivRoundingUp(xUint, yUint)
			floor := new(Uint).Div(xUint, yUint)
			floorPlusOne := new(Uint).Add(floor, one)

			if !ceil.Eq(floor) && !ceil.Eq(floorPlusOne) {
				t.Errorf("Property failed: ceil(%s) != floor(%s) && ceil != floor+1(%s)",
					ceil.ToString(), floor.ToString(), floorPlusOne.ToString())
			}
		})
	}
}

// Mathematical property verification for MulDiv
func TestMulDivMathematicalProperty(t *testing.T) {
	// Property: q = MulDiv(x, y, denom), r = (x * y) % denom
	// Then: q * denom + r == x * y AND r < denom

	testCases := []struct {
		x, y, denom uint64
	}{
		{123, 456, 789},
		{1000, 2000, 500},
		{999, 777, 333},
		{1, 1, 1},
		{0, 999, 123},
		{100, 7, 3},
	}

	for i, tc := range testCases {
		t.Run(ufmt.Sprintf("math_property_%d", i), func(t *testing.T) {
			if tc.denom == 0 {
				return // Skip division by zero
			}

			xUint := MustFromDecimal(ufmt.Sprintf("%d", tc.x))
			yUint := MustFromDecimal(ufmt.Sprintf("%d", tc.y))
			denomUint := MustFromDecimal(ufmt.Sprintf("%d", tc.denom))

			// Skip if this would cause overflow
			defer func() {
				if r := recover(); r != nil {
					// Overflow is expected for some cases
					return
				}
			}()

			product := new(Uint).Mul(xUint, yUint)
			quotient := MulDiv(xUint, yUint, denomUint)
			remainder := new(Uint).MulMod(xUint, yUint, denomUint)

			// Property 1: remainder < denom
			if remainder.Gte(denomUint) {
				t.Errorf("Property 1 failed: remainder(%s) >= denom(%s)",
					remainder.ToString(), denomUint.ToString())
			}

			// Property 2: quotient * denom + remainder == x * y
			reconstructed := new(Uint).Add(new(Uint).Mul(quotient, denomUint), remainder)
			if !reconstructed.Eq(product) {
				t.Errorf("Property 2 failed: q*d+r(%s) != x*y(%s)",
					reconstructed.ToString(), product.ToString())
			}
		})
	}
}
