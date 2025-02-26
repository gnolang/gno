package staker

import (
	"std"
	"testing"

	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/uassert"

	"gno.land/p/gnoswap/consts"
)

func TestGetOrigPkgAddr(t *testing.T) {
	std.TestSetOriginCaller(consts.STAKER_ADDR)
	origPkgAddr := GetOrigPkgAddr()
	if origPkgAddr != std.OriginCaller() {
		t.Errorf("Expected %v, got %v", std.OriginCaller(), origPkgAddr)
	}
}

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
		input       interface{}
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

func TestGetPrevRealm(t *testing.T) {
	tests := []struct {
		name         string
		originCaller std.Address
		expected     []string
	}{
		{
			name:         "Success - prevRealm is User",
			originCaller: consts.ADMIN,
			expected:     []string{"g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d", ""},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			std.TestSetOriginCaller(std.Address(tc.originCaller))
			got := getPrevRealm()
			uassert.Equal(t, got.Address().String(), tc.expected[0])
			uassert.Equal(t, got.PkgPath(), tc.expected[1])
		})
	}
}

func TestGetPrevAddr(t *testing.T) {
	tests := []struct {
		name         string
		originCaller std.Address
		expected     std.Address
	}{
		{
			name:         "Success - prev Address is User",
			originCaller: consts.ADMIN,
			expected:     "g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			std.TestSetOriginCaller(std.Address(tc.originCaller))
			got := getPrevAddr()
			uassert.Equal(t, got.String(), tc.expected.String())
		})
	}
}

func TestGetPrevAsString(t *testing.T) {
	tests := []struct {
		name         string
		originCaller std.Address
		expected     []string
	}{
		{
			name:         "Success - prev Realm of user info as string",
			originCaller: consts.ADMIN,
			expected:     []string{"g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d", ""},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			std.TestSetOriginCaller(std.Address(tc.originCaller))
			got1, got2 := getPrev()
			uassert.Equal(t, got1, tc.expected[0])
			uassert.Equal(t, got2, tc.expected[1])
		})
	}
}

func TestIsUserCall(t *testing.T) {
	tests := []struct {
		name          string
		originCaller  std.Address
		originPkgPath string
		expected      bool
	}{
		{
			name:         "Success - User Call",
			originCaller: consts.ADMIN,
			expected:     true,
		},
		{
			name:          "Failure - Not User Call",
			originCaller:  consts.ROUTER_ADDR,
			originPkgPath: consts.ROUTER_PATH,
			expected:      false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			std.TestSetOriginCaller(tc.originCaller)
			if !tc.expected {
				std.TestSetRealm(std.NewCodeRealm(tc.originPkgPath))
			}
			got := isUserCall()
			uassert.Equal(t, got, tc.expected)
		})
	}
}
