package staker

import (
	"std"
	"strings"
	"testing"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/gns"
	"gno.land/r/gnoswap/v1/gov/xgns"
	"gno.land/r/gnoswap/v1/rbac"
)

func TestDelegate(t *testing.T) {
	tests := []struct {
		name                 string
		inputToAddress       std.Address
		inputAmount          uint64
		expectedMintedAmount uint64
		expectedHasAbort     bool
		expectedAbortMsg     string
		expectedHasPanic     bool
		expectedPanicMsg     string
	}{
		{
			name:                 "delegate success",
			inputToAddress:       testutils.TestAddress("validator_1"),
			inputAmount:          1_000_000,
			expectedMintedAmount: 1_000_000,
			expectedHasAbort:     false,
			expectedAbortMsg:     "",
		},
		{
			name:                 "delegate below minimum amount",
			inputToAddress:       testutils.TestAddress("validator_1"),
			inputAmount:          999_999,
			expectedMintedAmount: 0,
			expectedHasAbort:     true,
			expectedAbortMsg:     "[GNOSWAP-GOV_STAKER-011] can not delegate less than minimum amount || minimum amount to delegate is 1000000 (requested:999999)",
		},
		{
			name:                 "delegate invalid amount",
			inputToAddress:       testutils.TestAddress("validator_1"),
			inputAmount:          1_000_001,
			expectedMintedAmount: 0,
			expectedHasAbort:     true,
			expectedAbortMsg:     "[GNOSWAP-GOV_STAKER-004] invalid amount || amount must be multiple of 1000000",
		},
		{
			name:                 "delegate invalid to address",
			inputToAddress:       "invalid_address",
			inputAmount:          1_000_000,
			expectedMintedAmount: 0,
			expectedHasAbort:     true,
			expectedAbortMsg:     "[GNOSWAP-GOV_STAKER-008] invalid address || invalid address invalid_address to delegate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			user := testutils.TestAddress("user")

			setupStakerRunning(t, true)

			func() {
				testing.SetRealm(std.NewCodeRealm(consts.EMISSION_PATH))
				testing.SkipHeights(100)

				gns.MintGns(cross, user) // 2M gns
			}()

			testing.SetOriginCaller(user)
			testing.SetRealm(std.NewUserRealm(user))

			// when
			gns.Approve(cross, govStaker, int64(tt.inputAmount))

			if tt.expectedHasPanic {
				uassert.PanicsWithMessage(t, tt.expectedPanicMsg, func() {
					Delegate(cross, tt.inputToAddress, tt.inputAmount, "")
				})
				return
			} else if tt.expectedHasAbort {
				uassert.AbortsWithMessage(t, tt.expectedAbortMsg, func() {
					Delegate(cross, tt.inputToAddress, tt.inputAmount, "")
				})
				return
			} else {
				Delegate(cross, tt.inputToAddress, tt.inputAmount, "")
			}

			// then
			mintedAmount := xgns.BalanceOf(user)
			uassert.Equal(t, uint64(mintedAmount), tt.expectedMintedAmount)
		})
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
			panicMsg:    "[GNOSWAP-GOV_STAKER-011] can not delegate less than minimum amount || minimum amount to delegate is 1000000 (requested:0)",
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
			setupStakerRunning(t, true)
			testing.SetRealm(userRealm)
			func(cur realm) {
				testing.SetRealm(std.NewCodeRealm("gno.land/r/gnoswap/v1/gov/staker"))
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

				Delegate(cur, tt.to, tt.amount, "")
			}(cross)
		})
	}
}

func TestRedelegate(t *testing.T) {
	tests := []struct {
		name                    string
		inputFromAddress        std.Address
		inputToAddress          std.Address
		inputAmount             uint64
		expectedDelegatedAmount uint64
		expectedHasPanic        bool
		expectedPanicMsg        string
		expectedHasAbort        bool
		expectedAbortMsg        string
	}{
		{
			name:                    "redelegate success",
			inputFromAddress:        testutils.TestAddress("validator_1"),
			inputToAddress:          testutils.TestAddress("validator_2"),
			inputAmount:             1_000_000,
			expectedDelegatedAmount: 1_000_000,
			expectedHasPanic:        false,
			expectedPanicMsg:        "",
			expectedHasAbort:        false,
			expectedAbortMsg:        "",
		},
		{
			name:                    "redelegate not enough balance",
			inputFromAddress:        testutils.TestAddress("validator_1"),
			inputToAddress:          testutils.TestAddress("validator_3"),
			inputAmount:             999_999,
			expectedDelegatedAmount: 0,
			expectedHasPanic:        false,
			expectedPanicMsg:        "",
			expectedHasAbort:        true,
			expectedAbortMsg:        "[GNOSWAP-GOV_STAKER-011] can not delegate less than minimum amount || minimum amount to redelegate is 1000000 (requested:999999)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			user := testutils.TestAddress("user")

			setupStakerRunning(t, true)

			func() {
				testing.SetRealm(std.NewCodeRealm(consts.EMISSION_PATH))
				testing.SkipHeights(100)

				gns.MintGns(cross, user) // 2M gns
			}()

			testing.SetOriginCaller(user)
			testing.SetRealm(std.NewUserRealm(user))

			delegatedAmount := uint64(1_000_000)
			gns.Approve(cross, govStaker, int64(delegatedAmount))
			Delegate(cross, tt.inputFromAddress, delegatedAmount, "")

			// when
			redelgateFn := func() {
				Redelegate(cross, tt.inputFromAddress, tt.inputToAddress, tt.inputAmount)
			}

			if tt.expectedHasPanic {
				uassert.PanicsWithMessage(t, tt.expectedPanicMsg, redelgateFn)
				return
			} else if tt.expectedHasAbort {
				uassert.AbortsWithMessage(t, tt.expectedAbortMsg, redelgateFn)
				return
			} else {
				redelgateFn()
			}

			// then
			toAddressDelegatedAmount := GetTotalDelegatedTo(tt.inputToAddress) // XXX: xgns balance not equals to delegated amount.
			uassert.Equal(t, tt.expectedDelegatedAmount, uint64(toAddressDelegatedAmount))
		})
	}
}

func TestUndelegate(t *testing.T) {
	setupStakerRunning(t, true)

	testing.SetRealm(std.NewCodeRealm(consts.EMISSION_PATH))
	func() {
		testing.SkipHeights(101)
		gns.MintGns(cross, user3Realm.Address()) // 2M gns

		testing.SetRealm(user3Realm)
		func() {
			testing.SetRealm(std.NewCodeRealm("gno.land/r/gnoswap/v1/gov/staker"))
			// to := makeFakeAddress("validator_1")
			amount := uint64(1_000_000)

			testing.SetRealm(user3Realm)
			testing.SetRealm(std.NewCodeRealm("gno.land/r/gnoswap/v1/gov/staker"))
			gns.Approve(cross, govStaker, int64(amount))
			// Delegate(to, amount, "")
			// Undelegate(to, amount)
		}()

		// if xgns.BalanceOf(user3Realm.Address()) == 1_000_000 {
		// 	t.Errorf("Expected user xGNS to be 0 after undelegating 1_000_000, got %d",
		// 		xgns.BalanceOf(user3Realm.Address()))
		// }

		// lockedList, exist := addrLockedGns.Get(user3Realm.Address().String())
		// if !exist {
		// 	t.Errorf("Expected lockedGNS to be created after Undelegate")
		// }
		// locked := lockedList.([]lockedGNS)
		// if len(locked) == 0 {
		// 	t.Errorf("Expected at least 1 lockedGNS after Undelegate")
		// }
		// if locked[0].amount != 1_000_000 {
		// 	t.Errorf("LockedGNS amount mismatch, got %d, want 1_000_000", locked[0].amount)
		// }
		// // check lockedAmount incremented?
		// if lockedAmount != 1_000_000 {
		// 	t.Errorf("lockedAmount expected 1_000_000, got %d", lockedAmount)
		// }

		// // below minimum => panic
		// {
		// 	defer func() {
		// 		if r := recover(); r == nil {
		// 			t.Errorf("Expected panic when undelegating below minimum")
		// 		}
		// 	}()
		// 	Undelegate(to, 999_999) // or 999_990, etc. if min = 1_000_000
		// }
	}()
}

func TestCollectUndelegatedGns(t *testing.T) {
	setupStakerRunning(t, true)

	testing.SetRealm(userRealm)
	func() {
		testing.SetRealm(std.NewCodeRealm("gno.land/r/gnoswap/v1/gov/staker"))
		// 1) no locked => expect 0
		addrLockedGns.Remove(userRealm.Address().String()) // ensure no locked
		collected := CollectUndelegatedGns(cross)
		if collected != 0 {
			t.Errorf("Expected 0 when no locked gns, got %d", collected)
		}

		// 2) add locked but time not passed => 0
		now := uint64(time.Now().Unix())
		locked := lockedGNS{
			amount: 100_000,
			unlock: now + SECONDS_IN_7_DAYS,
		}
		addrLockedGns.Set(userRealm.Address().String(), []lockedGNS{locked})
		lockedAmount = 100_000

		collected = CollectUndelegatedGns(cross)
		if collected != 0 {
			t.Errorf("Expected 0 if 7days not passed, got %d", collected)
		}
		// verify still in locked
		lockedList, exist := addrLockedGns.Get(userRealm.Address().String())
		if !exist {
			t.Errorf("Expected lockedGNS to remain after CollectUndelegatedGns")
		}
		lkList := lockedList.([]lockedGNS)
		if len(lkList) != 1 {
			t.Errorf("Locked list should remain, but length = %d", len(lkList))
		}

		// 3) set time => unlocked
		locked.unlock = uint64(time.Now().Unix()) - 1 // forcibly make it past time
		addrLockedGns.Set(userRealm.Address().String(), []lockedGNS{locked})

		// collected = CollectUndelegatedGns()
		// if collected != 100_000 {
		// 	t.Errorf("Expected 100_000 collected, got %d", collected)
		// }
		// // check locked removed from the tree
		// if _, exists := addrLockedGns.Get(userRealm.Address().String()); exists {
		// 	t.Errorf("Expected addrLockedGns key to be removed if empty after collecting all locked gns")
		// }
		// if lockedAmount != 0 {
		// 	t.Errorf("lockedAmount should have been decremented to 0, got %d", lockedAmount)
		// }
	}()
}

func TestCollectReward(t *testing.T) {
	setupStakerRunning(t, true)

	testing.SetRealm(user2Realm)
	func() {
		testing.SetRealm(std.NewCodeRealm(consts.EMISSION_PATH))
		testing.SkipHeights(100)
		gns.MintGns(cross, govStaker)

		testing.SetRealm(std.NewUserRealm(admin))
		ugnotFaucet(t, govStaker, 1_000_000)
		ugnotDeposit(t, govStaker, 1_000_000)
		ugnotFaucet(t, derivePkgAddr(wugnotPath), 1_000_000)
		ugnotFaucet(t, user2Realm.Address(), 1_000_000)
	}()

	testing.SetRealm(user2Realm)
	func() {
		testing.SetRealm(std.NewCodeRealm("gno.land/r/gnoswap/v1/gov/staker"))
		user := user2Realm.Address().String()

		rewardState.addStake(uint64(std.ChainHeight()), std.Address(user), 10, 0, make(map[string]uint64))

		// set a fake emission reward
		currentGNSBalance = 50_000
		currentProtocolFeeBalance[consts.WUGNOT_PATH] = 10_000

		testing.SetRealm(std.NewCodeRealm(consts.EMISSION_PATH))
		testing.SkipHeights(100)
		gns.MintGns(cross, user2Realm.Address())

		testing.SetRealm(user2Realm)
		testing.SkipHeights(100)

		// call CollectReward
		testing.SetOriginCaller(user2Realm.Address())
		func() {
			testing.SetRealm(std.NewCodeRealm("gno.land/r/gnoswap/v1/gov/staker"))
			CollectReward(cross)
		}()

		// // expect user emissionReward = 0
		// claimableAmount, protocolFeeClaimable := rewardState.CalculateReward(std.Address(user))
		// if claimableAmount != 0 {
		// 	t.Errorf("Expected userEmissionReward to be 0 after collect, got %d", claimableAmount)
		// }
		// if len(protocolFeeClaimable) != 0 {
		// 	t.Errorf("Expected protocolFeeClaimable to be empty after collect, got %d", len(protocolFeeClaimable))
		// }
		// for _, amount := range protocolFeeClaimable {
		// 	if amount != 0 {
		// 		t.Errorf("Expected protocolFeeClaimable to be 0 after collect, got %d", amount)
		// 	}
		// }
	}()

	// If GOV_STAKER_ADDR had less GNS than 50_000 => we expect panic
	// can test in separate subcase
}

func TestCollectRewardFromLaunchPad(t *testing.T) {
	tests := []struct {
		name             string
		inputOwner       std.Address
		expectedHasPanic bool
		expectedPanicMsg string
		expectedHasAbort bool
		expectedAbortMsg string
	}{
		{
			name:             "caller is launchpad",
			inputOwner:       consts.LAUNCHPAD_ADDR,
			expectedHasAbort: true,
			expectedAbortMsg: "[GNOSWAP-GOV_STAKER-008] invalid address || g1v9kxjcm9ta047h6lta047h6lta047h6lzd40gh is not project wallet from launchpad",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setupStakerRunning(t, true)
			testing.SetOriginCaller(tt.inputOwner)
			testing.SetRealm(std.NewCodeRealm("gno.land/r/gnoswap/v1/launchpad"))

			if tt.expectedHasPanic {
				uassert.PanicsWithMessage(t, tt.expectedPanicMsg, func() {
					CollectRewardFromLaunchPad(cross, userRealm.Address())
				})
			} else if tt.expectedHasAbort {
				uassert.AbortsWithMessage(t, tt.expectedAbortMsg, func() {
					CollectRewardFromLaunchPad(cross, userRealm.Address())
				})
			} else {
				CollectRewardFromLaunchPad(cross, userRealm.Address())
			}
		})
	}

	// Then set realm to a custom "launchpadRealm" whose .Address() == consts.LAUNCHPAD_ADDR
	// => call normal and see if distribution works
}

func resetStaker(t *testing.T) {
	t.Helper()

	totalDelegated = 0
	delegatorAmount = avl.NewTree()
	delegatedFromTo = avl.NewTree()
	delegatedTo = avl.NewTree()
}

func setupStakerRunning(t *testing.T, run bool) {
	t.Helper()
	resetStaker(t)
	setupStakerAdminPermission(t)

	adminAddr := testutils.TestAddress("admin")

	testing.SetOriginCaller(adminAddr)
	testing.SetRealm(std.NewUserRealm(adminAddr))

	SetRunning(cross, run)
}

func setupStakerAdminPermission(t *testing.T) {
	t.Helper()

	adminAddr := testutils.TestAddress("admin")

	testing.SetOriginCaller(adminAddr)
	testing.SetRealm(std.NewUserRealm(adminAddr))

	globalManager := rbac.GlobalManager()
	globalManager.UpdatePermission("admin", "access", func(addr std.Address) error {
		if addr != adminAddr {
			return ufmt.Errorf("caller is not admin")
		}

		return nil
	})
}
