package common

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"

	"gno.land/p/gnoswap/consts"
)

var (
	addr01 = testutils.TestAddress("addr01")
	addr02 = testutils.TestAddress("addr02")
)

func TestAssertCaller(t *testing.T) {
	t.Run("same caller", func(t *testing.T) {
		uassert.NoError(t, AssertCaller(addr01, addr01))
	})

	t.Run("different caller", func(t *testing.T) {
		uassert.Error(t, AssertCaller(addr01, addr02))
	})
}

func TestSatisfyCond(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		uassert.NoError(t, SatisfyCond(true))
	})

	t.Run("false", func(t *testing.T) {
		uassert.Error(t, SatisfyCond(false))
	})
}

func TestAdminOnly(t *testing.T) {
	t.Run("caller is admin", func(t *testing.T) {
		uassert.NoError(t, AdminOnly(consts.ADMIN))
	})

	t.Run("caller is not admin", func(t *testing.T) {
		uassert.Error(t, AdminOnly(addr01))
	})
}

func TestGovernanceOnly(t *testing.T) {
	t.Run("caller is governance", func(t *testing.T) {
		uassert.NoError(t, GovernanceOnly(consts.GOV_GOVERNANCE_ADDR))
	})

	t.Run("caller is not governance", func(t *testing.T) {
		uassert.Error(t, GovernanceOnly(addr01))
	})
}

func TestGovStakerOnly(t *testing.T) {
	t.Run("caller is gov staker", func(t *testing.T) {
		uassert.NoError(t, GovStakerOnly(consts.GOV_STAKER_ADDR))
	})

	t.Run("caller is not gov staker", func(t *testing.T) {
		uassert.Error(t, GovStakerOnly(addr01))
	})
}

func TestRouterOnly(t *testing.T) {
	t.Run("caller is router", func(t *testing.T) {
		uassert.NoError(t, RouterOnly(consts.ROUTER_ADDR))
	})

	t.Run("caller is not router", func(t *testing.T) {
		uassert.Error(t, RouterOnly(addr01))
	})
}

func TestPositionOnly(t *testing.T) {
	t.Run("caller is position", func(t *testing.T) {
		uassert.NoError(t, PositionOnly(consts.POSITION_ADDR))
	})

	t.Run("caller is not position", func(t *testing.T) {
		uassert.Error(t, PositionOnly(addr01))
	})
}

func TestStakerOnly(t *testing.T) {
	t.Run("caller is staker", func(t *testing.T) {
		uassert.NoError(t, StakerOnly(consts.STAKER_ADDR))
	})

	t.Run("caller is not staker", func(t *testing.T) {
		uassert.Error(t, StakerOnly(addr01))
	})
}

func TestLaunchpadOnly(t *testing.T) {
	t.Run("caller is launchpad", func(t *testing.T) {
		uassert.NoError(t, LaunchpadOnly(consts.LAUNCHPAD_ADDR))
	})

	t.Run("caller is not launchpad", func(t *testing.T) {
		uassert.Error(t, LaunchpadOnly(addr01))
	})
}

func TestEmissionOnly(t *testing.T) {
	t.Run("caller is emission", func(t *testing.T) {
		uassert.NoError(t, EmissionOnly(consts.EMISSION_ADDR))
	})

	t.Run("caller is not emission", func(t *testing.T) {
		uassert.Error(t, EmissionOnly(addr01))
	})
}

func TestUserOnly(t *testing.T) {
	t.Run("caller is user", func(t *testing.T) {
		uassert.NoError(t, UserOnly(std.NewUserRealm(addr01)))
	})

	t.Run("caller is not user", func(t *testing.T) {
		uassert.Error(t, UserOnly(std.NewCodeRealm("gno.land/r/realm")))
	})
}
