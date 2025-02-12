package position

import (
	"fmt"
	"std"
	"strings"
	"testing"

	"gno.land/p/demo/grc/grc721"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"
	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/common"
)

func assertPanic(t *testing.T, expectedMsg string, fn func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("expected panic but got none")
		} else if r != expectedMsg {
			t.Errorf("expected panic %v, got %v", expectedMsg, r)
		}
	}()
	fn()
}

func TestGetOrigPkgAddr(t *testing.T) {
	tests := []struct {
		name     string
		expected std.Address
	}{
		{
			name:     "Success - getOrigPkgAddr",
			expected: consts.POSITION_ADDR,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := GetOrigPkgAddr()
			uassert.Equal(t, got, tc.expected)
		})
	}
}

func TestAssertTokenExists(t *testing.T) {
	// Mock setup
	mockTokenId := uint64(1)
	mockInvalidTokenId := uint64(9999)

	t.Run("Token Exists - No Panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("unexpected panic for existing tokenId(%d): %v", mockTokenId, r)
			}
		}()
		assertTokenExists(mockTokenId) // Should not panic
	})

	t.Run("Token Does Not Exist - Panic Expected", func(t *testing.T) {
		expectedErr := ufmt.Sprintf("[GNOSWAP-POSITION-006] requested data not found || tokenId(%d) doesn't exist", mockInvalidTokenId)
		uassert.PanicsWithMessage(t, expectedErr, func() {
			assertTokenExists(mockInvalidTokenId) // Should panic
		})
	})

	t.Run("Boundary Case - TokenId Zero", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic for tokenId(0) but did not occur")
			}
		}()
		assertTokenExists(0) // Assume tokenId 0 does not exist
	})
}

func TestAssertOnlyOwnerOfToken(t *testing.T) {
	// Mock token ownership
	mockTokenId := uint64(1)
	mockOwner := admin
	mockCaller := alice
	//MakeMintPositionWithoutFee(t)

	t.Run("Token Owned by Caller - No Panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("unexpected panic for tokenId(%d) called by owner(%s): %v", mockTokenId, mockOwner, r)
			}
		}()
		assertOnlyOwnerOfToken(mockTokenId, mockOwner) // Should pass without panic
	})

	t.Run("Token Not Owned by Caller - Panic Expected", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic for tokenId(%d) called by unauthorized user(%s), but no panic occurred", mockTokenId, mockCaller)
			} else {
				expectedErr := ufmt.Sprintf("caller(%s) is not owner(%s) for tokenId(%d)", mockCaller, mockOwner, mockTokenId)
				if !strings.Contains(fmt.Sprintf("%v", r), expectedErr) {
					t.Errorf("unexpected error message: %v", r)
				}
			}
		}()
		assertOnlyOwnerOfToken(mockTokenId, mockCaller) // Should panic
	})

	t.Run("Non-Existent Token - Panic Expected", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("expected panic for non-existent tokenId(9999), but no panic occurred")
			}
		}()
		assertOnlyOwnerOfToken(9999, mockCaller) // Token 9999 does not exist, should panic
	})
}

func TestAssertOnlyUserOrStaker(t *testing.T) {
	tests := []struct {
		name         string
		originCaller std.Address
		expected     bool
	}{
		{
			name:         "Failure - Not User or Staker",
			originCaller: consts.ROUTER_ADDR,
			expected:     false,
		},
		{
			name:         "Success - User Call",
			originCaller: consts.ADMIN,
			expected:     true,
		},
		{
			name:         "Success - Staker Call",
			originCaller: consts.STAKER_ADDR,
			expected:     true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			std.TestSetOrigCaller(tc.originCaller)
			assertOnlyUserOrStaker(std.PrevRealm())
		})
	}
}

func TestAssertOnlyNotHalted(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
		panicMsg string
	}{
		{
			name:     "Failure - Halted",
			expected: false,
			panicMsg: "[GNOSWAP-COMMON-002] halted || GnoSwap is halted",
		},
		{
			name:     "Success - Not Halted",
			expected: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expected {
				uassert.NotPanics(t, func() {
					assertOnlyNotHalted()
				})
			} else {
				std.TestSetRealm(std.NewUserRealm(admin))
				common.SetHaltByAdmin(true)
				uassert.PanicsWithMessage(t, tc.panicMsg, func() {
					assertOnlyNotHalted()
				})
				common.SetHaltByAdmin(false)
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
			addr:     consts.ADMIN,
			expected: true,
		},
		{
			name:     "Failure - invalid address",
			addr:     "g1lmvrrrr4er2us84h2732sru76c9zl2nvknha8", // invalid length
			expected: false,
			errorMsg: "[GNOSWAP-POSITION-012] invalid address || (g1lmvrrrr4er2us84h2732sru76c9zl2nvknha8)",
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

func TestAssertOnlyValidAddressWith(t *testing.T) {
	tests := []struct {
		name     string
		addr     std.Address
		other    std.Address
		expected bool
		errorMsg string
	}{
		{
			name:     "Success - validation address check to compare with other address",
			addr:     consts.ADMIN,
			other:    std.Address("g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d"),
			expected: true,
		},
		{
			name:     "Failure - two address is different",
			addr:     "g1lmvrrrr4er2us84h2732sru76c9zl2nvknha8",
			other:    "g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d",
			expected: false,
			errorMsg: "[GNOSWAP-POSITION-012] invalid address || (g1lmvrrrr4er2us84h2732sru76c9zl2nvknha8)",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expected {
				uassert.NotPanics(t, func() {
					assertOnlyValidAddressWith(tc.addr, tc.other)
				})
			} else {
				uassert.PanicsWithMessage(t, tc.errorMsg, func() {
					assertOnlyValidAddressWith(tc.addr, tc.other)
				})
			}
		})
	}
}

func TestAssertValidNumberString(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectPanic bool
	}{
		// Valid Cases
		{
			name:        "Valid Positive Number",
			input:       "12345",
			expectPanic: false,
		},
		{
			name:        "Valid Negative Number",
			input:       "-98765",
			expectPanic: false,
		},
		{
			name:        "Zero",
			input:       "0",
			expectPanic: false,
		},
		{
			name:        "Negative Zero",
			input:       "-0",
			expectPanic: false,
		},

		// Invalid Cases
		{
			name:        "Empty String",
			input:       "",
			expectPanic: true,
		},
		{
			name:        "Alphabet in String",
			input:       "12a45",
			expectPanic: true,
		},
		{
			name:        "Special Characters",
			input:       "123@45",
			expectPanic: true,
		},
		{
			name:        "Leading Plus Sign",
			input:       "+12345",
			expectPanic: true,
		},
		{
			name:        "Multiple Negative Signs",
			input:       "--12345",
			expectPanic: true,
		},
		{
			name:        "Space in String",
			input:       "123 45",
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if !tt.expectPanic {
						t.Errorf("unexpected panic for input: %s, got: %v", tt.input, r)
					}
				} else {
					if tt.expectPanic {
						t.Errorf("expected panic but did not occur for input: %s", tt.input)
					}
				}
			}()

			// Test function
			assertValidNumberString(tt.input)
		})
	}
}

func TestDerivePkgAddr(t *testing.T) {
	pkgPath := "gno.land/r/gnoswap/v1/position"
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
			std.TestSetOrigCaller(std.Address(tc.originCaller))
			got1, got2 := getPrevAsString()
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
			std.TestSetOrigCaller(tc.originCaller)
			if !tc.expected {
				std.TestSetRealm(std.NewCodeRealm(tc.originPkgPath))
			}
			got := isUserCall()
			uassert.Equal(t, got, tc.expected)
		})
	}
}

func TestCheckDeadline(t *testing.T) {
	tests := []struct {
		name     string
		deadline int64
		now      int64
		expected string
	}{
		{
			name:     "Success - checkDeadline",
			deadline: 1234567890 + 100,
			now:      1234567890,
			expected: "",
		},
		{
			name:     "Failure - checkDeadline",
			deadline: 1234567890 - 100,
			now:      1234567890,
			expected: "[GNOSWAP-POSITION-007] transaction expired || transaction too old, now(1234567890) > deadline(1234567790)",
		},
		{
			name:     "Success - deadline equals now",
			deadline: 1234567890,
			now:      1234567890,
			expected: "",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expected != "" {
				uassert.PanicsWithMessage(t, tc.expected, func() {
					checkDeadline(tc.deadline)
				})
			} else {
				uassert.NotPanics(t, func() {
					checkDeadline(tc.deadline)
				})
			}
		})
	}
}

func TestTokenIdFrom(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		expected    string
		shouldPanic bool
	}{
		{
			name:        "Panic - nil",
			input:       nil,
			expected:    "[GNOSWAP-POSITION-005] invalid input data || tokenId is nil",
			shouldPanic: true,
		},
		{
			name:        "Panic - unsupported type",
			input:       float64(1),
			expected:    "[GNOSWAP-POSITION-005] invalid input data || unsupported tokenId type",
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
				got := tokenIdFrom(tc.input)
				uassert.Equal(t, tc.expected, string(got))
			} else {
				tokenIdFrom(tc.input)
			}
		})
	}
}

func TestExists(t *testing.T) {
	tests := []struct {
		name     string
		tokenId  uint64
		expected bool
	}{
		{
			name:     "Fail - not exists",
			tokenId:  300000,
			expected: false,
		},
		{
			name:     "Success - exists",
			tokenId:  1,
			expected: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := exists(tc.tokenId)
			uassert.Equal(t, tc.expected, got)
		})
	}
}

func TestIsOwner(t *testing.T) {
	tests := []struct {
		name     string
		tokenId  uint64
		addr     std.Address
		expected bool
	}{
		{
			name:     "Fail - is not owner",
			tokenId:  1,
			addr:     alice,
			expected: false,
		},
		{
			name:     "Success - is owner",
			tokenId:  1,
			addr:     admin,
			expected: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			//MakeMintPositionWithoutFee(t)
			got := isOwner(tc.tokenId, tc.addr)
			uassert.Equal(t, tc.expected, got)
		})
	}
}

func TestIsOperator(t *testing.T) {
	tests := []struct {
		name     string
		tokenId  uint64
		addr     std.Address
		expected bool
	}{
		{
			name:     "Fail - is not operator",
			tokenId:  1,
			addr:     alice,
			expected: false,
		},
		{
			name:     "Success - is operator",
			tokenId:  1,
			addr:     bob,
			expected: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expected {
				LPTokenApprove(t, admin, tc.addr, tc.tokenId)
			}
			got := isOperator(tc.tokenId, tc.addr)
			uassert.Equal(t, tc.expected, got)
		})
	}
}

func TestIsStaked(t *testing.T) {
	tests := []struct {
		name     string
		owner    std.Address
		operator std.Address
		tokenId  uint64
		expected bool
	}{
		{
			name:     "Fail - is not staked",
			owner:    bob,
			operator: alice,
			tokenId:  1,
			expected: false,
		},
		{
			name:     "Fail - is not exist tokenId",
			owner:    admin,
			operator: bob,
			tokenId:  100,
			expected: false,
		},
		{
			name:     "Success - is staked",
			owner:    admin,
			operator: admin,
			tokenId:  11,
			expected: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expected && tc.owner == tc.operator {
				MakeMintPositionWithoutFee(t)
				LPTokenApprove(t, tc.owner, consts.STAKER_ADDR, tc.tokenId)
				LPTokenStake(t, tc.owner, tc.tokenId)
			}
			got := isStaked(tokenIdFrom(tc.tokenId))
			uassert.Equal(t, tc.expected, got)
			if tc.expected && tc.owner == tc.operator {
				LPTokenUnStake(t, tc.owner, tc.tokenId, false)
			}
		})
	}
}

func TestIsOwnerOrOperator(t *testing.T) {
	tests := []struct {
		name     string
		owner    std.Address
		operator std.Address
		tokenId  uint64
		expected bool
	}{
		{
			name:     "Fail - is not owner or operator",
			owner:    admin,
			operator: alice,
			tokenId:  1,
			expected: false,
		},
		{
			name:     "Success - is operator",
			owner:    admin,
			operator: bob,
			tokenId:  1,
			expected: true,
		},
		{
			name:     "Success - is owner",
			owner:    admin,
			operator: admin,
			tokenId:  1,
			expected: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.expected && tc.owner != tc.operator {
				LPTokenApprove(t, tc.owner, tc.operator, tc.tokenId)
			}
			var got bool
			if tc.owner == tc.operator {
				got = isOwnerOrOperator(tc.owner, tc.tokenId)
			} else {
				got = isOwnerOrOperator(tc.operator, tc.tokenId)
			}
			uassert.Equal(t, tc.expected, got)
		})
	}
}

func TestIsOwnerOrOperatorWithStake(t *testing.T) {
	tests := []struct {
		name     string
		owner    std.Address
		operator std.Address
		tokenId  uint64
		isStake  bool
		expected bool
	}{
		{
			name:     "Fail - is not token staked",
			owner:    admin,
			operator: alice,
			tokenId:  10,
			isStake:  false,
			expected: false,
		},
		{
			name:     "Success - is token staked (position operator)",
			owner:    admin,
			operator: admin,
			tokenId:  10,
			isStake:  true,
			expected: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.isStake {
				tokenId, _, _, _ := MakeMintPositionWithoutFee(t)
				LPTokenApprove(t, tc.owner, consts.STAKER_ADDR, tc.tokenId)
				LPTokenStake(t, tc.owner, tc.tokenId)
			}
			got := isOwnerOrOperator(tc.operator, tc.tokenId)
			uassert.Equal(t, tc.expected, got)
		})
	}
}

func TestPoolKeyDivide(t *testing.T) {
	tests := []struct {
		name          string
		poolKey       string
		expectedPath0 string
		expectedPath1 string
		expectedFee   uint32
		expectedError string
		shouldPanic   bool
	}{
		{
			name:          "Fail - invalid poolKey",
			poolKey:       "gno.land/r/onbloc",
			expectedError: "[GNOSWAP-POSITION-005] invalid input data || invalid poolKey(gno.land/r/onbloc)",
			shouldPanic:   true,
		},
		{
			name:          "Success - split poolKey",
			poolKey:       "gno.land/r/gnoswap/v1/gns:gno.land/r/demo/wugnot:500",
			expectedPath0: gnsPath,
			expectedPath1: wugnotPath,
			expectedFee:   fee500,
			shouldPanic:   false,
		},
		{
			name:          "Fail -empty poolKey",
			poolKey:       "",
			expectedError: "[GNOSWAP-POSITION-005] invalid input data || invalid poolKey()",
			shouldPanic:   true,
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
						if r.(string) != tc.expectedError {
							t.Errorf(">>> %s: got panic %v, want %v", tc.name, r, tc.expectedError)
						}
					case error:
						if r.(error).Error() != tc.expectedError {
							t.Errorf(">>> %s: got panic %v, want %v", tc.name, r.(error).Error(), tc.expectedError)
						}
					default:
						t.Errorf(">>> %s: got panic %v, want %v", tc.name, r, tc.expectedError)
					}
				}
			}()

			if !tc.shouldPanic {
				gotToken0, gotToken1, gotFee := splitOf(tc.poolKey)
				uassert.Equal(t, tc.expectedPath0, gotToken0)
				uassert.Equal(t, tc.expectedPath1, gotToken1)
				uassert.Equal(t, tc.expectedFee, gotFee)
			} else {
				splitOf(tc.poolKey)
			}
		})
	}
}

func TestSplitOf(t *testing.T) {
	tests := []struct {
		name          string
		poolKey       string
		expectedPath0 string
		expectedPath1 string
		expectedFee   uint32
		expectedError string
		shouldPanic   bool
	}{
		{
			name:          "Fail - empty poolKey",
			poolKey:       "",
			expectedError: "[GNOSWAP-POSITION-005] invalid input data || invalid poolKey()",
			shouldPanic:   true,
		},
		{
			name:          "Fail - invalid delimiter",
			poolKey:       "gno.land/r/gnoswap:v1/gns:gno.land/r/demo/wugnot-500",
			expectedError: "[GNOSWAP-POSITION-005] invalid input data || invalid fee(gno.land/r/demo/wugnot-500)",
			shouldPanic:   true,
		},
		{
			name:          "Fail - non-numeric fee",
			poolKey:       "gno.land/r/gnoswap/v1/gns:gno.land/r/demo/wugnot:fee",
			expectedError: "[GNOSWAP-POSITION-005] invalid input data || invalid fee(fee)",
			shouldPanic:   true,
		},
		{
			name:          "Fail - missing fee part",
			poolKey:       "gno.land/r/gnoswap/v1/gns:gno.land/r/demo/wugnot:",
			expectedError: "[GNOSWAP-POSITION-005] invalid input data || invalid fee()",
			shouldPanic:   true,
		},
		{
			name:          "Fail - insufficient parts",
			poolKey:       "gno.land/r/gnoswap/v1/gns:gno.land/r/demo",
			expectedError: "[GNOSWAP-POSITION-005] invalid input data || invalid poolKey(gno.land/r/gnoswap/v1/gns:gno.land/r/demo)",
			shouldPanic:   true,
		},
		{
			name:          "Success - valid poolKey",
			poolKey:       "gno.land/r/gnoswap/v1/gns:gno.land/r/demo/wugnot:500",
			expectedPath0: "gno.land/r/gnoswap/v1/gns",
			expectedPath1: "gno.land/r/demo/wugnot",
			expectedFee:   500,
			shouldPanic:   false,
		},
		{
			name:          "Success - poolKey with large fee",
			poolKey:       "gno.land/r/gnoswap/v1/gns:gno.land/r/demo/wugnot:10000",
			expectedPath0: "gno.land/r/gnoswap/v1/gns",
			expectedPath1: "gno.land/r/demo/wugnot",
			expectedFee:   10000,
			shouldPanic:   false,
		},
		{
			name:          "Success - poolKey with minimal fee",
			poolKey:       "gno.land/r/gnoswap/v1/gns:gno.land/r/demo/wugnot:1",
			expectedPath0: "gno.land/r/gnoswap/v1/gns",
			expectedPath1: "gno.land/r/demo/wugnot",
			expectedFee:   1,
			shouldPanic:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldPanic {
				assertPanic(t, tc.expectedError, func() {
					splitOf(tc.poolKey)
				})
			} else {
				gotToken0, gotToken1, gotFee := splitOf(tc.poolKey)
				uassert.Equal(t, tc.expectedPath0, gotToken0, "Token0 mismatch")
				uassert.Equal(t, tc.expectedPath1, gotToken1, "Token1 mismatch")
				uassert.Equal(t, tc.expectedFee, gotFee, "Fee mismatch")
			}
		})
	}
}
