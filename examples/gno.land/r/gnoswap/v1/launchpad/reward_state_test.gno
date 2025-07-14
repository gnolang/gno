package launchpad

import (
	"testing"

	"gno.land/p/demo/uassert"
	u256 "gno.land/p/gnoswap/uint256"
)

func TestRewardState_IsClaimable(t *testing.T) {
	tests := []struct {
		name                string
		claimableHeight     int64
		currentHeight       int64
		expectedIsClaimable bool
		expectedHasError    bool
		expectedError       string
	}{
		{
			name:                "is claimable return false when current height is before claimable height",
			claimableHeight:     100,
			currentHeight:       50,
			expectedIsClaimable: false,
			expectedHasError:    false,
		},
		{
			name:                "is claimable return true when current height equals claimable height",
			claimableHeight:     100,
			currentHeight:       100,
			expectedIsClaimable: true,
			expectedHasError:    false,
		},
		{
			name:                "is claimable return true when current height is after claimable height",
			claimableHeight:     100,
			currentHeight:       150,
			expectedIsClaimable: true,
			expectedHasError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			state := &RewardState{
				claimableBlockTimeInfo: &BlockTimeInfo{
					blockHeight: tt.claimableHeight,
				},
			}

			// Execute
			result := state.IsClaimable(tt.currentHeight)

			// Verify
			uassert.Equal(t, result, tt.expectedIsClaimable)
		})
	}
}

func TestRewardState_CalculateReward(t *testing.T) {
	tests := []struct {
		name                    string
		priceDebt               *u256.Uint
		depositAmount           int64
		claimedAmount           int64
		accumRewardPerDeposit   *u256.Uint
		expectedReward          int64
		expectedClaimableReward int64
		expectedHasError        bool
		expectedError           string
	}{
		{
			name:                    "calculate reward return correct reward for first claim",
			priceDebt:               u256.NewUintFromInt64(100),
			depositAmount:           1000,
			claimedAmount:           0,
			accumRewardPerDeposit:   u256.NewUintFromInt64(200),
			expectedReward:          100 * 1000,
			expectedClaimableReward: 100 * 1000,
			expectedHasError:        false,
		},
		{
			name:                    "calculate reward return correct reward with previous claims",
			priceDebt:               u256.NewUintFromInt64(100),
			depositAmount:           1000,
			claimedAmount:           50,
			accumRewardPerDeposit:   u256.NewUintFromInt64(200),
			expectedReward:          100 * 1000,
			expectedClaimableReward: 100 * 1000 - 50,
			expectedHasError:        false,
		},
		{
			name:                    "calculate reward return zero when no new rewards",
			priceDebt:               u256.NewUintFromInt64(200),
			depositAmount:           1000,
			claimedAmount:           0,
			accumRewardPerDeposit:   u256.NewUintFromInt64(200),
			expectedReward:          0 * 1000,
			expectedClaimableReward: 0 * 1000,
			expectedHasError:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			state := &RewardState{
				priceDebtX128: u256.Zero().Lsh(tt.priceDebt, 128),
				depositAmount: tt.depositAmount,
				claimedAmount: tt.claimedAmount,
			}

			// Execute
			accumRewardPerDepositX128 := u256.Zero().Lsh(tt.accumRewardPerDeposit, 128)
			resultReward := state.calculateReward(accumRewardPerDepositX128)
			resultClaimableReward := state.calculateClaimableReward(accumRewardPerDepositX128)

			// Verify
			uassert.Equal(t, resultReward, tt.expectedReward)
			uassert.Equal(t, resultClaimableReward, tt.expectedClaimableReward)
		})
	}
}

func TestNewRewardState(t *testing.T) {
	tests := []struct {
		name                  string
		priceDebt             *u256.Uint
		depositAmount         int64
		distributeStartHeight int64
		distributeEndHeight   int64
		claimableTime         int64
		claimableHeight       int64
		expectedHasError      bool
		expectedError         string
	}{
		{
			name:                  "new reward state create valid state with correct values",
			priceDebt:             u256.NewUintFromInt64(100),
			depositAmount:         1000,
			distributeStartHeight: 100,
			distributeEndHeight:   200,
			claimableTime:         1000,
			claimableHeight:       100,
			expectedHasError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			claimableBlockTimeInfo := &BlockTimeInfo{
				blockTime:   tt.claimableTime,
				blockHeight: tt.claimableHeight,
			}

			accumulatedRewardPerDepositX128 := u256.Zero().Lsh(tt.priceDebt, 128)
			// Execute
			state := NewRewardState(
				accumulatedRewardPerDepositX128,
				tt.depositAmount,
				tt.distributeStartHeight,
				tt.distributeEndHeight,
				claimableBlockTimeInfo,
			)

			// Verify
			statePriceDebt := u256.Zero().Rsh(state.PriceDebtX128(), 128).Int64()
			uassert.Equal(t, tt.priceDebt.Int64(), statePriceDebt)
			uassert.Equal(t, tt.depositAmount, state.DepositAmount())
			uassert.Equal(t, tt.distributeStartHeight, state.DistributeStartHeight())
			uassert.Equal(t, tt.distributeEndHeight, state.DistributeEndHeight())
			uassert.Equal(t, tt.claimableHeight, state.ClaimableHeight())
			uassert.Equal(t, tt.claimableTime, state.ClaimableTime())
		})
	}
}
