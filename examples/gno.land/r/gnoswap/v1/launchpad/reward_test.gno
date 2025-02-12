package launchpad

import (
	"std"
	"testing"
	"time"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
	"gno.land/p/demo/ufmt"
	u256 "gno.land/p/gnoswap/uint256"
)

func TestValidateRewardCollection(t *testing.T) {
	currentHeight := uint64(100)

	tests := []struct {
		name        string
		deposit     Deposit
		height      uint64
		shouldError bool
		errorMsg    string
	}{
		{
			name: "invalid deposit Id",
			deposit: Deposit{
				claimableHeight:   50, // Less than current height
				rewardCollectTime: 0,
			},
			height:      currentHeight,
			shouldError: true,
			errorMsg:    "[GNOSWAP-LAUNCHPAD-020] not exist deposit || depositId()",
		},
		{
			name: "Reward not yet claimable",
			deposit: Deposit{
				claimableHeight:   150, // Greater than current height
				rewardCollectTime: 0,
			},
			height:      currentHeight,
			shouldError: true,
			errorMsg:    "reward not yet claimable",
		},
		{
			name: "Reward already collected",
			deposit: Deposit{
				claimableHeight:   50, // Less than current height
				rewardCollectTime: 100,
			},
			height:      currentHeight,
			shouldError: true,
			errorMsg:    "[GNOSWAP-LAUNCHPAD-023] already collected",
		},
		{
			name: "Valid case",
			deposit: Deposit{
				id:                "depositId",
				depositor:         "g1wymu47drhr0kuq2098m792lytgtj2nyx77yrsm",
				claimableHeight:   50, // Less than current height
				rewardCollectTime: 0,
			},
			height:      currentHeight,
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.shouldError {
				updateDeposit(tt.deposit)
			}
			err := validateRewardCollection(tt.deposit, tt.height)
			if tt.shouldError {
				uassert.Error(t, err)
			} else {
				uassert.NoError(t, err)
			}
		})
	}
}

func TestCalculateDepositRatioX128(t *testing.T) {
	tests := []struct {
		name       string
		tierAmount uint64
		amount     uint64
		expected   string
	}{
		{
			name:       "Equal amounts",
			tierAmount: 1000,
			amount:     1000,
			expected:   "340282366920938463463374607431768211456", // 1.0 in Q128
		},
		{
			name:       "Half amount",
			tierAmount: 1000,
			amount:     500,
			expected:   "170141183460469231731687303715884105728", // 0.5 in Q128
		},
		{
			name:       "Double amount",
			tierAmount: 500,
			amount:     1000,
			expected:   "680564733841876926926749214863536422912", // 2.0 in Q128
		},
		{
			name:       "Zero amount",
			tierAmount: 1000,
			amount:     0,
			expected:   "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ratio := calcDepositRatioX128(tt.tierAmount, tt.amount)
			expected := u256.MustFromDecimal(tt.expected)
			uassert.Equal(t, 0, ratio.Cmp(expected))
		})
	}
}

func TestProcessDepositReward(t *testing.T) {
	tests := []struct {
		name           string
		deposit        Deposit
		rewardX128     *u256.Uint
		tierAmount     uint64
		expectedReward uint64
		shouldError    bool
	}{
		{
			name: "Normal reward calculation",
			deposit: Deposit{
				amount: 1000,
				//rewardAmount: 0,
			},
			rewardX128:     u256.NewUint(1000).Mul(u256.NewUint(1000), q128),
			tierAmount:     2000,
			expectedReward: 500, // (1000/2000) * 1000 = 500
			shouldError:    false,
		},
		{
			name: "Zero tier amount",
			deposit: Deposit{
				amount: 1000,
				//rewardAmount: 0,
			},
			rewardX128:     u256.NewUint(1000).Mul(u256.NewUint(1000), q128),
			tierAmount:     0,
			expectedReward: 0,
			shouldError:    true,
		},
		{
			name: "Reward is accumulated",
			deposit: Deposit{
				amount: 1000,
				//rewardAmount: 100,
			},
			rewardX128:     u256.NewUint(1000).Mul(u256.NewUint(1000), q128),
			tierAmount:     2000,
			expectedReward: 500, // 600, // Existing 100 + New reward 500
			shouldError:    false,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewRewardState(tt.rewardX128, 1000, 2000)
			depositId := ufmt.Sprintf("depositId-%d", i)
			state.addStake(1000, depositId, tt.deposit.amount)
			state.TotalStake = tt.tierAmount // force overriding total stake
			state.finalize(1001)

			reward := state.CalculateReward(depositId)
			uassert.Equal(t, tt.expectedReward, reward)
		})
	}
}

func TestCalculateTierRewards(t *testing.T) {
	tests := []struct {
		name               string
		tier               Tier
		currentHeight      uint64
		lastCalcHeight     uint64
		expectedRewardX128 string
		expectedReward     uint64
		shouldBeZero       bool
	}{
		{
			name: "tierAmountPerBlockX96 is 0",
			tier: Tier{
				tierAmountPerBlockX128: u256.Zero(),
				ended:                  TimeInfo{height: 1000},
			},
			currentHeight:  500,
			lastCalcHeight: 400,
			shouldBeZero:   true,
		},
		{
			name: "sinceLast is 0",
			tier: Tier{
				tierAmountPerBlockX128: u256.NewUint(1000),
				ended:                  TimeInfo{height: 1000},
			},
			currentHeight:  500,
			lastCalcHeight: 500,
			shouldBeZero:   true,
		},
		{
			name: "Current height exceeds end height",
			tier: Tier{
				tierAmountPerBlockX128: u256.NewUint(1000),
				ended:                  TimeInfo{height: 450},
			},
			currentHeight:      500,
			lastCalcHeight:     400,
			expectedRewardX128: "50000", // 1000 * (450-400)
			expectedReward:     0,       // (50000 / Q96)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rewardX128, reward, err := calculateTierRewards(tt.tier, tt.currentHeight, tt.lastCalcHeight)

			uassert.NoError(t, err)

			if tt.shouldBeZero {
				uassert.Equal(t, true, rewardX128.IsZero())
				uassert.Equal(t, uint64(0), reward)
			} else {
				expected := u256.MustFromDecimal(tt.expectedRewardX128)
				uassert.Equal(t, 0, rewardX128.Cmp(expected))
				uassert.Equal(t, tt.expectedReward, reward)
			}
		})
	}
}

func TestCollectRewardByDepositId_FailCases(t *testing.T) {
	setup := func() {
		projects = make(map[string]Project)
		deposits = make(map[string]Deposit)
		depositsByUserByProject = make(map[std.Address]map[string][]string)
	}

	t.Run("ID not found", func(t *testing.T) {
		setup()

		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic but not")
			}
		}()

		CollectRewardByDepositId("non_existent_id")
	})

	t.Run("Claimable height before", func(t *testing.T) {
		setup()

		projectId := "test_project"
		depositId := "test_deposit"

		deposits[depositId] = Deposit{
			id:              depositId,
			projectId:       projectId,
			claimableHeight: uint64(std.GetHeight()) + 1000,
		}

		uassert.PanicsWithMessage(t, "[GNOSWAP-LAUNCHPAD-024] not yet claim reward || need block >= 1123, current=123", func() {
			CollectRewardByDepositId(depositId)
		})
	})
}

func TestCollectRewardByDepositId_OwnerCheck(t *testing.T) {
	userA := testutils.TestAddress("userA")
	userB := testutils.TestAddress("userB")

	project := createTestProject(t)
	projects[project.id] = project

	depositA := Deposit{
		id:                "depA",
		projectId:         project.id,
		tier:              "30",
		depositor:         userA,
		claimableHeight:   133,
		claimableTime:     uint64(time.Now().Unix()) + 1000,
		rewardCollectTime: 0,
	}

	updateDeposit(depositA)

	depositB := Deposit{
		id:                "depB",
		projectId:         project.id,
		tier:              "30",
		depositor:         userB,
		claimableHeight:   10,
		claimableTime:     uint64(time.Now().Unix()) + 1000,
		rewardCollectTime: 0,
	}

	updateDeposit(depositB)

	t.Run("UserA tries depositB => panic", func(t *testing.T) {
		std.TestSetRealm(std.NewUserRealm(userA))
		defer func() {
			r := recover()
			if r == nil {
				t.Errorf("expected panic, got none")
			} else {
				uassert.Equal(t, "[GNOSWAP-LAUNCHPAD-026] invalid owner || caller(g1w4ek2ujpta047h6lta047h6lta047h6l55c22z) does not own depositId(depB)", r.(string))
			}
		}()
		CollectRewardByDepositId("depB")
	})

	t.Run("UserA depositA => not claimable yet", func(t *testing.T) {
		std.TestSetRealm(std.NewUserRealm(userA))
		// height=0, depositA.claimableHeight=10
		defer func() {
			r := recover()
			if r == nil {
				t.Errorf("expected panic for not claimable, got none")
			} else {
				uassert.Equal(t, "[GNOSWAP-LAUNCHPAD-024] not yet claim reward || need block >= 133, current=123", r.(string))
			}
		}()
		CollectRewardByDepositId("depA")
	})

	std.TestSkipHeights(11) // now height=11 >= depositA.claimableHeight=10

	depA := deposits["depA"]
	depA.claimableTime = 0
	deposits["depA"] = depA

	t.Run("UserA depositA => claim success", func(t *testing.T) {
		std.TestSetRealm(std.NewUserRealm(userA))

		rewardPerDeposit := u256.NewUint(1)
		rewardPerDepositX128 := u256.Zero().Lsh(rewardPerDeposit, 128)
		rewardPerBlock := u256.NewUint(100)
		rewardPerBlock = u256.Zero().Lsh(rewardPerBlock, 128)
		mockReward := NewReward(rewardPerDepositX128, 123, 10000)
		mockReward.info.Set("depA", &RewardInfo{
			startHeight:   123,
			priceDebt:     u256.Zero(),
			depositAmount: 100,
			claimed:       0,
		})
		mockTier := &Tier{
			tierAmountPerBlockX128: rewardPerBlock,
			totalDepositAmount:     100,
			ended:                  TimeInfo{height: 10000},
		}
		mockTier.setReward(*mockReward)
		project.setTier(TIER30, *mockTier)

		amount := CollectRewardByDepositId("depA")
		uassert.Equal(t, uint64(100), amount)
	})
}
