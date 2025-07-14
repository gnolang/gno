package launchpad

import (
	"std"
	"testing"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/uassert"
	u256 "gno.land/p/gnoswap/uint256"
)

func TestRewardManager_IsInitialized(t *testing.T) {
	tests := []struct {
		name             string
		rewardStates     map[string]*RewardState
		expectedResult   bool
		expectedHasError bool
		expectedError    string
	}{
		{
			name:             "is initialized return false for empty rewards",
			rewardStates:     map[string]*RewardState{},
			expectedResult:   false,
			expectedHasError: false,
		},
		{
			name: "is initialized return true when rewards exist",
			rewardStates: map[string]*RewardState{
				"test": NewRewardState(u256.Zero(), 1000, 100, 200, &BlockTimeInfo{}),
			},
			expectedResult:   true,
			expectedHasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tree := avl.NewTree()
			for k, v := range tt.rewardStates {
				tree.Set(k, v)
			}
			manager := &RewardManager{rewards: tree}

			// Execute
			result := manager.IsInitialized()

			// Verify
			uassert.Equal(t, result, tt.expectedResult)
		})
	}
}

func TestRewardManager_UpdateRewardPerDepositX128(t *testing.T) {
	tests := []struct {
		name                                string
		totalDistributeAmount               int64
		distributeStartHeight               int64
		distributeEndHeight                 int64
		currentHeight                       int64
		rewardHeight                        int64
		expectedAccumulatedRewardPerDeposit string
		expectedHasError                    bool
		expectedError                       string
	}{
		{
			name:                                "update reward per deposit succeed with valid height",
			totalDistributeAmount:               1000,
			distributeStartHeight:               100,
			distributeEndHeight:                 200,
			currentHeight:                       150,
			rewardHeight:                        100,
			expectedAccumulatedRewardPerDeposit: "0",
			expectedHasError:                    false,
		},
		{
			name:                                "update reward per deposit fail with invalid height",
			totalDistributeAmount:               1000,
			distributeStartHeight:               100,
			distributeEndHeight:                 200,
			currentHeight:                       50,
			rewardHeight:                        0,
			expectedAccumulatedRewardPerDeposit: "0",
			expectedHasError:                    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			manager := NewRewardManager(
				tt.totalDistributeAmount,
				tt.distributeStartHeight,
				tt.distributeEndHeight,
				tt.currentHeight,
			)

			// Execute
			err := manager.updateRewardPerDepositX128(tt.totalDistributeAmount, tt.rewardHeight)

			// Verify
			if tt.expectedHasError {
				uassert.ErrorContains(t, err, tt.expectedError)
			} else {
				uassert.Equal(t, tt.expectedAccumulatedRewardPerDeposit, manager.AccumulatedRewardPerDepositX128().ToString())
			}
		})
	}
}

func TestRewardManager_AddRewardStateByDeposit(t *testing.T) {
	tests := []struct {
		name                  string
		deposit               *Deposit
		totalDistributeAmount int64
		distributeStart       int64
		distributeEnd         int64
		averageBlockTimeMs    int64
		existingDeposits      []*Deposit
		collectHeight         int64
		expectedStartHeight   int64
		expectedReward        int64
		expectedHasError      bool
		expectedError         string
	}{
		{
			name: "add reward state succeed for first deposit",
			deposit: NewDeposit(
				"test",
				30,
				"",
				1000,
				&BlockTimeInfo{blockTime: 1000, blockHeight: 150},
				&BlockTimeInfo{blockTime: 1000, blockHeight: 150},
			),
			totalDistributeAmount: 1000,
			distributeStart:       150,
			distributeEnd:         250,
			averageBlockTimeMs:    2000,
			existingDeposits:      []*Deposit{},
			collectHeight:         250,
			expectedStartHeight:   150,
			expectedReward:        1000,
			expectedHasError:      false,
		},
		{
			name: "add reward state succeed for subsequent deposit",
			deposit: NewDeposit(
				"test",
				30,
				"",
				1000,
				&BlockTimeInfo{blockTime: 1000, blockHeight: 150},
				&BlockTimeInfo{blockTime: 1000, blockHeight: 250},
			),
			totalDistributeAmount: 1000,
			distributeStart:       150,
			distributeEnd:         250,
			averageBlockTimeMs:    2000,
			existingDeposits: []*Deposit{
				NewDeposit(
					"existing",
					500,
					"",
					500,
					&BlockTimeInfo{blockTime: 1000, blockHeight: 150},
					&BlockTimeInfo{blockTime: 1000, blockHeight: 250},
				),
			},
			collectHeight:       250,
			expectedStartHeight: 150,
			expectedReward:      666,
			expectedHasError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			currentHeight := std.ChainHeight()

			manager := NewRewardManager(
				tt.totalDistributeAmount,
				tt.distributeStart,
				tt.distributeEnd,
				currentHeight,
			)

			totalDepositAmount := int64(0)

			for _, deposit := range tt.existingDeposits {
				totalDepositAmount += int64(deposit.depositAmount)
				manager.updateRewardPerDepositX128(totalDepositAmount, deposit.DepositHeight())
				manager.addRewardStateByDeposit(deposit, tt.averageBlockTimeMs)
			}

			// Execute
			totalDepositAmount += tt.deposit.depositAmount
			manager.updateRewardPerDepositX128(totalDepositAmount, tt.deposit.DepositHeight())
			rewardState := manager.addRewardStateByDeposit(tt.deposit, tt.averageBlockTimeMs)

			manager.updateRewardPerDepositX128(totalDepositAmount, tt.collectHeight)
			reward, err := manager.collectReward(tt.deposit.ID(), tt.collectHeight)
			if err != nil {
				uassert.NoError(t, err)
			}

			uassert.Equal(t, tt.expectedStartHeight, rewardState.DistributeStartHeight())
			uassert.Equal(t, tt.expectedReward, reward)
		})
	}
}

func TestRewardManager_CollectReward(t *testing.T) {
	tests := []struct {
		name                  string
		totalDistributeAmount int64
		distributeStartHeight int64
		distributeEndHeight   int64
		currentHeight         int64
		depositId             string
		deposit               *Deposit
		averageBlockTimeMs    int64
		existingDeposits      []*Deposit
		collectHeight         int64
		expectedAmount        int64
		expectedHasError      bool
		expectedError         string
	}{
		{
			name:                  "collect reward succeed with valid deposit",
			totalDistributeAmount: 1000,
			distributeStartHeight: 100,
			distributeEndHeight:   200,
			currentHeight:         150,
			deposit: NewDeposit(
				"test",
				30,
				"",
				1000,
				&BlockTimeInfo{blockTime: 1000, blockHeight: 150},
				&BlockTimeInfo{blockTime: 1000, blockHeight: 200},
			),
			depositId:          "test:30::150",
			averageBlockTimeMs: 2000,
			existingDeposits:   []*Deposit{},
			collectHeight:      200,
			expectedAmount:     1000,
			expectedHasError:   false,
		},
		{
			name: "collect reward fail with nonexistent deposit",
			deposit: NewDeposit(
				"test",
				30,
				"",
				1000,
				&BlockTimeInfo{blockTime: 1000, blockHeight: 150},
				&BlockTimeInfo{blockTime: 1000, blockHeight: 250},
			),
			depositId:          "nonexistent",
			collectHeight:      150,
			averageBlockTimeMs: 2000,
			existingDeposits:   []*Deposit{},
			expectedAmount:     0,
			expectedHasError:   true,
			expectedError:      "[GNOSWAP-LAUNCHPAD-020] not exist deposit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testing.SetHeight(123)

			// Setup
			manager := NewRewardManager(
				tt.totalDistributeAmount,
				tt.distributeStartHeight,
				tt.distributeEndHeight,
				tt.currentHeight,
			)

			for _, deposit := range tt.existingDeposits {
				manager.updateRewardPerDepositX128(tt.totalDistributeAmount, deposit.DepositHeight())
				manager.addRewardStateByDeposit(deposit, tt.averageBlockTimeMs)
			}

			manager.updateRewardPerDepositX128(tt.totalDistributeAmount, tt.deposit.DepositHeight())
			manager.addRewardStateByDeposit(tt.deposit, tt.averageBlockTimeMs)

			// Execute
			manager.updateRewardPerDepositX128(tt.totalDistributeAmount, tt.collectHeight)
			amount, err := manager.collectReward(tt.depositId, tt.collectHeight)

			// Verify
			if tt.expectedHasError {
				uassert.ErrorContains(t, err, tt.expectedError)
			} else {
				uassert.NoError(t, err)
				uassert.Equal(t, amount, tt.expectedAmount)
			}
		})
	}
}

func TestNewRewardManager(t *testing.T) {
	tests := []struct {
		name                         string
		totalDistributeAmount        int64
		distributeStartHeight        int64
		distributeEndHeight          int64
		distributeAmountPerBlockX128 *u256.Uint
		currentHeight                int64
		expectedHasError             bool
		expectedError                string
	}{
		{
			name:                         "new reward manager create valid manager with correct values",
			totalDistributeAmount:        1000,
			distributeStartHeight:        100,
			distributeEndHeight:          200,
			distributeAmountPerBlockX128: u256.NewUintFromInt64(100),
			currentHeight:                100,
			expectedHasError:             false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			manager := NewRewardManager(
				tt.totalDistributeAmount,
				tt.distributeStartHeight,
				tt.distributeEndHeight,
				tt.currentHeight,
			)

			// Verify
			uassert.NotEqual(t, manager.rewards, nil)
		})
	}
}
