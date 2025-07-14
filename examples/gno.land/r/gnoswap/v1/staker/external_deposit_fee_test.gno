package staker

import (
	"std"
	"testing"

	"gno.land/r/gnoswap/v1/rbac"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"
)

func TestGetDepositGnsAmount(t *testing.T) {
	expected := int64(1_000_000_000)
	actual := GetDepositGnsAmount()

	if actual != expected {
		t.Errorf("GetDepositGnsAmount() = %d; want %d", actual, expected)
	}
}

func TestSetDepositGnsAmountByAdmin(t *testing.T) {
	tests := []struct {
		name             string
		caller           std.Address
		inputAmount      int64
		expectedAmount   int64
		expectedHasAbort bool
		expectedAbortMsg string
		expectedHasPanic bool
		expectedPanicMsg string
	}{
		{
			name:             "set deposit amount by admin success",
			caller:           testutils.TestAddress("admin"),
			inputAmount:      2_000_000_000,
			expectedAmount:   2_000_000_000,
			expectedHasAbort: false,
			expectedAbortMsg: "",
			expectedHasPanic: false,
			expectedPanicMsg: "",
		},
		{
			name:             "set deposit amount by admin failed by unauthorized caller",
			caller:           testutils.TestAddress("user1"),
			inputAmount:      2_000_000_000,
			expectedAmount:   0,
			expectedHasPanic: false,
			expectedPanicMsg: "",
			expectedHasAbort: true,
			expectedAbortMsg: "caller(g1w4ek2u33ta047h6lta047h6lta047h6ldvdwpn) has no permission",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			depositGnsAmount = 0
			setupExternalDepositFeeAdminPermission(t)

			testing.SetOriginCaller(tt.caller)

			// when
			if tt.expectedHasPanic {
				uassert.PanicsWithMessage(t, tt.expectedPanicMsg, func() {
					SetDepositGnsAmountByAdmin(cross, tt.inputAmount)
				})
			} else if tt.expectedHasAbort {
				uassert.AbortsWithMessage(t, tt.expectedAbortMsg, func() {
					SetDepositGnsAmountByAdmin(cross, tt.inputAmount)
				})
			} else {
				SetDepositGnsAmountByAdmin(cross, tt.inputAmount)
			}

			// then
			actual := GetDepositGnsAmount()
			uassert.Equal(t, tt.expectedAmount, actual)
		})
	}
}

func TestSetDepositGnsAmount(t *testing.T) {
	tests := []struct {
		name             string
		caller           std.Address
		inputAmount      int64
		expectedAmount   int64
		expectedHasAbort bool
		expectedAbortMsg string
		expectedHasPanic bool
		expectedPanicMsg string
	}{
		{
			name:             "set deposit success by governance",
			caller:           testutils.TestAddress("governance"),
			inputAmount:      3_000_000_000,
			expectedAmount:   3_000_000_000,
			expectedHasAbort: false,
			expectedAbortMsg: "",
			expectedHasPanic: false,
			expectedPanicMsg: "",
		},
		{
			name:             "set deposit failed by non governance",
			caller:           testutils.TestAddress("non-governance"),
			inputAmount:      3_000_000_000,
			expectedAmount:   0,
			expectedHasPanic: false,
			expectedPanicMsg: "",
			expectedHasAbort: true,
			expectedAbortMsg: "caller(g1dehkutt8damx2unwv9hxxe2lta047h6l9v0nxc) has no permission",
		},
		{
			name:             "set deposit amount failed by unauthorized caller",
			caller:           testutils.TestAddress("unauthorized"),
			inputAmount:      3_000_000_000,
			expectedAmount:   0,
			expectedHasPanic: false,
			expectedPanicMsg: "",
			expectedHasAbort: true,
			expectedAbortMsg: "caller(g1w4hxzat5dphhy6t6v4j97h6lta047h6l0wjddg) has no permission",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			depositGnsAmount = 0
			setupExternalDepositFeeGovernancePermission(t)

			testing.SetOriginCaller(tt.caller)

			// when
			if tt.expectedHasPanic {
				uassert.PanicsWithMessage(t, tt.expectedPanicMsg, func() {
					SetDepositGnsAmount(cross, tt.inputAmount)
				})
			} else if tt.expectedHasAbort {
				uassert.AbortsWithMessage(t, tt.expectedAbortMsg, func() {
					SetDepositGnsAmount(cross, tt.inputAmount)
				})
			} else {
				SetDepositGnsAmount(cross, tt.inputAmount)
			}

			// then
			actual := GetDepositGnsAmount()
			uassert.Equal(t, tt.expectedAmount, actual)
		})
	}
}

func TestSetMinimumRewardAmount(t *testing.T) {
	tests := []struct {
		name                        string
		caller                      std.Address
		inputMinimumRewardAmount    int64
		expectedMinimumRewardAmount int64
		expectedHasAbort            bool
		expectedAbortMsg            string
		expectedHasPanic            bool
		expectedPanicMsg            string
	}{
		{
			name:                        "set minimum reward amount success by governance",
			caller:                      testutils.TestAddress("governance"),
			inputMinimumRewardAmount:    1_000_000_000,
			expectedMinimumRewardAmount: 1_000_000_000,
			expectedHasAbort:            false,
			expectedAbortMsg:            "",
			expectedHasPanic:            false,
			expectedPanicMsg:            "",
		},
		{
			name:                        "set minimum reward amount failed by non governance",
			caller:                      testutils.TestAddress("non-governance"),
			inputMinimumRewardAmount:    1_000_000_000,
			expectedMinimumRewardAmount: 0,
			expectedHasPanic:            false,
			expectedPanicMsg:            "",
			expectedHasAbort:            true,
			expectedAbortMsg:            "caller(g1dehkutt8damx2unwv9hxxe2lta047h6l9v0nxc) has no permission",
		},
		{
			name:                        "set minimum reward amount failed by unauthorized caller",
			caller:                      testutils.TestAddress("unauthorized"),
			inputMinimumRewardAmount:    1_000_000_000,
			expectedMinimumRewardAmount: 0,
			expectedHasPanic:            false,
			expectedPanicMsg:            "",
			expectedHasAbort:            true,
			expectedAbortMsg:            "caller(g1w4hxzat5dphhy6t6v4j97h6lta047h6l0wjddg) has no permission",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			minimumRewardAmount = 0
			setupExternalDepositFeeGovernancePermission(t)

			testing.SetOriginCaller(tt.caller)

			// when
			if tt.expectedHasPanic {
				uassert.PanicsWithMessage(t, tt.expectedPanicMsg, func() {
					SetMinimumRewardAmount(cross, tt.inputMinimumRewardAmount)
				})
			} else if tt.expectedHasAbort {
				uassert.AbortsWithMessage(t, tt.expectedAbortMsg, func() {
					SetMinimumRewardAmount(cross, tt.inputMinimumRewardAmount)
				})
			} else {
				SetMinimumRewardAmount(cross, tt.inputMinimumRewardAmount)
			}

			// then
			actual := GetMinimumRewardAmount()
			uassert.Equal(t, tt.expectedMinimumRewardAmount, actual)
		})
	}
}

func setupExternalDepositFeeAdminPermission(t *testing.T) {
	t.Helper()

	addr := testutils.TestAddress("admin")
	setupExternalDepositFeePermission(t, addr, "admin")
}

func setupExternalDepositFeeGovernancePermission(t *testing.T) {
	t.Helper()

	addr := testutils.TestAddress("governance")
	setupExternalDepositFeePermission(t, addr, "governance")
}

func setupExternalDepositFeePermission(t *testing.T, caller std.Address, permissionName string) {
	t.Helper()

	testing.SetOriginCaller(caller)
	testing.SetRealm(std.NewUserRealm(caller))

	globalManager := rbac.GlobalManager()
	globalManager.UpdatePermission(permissionName, "access", func(addr std.Address) error {
		if addr != caller {
			return ufmt.Errorf("caller is not %s", addr)
		}

		return nil
	})
}
