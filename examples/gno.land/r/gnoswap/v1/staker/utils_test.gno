package staker

import (
	"testing"

	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/uassert"
)

func TestPoolPathAlign(t *testing.T) {
	tests := []struct {
		input       string
		expected    string
		shouldPanic bool
	}{
		// Valid cases
		{"baz:bar:500", "bar:baz:500", false},
		{"bar:baz:500", "bar:baz:500", false},
		{"foo:bar:300", "bar:foo:300", false},
		{"bar:foo:300", "bar:foo:300", false},

		// Invalid cases
		{"invalid:path", "", true}, // Missing fee
		{"bar:baz", "", true},      // Too few components
		{"", "", true},             // Empty string
	}

	for _, tt := range tests {
		if tt.shouldPanic {
			// Test for panic
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("poolPathAlign(%s) did not panic as expected", tt.input)
				}
			}()
			_ = poolPathAlign(tt.input)
		} else {
			// Test normal cases
			result := poolPathAlign(tt.input)
			if result != tt.expected {
				t.Errorf("poolPathAlign(%s) = %s; want %s", tt.input, result, tt.expected)
			}
		}
	}
}

func TestPoolPathDivide(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected0   string
		expected1   string
		expectedFee string
		shouldPanic bool
	}{
		// Valid cases
		{"Valid pool path", "bar:baz:500", "bar", "baz", "500", false},
		{"Another valid pool path", "foo:bar:300", "foo", "bar", "300", false},

		// Invalid cases
		{"Missing fee", "bar:baz", "", "", "", true},
		{"Too few components", "bar", "", "", "", true},
		{"Empty string", "", "", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				// Test for panic cases
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("poolPathDivide(%s) did not panic as expected", tt.input)
					}
				}()
				_, _, _ = poolPathDivide(tt.input)
			} else {
				// Test for normal cases
				pToken0, pToken1, fee := poolPathDivide(tt.input)
				if pToken0 != tt.expected0 || pToken1 != tt.expected1 || fee != tt.expectedFee {
					t.Errorf(
						"poolPathDivide(%s) = (%s, %s, %s); want (%s, %s, %s)",
						tt.input, pToken0, pToken1, fee, tt.expected0, tt.expected1, tt.expectedFee,
					)
				}
			}
		})
	}
}

func TestTid(t *testing.T) {
	tests := []struct {
		name        string
		input       any
		expected    string
		shouldPanic bool
	}{
		{
			name:        "Panic - nil",
			input:       nil,
			expected:    "[GNOSWAP-STAKER-022] requested data not found || positionId is nil",
			shouldPanic: true,
		},
		{
			name:        "Panic - unsupported type",
			input:       float64(1),
			expected:    "[GNOSWAP-STAKER-007] invalid input data || unsupported positionId type(unknown)",
			shouldPanic: true,
		},
		{
			name:        "Success - string",
			input:       "1",
			expected:    "1",
			shouldPanic: false,
		},
		{
			name:        "Success - int",
			input:       int(1),
			expected:    "1",
			shouldPanic: false,
		},
		{
			name:        "Success - uint64",
			input:       uint64(1),
			expected:    "1",
			shouldPanic: false,
		},
		{
			name:        "Success - grc721.TokenID",
			input:       grc721.TokenID("1"),
			expected:    "1",
			shouldPanic: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if r == nil {
					if tc.shouldPanic {
						t.Errorf(">>> %s: expected panic but got none", tc.name)
						return
					}
				} else {
					switch r.(type) {
					case string:
						if r.(string) != tc.expected {
							t.Errorf(">>> %s: got panic %v, want %v", tc.name, r, tc.expected)
						}
					case error:
						if r.(error).Error() != tc.expected {
							t.Errorf(">>> %s: got panic %v, want %v", tc.name, r.(error).Error(), tc.expected)
						}
					default:
						t.Errorf(">>> %s: got panic %v, want %v", tc.name, r, tc.expected)
					}
				}
			}()

			if !tc.shouldPanic {
				got := positionIdFrom(tc.input)
				uassert.Equal(t, tc.expected, string(got))
			} else {
				positionIdFrom(tc.input)
			}
		})
	}
}
