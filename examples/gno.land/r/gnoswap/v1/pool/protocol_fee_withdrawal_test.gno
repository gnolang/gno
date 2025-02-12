package pool

import (
	"std"
	"strconv"
	"strings"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/gnoswap/consts"
	pn "gno.land/r/gnoswap/v1/position"
)

func TestHandleWithdrawalFee(t *testing.T) {
	var (
		admin       = consts.ADMIN
		position    = consts.POSITION_ADDR
		pool        = consts.POOL_ADDR
		protocolFee = consts.PROTOCOL_FEE_ADDR
		alice       = testutils.TestAddress("alice")
	)
	tests := []struct {
		name        string
		action      func(t *testing.T)
		verify      func(t *testing.T) (string, string)
		expected    string
		shouldPanic bool
	}{
		{
			name: "Panic if caller is not position contract",
			action: func(t *testing.T) {
				std.TestSetOrigCaller(admin)
				HandleWithdrawalFee(0, "gno.land/r/onbloc/foo", "0", "gno.land/r/onbloc/foo", "0", "", admin)
			},
			verify:      nil,
			expected:    "[GNOSWAP-POOL-001] caller has no permission || withdrawal_fee.gno__HandleWithdrawalFee() || only position(g1q646ctzhvn60v492x8ucvyqnrj2w30cwh6efk5) can call this function, called from g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d",
			shouldPanic: true,
		},
		{
			name: "Panic if pkgPath is not registered",
			action: func(t *testing.T) {
				std.TestSetRealm(std.NewUserRealm(position))
				HandleWithdrawalFee(0, "pkgPath", "1000", "pkgPath", "1000", "poolPath", admin)
			},
			verify:      nil,
			expected:    "[GNOSWAP-COMMON-004] token is not registered || token(pkgPath)",
			shouldPanic: true,
		},
		{
			name: "Panic if spender has no approved balance",
			action: func(t *testing.T) {
				InitialisePoolTest(t)
				std.TestSetRealm(std.NewUserRealm(position))
				poolPath := GetPoolPath(wugnotPath, gnsPath, fee3000)
				if !pools.Has(poolPath) {
					panic("pool not found")
				}
				TokenApprove(t, wugnotPath, alice, protocolFee, uint64(0))
				TokenApprove(t, gnsPath, alice, protocolFee, uint64(0))
				HandleWithdrawalFee(1, wugnotPath, "1000", gnsPath, "1000", poolPath, alice)
			},
			verify:      nil,
			expected:    "[GNOSWAP-POOL-021] token transfer failed || insufficient allowance",
			shouldPanic: true,
		},
		{
			name: "Success call by position contract",
			action: func(t *testing.T) {
				InitialisePoolTest(t)
				std.TestSetRealm(std.NewUserRealm(alice))
				TokenApprove(t, wugnotPath, alice, pool, uint64(1000))
				TokenApprove(t, gnsPath, alice, pool, uint64(1000))
				pn.Mint(
					wugnotPath,
					gnsPath,
					fee3000,
					int32(1020),
					int32(5040),
					"1000",
					"1000",
					"0",
					"0",
					max_timeout,
					alice,
					alice,
				)
			},
			verify: func(t *testing.T) (string, string) {
				std.TestSetRealm(std.NewUserRealm(alice))
				TokenApprove(t, wugnotPath, alice, pool, uint64(10000000000))
				TokenApprove(t, gnsPath, alice, pool, uint64(1000000000))
				std.TestSetRealm(std.NewUserRealm(position))
				poolPath := GetPoolPath(wugnotPath, gnsPath, fee3000)
				return HandleWithdrawalFee(2, wugnotPath, "1000", gnsPath, "1000", poolPath, alice)
			},
			expected:    "990,990",
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
				tc.action(t)
				if tc.verify != nil {
					gotAfterAmount0, gotAfterAmount1 := tc.verify(t)
					expected := strings.Split(tc.expected, ",")
					uassert.Equal(t, gotAfterAmount0, expected[0])
					uassert.Equal(t, gotAfterAmount1, expected[1])
				}
			} else {
				tc.action(t)
			}
		})
	}
}

func TestSetWithdrawalFee(t *testing.T) {
	var (
		admin = consts.ADMIN
		alice = testutils.TestAddress("alice")
	)
	tests := []struct {
		name        string
		action      func(t *testing.T)
		verify      func(t *testing.T) string
		expected    string
		shouldPanic bool
	}{
		{
			name: "Panic call to set withdrawal fee setFee by non-admin (user)",
			action: func(t *testing.T) {
				const newFee = uint64(200)
				userRealm := std.NewUserRealm(alice)
				std.TestSetRealm(userRealm)
				{
					SetWithdrawalFee(newFee)
				}
			},
			verify:      nil,
			expected:    "caller(g1v9kxjcm9ta047h6lta047h6lta047h6lzd40gh) has no permission",
			shouldPanic: true,
		},
		{
			name: "Panic call to set withdrawal fee by admin",
			action: func(t *testing.T) {
				const newFee = uint64(200)
				adminRealm := std.NewUserRealm(admin)
				std.TestSetRealm(adminRealm)
				{
					SetWithdrawalFee(newFee)
				}
			},
			verify:      nil,
			expected:    "caller(g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d) has no permission",
			shouldPanic: true,
		},
		{
			name: "Success call to set withdrawal fee by governance",
			action: func(t *testing.T) {
				const newFee = uint64(200)
				govRealm := std.NewUserRealm(consts.GOV_GOVERNANCE_ADDR)
				std.TestSetRealm(govRealm)
				{
					SetWithdrawalFee(newFee)
				}
			},
			verify: func(t *testing.T) string {
				return strconv.FormatUint(GetWithdrawalFee(), 10)
			},
			expected:    strconv.FormatUint(200, 10),
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
				tc.action(t)
				if tc.verify != nil {
					gotWithdrawalFee := tc.verify(t)
					uassert.Equal(t, gotWithdrawalFee, tc.expected)
				}
			} else {
				tc.action(t)
			}
		})
	}
}

func TestSetWithdrawalFeeByAdmin(t *testing.T) {
	var (
		admin      = consts.ADMIN
		alice      = testutils.TestAddress("alice")
		governance = consts.GOV_GOVERNANCE_ADDR
	)
	tests := []struct {
		name        string
		action      func(t *testing.T)
		verify      func(t *testing.T) string
		expected    string
		shouldPanic bool
	}{
		{
			name: "Panic call to set withdrawal fee by non-admin (user)",
			action: func(t *testing.T) {
				const newFee = uint64(100)
				userRealm := std.NewUserRealm(alice)
				std.TestSetRealm(userRealm)
				{
					SetWithdrawalFeeByAdmin(newFee)
				}
			},
			verify:      nil,
			expected:    "caller(g1v9kxjcm9ta047h6lta047h6lta047h6lzd40gh) has no permission",
			shouldPanic: true,
		},
		{
			name: "Panic call to set withdrawal fee by non-admin (gov contract)",
			action: func(t *testing.T) {
				const newFee = uint64(100)
				govRealm := std.NewUserRealm(governance)
				std.TestSetRealm(govRealm)
				{
					SetWithdrawalFeeByAdmin(newFee)
				}
			},
			verify:      nil,
			expected:    "caller(g17s8w2ve7k85fwfnrk59lmlhthkjdted8whvqxd) has no permission",
			shouldPanic: true,
		},
		{
			name: "Success call to set withdrawal fee by admin",
			action: func(t *testing.T) {
				const newFee = uint64(100)
				adminRealm := std.NewUserRealm(admin)
				std.TestSetRealm(adminRealm)
				{
					SetWithdrawalFeeByAdmin(newFee)
				}
			},
			verify: func(t *testing.T) string {
				return strconv.FormatUint(GetWithdrawalFee(), 10)
			},
			expected:    strconv.FormatUint(100, 10),
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
				tc.action(t)
				if tc.verify != nil {
					gotWithdrawalFee := tc.verify(t)
					uassert.Equal(t, gotWithdrawalFee, tc.expected)
				}
			} else {
				tc.action(t)
			}
		})
	}
}
