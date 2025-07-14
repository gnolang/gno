package community_pool

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"

	"gno.land/p/gnoswap/consts"

	"gno.land/r/gnoswap/v1/access"
	"gno.land/r/gnoswap/v1/gns"
	"gno.land/r/gnoswap/v1/halt"
)

var (
	adminAddr, _ = access.GetAddress(access.ROLE_ADMIN)
	adminRealm   = std.NewUserRealm(adminAddr)

	communityPoolAddr, _ = access.GetAddress(access.ROLE_COMMUNITY_POOL)

	govRealm = std.NewCodeRealm(consts.GOV_GOVERNANCE_PATH)

	dummyCaller   = std.NewUserRealm(testutils.TestAddress("dummyCaller"))
	dummyReceiver = testutils.TestAddress("dummyReceiver")
)

func resetCommunityPoolState(t *testing.T) {
	testing.SetRealm(adminRealm)
	halt.SetHaltByAdmin(cross, false)
	gns.Transfer(cross, communityPoolAddr, 10_000)
}

func TestTransferTokenByAdmin(t *testing.T) {
	tests := []struct {
		name        string
		setup       func()
		caller      std.Realm
		tokenPath   string
		to          std.Address
		amount      int64
		shouldPanic bool
		panicMsg    string
	}{
		{
			name: "panic if halted",
			setup: func() {
				testing.SetRealm(adminRealm)
				halt.SetHaltByAdmin(cross, true)
			},
			caller:      adminRealm,
			tokenPath:   consts.GNS_PATH,
			to:          dummyReceiver,
			amount:      1000,
			shouldPanic: true,
			panicMsg:    "[GNOSWAP-COMMUNITY_POOL-005] halted || community pool is currently halted",
		},
		{
			name: "panic if not admin",
			setup: func() {
				testing.SetRealm(adminRealm)
				halt.SetHaltByAdmin(cross, false)
			},
			caller:      dummyCaller,
			tokenPath:   consts.GNS_PATH,
			to:          dummyReceiver,
			amount:      1000,
			shouldPanic: true,
			panicMsg:    "caller(g1v36k6mtegdskcmr9wf047h6lta047h6ljmnks7) has no permission",
		},
		{
			name: "success if enough balance",
			setup: func() {
				testing.SetRealm(adminRealm)
				gns.Transfer(cross, communityPoolAddr, 10_000)
			},
			caller:    adminRealm,
			tokenPath: consts.GNS_PATH,
			to:        dummyReceiver,
			amount:    10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetCommunityPoolState(t)

			if tt.setup != nil {
				tt.setup()
			}

			if tt.caller != (std.Realm{}) {
				testing.SetRealm(tt.caller)
			}

			if tt.shouldPanic {
				uassert.AbortsWithMessage(t, tt.panicMsg, func() {
					TransferTokenByAdmin(cross, tt.tokenPath, tt.to, tt.amount)
				})
			} else {
				receiverOldBalance := gns.BalanceOf(tt.to)
				TransferTokenByAdmin(cross, tt.tokenPath, tt.to, tt.amount)
				receiverNewBalance := gns.BalanceOf(tt.to)
				uassert.Equal(t, receiverNewBalance-receiverOldBalance, tt.amount)
			}
		})
	}
}

func TestTransferToken(t *testing.T) {
	tests := []struct {
		name        string
		setup       func()
		caller      std.Realm
		tokenPath   string
		to          std.Address
		amount      int64
		shouldPanic bool
		panicMsg    string
	}{
		{
			name: "panic if halted",
			setup: func() {
				testing.SetRealm(adminRealm)
				halt.SetHaltByAdmin(cross, true)
			},
			tokenPath:   consts.GNS_PATH,
			to:          dummyReceiver,
			amount:      1000,
			shouldPanic: true,
			panicMsg:    "[GNOSWAP-COMMUNITY_POOL-005] halted || community pool is currently halted",
		},
		{
			name: "panic if not governance",
			setup: func() {
				testing.SetRealm(adminRealm)
				halt.SetHaltByAdmin(cross, false)
			},
			tokenPath:   consts.GNS_PATH,
			to:          dummyReceiver,
			amount:      1000,
			shouldPanic: true,
			panicMsg:    "caller(g100fnnlz5eh87p5hvwt8pf279lxaelm8k8md049) has no permission",
		},
		{
			name: "governance can transfer community pool token",
			setup: func() {
				testing.SetRealm(adminRealm)
				gns.Transfer(cross, communityPoolAddr, 10_000)
			},
			caller:    govRealm,
			tokenPath: consts.GNS_PATH,
			to:        dummyReceiver,
			amount:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetCommunityPoolState(t)

			if tt.setup != nil {
				tt.setup()
			}

			if tt.caller != (std.Realm{}) {
				testing.SetRealm(tt.caller)
			}

			if tt.shouldPanic {
				uassert.AbortsWithMessage(t, tt.panicMsg, func() {
					TransferToken(cross, tt.tokenPath, tt.to, tt.amount)
				})
			} else {
				receiverOldBalance := gns.BalanceOf(tt.to)
				TransferToken(cross, tt.tokenPath, tt.to, tt.amount)
				receiverNewBalance := gns.BalanceOf(tt.to)
				uassert.Equal(t, receiverNewBalance-receiverOldBalance, tt.amount)
			}
		})
	}
}
