package int256

// func TestSetInt64(t *testing.T) {
// 	tests := []struct {
// 		x    int64
// 		want string
// 	}{
// 		{0, "0"},
// 		{1, "1"},
// 		{-1, "-1"},
// 		{9223372036854775807, "9223372036854775807"},
// 		{-9223372036854775808, "-9223372036854775808"},
// 	}

// 	for _, tc := range tests {
// 		var z Int
// 		z.SetInt64(tc.x)

// 		got := z.ToString()
// 		if got != tc.want {
// 			t.Errorf("SetInt64(%d) = %s, want %s", tc.x, got, tc.want)
// 		}
// 	}
// }

// func TestSetUint64(t *testing.T) {
// 	tests := []struct {
// 		x    uint64
// 		want string
// 	}{
// 		{0, "0"},
// 		{1, "1"},
// 	}

// 	for _, tc := range tests {
// 		var z Int
// 		z.SetUint64(tc.x)

// 		got := z.ToString()
// 		if got != tc.want {
// 			t.Errorf("SetUint64(%d) = %s, want %s", tc.x, got, tc.want)
// 		}
// 	}
// }

// func TestUint64(t *testing.T) {
// 	tests := []struct {
// 		x    string
// 		want uint64
// 	}{
// 		{"0", 0},
// 		{"1", 1},
// 		{"9223372036854775807", 9223372036854775807},
// 		{"9223372036854775808", 9223372036854775808},
// 		{"18446744073709551615", 18446744073709551615},
// 		{"18446744073709551616", 0},
// 		{"18446744073709551617", 1},
// 		{"-1", 1},
// 		{"-18446744073709551615", 18446744073709551615},
// 		{"-18446744073709551616", 0},
// 		{"-18446744073709551617", 1},
// 	}

// 	for _, tc := range tests {
// 		z := MustFromDecimal(tc.x)

// 		got := z.Uint64()
// 		if got != tc.want {
// 			t.Errorf("Uint64(%s) = %d, want %d", tc.x, got, tc.want)
// 		}
// 	}
// }

// func TestInt64(t *testing.T) {
// 	tests := []struct {
// 		x    string
// 		want int64
// 	}{
// 		{"0", 0},
// 		{"1", 1},
// 		{"9223372036854775807", 9223372036854775807},
// 		{"18446744073709551616", 0},
// 		{"18446744073709551617", 1},
// 		{"-1", -1},
// 		{"-9223372036854775808", -9223372036854775808},
// 	}

// 	for _, tc := range tests {
// 		z := MustFromDecimal(tc.x)

// 		got := z.Int64()
// 		if got != tc.want {
// 			t.Errorf("Uint64(%s) = %d, want %d", tc.x, got, tc.want)
// 		}
// 	}
// }

// func TestNeg(t *testing.T) {
// 	tests := []struct {
// 		x    string
// 		want string
// 	}{
// 		{"0", "0"},
// 		{"1", "-1"},
// 		{"-1", "1"},
// 		{"9223372036854775807", "-9223372036854775807"},
// 		{"-18446744073709551615", "18446744073709551615"},
// 	}

// 	for _, tc := range tests {
// 		z := MustFromDecimal(tc.x)
// 		z.Neg(z)

// 		got := z.ToString()
// 		if got != tc.want {
// 			t.Errorf("Neg(%s) = %s, want %s", tc.x, got, tc.want)
// 		}
// 	}
// }

// func TestSet(t *testing.T) {
// 	tests := []struct {
// 		x    string
// 		want string
// 	}{
// 		{"0", "0"},
// 		{"1", "1"},
// 		{"-1", "-1"},
// 		{"9223372036854775807", "9223372036854775807"},
// 		{"-18446744073709551615", "-18446744073709551615"},
// 	}

// 	for _, tc := range tests {
// 		z := MustFromDecimal(tc.x)
// 		z.Set(z)

// 		got := z.ToString()
// 		if got != tc.want {
// 			t.Errorf("Set(%s) = %s, want %s", tc.x, got, tc.want)
// 		}
// 	}
// }

// func TestSetUint256(t *testing.T) {
// 	tests := []struct {
// 		x    string
// 		want string
// 	}{
// 		{"0", "0"},
// 		{"1", "1"},
// 		{"9223372036854775807", "9223372036854775807"},
// 		{"18446744073709551615", "18446744073709551615"},
// 	}

// 	for _, tc := range tests {
// 		got := New()

// 		z := uint256.MustFromDecimal(tc.x)
// 		got.SetUint256(z)

// 		if got.ToString() != tc.want {
// 			t.Errorf("SetUint256(%s) = %s, want %s", tc.x, got.ToString(), tc.want)
// 		}
// 	}
// }

// func TestToString(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		setup    func() *Int
// 		expected string
// 	}{
// 		{
// 			name: "Zero from subtraction",
// 			setup: func() *Int {
// 				minusThree := MustFromDecimal("-3")
// 				three := MustFromDecimal("3")
// 				return Zero().Add(minusThree, three)
// 			},
// 			expected: "0",
// 		},
// 		{
// 			name: "Zero from right shift",
// 			setup: func() *Int {
// 				return Zero().Rsh(One(), 1234)
// 			},
// 			expected: "0",
// 		},
// 		{
// 			name: "Positive number",
// 			setup: func() *Int {
// 				return MustFromDecimal("42")
// 			},
// 			expected: "42",
// 		},
// 		{
// 			name: "Negative number",
// 			setup: func() *Int {
// 				return MustFromDecimal("-42")
// 			},
// 			expected: "-42",
// 		},
// 		{
// 			name: "Large positive number",
// 			setup: func() *Int {
// 				return MustFromDecimal("115792089237316195423570985008687907853269984665640564039457584007913129639935")
// 			},
// 			expected: "115792089237316195423570985008687907853269984665640564039457584007913129639935",
// 		},
// 		{
// 			name: "Large negative number",
// 			setup: func() *Int {
// 				return MustFromDecimal("-115792089237316195423570985008687907853269984665640564039457584007913129639935")
// 			},
// 			expected: "-115792089237316195423570985008687907853269984665640564039457584007913129639935",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			z := tt.setup()
// 			result := z.ToString()
// 			if result != tt.expected {
// 				t.Errorf("ToString() = %s, want %s", result, tt.expected)
// 			}
// 		})
// 	}
// }
