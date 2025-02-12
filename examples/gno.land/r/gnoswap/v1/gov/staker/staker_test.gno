package staker

import (
	"std"
	"strings"
	"testing"
	"time"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/gns"
	"gno.land/r/gnoswap/v1/gov/xgns"
)

func TestDelegate(t *testing.T) {
	std.TestSetOrigCaller(consts.ADMIN)
	SetRunning(true)

	std.TestSetRealm(userRealm)
	{
		std.TestSetRealm(std.NewCodeRealm(consts.EMISSION_PATH))
		std.TestSkipHeights(100)
		gns.MintGns(userRealm.Addr()) // 2M gns

		std.TestSetRealm(userRealm)
		to := makeFakeAddress("validator_1")
		amount := uint64(1_000_000)

		gns.Approve(consts.GOV_STAKER_ADDR, amount)
		Delegate(to, amount)

		minted := xgns.BalanceOf(userRealm.Addr())
		if minted != amount {
			t.Errorf("Delegate minted xGNS = %d, want %d", minted, amount)
		}
		// verify "delegate" avl state, or any event logs, etc.
	}

	// 4) below minimum
	{
		to := makeFakeAddress("validator_2")
		amount := uint64(999_999)

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic when delegating below minimumAmount")
			}
		}()
		Delegate(to, amount)
	}

	// 5) invalid to address
	{
		to := invalidAddr
		amount := uint64(1_000_000)

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic with invalid 'to' address")
			}
		}()
		Delegate(to, amount)
	}

	// 6) not enough GNS balance user
	{
		// user GNS = 0 now (they delegated all away above)
		to := makeFakeAddress("validator_3")
		amount := uint64(2_000_000)

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic due to not enough GNS balance in user address")
			}
		}()
		Delegate(to, amount)
	}

	// 7) running = false => optional check
	{
		std.TestSetOrigCaller(consts.ADMIN)
		SetRunning(false)
		// delegate?
		// depending on code logic, might skip cleanDelegationStatHistory() or do something else
		// if there's logic that forbids delegation when not running, test that
		// For now, assume it still works but doesn't clean.
		// Restore running to true for subsequent tests.
		std.TestSetOrigCaller(consts.ADMIN)
		SetRunning(true)
	}

	{
		to := makeFakeAddress("validator_2")
		amount := uint64(1999_999)

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic when delegating below minimumAmount")
			}
		}()
		Delegate(to, amount)
	}

	{
		to := makeFakeAddress("validator_2")
		amount := uint64(2000_000)

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic when delegating below minimumAmount")
			}
		}()
		Delegate(to, amount)
	}
}

func TestDelegate_Boundary_Values(t *testing.T) {
	tests := []struct {
		name        string
		to          std.Address
		amount      uint64
		expectPanic bool
		panicMsg    string
	}{
		{
			name:        "delegate zero amount",
			to:          makeFakeAddress("validator_1"),
			amount:      0,
			expectPanic: true,
			panicMsg:    "delegation amount cannot be 0",
		},
		{
			name:        "delegate max uint64",
			to:          makeFakeAddress("validator_1"),
			amount:      ^uint64(0), // max uint64
			expectPanic: true,
			panicMsg:    "[GNOSWAP-GOV_STAKER-004] invalid amount || amount must be multiple of 1000000",
		},
		{
			name:        "delegate near max uint64",
			to:          makeFakeAddress("validator_1"),
			amount:      ^uint64(0) - 1000,
			expectPanic: true,
			panicMsg:    "[GNOSWAP-GOV_STAKER-004] invalid amount || amount must be multiple of 1000000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			std.TestSetOrigCaller(consts.ADMIN)
			SetRunning(true)
			std.TestSetRealm(userRealm)

			if tt.expectPanic {
				defer func() {
					r := recover()
					if r == nil {
						t.Errorf("Expected panic but got none")
					}
					if tt.panicMsg != "" && r != nil {
						if msg, ok := r.(string); !ok || !strings.Contains(msg, tt.panicMsg) {
							t.Errorf("Expected panic message containing '%s', got '%v'", tt.panicMsg, r)
						}
					}
				}()
			}

			Delegate(tt.to, tt.amount)
		})
	}
}

/*
func TestEmptyRewardPool(t *testing.T) {
	tests := []struct {
		name           string
		setupFn        func()
		expectPanic    bool
		expectPanicMsg string
		checkFn        func(t *testing.T)
	}{
		{
			name: "collect with empty reward pool",
			setupFn: func() {
				//userEmissionReward.Remove(userRealm.Addr().String())
			},
			expectPanic: false,
			checkFn: func(t *testing.T) {
				if v, exists := userEmissionReward.Get(userRealm.Addr().String()); exists {
					t.Errorf("Expected userEmissionReward to be removed, but got %d", v.(uint64))
				}
			},
		},
		{
			name: "collect with empty protocol fee rewards",
			setupFn: func() {
				//userProtocolFeeReward.Remove(userRealm.Addr().String())
			},
			expectPanic: false,
			checkFn: func(t *testing.T) {
				if v, exists := userProtocolFeeReward.Get(userRealm.Addr().String()); exists {
					t.Errorf("Expected userProtocolFeeReward to be removed, but got %d", v.(uint64))
				}
			},
		},
		{
			name: "collect with empty staker GNS balance",
			setupFn: func() {
				std.TestSetRealm(std.NewCodeRealm(consts.GOV_STAKER_PATH))
				userEmissionReward.Set(userRealm.Addr().String(), uint64(1000))
			},
			expectPanic:    true,
			expectPanicMsg: "not enough GNS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "collect with empty staker GNS balance" {
				t.Skip("run this test separately")
			}
			std.TestSetRealm(userRealm)

			if tt.setupFn != nil {
				tt.setupFn()
			}

			if tt.expectPanic {
				defer func() {
					r := recover()
					if r == nil {
						t.Errorf("Expected panic but got none")
					}
					if tt.expectPanicMsg != "" && r != nil {
						if msg, ok := r.(string); !ok || !strings.Contains(msg, tt.expectPanicMsg) {
							t.Errorf("Expected panic message containing '%s', got '%v'", tt.expectPanicMsg, r)
						}
					}
				}()
			}

			CollectReward()

			if tt.checkFn != nil {
				tt.checkFn(t)
			}
		})
	}
}
*/
/*
func TestProtocolFee(t *testing.T) {
	tests := []struct {
		name        string
		setupFn     func()
		expectPanic bool
		panicMsg    string
		checkFn     func(t *testing.T)
	}{
		{
			name: "collect max uint64 protocol fee",
			setupFn: func() {
				//tree := avl.NewTree()
				//tree.Set(consts.WUGNOT_PATH, uint64(^uint64(0)))
				//userProtocolFeeReward.Set(userRealm.Addr().String(), tree)
				currentProtocolFeeBalance[consts.WUGNOT_PATH] = ^uint64(0)
			},
			expectPanic: true,
			panicMsg:    "insufficient balance",
		},
		{
			name: "collect multiple empty token balances",
			setupFn: func() {
				//tree := avl.NewTree()
				//tree.Set(consts.WUGNOT_PATH, uint64(1000))
				//tree.Set("some/other/token", uint64(2000))
				//userProtocolFeeReward.Set(userRealm.Addr().String(), tree)
				currentProtocolFeeBalance[consts.WUGNOT_PATH] = 1000
				currentProtocolFeeBalance["some/other/token"] = 2000
			},
			expectPanic: true,
			panicMsg:    "insufficient balance",
		},
		{
			name: "collect with zero amounts",
			setupFn: func() {
				//tree := avl.NewTree()
				//tree.Set(consts.WUGNOT_PATH, uint64(0))
				//tree.Set("some/other/token", uint64(0))
				//userProtocolFeeReward.Set(userRealm.Addr().String(), tree)
				currentProtocolFeeBalance[consts.WUGNOT_PATH] = 0
				currentProtocolFeeBalance["some/other/token"] = 0
			},
			expectPanic: false,
			checkFn: func(t *testing.T) {
				val, exists := userProtocolFeeReward.Get(userRealm.Addr().String())
				if !exists {
					t.Errorf("Expected protocol fee reward entry to exist")
					return
				}
				tree := val.(*avl.Tree)
				wugnotAmt, _ := tree.Get(consts.WUGNOT_PATH)
				if wugnotAmt.(uint64) != 0 {
					t.Errorf("Expected WUGNOT amount to be 0, got %d", wugnotAmt)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			std.TestSetRealm(userRealm)

			if tt.setupFn != nil {
				tt.setupFn()
			}

			if tt.expectPanic {
				defer func() {
					r := recover()
					if r == nil {
						t.Errorf("Expected panic but got none")
					}
					if tt.panicMsg != "" && r != nil {
						if msg, ok := r.(string); !ok || !strings.Contains(strings.ToLower(msg), strings.ToLower(tt.panicMsg)) {
							t.Errorf("Expected panic message containing '%s', got '%v'", tt.panicMsg, r)
						}
					}
				}()
			}

			CollectReward()

			if tt.checkFn != nil {
				tt.checkFn(t)
			}
		})
	}
}
*/
func TestRedelegate(t *testing.T) {
	std.TestSetRealm(std.NewCodeRealm(consts.EMISSION_PATH))
	std.TestSkipHeights(100)
	gns.MintGns(userRealm.Addr()) // 2M gns

	// user has xGNS from previous test (some minted)
	// try re-delegating
	std.TestSetRealm(userRealm)
	from := userRealm.Addr()
	to := makeFakeAddress("validator_1")
	amount := uint64(1_000_000)

	std.TestSetOrigCaller(userRealm.Addr())
	gns.Approve(consts.GOV_STAKER_ADDR, amount)
	Delegate(from, amount)

	std.TestSetOrigCaller(userRealm.Addr())
	gns.Approve(consts.GOV_STAKER_ADDR, amount)
	Redelegate(from, to, amount)

	// check data: user xGNS must remain the same, but from-> to shift in staker structure?

	// not enough xGNS => panic
	{
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic if user tries to re-delegate more than xGNS balance")
			}
		}()
		Redelegate(from, to, 999_999_999)
	}
}

func TestUndelegate(t *testing.T) {
	std.TestSetRealm(std.NewCodeRealm(consts.EMISSION_PATH))
	std.TestSkipHeights(101)
	gns.MintGns(user3Realm.Addr()) // 2M gns

	std.TestSetRealm(user3Realm)
	to := makeFakeAddress("validator_1")
	amount := uint64(1_000_000)

	std.TestSetOrigCaller(user3Realm.Addr())
	gns.Approve(consts.GOV_STAKER_ADDR, amount)
	Delegate(to, amount)
	Undelegate(to, amount)

	if xgns.BalanceOf(user3Realm.Addr()) == 1_000_000 {
		t.Errorf("Expected user xGNS to be 0 after undelegating 1_000_000, got %d",
			xgns.BalanceOf(user3Realm.Addr()))
	}

	lockedList, exist := addrLockedGns.Get(user3Realm.Addr().String())
	if !exist {
		t.Errorf("Expected lockedGNS to be created after Undelegate")
	}
	locked := lockedList.([]lockedGNS)
	if len(locked) == 0 {
		t.Errorf("Expected at least 1 lockedGNS after Undelegate")
	}
	if locked[0].amount != 1_000_000 {
		t.Errorf("LockedGNS amount mismatch, got %d, want 1_000_000", locked[0].amount)
	}
	// check lockedAmount incremented?
	if lockedAmount != 1_000_000 {
		t.Errorf("lockedAmount expected 1_000_000, got %d", lockedAmount)
	}

	// below minimum => panic
	{
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic when undelegating below minimum")
			}
		}()
		Undelegate(to, 999_999) // or 999_990, etc. if min = 1_000_000
	}
}

func TestCollectUndelegatedGns(t *testing.T) {
	std.TestSetRealm(userRealm)

	// 1) no locked => expect 0
	addrLockedGns.Remove(userRealm.Addr().String()) // ensure no locked
	collected := CollectUndelegatedGns()
	if collected != 0 {
		t.Errorf("Expected 0 when no locked gns, got %d", collected)
	}

	// 2) add locked but time not passed => 0
	now := uint64(time.Now().Unix())
	locked := lockedGNS{
		amount: 100_000,
		unlock: now + SECONDS_IN_7_DAYS,
	}
	addrLockedGns.Set(userRealm.Addr().String(), []lockedGNS{locked})
	lockedAmount = 100_000

	collected = CollectUndelegatedGns()
	if collected != 0 {
		t.Errorf("Expected 0 if 7days not passed, got %d", collected)
	}
	// verify still in locked
	lockedList, exist := addrLockedGns.Get(userRealm.Addr().String())
	if !exist {
		t.Errorf("Expected lockedGNS to remain after CollectUndelegatedGns")
	}
	lkList := lockedList.([]lockedGNS)
	if len(lkList) != 1 {
		t.Errorf("Locked list should remain, but length = %d", len(lkList))
	}

	// 3) set time => unlocked
	locked.unlock = uint64(time.Now().Unix()) - 1 // forcibly make it past time
	addrLockedGns.Set(userRealm.Addr().String(), []lockedGNS{locked})

	collected = CollectUndelegatedGns()
	if collected != 100_000 {
		t.Errorf("Expected 100_000 collected, got %d", collected)
	}
	// check locked removed from the tree
	if _, exists := addrLockedGns.Get(userRealm.Addr().String()); exists {
		t.Errorf("Expected addrLockedGns key to be removed if empty after collecting all locked gns")
	}
	if lockedAmount != 0 {
		t.Errorf("lockedAmount should have been decremented to 0, got %d", lockedAmount)
	}
}

func TestCollectReward(t *testing.T) {
	std.TestSetRealm(user2Realm)
	{
		std.TestSetRealm(std.NewCodeRealm(consts.EMISSION_PATH))
		std.TestSkipHeights(100)
		gns.MintGns(consts.GOV_STAKER_ADDR)

		std.TestSetRealm(std.NewUserRealm(consts.ADMIN))
		ugnotFaucet(t, consts.GOV_STAKER_ADDR, 1_000_000)
		ugnotDeposit(t, consts.GOV_STAKER_ADDR, 1_000_000)
		ugnotFaucet(t, derivePkgAddr(wugnotPath), 1_000_000)
		ugnotFaucet(t, user2Realm.Addr(), 1_000_000)
	}

	std.TestSetRealm(user2Realm)
	user := user2Realm.Addr().String()

	rewardState.AddStake(uint64(std.GetHeight()), std.Address(user), 10, 0, make(map[string]uint64))

	// set a fake emission reward
	//userEmissionReward.Set(user, uint64(50_000))
	currentGNSBalance = 50_000
	// set a fake protocol fee reward
	//tree := avl.NewTree()
	//tree.Set(consts.WUGNOT_PATH, uint64(10_000))
	//userProtocolFeeReward.Set(user, tree)
	currentProtocolFeeBalance[consts.WUGNOT_PATH] = 10_000

	std.TestSetRealm(std.NewCodeRealm(consts.EMISSION_PATH))
	std.TestSkipHeights(100)
	gns.MintGns(user2Realm.Addr())

	std.TestSetRealm(user2Realm)
	std.TestSkipHeights(100)

	// call CollectReward
	std.TestSetOrigCaller(user2Realm.Addr())
	CollectReward()

	// expect user emissionReward = 0
	claimableAmount, protocolFeeClaimable := rewardState.CalculateReward(std.Address(user))
	if claimableAmount != 0 {
		t.Errorf("Expected userEmissionReward to be 0 after collect, got %d", claimableAmount)
	}
	if len(protocolFeeClaimable) != 0 {
		t.Errorf("Expected protocolFeeClaimable to be empty after collect, got %d", len(protocolFeeClaimable))
	}
	for _, amount := range protocolFeeClaimable {
		if amount != 0 {
			t.Errorf("Expected protocolFeeClaimable to be 0 after collect, got %d", amount)
		}
	}

	// If GOV_STAKER_ADDR had less GNS than 50_000 => we expect panic
	// can test in separate subcase
}

func TestCollectRewardFromLaunchPad(t *testing.T) {
	// set realm to LAUNCHPAD_ADDR?
	// or we do a quick scenario: if current caller != LAUNCHPAD_ADDR => panic
	// => check it with a user realm to ensure panic
	std.TestSetRealm(userRealm)
	{
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected panic, because caller is not launchpad")
			}
		}()
		CollectRewardFromLaunchPad(userRealm.Addr())
	}

	// Then set realm to a custom "launchpadRealm" whose .Addr() == consts.LAUNCHPAD_ADDR
	// => call normal and see if distribution works
}
