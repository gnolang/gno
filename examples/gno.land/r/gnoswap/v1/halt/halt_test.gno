package halt

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"

	"gno.land/p/gnoswap/consts"
	phalt "gno.land/p/gnoswap/halt"
)

var (
	adminRealm = std.NewUserRealm(consts.ADMIN)
	govRealm   = std.NewCodeRealm(consts.GOV_GOVERNANCE_PATH)
	userRealm  = std.NewCodeRealm("gno.land/r/demo/users") // for testing non-governance, non-admin caller
)

func TestHalts(t *testing.T) {
	t.Run("GetHalt() initial value", func(t *testing.T) {
		// Initially we're in TestnetSafeMode which should report as halted
		uassert.True(t, GetHalt())
	})

	t.Run("Operation specific checks in TestnetSafeMode", func(t *testing.T) {
		// Governance operations should be allowed in TestnetSafeMode
		err := IsHalted(OpTypeGovernance)
		uassert.NoError(t, err)

		// Withdrawals should be disabled in TestnetSafeMode
		err = IsHalted(OpTypeWithdraw)
		uassert.Error(t, err)

		// Swaps should be disabled in TestnetSafeMode
		err = IsHalted(OpTypeSwap)
		uassert.Error(t, err)
	})
}

func TestSetHaltLevelByAdmin(t *testing.T) {
	t.Run("with non-admin privilege, panics", func(t *testing.T) {
		std.TestSetRealm(userRealm)
		err := SetHaltLevelByAdmin(phalt.LvNoHalt)
		uassert.Error(t, err)
	})

	t.Run("with admin privilege, success", func(t *testing.T) {
		std.TestSetRealm(adminRealm)

		// Initially we're in TestnetSafeMode
		uassert.True(t, GetHalt())
		if GetCurrentHaltLevel() != LvTestnetSafeMode {
			t.Fatalf("GetCurrentHaltLevel() = %d, want %d", GetCurrentHaltLevel(), LvTestnetSafeMode)
		}

		// Change to NoHalt
		err := SetHaltLevelByAdmin(phalt.LvNoHalt)
		uassert.NoError(t, err)
		if GetCurrentHaltLevel() != phalt.LvNoHalt {
			t.Fatalf("GetCurrentHaltLevel() = %d, want %d", GetCurrentHaltLevel(), phalt.LvNoHalt)
		}

		// Back to CompleteHalt
		err = SetHaltLevelByAdmin(phalt.LvCompleteHalt)
		uassert.NoError(t, err)
		if GetCurrentHaltLevel() != phalt.LvCompleteHalt {
			t.Fatalf("GetCurrentHaltLevel() = %d, want %d", GetCurrentHaltLevel(), phalt.LvCompleteHalt)
		}
	})
}

func TestSetHalt(t *testing.T) {
	t.Run("with admin (non-governance) privilege, should fail", func(t *testing.T) {
		std.TestSetRealm(adminRealm)
		err := SetHalt(false)
		uassert.Error(t, err)
	})

	t.Run("with governance privilege, success", func(t *testing.T) {
		std.TestSetRealm(govRealm)

		// Currently in CompleteHalt from previous test
		uassert.True(t, GetHalt())

		err := SetHalt(false) // try to set to NoHalt
		uassert.NoError(t, err)
		if GetCurrentHaltLevel() != phalt.LvNoHalt {
			t.Fatalf("GetCurrentHaltLevel() = %d, want %d", GetCurrentHaltLevel(), phalt.LvNoHalt)
		}
	})
}

type haltTestCases struct {
	opType    phalt.OpType
	wantError bool
}

func TestEnableOperationsForTestnet(t *testing.T) {
	// Setup test with admin privileges
	std.TestSetRealm(adminRealm)

	// First set to TestnetSafeMode
	err := SetHaltLevelByAdmin(LvTestnetSafeMode)
	uassert.NoError(t, err)
	if GetCurrentHaltLevel() != LvTestnetSafeMode {
		t.Fatalf("GetCurrentHaltLevel() = %d, want %d", GetCurrentHaltLevel(), LvTestnetSafeMode)
	}

	// Verify initially all operations except governance are disabled
	isHaltedTests := []haltTestCases{
		{phalt.OpTypeWithdraw, true},
		{phalt.OpTypeSwap, true},
		{phalt.OpTypeLiquidity, true},
		{phalt.OpTypeGovernance, false},
	}

	for _, tt := range isHaltedTests {
		err := IsHalted(tt.opType)
		if tt.wantError {
			uassert.Error(t, err)
			return
		}
		uassert.NoError(t, err)
	}

	t.Run("Enable withdrawals", func(t *testing.T) {
		err := EnableWithdrawalsForTestnet()
		uassert.NoError(t, err)
		tests := []haltTestCases{
			{phalt.OpTypeWithdraw, false}, // now enabled
			{phalt.OpTypeSwap, true},      // still disabled
			{phalt.OpTypeLiquidity, true}, // still disabled
		}

		for _, tt := range tests {
			err := IsHalted(tt.opType)
			if tt.wantError {
				uassert.Error(t, err)
				return
			}
			uassert.NoError(t, err)
		}
	})

	t.Run("Enable swaps", func(t *testing.T) {
		err := EnableSwapsForTestnet()
		uassert.NoError(t, err)

		tests := []haltTestCases{
			{phalt.OpTypeSwap, false},     // now enabled
			{phalt.OpTypeWithdraw, false}, // now enabled
			{phalt.OpTypeLiquidity, true}, // still disabled
		}

		for _, tt := range tests {
			err := IsHalted(tt.opType)
			if tt.wantError {
				uassert.Error(t, err)
				return
			}
			uassert.NoError(t, err)
		}
	})

	t.Run("Enable liquidity", func(t *testing.T) {
		err := EnableLiquidityForTestnet()
		uassert.NoError(t, err)

		tests := []haltTestCases{
			{phalt.OpTypeLiquidity, false},  // now enabled
			{phalt.OpTypeSwap, false},       // now enabled
			{phalt.OpTypeWithdraw, false},   // now enabled
			{phalt.OpTypeGovernance, false}, // now enabled
		}

		for _, tt := range tests {
			err := IsHalted(tt.opType)
			if tt.wantError {
				uassert.Error(t, err)
				return
			}
			uassert.NoError(t, err)
		}

		// Despite all operations being enabled, we're still in TestnetSafeMode
		// So `GetHalt()` should still return true.
		uassert.True(t, GetHalt())
	})

	t.Run("Move to NoHalt after testing", func(t *testing.T) {
		// Finally transition to NoHalt
		err := SetHaltLevelByAdmin(phalt.LvNoHalt)
		uassert.NoError(t, err)

		// Now GetHalt() should return false
		uassert.False(t, GetHalt())

		if GetCurrentHaltLevel() != phalt.LvNoHalt {
			t.Fatalf("GetCurrentHaltLevel() = %d, want %d", GetCurrentHaltLevel(), phalt.LvNoHalt)
		}
	})
}
