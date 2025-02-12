package community_pool

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/common"

	"gno.land/r/gnoswap/v1/gns"
)

var (
	adminAddr  = consts.ADMIN
	adminRealm = std.NewUserRealm(adminAddr)

	govRealm = std.NewCodeRealm(consts.GOV_GOVERNANCE_PATH)

	dummyCaller   = std.NewUserRealm(testutils.TestAddress("dummyCaller"))
	dummyReceiver = testutils.TestAddress("dummyReceiver")
)

func TestTransferTokenByAdmin(t *testing.T) {
	tests := []struct {
		name        string
		setup       func()
		caller      std.Realm
		tokenPath   string
		to          std.Address
		amount      uint64
		shouldPanic bool
		panicMsg    string
	}{
		{
			name: "panic if halted",
			setup: func() {
				std.TestSetRealm(adminRealm)
				common.SetHaltByAdmin(true)
			},
			tokenPath:   consts.GNS_PATH,
			to:          dummyReceiver,
			amount:      1000,
			shouldPanic: true,
			panicMsg:    "[GNOSWAP-COMMON-002] halted || GnoSwap is halted",
		},
		{
			name: "panic if not admin",
			setup: func() {
				std.TestSetRealm(adminRealm)
				common.SetHaltByAdmin(false)
			},
			tokenPath:   consts.GNS_PATH,
			to:          dummyReceiver,
			amount:      1000,
			shouldPanic: true,
			panicMsg:    "caller(g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm) has no permission",
		},
		{
			name: "panic if not enough balance",
			setup: func() {
				std.TestSetRealm(adminRealm)
				gns.Transfer(consts.COMMUNITY_POOL_ADDR, 10_000)
			},
			caller:      adminRealm,
			tokenPath:   consts.GNS_PATH,
			to:          dummyReceiver,
			amount:      10_001,
			shouldPanic: true,
			panicMsg:    "insufficient balance",
		},
		{
			name:      "success if enough balance",
			caller:    adminRealm,
			tokenPath: consts.GNS_PATH,
			to:        dummyReceiver,
			amount:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			if tt.caller != (std.Realm{}) {
				std.TestSetRealm(tt.caller)
			}

			if tt.shouldPanic {
				uassert.PanicsWithMessage(t, tt.panicMsg, func() {
					TransferTokenByAdmin(tt.tokenPath, tt.to, tt.amount)
				})
			} else {
				receiverOldBalance := gns.BalanceOf(tt.to)
				TransferTokenByAdmin(tt.tokenPath, tt.to, tt.amount)
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
		amount      uint64
		shouldPanic bool
		panicMsg    string
	}{
		{
			name: "panic if halted",
			setup: func() {
				std.TestSetRealm(adminRealm)
				common.SetHaltByAdmin(true)
			},
			tokenPath:   consts.GNS_PATH,
			to:          dummyReceiver,
			amount:      1000,
			shouldPanic: true,
			panicMsg:    "[GNOSWAP-COMMON-002] halted || GnoSwap is halted",
		},
		{
			name: "panic if not governance",
			setup: func() {
				std.TestSetRealm(adminRealm)
				common.SetHaltByAdmin(false)
			},
			tokenPath:   consts.GNS_PATH,
			to:          dummyReceiver,
			amount:      1000,
			shouldPanic: true,
			panicMsg:    "caller(g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm) has no permission",
		},
		{
			name: "governance can't transfer community pool token",
			setup: func() {
				std.TestSetRealm(adminRealm)
			},
			caller:      govRealm,
			tokenPath:   consts.GNS_PATH,
			to:          dummyReceiver,
			amount:      10_001,
			shouldPanic: true,
			panicMsg:    "insufficient balance",
		},
		{
			name:      "governance can transfer community pool token",
			caller:    govRealm,
			tokenPath: consts.GNS_PATH,
			to:        dummyReceiver,
			amount:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}

			if tt.caller != (std.Realm{}) {
				std.TestSetRealm(tt.caller)
			}

			if tt.shouldPanic {
				uassert.PanicsWithMessage(t, tt.panicMsg, func() {
					TransferToken(tt.tokenPath, tt.to, tt.amount)
				})
			} else {
				receiverOldBalance := gns.BalanceOf(tt.to)
				TransferToken(tt.tokenPath, tt.to, tt.amount)
				receiverNewBalance := gns.BalanceOf(tt.to)
				uassert.Equal(t, receiverNewBalance-receiverOldBalance, tt.amount)
			}
		})
	}
}

func TestPrivateTransferToken(t *testing.T) {
	tests := []struct {
		tokenPath   string
		to          std.Address
		amount      uint64
		shouldPanic bool
		panicMsg    string
	}{
		{
			tokenPath:   "not_registered_token",
			to:          dummyReceiver,
			amount:      1,
			shouldPanic: true,
			panicMsg:    "unknown token: not_registered_token",
		},
		{
			tokenPath:   consts.GNS_PATH,
			to:          dummyReceiver,
			amount:      1,
			shouldPanic: false,
		},
	}

	for _, tt := range tests {
		if tt.shouldPanic {
			uassert.PanicsWithMessage(t, tt.panicMsg, func() {
				transferToken(tt.tokenPath, tt.to, tt.amount)
			})
		} else {
			transferToken(tt.tokenPath, tt.to, tt.amount)
		}
	}
}
