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
		{
			name:     "compare max u256 with zero",
			num1:     "115792089237316195423570985008687907853269984665640564039457584007913129639935",
			num2:     "0",
			expected: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			num1 := u256.MustFromDecimal(tt.num1)
			num2 := u256.MustFromDecimal(tt.num2)
			expected := u256.MustFromDecimal(tt.expected)

			uassert.Equal(t, expected.ToString(), u256Min(num1, num2).ToString())
		})
	}
}

func TestIsUserCall(t *testing.T) {
	tests := []struct {
		name     string
		action   func() bool
		expected bool
	}{
		{
			name: "called from user",
			action: func() bool {
				userRealm := std.NewUserRealm(std.Address("user"))
				std.TestSetRealm(userRealm)
				return isUserCall()
			},
			expected: true,
		},
		{
			name: "called from realm",
			action: func() bool {
				fromRealm := std.NewCodeRealm("gno.land/r/realm")
				std.TestSetRealm(fromRealm)
				return isUserCall()
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uassert.Equal(t, tt.expected, tt.action())
		})
	}
}

func TestGetPrevAsString(t *testing.T) {
	tests := []struct {
		name            string
		action          func() (string, string)
		expectedAddr    string
		expectedPkgPath string
	}{
		{
			name: "user call",
			action: func() (string, string) {
				userRealm := std.NewUserRealm(std.Address("user"))
				std.TestSetRealm(userRealm)
				return getPrevAsString()
			},
			expectedAddr:    "user",
			expectedPkgPath: "",
		},
		{
			name: "code call",
			action: func() (string, string) {
				codeRealm := std.NewCodeRealm("gno.land/r/demo/realm")
				std.TestSetRealm(codeRealm)
				return getPrevAsString()
			},
			expectedAddr:    std.DerivePkgAddr("gno.land/r/demo/realm").String(),
			expectedPkgPath: "gno.land/r/demo/realm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addr, pkgPath := tt.action()
			uassert.Equal(t, tt.expectedAddr, addr)
			uassert.Equal(t, tt.expectedPkgPath, pkgPath)
		})
	}
}

func TestSafeConvertToUint64(t *testing.T) {
	tests := []struct {
		name      string
		value     *u256.Uint
		wantRes   uint64
		wantPanic bool
	}{
		{"normal conversion", u256.NewUint(123), 123, false},
		{"overflow", u256.MustFromDecimal(consts.MAX_UINT128), 0, true},
		{"max uint64", u256.NewUint(1<<64 - 1), 1<<64 - 1, false},
		{"zero", u256.NewUint(0), 0, false},
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

			res := safeConvertToUint64(tt.value)
			if res != tt.wantRes {
				t.Errorf("safeConvertToUint64() = %v, want %v", res, tt.wantRes)
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
	}{
		{"normal conversion", "170141183460469231731687303715884105727", "170141183460469231731687303715884105727", false},
		{"overflow", "170141183460469231731687303715884105728", "", true},
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
				t.Errorf("safeConvertToUint64() = %v, want %v", res, tt.wantRes)
			}
		})
	}
}

func TestDerivePkgAddr(t *testing.T) {
	var (
		pkgPath = "gno.land/r/gnoswap/v1/position"
	)
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Success - derivePkgAddr",
			input:    pkgPath,
			expected: "g1q646ctzhvn60v492x8ucvyqnrj2w30cwh6efk5",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := derivePkgAddr(tc.input)
			uassert.Equal(t, got.String(), tc.expected)
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
			std.TestSetOrigCaller(std.Address(tc.originCaller))
			got := getPrevRealm()
			uassert.Equal(t, got.Addr().String(), tc.expected[0])
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
			std.TestSetOrigCaller(std.Address(tc.originCaller))
			got := getPrevAddr()
			uassert.Equal(t, got.String(), tc.expected.String())
		})
	}
}

func TestCheckOverFlowInt128(t *testing.T) {
	tests := []struct {
		name        string
		input       *i256.Int
		shouldPanic bool
		expected    string
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
			input:       i256.MustFromDecimal(consts.MAX_INT256),
			shouldPanic: true,
			expected:    "[GNOSWAP-POOL-026] overflow: amount(170141183460469231731687303715884105728) overflows int128 range",
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
					uassert.Equal(t, tt.expected, r)
					t.Errorf("Expected panic but none occurred")
				}
			}()
			checkOverFlowInt128(tt.input)
		})
	}
}

func TestCheckTickSpacing(t *testing.T) {
	tests := []struct {
		name        string
		tick        int32
		tickSpacing int32
		shouldPanic bool
		expected    string
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
			expected:    "[GNOSWAP-POOL-022] invalid tick and tick spacing requested || tick(15) MOD tickSpacing(10) != 0(5)",
		},
		{
			name:        "Invalid tick - negative tick",
			tick:        -35,
			tickSpacing: 20,
			shouldPanic: true,
			expected:    "[GNOSWAP-POOL-022] invalid tick and tick spacing requested || tick(-35) MOD tickSpacing(20) != 0(-15)",
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
					uassert.Equal(t, tt.expected, r)
				}
			}()
			checkTickSpacing(tt.tick, tt.tickSpacing)
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
			addr:     consts.ADMIN,
			expected: true,
		},
		{
			name:     "Failure - invalid address",
			addr:     "g1lmvrrrr4er2us84h2732sru76c9zl2nvknha8", // invalid length
			expected: false,
			errorMsg: "[GNOSWAP-POOL-023] invalid address || (g1lmvrrrr4er2us84h2732sru76c9zl2nvknha8)",
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
