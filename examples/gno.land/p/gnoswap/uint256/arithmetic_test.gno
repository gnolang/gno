package uint256

import (
	"testing"
)

// TestAdd verifies addition operations
func TestAdd(t *testing.T) {
	tests := []struct {
		name string
		x    string
		y    string
		want string
	}{
		// Basic addition
		{
			name: "zero_plus_zero",
			x:    "0",
			y:    "0",
			want: "0",
		},
		{
			name: "zero_plus_one",
			x:    "0",
			y:    "1",
			want: "1",
		},
		{
			name: "one_plus_one",
			x:    "1",
			y:    "1",
			want: "2",
		},
		{
			name: "ten_plus_ten",
			x:    "10",
			y:    "10",
			want: "20",
		},

		// uint64 boundary
		{
			name: "uint64_max_plus_one",
			x:    "18446744073709551615",
			y:    "1",
			want: "18446744073709551616",
		},
		{
			name: "uint64_max_plus_uint64_max",
			x:    "18446744073709551615",
			y:    "18446744073709551615",
			want: "36893488147419103230",
		},

		// Carry propagation
		{
			name: "carry_through_single_word",
			x:    "0xFFFFFFFFFFFFFFFF",
			y:    "1",
			want: "0x10000000000000000",
		},
		{
			name: "carry_through_multiple_words",
			x:    "0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
			y:    "1",
			want: "0x100000000000000000000000000000000",
		},

		// Maximum value edge cases
		{
			name: "max_plus_zero",
			x:    MAX_UINT256,
			y:    "0",
			want: MAX_UINT256,
		},
		{
			name: "max_plus_one_wraps_to_zero",
			x:    MAX_UINT256,
			y:    "1",
			want: "0",
		},
		{
			name: "half_max_plus_half_max_plus_one",
			x:    "57896044618658097711785492504343953926634992332820282019728792003956564819967",
			y:    "57896044618658097711785492504343953926634992332820282019728792003956564819968",
			want: MAX_UINT256,
		},

		// Additional boundary cases
		{
			name: "max_minus_one_plus_one",
			x:    "115792089237316195423570985008687907853269984665640564039457584007913129639934",
			y:    "1",
			want: MAX_UINT256,
		},
		{
			name: "max_minus_one_plus_two",
			x:    "115792089237316195423570985008687907853269984665640564039457584007913129639934",
			y:    "2",
			want: "0",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			x := parseUintT(t, tc.x)
			y := parseUintT(t, tc.y)
			want := parseUintT(t, tc.want)

			got := new(Uint).Add(x, y)
			if !got.Eq(want) {
				t.Errorf("Add(%s, %s) = %s, want %s", tc.x, tc.y, got.ToString(), tc.want)
			}
		})
	}
}

// TestAddOverflow verifies addition with overflow detection
func TestAddOverflow(t *testing.T) {
	tests := []struct {
		name     string
		x        string
		y        string
		want     string
		overflow bool
	}{
		{
			name:     "no_overflow_small_numbers",
			x:        "100",
			y:        "200",
			want:     "300",
			overflow: false,
		},
		{
			name:     "no_overflow_at_boundary",
			x:        MAX_UINT256,
			y:        "0",
			want:     MAX_UINT256,
			overflow: false,
		},
		{
			name:     "overflow_max_plus_one",
			x:        MAX_UINT256,
			y:        "1",
			want:     "0",
			overflow: true,
		},
		{
			name:     "overflow_max_plus_max",
			x:        MAX_UINT256,
			y:        MAX_UINT256,
			want:     "115792089237316195423570985008687907853269984665640564039457584007913129639934",
			overflow: true,
		},
		{
			name:     "overflow_with_carry_propagation",
			x:        "0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
			y:        "1",
			want:     "0",
			overflow: true,
		},
		{
			name:     "no_overflow_half_max_times_two",
			x:        "57896044618658097711785492504343953926634992332820282019728792003956564819967",
			y:        "57896044618658097711785492504343953926634992332820282019728792003956564819967",
			want:     "115792089237316195423570985008687907853269984665640564039457584007913129639934",
			overflow: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			x := parseUint(tc.x)
			y := parseUint(tc.y)
			want := parseUint(tc.want)

			got, overflow := new(Uint).AddOverflow(x, y)
			if !got.Eq(want) {
				t.Errorf("AddOverflow(%s, %s) result = %s, want %s", tc.x, tc.y, got.ToString(), tc.want)
			}
			if overflow != tc.overflow {
				t.Errorf("AddOverflow(%s, %s) overflow = %v, want %v", tc.x, tc.y, overflow, tc.overflow)
			}
		})
	}
}

// TestSub verifies subtraction operations
func TestSub(t *testing.T) {
	tests := []struct {
		name string
		x    string
		y    string
		want string
	}{
		// Basic subtraction
		{
			name: "one_minus_zero",
			x:    "1",
			y:    "0",
			want: "1",
		},
		{
			name: "one_minus_one",
			x:    "1",
			y:    "1",
			want: "0",
		},
		{
			name: "thousand_minus_hundred",
			x:    "31337",
			y:    "1337",
			want: "30000",
		},

		// Underflow cases (wraps around)
		{
			name: "zero_minus_one_wraps_to_max",
			x:    "0",
			y:    "1",
			want: MAX_UINT256,
		},
		{
			name: "small_minus_large",
			x:    "2",
			y:    "3",
			want: "115792089237316195423570985008687907853269984665640564039457584007913129639935",
		},

		// Maximum value cases
		{
			name: "max_minus_zero",
			x:    MAX_UINT256,
			y:    "0",
			want: MAX_UINT256,
		},
		{
			name: "max_minus_max",
			x:    MAX_UINT256,
			y:    MAX_UINT256,
			want: "0",
		},

		// Borrow propagation
		{
			name: "borrow_propagation_single_word",
			x:    "0x10000000000000000",
			y:    "1",
			want: "0xFFFFFFFFFFFFFFFF",
		},
		{
			name: "borrow_propagation_multiple_words",
			x:    "0x100000000000000000000000000000000",
			y:    "1",
			want: "0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
		},
		{
			name: "borrow_propagation_all_words",
			x:    "0", // Start with 0 and add 1, then subtract 1 to test borrow
			y:    "1",
			want: MAX_UINT256,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			x := parseUint(tc.x)
			y := parseUint(tc.y)
			want := parseUint(tc.want)

			got := new(Uint).Sub(x, y)
			if !got.Eq(want) {
				t.Errorf("Sub(%s, %s) = %s, want %s", tc.x, tc.y, got.ToString(), tc.want)
			}
		})
	}
}

// TestSubOverflow verifies subtraction with underflow detection
func TestSubOverflow(t *testing.T) {
	tests := []struct {
		name      string
		x         string
		y         string
		want      string
		underflow bool
	}{
		{
			name:      "no_underflow_normal_case",
			x:         "1000",
			y:         "100",
			want:      "900",
			underflow: false,
		},
		{
			name:      "no_underflow_equal_values",
			x:         "12345",
			y:         "12345",
			want:      "0",
			underflow: false,
		},
		{
			name:      "underflow_zero_minus_one",
			x:         "0",
			y:         "1",
			want:      MAX_UINT256,
			underflow: true,
		},
		{
			name:      "underflow_small_minus_large",
			x:         "100",
			y:         "1000",
			want:      "115792089237316195423570985008687907853269984665640564039457584007913129639036",
			underflow: true,
		},
		{
			name:      "underflow_one_minus_max",
			x:         "1",
			y:         MAX_UINT256,
			want:      "2",
			underflow: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			x := parseUint(tc.x)
			y := parseUint(tc.y)
			want := parseUint(tc.want)

			got, underflow := new(Uint).SubOverflow(x, y)
			if !got.Eq(want) {
				t.Errorf("SubOverflow(%s, %s) result = %s, want %s", tc.x, tc.y, got.ToString(), tc.want)
			}
			if underflow != tc.underflow {
				t.Errorf("SubOverflow(%s, %s) underflow = %v, want %v", tc.x, tc.y, underflow, tc.underflow)
			}
		})
	}
}

// TestMul verifies multiplication operations
func TestMul(t *testing.T) {
	tests := []struct {
		name string
		x    string
		y    string
		want string
	}{
		// Basic multiplication
		{
			name: "zero_times_zero",
			x:    "0",
			y:    "0",
			want: "0",
		},
		{
			name: "one_times_one",
			x:    "1",
			y:    "1",
			want: "1",
		},
		{
			name: "ten_times_ten",
			x:    "10",
			y:    "10",
			want: "100",
		},

		// uint64 overflow but not uint256
		{
			name: "uint64_max_times_two",
			x:    "18446744073709551615",
			y:    "2",
			want: "36893488147419103230",
		},
		{
			name: "large_numbers_no_overflow",
			x:    "0xFFFFFFFFFFFFFFFF",
			y:    "0xFFFFFFFFFFFFFFFF",
			want: "0xFFFFFFFFFFFFFFFE0000000000000001",
		},

		// Maximum value cases
		{
			name: "max_times_zero",
			x:    MAX_UINT256,
			y:    "0",
			want: "0",
		},
		{
			name: "max_times_one",
			x:    MAX_UINT256,
			y:    "1",
			want: MAX_UINT256,
		},

		// Q128 related (sqrt of max uint256)
		{
			name: "q128_minus_one_squared",
			x:    "340282366920938463463374607431768211455",
			y:    "340282366920938463463374607431768211455",
			want: "115792089237316195423570985008687907852589419931798687112530834793049593217025",
		},

		// Q96 related (Uniswap V3 price)
		{
			name: "q96_squared",
			x:    "79228162514264337593543950336",
			y:    "79228162514264337593543950336",
			want: "6277101735386680763835789423207666416102355444464034512896",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			x := parseUint(tc.x)
			y := parseUint(tc.y)
			want := parseUint(tc.want)

			got := new(Uint).Mul(x, y)
			if !got.Eq(want) {
				t.Errorf("Mul(%s, %s) = %s, want %s", tc.x, tc.y, got.ToString(), tc.want)
			}
		})
	}
}

// TestMulOverflow verifies multiplication with overflow detection
func TestMulOverflow(t *testing.T) {
	tests := []struct {
		name     string
		x        string
		y        string
		want     string
		overflow bool
	}{
		{
			name:     "no_overflow_small_numbers",
			x:        "1000",
			y:        "2000",
			want:     "2000000",
			overflow: false,
		},
		{
			name:     "no_overflow_at_boundary",
			x:        "340282366920938463463374607431768211455",
			y:        "340282366920938463463374607431768211455",
			want:     "115792089237316195423570985008687907852589419931798687112530834793049593217025",
			overflow: false,
		},
		{
			name:     "overflow_q128_squared",
			x:        "340282366920938463463374607431768211456",
			y:        "340282366920938463463374607431768211456",
			want:     "0",
			overflow: true,
		},
		{
			name:     "overflow_max_times_two",
			x:        MAX_UINT256,
			y:        "2",
			want:     "115792089237316195423570985008687907853269984665640564039457584007913129639934",
			overflow: true,
		},
		{
			name:     "overflow_max_times_max",
			x:        MAX_UINT256,
			y:        MAX_UINT256,
			want:     "1",
			overflow: true,
		},
		{
			name:     "no_overflow_large_numbers",
			x:        "340282366920938463463374607431768211455", // 2^128 - 1
			y:        "340282366920938463463374607431768211456", // 2^128
			want:     "115792089237316195423570985008687907852929702298719625575994209400481361428480",
			overflow: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			x := parseUint(tc.x)
			y := parseUint(tc.y)
			want := parseUint(tc.want)

			got, overflow := new(Uint).MulOverflow(x, y)
			if !got.Eq(want) {
				t.Errorf("MulOverflow(%s, %s) result = %s, want %s", tc.x, tc.y, got.ToString(), tc.want)
			}
			if overflow != tc.overflow {
				t.Errorf("MulOverflow(%s, %s) overflow = %v, want %v", tc.x, tc.y, overflow, tc.overflow)
			}
		})
	}
}

// TestDiv verifies division operations
func TestDiv(t *testing.T) {
	tests := []struct {
		name    string
		x       string
		y       string
		want    string
		wantErr bool
	}{
		// Basic division
		{
			name:    "ten_div_two",
			x:       "10",
			y:       "2",
			want:    "5",
			wantErr: false,
		},
		{
			name:    "exact_division",
			x:       "31337",
			y:       "3",
			want:    "10445",
			wantErr: false,
		},
		{
			name:    "division_with_remainder",
			x:       "10",
			y:       "3",
			want:    "3",
			wantErr: false,
		},

		// Special cases
		{
			name:    "zero_div_nonzero",
			x:       "0",
			y:       "31337",
			want:    "0",
			wantErr: false,
		},
		{
			name:    "smaller_div_larger",
			x:       "2",
			y:       "31337",
			want:    "0",
			wantErr: false,
		},
		{
			name:    "equal_div_equal",
			x:       "31337",
			y:       "31337",
			want:    "1",
			wantErr: false,
		},

		// Division by zero
		{
			name:    "div_by_zero",
			x:       "31337",
			y:       "0",
			want:    "",
			wantErr: true,
		},
		{
			name:    "zero_div_zero",
			x:       "0",
			y:       "0",
			want:    "",
			wantErr: true,
		},

		// Large numbers
		{
			name:    "max_div_one",
			x:       MAX_UINT256,
			y:       "1",
			want:    MAX_UINT256,
			wantErr: false,
		},
		{
			name:    "max_div_two",
			x:       MAX_UINT256,
			y:       "2",
			want:    "57896044618658097711785492504343953926634992332820282019728792003956564819967",
			wantErr: false,
		},
		{
			name:    "max_div_max",
			x:       MAX_UINT256,
			y:       MAX_UINT256,
			want:    "1",
			wantErr: false,
		},
		{
			name:    "large_div_large",
			x:       "0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
			y:       "0xFFFFFFFFFFFFFFFF",
			want:    "0x10000000000000001",
			wantErr: false,
		},

		// Verify truncation (rounds down)
		{
			name:    "verify_truncation",
			x:       "1000000000000000001",
			y:       "1000000000000000000",
			want:    "1",
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			x := parseUint(tc.x)
			y := parseUint(tc.y)

			got := new(Uint).Div(x, y)

			if tc.wantErr {
				// Division by zero returns 0
				if !got.IsZero() {
					t.Errorf("Div(%s, %s) expected zero for division by zero", tc.x, tc.y)
				}
				return
			}

			want := parseUint(tc.want)
			if !got.Eq(want) {
				t.Errorf("Div(%s, %s) = %s, want %s", tc.x, tc.y, got.ToString(), tc.want)
			}
		})
	}
}

// TestMod verifies modulo operations
func TestMod(t *testing.T) {
	tests := []struct {
		name    string
		x       string
		y       string
		want    string
		wantErr bool
	}{
		// Basic modulo
		{
			name:    "ten_mod_three",
			x:       "10",
			y:       "3",
			want:    "1",
			wantErr: false,
		},
		{
			name:    "exact_division_mod",
			x:       "100",
			y:       "10",
			want:    "0",
			wantErr: false,
		},
		{
			name:    "large_mod_calculation",
			x:       "31337",
			y:       "3",
			want:    "2",
			wantErr: false,
		},

		// Special cases
		{
			name:    "zero_mod_nonzero",
			x:       "0",
			y:       "31337",
			want:    "0",
			wantErr: false,
		},
		{
			name:    "smaller_mod_larger",
			x:       "2",
			y:       "31337",
			want:    "2",
			wantErr: false,
		},
		{
			name:    "equal_mod_equal",
			x:       "31337",
			y:       "31337",
			want:    "0",
			wantErr: false,
		},

		// Modulo by zero
		{
			name:    "mod_by_zero",
			x:       "31337",
			y:       "0",
			want:    "",
			wantErr: true,
		},
		{
			name:    "zero_mod_zero",
			x:       "0",
			y:       "0",
			want:    "",
			wantErr: true,
		},

		// Large numbers
		{
			name:    "max_mod_small",
			x:       MAX_UINT256,
			y:       "1000000",
			want:    "639935",
			wantErr: false,
		},
		{
			name:    "power_of_two_mod",
			x:       "0xFFFFFFFFFFFFFFFF",
			y:       "0x10000000000000000",
			want:    "0xFFFFFFFFFFFFFFFF",
			wantErr: false,
		},
		{
			name:    "max_mod_large_prime",
			x:       MAX_UINT256,
			y:       "115792089237316195423570985008687907853269984665640564039457584007913129639747", // Large prime
			want:    "188",
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			x := parseUint(tc.x)
			y := parseUint(tc.y)

			got := new(Uint).Mod(x, y)

			if tc.wantErr {
				// Modulo by zero returns 0
				if !got.IsZero() {
					t.Errorf("Mod(%s, %s) expected zero for modulo by zero", tc.x, tc.y)
				}
				return
			}

			want := parseUint(tc.want)
			if !got.Eq(want) {
				t.Errorf("Mod(%s, %s) = %s, want %s", tc.x, tc.y, got.ToString(), tc.want)
			}
		})
	}
}

// TestDivMod verifies combined division and modulo operations
func TestDivMod(t *testing.T) {
	tests := []struct {
		name    string
		x       string
		y       string
		wantDiv string
		wantMod string
		wantErr bool
	}{
		// Basic cases
		{
			name:    "simple_divmod",
			x:       "10",
			y:       "3",
			wantDiv: "3",
			wantMod: "1",
			wantErr: false,
		},
		{
			name:    "exact_divmod",
			x:       "100",
			y:       "10",
			wantDiv: "10",
			wantMod: "0",
			wantErr: false,
		},
		{
			name:    "large_divmod",
			x:       "31337",
			y:       "3",
			wantDiv: "10445",
			wantMod: "2",
			wantErr: false,
		},

		// Special cases
		{
			name:    "zero_divmod_nonzero",
			x:       "0",
			y:       "31337",
			wantDiv: "0",
			wantMod: "0",
			wantErr: false,
		},
		{
			name:    "smaller_divmod_larger",
			x:       "2",
			y:       "31337",
			wantDiv: "0",
			wantMod: "2",
			wantErr: false,
		},
		{
			name:    "equal_divmod_equal",
			x:       "31337",
			y:       "31337",
			wantDiv: "1",
			wantMod: "0",
			wantErr: false,
		},

		// Division by zero
		{
			name:    "divmod_by_zero",
			x:       "31337",
			y:       "0",
			wantDiv: "",
			wantMod: "",
			wantErr: true,
		},

		// Large numbers
		{
			name:    "max_divmod_small",
			x:       MAX_UINT256,
			y:       "1000000",
			wantDiv: "115792089237316195423570985008687907853269984665640564039457584007913129",
			wantMod: "639935",
			wantErr: false,
		},

		// Power of two optimization cases
		{
			name:    "divmod_by_2",
			x:       "31337",
			y:       "2",
			wantDiv: "15668",
			wantMod: "1",
			wantErr: false,
		},
		{
			name:    "divmod_by_4",
			x:       "31337",
			y:       "4",
			wantDiv: "7834",
			wantMod: "1",
			wantErr: false,
		},
		{
			name:    "divmod_by_256",
			x:       "0xABCDEF1234567890",
			y:       "0x100",
			wantDiv: "0xABCDEF12345678",
			wantMod: "0x90",
			wantErr: false,
		},
		{
			name:    "divmod_by_2_64",
			x:       "0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
			y:       "0x10000000000000000", // 2^64
			wantDiv: "0xFFFFFFFFFFFFFFFF",
			wantMod: "0xFFFFFFFFFFFFFFFF",
			wantErr: false,
		},
		{
			name:    "max_divmod_power_of_two",
			x:       MAX_UINT256,
			y:       "0x10000000000000000", // 2^64
			wantDiv: "6277101735386680763835789423207666416102355444464034512895",
			wantMod: "0xFFFFFFFFFFFFFFFF",
			wantErr: false,
		},
		{
			name:    "divmod_by_2_128",
			x:       MAX_UINT256,
			y:       "340282366920938463463374607431768211456", // 2^128
			wantDiv: "340282366920938463463374607431768211455",
			wantMod: "340282366920938463463374607431768211455",
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			x := parseUint(tc.x)
			y := parseUint(tc.y)

			gotDiv := new(Uint)
			gotMod := new(Uint)
			gotDiv, gotMod = gotDiv.DivMod(x, y, gotMod)

			if tc.wantErr {
				// Division by zero returns (0, 0)
				if !gotDiv.IsZero() || !gotMod.IsZero() {
					t.Errorf("DivMod(%s, %s) expected (0, 0) for division by zero", tc.x, tc.y)
				}
				return
			}

			wantDiv := parseUint(tc.wantDiv)
			wantMod := parseUint(tc.wantMod)

			if !gotDiv.Eq(wantDiv) {
				t.Errorf("DivMod(%s, %s) div = %s, want %s", tc.x, tc.y, gotDiv.ToString(), tc.wantDiv)
			}
			if !gotMod.Eq(wantMod) {
				t.Errorf("DivMod(%s, %s) mod = %s, want %s", tc.x, tc.y, gotMod.ToString(), tc.wantMod)
			}
		})
	}
}

// TestMulMod verifies modular multiplication operations
func TestMulMod(t *testing.T) {
	tests := []struct {
		name    string
		x       string
		y       string
		m       string
		want    string
		wantErr bool
	}{
		// Basic cases
		{
			name:    "simple_mulmod",
			x:       "10",
			y:       "20",
			m:       "7",
			want:    "4", // (10 * 20) % 7 = 200 % 7 = 4
			wantErr: false,
		},
		{
			name:    "exact_mulmod",
			x:       "10",
			y:       "20",
			m:       "200",
			want:    "0",
			wantErr: false,
		},

		// Edge cases
		{
			name:    "zero_times_any_mod",
			x:       "0",
			y:       "12345",
			m:       "100",
			want:    "0",
			wantErr: false,
		},
		{
			name:    "one_times_any_mod",
			x:       "1",
			y:       "12345",
			m:       "100",
			want:    "45",
			wantErr: false,
		},
		{
			name:    "mulmod_by_zero",
			x:       "10",
			y:       "20",
			m:       "0",
			want:    "",
			wantErr: true,
		},

		// Fast path: small product
		{
			name:    "fast_path_small_product",
			x:       "1000000",
			y:       "2000000",
			m:       "999999",
			want:    "2", // (1000000 * 2000000) % 999999 = 2
			wantErr: false,
		},
		{
			name:    "fast_path_exact_256_bits",
			x:       "0x1FFFFFFFFFFFFFFFF", // Just over 2^64
			y:       "2",
			m:       "0xFFFFFFFFFFFFFFFF", // 2^64 - 1
			want:    "2",
			wantErr: false,
		},

		// Large numbers that would overflow in regular multiplication
		{
			name:    "large_mulmod_no_overflow",
			x:       "340282366920938463463374607431768211456", // 2^128
			y:       "340282366920938463463374607431768211456",
			m:       MAX_UINT256,
			want:    "1",
			wantErr: false,
		},

		// Reduce4 path: large modulus
		{
			name:    "reduce4_path_large_modulus",
			x:       MAX_UINT256,
			y:       MAX_UINT256,
			m:       "57896044618658097711785492504343953926634992332820282019728792003956564819968", // 2^255
			want:    "1",
			wantErr: false,
		},
		{
			name:    "reduce4_path_exact_division",
			x:       "340282366920938463463374607431768211456",
			y:       "340282366920938463463374607431768211456",
			m:       "115792089237316195423570985008687907852589419931798687112530834793049593217024",
			want:    "680564733841876926926749214863536422912",
			wantErr: false,
		},

		// Q96 price calculation (corrected)
		{
			name:    "q96_price_calculation",
			x:       "79228162514264337593543950336", // 2^96
			y:       "1000000000000000000",           // 1e18
			m:       "79228162514264337593543950335", // 2^96 - 1
			want:    "1000000000000000000",           // Corrected from 1000000000000000001
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			x := parseUint(tc.x)
			y := parseUint(tc.y)
			m := parseUint(tc.m)

			got := new(Uint).MulMod(x, y, m)

			if tc.wantErr {
				// MulMod by zero returns 0
				if !got.IsZero() {
					t.Errorf("MulMod(%s, %s, %s) expected zero for modulo by zero", tc.x, tc.y, tc.m)
				}
				return
			}

			want := parseUint(tc.want)
			if !got.Eq(want) {
				t.Errorf("MulMod(%s, %s, %s) = %s, want %s", tc.x, tc.y, tc.m, got.ToString(), tc.want)
			}
		})
	}
}

// TestNeg verifies negation operations
func TestNeg(t *testing.T) {
	tests := []struct {
		name string
		x    string
		want string
	}{
		{
			name: "neg_zero",
			x:    "0",
			want: "0",
		},
		{
			name: "neg_one",
			x:    "1",
			want: "115792089237316195423570985008687907853269984665640564039457584007913129639935",
		},
		{
			name: "neg_two",
			x:    "2",
			want: "115792089237316195423570985008687907853269984665640564039457584007913129639934",
		},
		{
			name: "neg_large",
			x:    "31337",
			want: "115792089237316195423570985008687907853269984665640564039457584007913129608599",
		},
		{
			name: "neg_neg_large",
			x:    "115792089237316195423570985008687907853269984665640564039457584007913129608599",
			want: "31337",
		},
		{
			name: "neg_max",
			x:    MAX_UINT256,
			want: "1",
		},
		{
			name: "neg_half",
			x:    "57896044618658097711785492504343953926634992332820282019728792003956564819968", // 2^255
			want: "57896044618658097711785492504343953926634992332820282019728792003956564819968",
		},
		{
			name: "neg_max_minus_one",
			x:    "115792089237316195423570985008687907853269984665640564039457584007913129639934",
			want: "2",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			x := parseUint(tc.x)
			want := parseUint(tc.want)

			got := new(Uint).Neg(x)
			if !got.Eq(want) {
				t.Errorf("Neg(%s) = %s, want %s", tc.x, got.ToString(), tc.want)
			}
		})
	}
}

// TestExp verifies exponentiation operations
func TestExp(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		exponent string
		want     string
	}{
		// Basic cases
		{
			name:     "zero_power_zero",
			base:     "0",
			exponent: "0",
			want:     "1", // Mathematical convention
		},
		{
			name:     "any_power_zero",
			base:     "31337",
			exponent: "0",
			want:     "1",
		},
		{
			name:     "zero_power_any",
			base:     "0",
			exponent: "31337",
			want:     "0",
		},
		{
			name:     "one_power_any",
			base:     "1",
			exponent: "31337",
			want:     "1",
		},
		{
			name:     "two_power_three",
			base:     "2",
			exponent: "3",
			want:     "8",
		},

		// Powers of 2
		{
			name:     "two_power_64",
			base:     "2",
			exponent: "64",
			want:     "18446744073709551616",
		},
		{
			name:     "two_power_96",
			base:     "2",
			exponent: "96",
			want:     "79228162514264337593543950336",
		},
		{
			name:     "two_power_128",
			base:     "2",
			exponent: "128",
			want:     "340282366920938463463374607431768211456",
		},
		{
			name:     "two_power_255",
			base:     "2",
			exponent: "255",
			want:     "57896044618658097711785492504343953926634992332820282019728792003956564819968",
		},
		{
			name:     "two_power_256_wraps",
			base:     "2",
			exponent: "256",
			want:     "0",
		},

		// Non-power-of-two cases
		{
			name:     "three_power_small",
			base:     "3",
			exponent: "5",
			want:     "243",
		},
		{
			name:     "three_power_large",
			base:     "3",
			exponent: "100",
			want:     "515377520732011331036461129765621272702107522001",
		},
		{
			name:     "five_power_fifty",
			base:     "5",
			exponent: "50",
			want:     "88817841970012523233890533447265625",
		},
		{
			name:     "seven_power_thirty",
			base:     "7",
			exponent: "30",
			want:     "22539340290692258087863249",
		},

		// Other cases
		{
			name:     "large_base_small_exp",
			base:     "31337",
			exponent: "3",
			want:     "30773171189753",
		},
		{
			name:     "large_base_large_exp",
			base:     "12345678901234567890",
			exponent: "3",
			want:     "1881676372353657772490265749424677022198701224860897069000",
		},
		{
			name:     "max_base_exp_one",
			base:     MAX_UINT256,
			exponent: "1",
			want:     MAX_UINT256,
		},
		{
			name:     "max_base_exp_two",
			base:     MAX_UINT256,
			exponent: "2",
			want:     "1", // Wraps around
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			base := parseUint(tc.base)
			exponent := parseUint(tc.exponent)
			want := parseUint(tc.want)

			got := new(Uint).Exp(base, exponent)
			if !got.Eq(want) {
				t.Errorf("Exp(%s, %s) = %s, want %s", tc.base, tc.exponent, got.ToString(), tc.want)
			}
		})
	}
}

// TestMustMul verifies MustMul panic behavior
func TestMustMul(t *testing.T) {
	t.Run("normal_multiplication", func(t *testing.T) {
		x := NewUint(1000)
		y := NewUint(2000)
		result := new(Uint).MustMul(x, y)
		if result.ToString() != "2000000" {
			t.Errorf("MustMul(1000, 2000) = %s, want 2000000", result.ToString())
		}
	})

	t.Run("large_valid_multiplication", func(t *testing.T) {
		x := parseUint("340282366920938463463374607431768211455") // 2^128 - 1
		y := NewUint(2)
		result := new(Uint).MustMul(x, y)
		if result.ToString() != "680564733841876926926749214863536422910" {
			t.Errorf("MustMul large valid failed")
		}
	})

	t.Run("overflow_panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("MustMul should panic on overflow")
			}
		}()

		max := MustFromDecimal(MAX_UINT256)
		new(Uint).MustMul(max, NewUint(2))
	})
}

// TestIsOverflow verifies overflow bit detection
func TestIsOverflow(t *testing.T) {
	tests := []struct {
		name     string
		input    *Uint
		expected bool
	}{
		{
			name:     "zero_not_overflow",
			input:    &Uint{arr: [4]uint64{0, 0, 0, 0}},
			expected: false,
		},
		{
			name:     "max_value_not_overflow",
			input:    &Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0) >> 1}},
			expected: false,
		},
		{
			name:     "bit_255_set_is_overflow",
			input:    &Uint{arr: [4]uint64{0, 0, 0, uint64(1) << 63}},
			expected: true,
		},
		{
			name:     "all_bits_set_is_overflow",
			input:    &Uint{arr: [4]uint64{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}},
			expected: true,
		},
		{
			name:     "max_plus_one_is_overflow",
			input:    &Uint{arr: [4]uint64{0, 0, 0, uint64(1) << 63}},
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.input.IsOverflow(); got != tc.expected {
				t.Errorf("IsOverflow() = %v, want %v", got, tc.expected)
			}
		})
	}
}

// TestDivisionByZero verifies all division by zero cases return error
func TestDivisionByZero(t *testing.T) {
	testCases := []struct {
		name string
		fn   func() *Uint
	}{
		{
			name: "div_by_zero",
			fn: func() *Uint {
				x := MustFromDecimal("12345")
				y := MustFromDecimal("0")
				return new(Uint).Div(x, y)
			},
		},
		{
			name: "mod_by_zero",
			fn: func() *Uint {
				x := MustFromDecimal("12345")
				y := MustFromDecimal("0")
				return new(Uint).Mod(x, y)
			},
		},
		{
			name: "mulmod_by_zero",
			fn: func() *Uint {
				x := MustFromDecimal("12345")
				y := MustFromDecimal("67890")
				m := MustFromDecimal("0")
				return new(Uint).MulMod(x, y, m)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.fn()
			if !result.IsZero() {
				t.Errorf("%s should return zero but returned %s", tc.name, result.ToString())
			}
		})
	}
}

// TestMustOperations verifies that Must* functions panic appropriately
func TestMustOperations(t *testing.T) {
	t.Run("MustDiv_panics_on_zero", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustDiv should panic on division by zero")
			}
		}()
		x := NewUint(100)
		y := Zero()
		new(Uint).MustDiv(x, y)
	})

	t.Run("MustMod_panics_on_zero", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustMod should panic on modulo by zero")
			}
		}()
		x := NewUint(100)
		y := Zero()
		new(Uint).MustMod(x, y)
	})

	t.Run("MustMulMod_panics_on_zero", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustMulMod should panic on modulo by zero")
			}
		}()
		x := NewUint(10)
		y := NewUint(20)
		m := Zero()
		new(Uint).MustMulMod(x, y, m)
	})

	t.Run("MustDivMod_panics_on_zero", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustDivMod should panic on division by zero")
			}
		}()
		x := NewUint(100)
		y := Zero()
		new(Uint).MustDivMod(x, y, new(Uint))
	})

	t.Run("MustMul_panics_on_overflow", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustMul should panic on overflow")
			}
		}()
		x := MustFromDecimal("340282366920938463463374607431768211456") // 2^128
		y := MustFromDecimal("340282366920938463463374607431768211456")
		new(Uint).MustMul(x, y)
	})

	// Test successful Must operations
	t.Run("MustDiv_success", func(t *testing.T) {
		x := NewUint(100)
		y := NewUint(10)
		result := new(Uint).MustDiv(x, y)
		if !result.Eq(NewUint(10)) {
			t.Error("MustDiv failed")
		}
	})

	t.Run("MustMul_success", func(t *testing.T) {
		x := NewUint(100)
		y := NewUint(200)
		result := new(Uint).MustMul(x, y)
		if !result.Eq(NewUint(20000)) {
			t.Error("MustMul failed")
		}
	})
}

// TestMulModPaths verifies both fast and reduce4 paths in MulMod
func TestMulModPaths(t *testing.T) {
	tests := []struct {
		name string
		x    string
		y    string
		m    string
		want string
		path string // "fast" or "reduce4"
	}{
		// Fast path: x*y fits in 256 bits, m has no high word
		{
			name: "fast_path_small_product",
			x:    "1000000",
			y:    "2000000",
			m:    "999999",
			want: "2",
			path: "fast",
		},
		{
			name: "fast_path_exact_256_bits",
			x:    "0x1FFFFFFFFFFFFFFFF", // Just over 2^64
			y:    "2",
			m:    "0xFFFFFFFFFFFFFFFF", // 2^64 - 1
			want: "2",
			path: "fast",
		},
		{
			name: "boundary_case_m_arr3_zero",
			x:    "340282366920938463463374607431768211456", // Q128
			y:    "340282366920938463463374607431768211456",
			m:    "340282366920938463463374607431768211455", // m.arr[3] != 0
			want: "1",
			path: "fast", // Takes fast path because m.arr[3] == 0
		},

		// Reduce4 path: x*y > 2^256 and m.arr[3] != 0
		{
			name: "reduce4_path_max_inputs",
			x:    MAX_UINT256,
			y:    MAX_UINT256,
			m:    "57896044618658097711785492504343953926634992332820282019728792003956564819968", // 2^255
			want: "1",
			path: "reduce4",
		},
		{
			name: "reduce4_path_exact_division",
			x:    "340282366920938463463374607431768211456",
			y:    "340282366920938463463374607431768211456",
			m:    "115792089237316195423570985008687907852589419931798687112530834793049593217024",
			want: "680564733841876926926749214863536422912",
			path: "reduce4",
		},
		{
			name: "reduce4_path_q96_calculation",
			x:    "79228162514264337593543950336", // 2^96
			y:    "79228162514264337593543950336",
			m:    "6277101735386680763835789423207666416102355444464034512895", // Just under result
			want: "1",
			path: "reduce4",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			x := parseUint(tc.x)
			y := parseUint(tc.y)
			m := parseUint(tc.m)
			want := parseUint(tc.want)

			got := new(Uint).MulMod(x, y, m)

			if !got.Eq(want) {
				t.Errorf("MulMod(%s, %s, %s) = %s, want %s",
					tc.x, tc.y, tc.m, got.ToString(), tc.want)
			}
		})
	}
}

// TestDivModPowerOfTwo verifies optimized division by powers of two
func TestDivModPowerOfTwo(t *testing.T) {
	tests := []struct {
		name    string
		x       string
		y       string
		wantDiv string
		wantMod string
	}{
		{
			name:    "divmod_by_2",
			x:       "31337",
			y:       "2",
			wantDiv: "15668",
			wantMod: "1",
		},
		{
			name:    "divmod_by_4",
			x:       "31337",
			y:       "4",
			wantDiv: "7834",
			wantMod: "1",
		},
		{
			name:    "divmod_by_8",
			x:       "31337",
			y:       "8",
			wantDiv: "3917",
			wantMod: "1",
		},
		{
			name:    "divmod_by_256",
			x:       "0xABCDEF1234567890",
			y:       "0x100",
			wantDiv: "0xABCDEF12345678",
			wantMod: "0x90",
		},
		{
			name:    "divmod_by_2_32",
			x:       "0xFFFFFFFFFFFFFFFF",
			y:       "0x100000000", // 2^32
			wantDiv: "0xFFFFFFFF",
			wantMod: "0xFFFFFFFF",
		},
		{
			name:    "divmod_by_2_64",
			x:       "0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
			y:       "0x10000000000000000", // 2^64
			wantDiv: "0xFFFFFFFFFFFFFFFF",
			wantMod: "0xFFFFFFFFFFFFFFFF",
		},
		{
			name:    "max_divmod_power_of_two",
			x:       MAX_UINT256,
			y:       "0x10000000000000000", // 2^64
			wantDiv: "6277101735386680763835789423207666416102355444464034512895",
			wantMod: "0xFFFFFFFFFFFFFFFF",
		},
		{
			name:    "divmod_by_2_96",
			x:       MAX_UINT256,
			y:       "79228162514264337593543950336", // 2^96
			wantDiv: "1461501637330902918203684832716283019655932542975",
			wantMod: "79228162514264337593543950335",
		},
		{
			name:    "divmod_by_2_128",
			x:       MAX_UINT256,
			y:       "340282366920938463463374607431768211456", // 2^128
			wantDiv: "340282366920938463463374607431768211455",
			wantMod: "340282366920938463463374607431768211455",
		},
		{
			name:    "divmod_by_2_192",
			x:       MAX_UINT256,
			y:       "6277101735386680763835789423207666416102355444464034512896", // 2^192
			wantDiv: "18446744073709551615",
			wantMod: "6277101735386680763835789423207666416102355444464034512895",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			x := parseUint(tc.x)
			y := parseUint(tc.y)
			wantDiv := parseUint(tc.wantDiv)
			wantMod := parseUint(tc.wantMod)

			gotDiv := new(Uint)
			gotMod := new(Uint)
			gotDiv, gotMod = gotDiv.DivMod(x, y, gotMod)

			if !gotDiv.Eq(wantDiv) {
				t.Errorf("DivMod div = %s, want %s", gotDiv.ToString(), tc.wantDiv)
			}
			if !gotMod.Eq(wantMod) {
				t.Errorf("DivMod mod = %s, want %s", gotMod.ToString(), tc.wantMod)
			}
		})
	}
}

// TestChainedOperations verifies DEX-like operation chains
func TestChainedOperations(t *testing.T) {
	tests := []struct {
		name string
		ops  []struct {
			op  string
			arg string
		}
		start string
		want  string
	}{
		{
			name:  "swap_fee_calculation",
			start: "1000000000000000000", // 1 token
			ops: []struct {
				op  string
				arg string
			}{
				{op: "mul", arg: "997"},  // 0.3% fee
				{op: "div", arg: "1000"}, // fee calculation
			},
			want: "997000000000000000", // 0.997 tokens after fee
		},
		{
			name:  "liquidity_calculation",
			start: "1000000000000000000000", // 1000 tokens
			ops: []struct {
				op  string
				arg string
			}{
				{op: "mul", arg: "79228162514264337593543950336"}, // Multiply by Q96
				{op: "div", arg: "1000000000000000000"},           // Normalize
			},
			want: "79228162514264337593543950336000", // Result in Q96 format
		},
		{
			name:  "complex_price_impact",
			start: "1000000000000000000", // 1 token
			ops: []struct {
				op  string
				arg string
			}{
				{op: "mul", arg: "997"}, // Fee
				{op: "div", arg: "1000"},
				{op: "mul", arg: "79228162514264337593543950336"}, // Price in Q96
				{op: "div", arg: "79228162514264337593543950336"}, // Normalize back
			},
			want: "997000000000000000", // Same as fee calculation
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := parseUint(tc.start)

			for _, op := range tc.ops {
				arg := parseUint(op.arg)
				switch op.op {
				case "add":
					result = new(Uint).Add(result, arg)
				case "sub":
					result = new(Uint).Sub(result, arg)
				case "mul":
					result = new(Uint).Mul(result, arg)
				case "div":
					result = new(Uint).Div(result, arg)
				}
			}

			want := parseUint(tc.want)
			if !result.Eq(want) {
				t.Errorf("Chained operations = %s, want %s", result.ToString(), tc.want)
			}
		})
	}
}

// Division rounding consistency test
func TestDivisionRoundingConsistency(t *testing.T) {
	tests := []struct {
		name    string
		x       string
		y       string
		wantDiv string
		wantMod string
	}{
		{
			name:    "rounds_down_small",
			x:       "10",
			y:       "3",
			wantDiv: "3",
			wantMod: "1",
		},
		{
			name:    "rounds_down_large",
			x:       "999999999999999999",
			y:       "1000000000000000000",
			wantDiv: "0",
			wantMod: "999999999999999999",
		},
		{
			name:    "exact_division",
			x:       "1000000000000000000",
			y:       "1000000000000000",
			wantDiv: "1000",
			wantMod: "0",
		},
		{
			name:    "almost_two",
			x:       "1999999999999999999",
			y:       "1000000000000000000",
			wantDiv: "1",
			wantMod: "999999999999999999",
		},
		{
			name:    "large_division_rounds_down",
			x:       MAX_UINT256,
			y:       "1000000000000000001",
			wantDiv: "115792089237316195307778895771371712545491088894268851493966",
			wantMod: "495113644278145969",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x := parseUint(tt.x)
			y := parseUint(tt.y)

			gotDiv := new(Uint).Div(x, y)
			gotMod := new(Uint).Mod(x, y)

			wantDiv := parseUint(tt.wantDiv)
			wantMod := parseUint(tt.wantMod)

			if !gotDiv.Eq(wantDiv) {
				t.Errorf("Div(%s, %s) = %s, want %s", tt.x, tt.y, gotDiv.ToString(), tt.wantDiv)
			}
			if !gotMod.Eq(wantMod) {
				t.Errorf("Mod(%s, %s) = %s, want %s", tt.x, tt.y, gotMod.ToString(), tt.wantMod)
			}

			// Verify division always rounds down
			// x = y * div + mod, and 0 <= mod < y
			reconstructed := new(Uint).Mul(y, gotDiv)
			reconstructed = new(Uint).Add(reconstructed, gotMod)
			if !reconstructed.Eq(x) {
				t.Errorf("Division property violated: y*div+mod != x")
			}
			if gotMod.Gte(y) {
				t.Errorf("Modulo >= divisor: mod=%s, y=%s", gotMod.ToString(), tt.y)
			}
		})
	}
}

// Test extreme values combinations
func TestExtremeValues(t *testing.T) {
	extremeValues := []struct {
		name  string
		value string
	}{
		{"one", "1"},
		{"two", "2"},
		{"max_minus_two", "115792089237316195423570985008687907853269984665640564039457584007913129639933"},
		{"max_minus_one", "115792089237316195423570985008687907853269984665640564039457584007913129639934"},
		{"max", MAX_UINT256},
	}

	for _, x := range extremeValues {
		for _, y := range extremeValues {
			t.Run(x.name+"_with_"+y.name, func(t *testing.T) {
				xu := parseUint(x.value)
				yu := parseUint(y.value)

				// Test all operations don't panic unexpectedly
				_ = new(Uint).Add(xu, yu)
				_ = new(Uint).Sub(xu, yu)
				_ = new(Uint).Mul(xu, yu)

				if !yu.IsZero() {
					_ = new(Uint).Div(xu, yu)
					_ = new(Uint).Mod(xu, yu)

					// DivMod
					quotient := new(Uint)
					remainder := new(Uint)
					quotient, remainder = quotient.DivMod(xu, yu, remainder)

					// Verify DivMod consistency
					divResult := new(Uint).Div(xu, yu)
					modResult := new(Uint).Mod(xu, yu)
					if !quotient.Eq(divResult) || !remainder.Eq(modResult) {
						t.Errorf("DivMod inconsistent with Div/Mod")
					}
				}
			})
		}
	}
}

// Benchmark functions
func BenchmarkAdd(b *testing.B) {
	x := MustFromDecimal("123456789012345678901234567890123456789")
	y := MustFromDecimal("987654321098765432109876543210987654321")
	z := new(Uint)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		z.Add(x, y)
	}
}

func BenchmarkMul(b *testing.B) {
	x := MustFromDecimal("123456789012345678901234567890")
	y := MustFromDecimal("987654321098765432109876543210")
	z := new(Uint)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		z.Mul(x, y)
	}
}

func BenchmarkDiv(b *testing.B) {
	x := MustFromDecimal("123456789012345678901234567890123456789012345678901234567890")
	y := MustFromDecimal("123456789012345678901234567890")
	z := new(Uint)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		z.Div(x, y)
	}
}

func BenchmarkExp(b *testing.B) {
	base := NewUint(2)
	exp := NewUint(100)
	z := new(Uint)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		z.Exp(base, exp)
	}
}

func BenchmarkMulMod(b *testing.B) {
	x := MustFromDecimal("123456789012345678901234567890")
	y := MustFromDecimal("987654321098765432109876543210")
	m := MustFromDecimal("1000000000000000000000000000000")
	z := new(Uint)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		z.MulMod(x, y, m)
	}
}
