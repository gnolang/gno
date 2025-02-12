package pool

import (
	"std"
	"strconv"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/gnoswap/consts"
)

func TestSetPoolCreationFee(t *testing.T) {
	var (
		admin = std.Address("g1lmvrrrr4er2us84h2732sru76c9zl2nvknha8c")
	)

	tests := []struct {
		name        string
		action      func()
		verify      func() string
		expected    string
		shouldPanic bool
	}{
		{
			name: "Panic call by non-governance",
			action: func() {
				const newFee = 2_000_000_000
				std.TestSetOrigCaller(admin)
				SetPoolCreationFee(newFee)
			},
			verify:      nil,
			expected:    "caller(g1lmvrrrr4er2us84h2732sru76c9zl2nvknha8c) has no permission",
			shouldPanic: true,
		},
		{
			name: "Success call by governance",
			action: func() {
				const newFee = 2_000_000_000
				govRealm := std.NewUserRealm(consts.GOV_GOVERNANCE_ADDR)
				std.TestSetRealm(govRealm)
				SetPoolCreationFee(newFee)
			},
			verify: func() string {
				return strconv.FormatUint(GetPoolCreationFee(), 10)
			},
			expected:    strconv.FormatUint(2_000_000_000, 10),
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
					if r.(string) != tc.expected {
						t.Errorf(">>> %s: got panic %v, want %v", tc.name, r, tc.expected)
					}
				}
			}()

			if !tc.shouldPanic {
				tc.action()
				if tc.verify != nil {
					got := tc.verify()
					uassert.Equal(t, got, tc.expected)
				}
			} else {
				tc.action()
			}
		})
	}
}

func TestSetPoolCreationFeeByAdmin(t *testing.T) {
	var (
		admin = consts.ADMIN
		alice = testutils.TestAddress("alice")
	)

	tests := []struct {
		name        string
		action      func()
		verify      func() string
		expected    string
		shouldPanic bool
	}{
		{
			name: "Panic call by non-admin (gov contract)",
			action: func() {
				const newFee = 2_000_000_000
				govRealm := std.NewUserRealm(consts.GOV_GOVERNANCE_ADDR)
				std.TestSetRealm(govRealm)
				{
					SetPoolCreationFeeByAdmin(newFee)
				}
			},
			verify:      nil,
			expected:    "caller(g17s8w2ve7k85fwfnrk59lmlhthkjdted8whvqxd) has no permission",
			shouldPanic: true,
		},
		{
			name: "Panic call by non-admin (user)",
			action: func() {
				const newFee = 2_000_000_000
				std.TestSetOrigCaller(alice)
				{
					SetPoolCreationFeeByAdmin(newFee)
				}
			},
			verify:      nil,
			expected:    "caller(g1v9kxjcm9ta047h6lta047h6lta047h6lzd40gh) has no permission",
			shouldPanic: true,
		},
		{
			name: "Success call by admin",
			action: func() {
				const newFee = 2_000_000_000
				std.TestSetOrigCaller(admin)
				{
					SetPoolCreationFeeByAdmin(newFee)
				}
			},
			verify: func() string {
				return strconv.FormatUint(GetPoolCreationFee(), 10)
			},
			expected:    strconv.FormatUint(2_000_000_000, 10),
			shouldPanic: false,
		},
		{
			name: "Success call by admin (rollback)",
			action: func() {
				const newFee = 1_000_000_000
				std.TestSetOrigCaller(admin)
				{
					SetPoolCreationFeeByAdmin(newFee)
				}
			},
			verify: func() string {
				return strconv.FormatUint(GetPoolCreationFee(), 10)
			},
			expected:    strconv.FormatUint(1_000_000_000, 10),
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
					if r.(string) != tc.expected {
						t.Errorf(">>> %s: got panic %v, want %v", tc.name, r, tc.expected)
					}
				}
			}()

			if !tc.shouldPanic {
				tc.action()
				if tc.verify != nil {
					got := tc.verify()
					uassert.Equal(t, got, tc.expected)
				}
			} else {
				tc.action()
			}
		})
	}
}
