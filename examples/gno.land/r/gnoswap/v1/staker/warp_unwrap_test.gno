package staker

import (
	"std"
	"strconv"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/gnoswap/consts"
)

func TestWrap(t *testing.T) {
	user1Addr := testutils.TestAddress("user1")
	tests := []struct {
		name        string
		action      func()
		verify      func() uint64
		expected    string
		shouldPanic bool
	}{
		{
			name: "Failure - Amount less than minimum",
			action: func() {
				wrap(999)
			},
			verify:      nil,
			expected:    "[GNOSWAP-STAKER-006] can not wrapless than minimum amount || amount(999) < minimum(1000)",
			shouldPanic: true,
		},
		{
			name: "Failure - Zero amount",
			action: func() {
				wrap(0)
			},
			verify:      nil,
			expected:    "[GNOSWAP-STAKER-005] wrap, unwrap failed || cannot wrap 0 ugnot",
			shouldPanic: true,
		},
		{
			name: "Success - Valid amount",
			action: func() {
				std.TestSetRealm(std.NewUserRealm(user1Addr))
				ugnotFaucet(t, user1Addr, 1_000)
				ugnotFaucet(t, consts.STAKER_ADDR, 1_000)
				std.TestSetRealm(std.NewUserRealm(user1Addr))
				wrap(1_000)
			},
			verify: func() uint64 {
				return TokenBalance(t, wugnotPath, user1Addr)
			},
			expected:    "1000",
			shouldPanic: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					switch v := r.(type) {
					case string:
						if v != tc.expected {
							t.Errorf("Expected panic: %s, got: %s", tc.expected, v)
						}
					case error:
						if v.Error() != tc.expected {
							t.Errorf("Expected panic: %s, got: %s", tc.expected, v.Error())
						}
					default:
						t.Errorf("Unexpected panic type: %v", r)
					}
				}
			}()

			if tc.shouldPanic {
				tc.action()
			} else {
				tc.action()
				if tc.verify != nil {
					balance := tc.verify()
					if strconv.FormatUint(balance, 10) != tc.expected {
						t.Errorf("Expected balance: %s, got: %d", tc.expected, balance)
					}
				}
			}
		})
	}
}

func TestUnwrap(t *testing.T) {
	user2Addr := testutils.TestAddress("user2")
	tests := []struct {
		name        string
		action      func()
		verify      func() uint64
		expected    string
		shouldPanic bool
	}{
		{
			name: "Failure - Zero amount",
			action: func() {
				unwrap(0)
			},
			verify:      nil,
			expected:    "",
			shouldPanic: false, // No panic as zero amount is ignored
		},
		{
			name: "Success - Valid amount",
			action: func() {
				std.TestSetRealm(std.NewUserRealm(user2Addr))
				ugnotFaucet(t, user2Addr, 1_000)
				ugnotFaucet(t, consts.STAKER_ADDR, 1_000)
				std.TestSetRealm(std.NewUserRealm(user2Addr))
				wrap(1_000)
				std.TestSetRealm(std.NewUserRealm(user2Addr))
				wugnotApprove(t, user2Addr, consts.STAKER_ADDR, 1_000)
				unwrap(1_000)
			},
			verify: func() uint64 {
				return TokenBalance(t, wugnotPath, user2Addr)
			},
			expected:    "0",
			shouldPanic: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					switch v := r.(type) {
					case string:
						if v != tc.expected {
							t.Errorf("Expected panic: %s, got: %s", tc.expected, v)
						}
					case error:
						if v.Error() != tc.expected {
							t.Errorf("Expected panic: %s, got: %s", tc.expected, v.Error())
						}
					default:
						t.Errorf("Unexpected panic type: %v", r)
					}
				}
			}()

			if tc.shouldPanic {
				tc.action()
			} else {
				tc.action()
				if tc.verify != nil {
					balance := tc.verify()
					if strconv.FormatUint(balance, 10) != tc.expected {
						t.Errorf("Expected balance: %s, got: %d", tc.expected, balance)
					}
				}
			}
		})
	}
}
