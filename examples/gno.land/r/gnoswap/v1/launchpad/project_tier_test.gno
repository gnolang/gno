package launchpad

import (
	"testing"

	"gno.land/p/demo/uassert"
	u256 "gno.land/p/gnoswap/uint256"
)

func TestProjectTier_Getters(t *testing.T) {
	tests := []struct {
		name                         string
		id                           string
		totalDistributeAmount        int64
		distributeAmountPerBlockX128 *u256.Uint
		startTime                    int64
		startHeight                  int64
		endTime                      int64
		endHeight                    int64
		totalDepositAmount           int64
		totalWithdrawAmount          int64
		totalDepositCount            int64
		totalWithdrawCount           int64
		totalCollectedAmount         int64
		getterFunc                   string
		expectedValue                interface{}
	}{
		{
			name:                         "project tier get id",
			id:                           "test:180",
			totalDistributeAmount:        1000,
			distributeAmountPerBlockX128: u256.NewUintFromInt64(100),
			startTime:                    1000,
			startHeight:                  100,
			endTime:                      2000,
			endHeight:                    200,
			totalDepositAmount:           500,
			totalWithdrawAmount:          200,
			totalDepositCount:            5,
			totalWithdrawCount:           2,
			totalCollectedAmount:         300,
			getterFunc:                   "ID",
			expectedValue:                "test:180",
		},
		{
			name:                         "project tier get total distribute amount",
			id:                           "test:180",
			totalDistributeAmount:        1000,
			distributeAmountPerBlockX128: u256.NewUintFromInt64(100),
			startTime:                    1000,
			startHeight:                  100,
			endTime:                      2000,
			endHeight:                    200,
			totalDepositAmount:           500,
			totalWithdrawAmount:          200,
			totalDepositCount:            5,
			totalWithdrawCount:           2,
			totalCollectedAmount:         300,
			getterFunc:                   "TotalDistributeAmount",
			expectedValue:                int64(1000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tier := &ProjectTier{
				id:                           tt.id,
				totalDistributeAmount:        tt.totalDistributeAmount,
				distributeAmountPerBlockX128: tt.distributeAmountPerBlockX128,
				startBlockTimeInfo: &BlockTimeInfo{
					blockTime:   tt.startTime,
					blockHeight: tt.startHeight,
				},
				endBlockTimeInfo: &BlockTimeInfo{
					blockTime:   tt.endTime,
					blockHeight: tt.endHeight,
				},
				totalDepositAmount:   tt.totalDepositAmount,
				totalWithdrawAmount:  tt.totalWithdrawAmount,
				totalDepositCount:    tt.totalDepositCount,
				totalWithdrawCount:   tt.totalWithdrawCount,
				totalCollectedAmount: tt.totalCollectedAmount,
			}

			// Execute and verify
			var result interface{}
			switch tt.getterFunc {
			case "ID":
				result = tier.ID()
			case "TotalDistributeAmount":
				result = tier.TotalDistributeAmount()
			case "TotalCollectedAmount":
				result = tier.TotalCollectedAmount()
			case "TotalDepositAmount":
				result = tier.TotalDepositAmount()
			case "TotalWithdrawAmount":
				result = tier.TotalWithdrawAmount()
			case "TotalDepositCount":
				result = tier.TotalDepositCount()
			case "TotalWithdrawCount":
				result = tier.TotalWithdrawCount()
			case "CurrentDepositCount":
				result = tier.CurrentDepositCount()
			case "CurrentDepositAmount":
				result = tier.CurrentDepositAmount()
			}
			uassert.Equal(t, result, tt.expectedValue)
		})
	}
}

func TestProjectTier_Status(t *testing.T) {
	tests := []struct {
		name             string
		startHeight      int64
		endHeight        int64
		currentHeight    int64
		expectedIsActive bool
		expectedIsEnded  bool
		expectedHasError bool
		expectedError    string
	}{
		{
			name:             "project tier status inactive and not ended when before start",
			startHeight:      100,
			endHeight:        200,
			currentHeight:    50,
			expectedIsActive: false,
			expectedIsEnded:  false,
		},
		{
			name:             "project tier status active and not ended during active period",
			startHeight:      100,
			endHeight:        200,
			currentHeight:    150,
			expectedIsActive: true,
			expectedIsEnded:  false,
		},
		{
			name:             "project tier status inactive and ended after end",
			startHeight:      100,
			endHeight:        200,
			currentHeight:    250,
			expectedIsActive: false,
			expectedIsEnded:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tier := &ProjectTier{
				startBlockTimeInfo: &BlockTimeInfo{
					blockHeight: tt.startHeight,
				},
				endBlockTimeInfo: &BlockTimeInfo{
					blockHeight: tt.endHeight,
				},
			}

			// Execute and verify
			isActive := tier.isActivated(tt.currentHeight)
			isEnded := tier.isEnded(tt.currentHeight)

			uassert.Equal(t, isActive, tt.expectedIsActive)
			uassert.Equal(t, isEnded, tt.expectedIsEnded)
		})
	}
}

func TestProjectTier_DepositAndWithdraw(t *testing.T) {
	tests := []struct {
		name                   string
		initialDepositAmount   int64
		initialWithdrawAmount  int64
		initialDepositCount    int64
		initialWithdrawCount   int64
		depositAmount          int64
		operation              string
		expectedDepositAmount  int64
		expectedWithdrawAmount int64
		expectedDepositCount   int64
		expectedWithdrawCount  int64
		expectedCurrentAmount  int64
		expectedCurrentCount   int64
		expectedHasError       bool
		expectedError          string
	}{
		{
			name:                   "project tier deposit update amounts and counts correctly",
			initialDepositAmount:   1000,
			initialWithdrawAmount:  0,
			initialDepositCount:    2,
			initialWithdrawCount:   0,
			depositAmount:          500,
			operation:              "deposit",
			expectedDepositAmount:  1500,
			expectedWithdrawAmount: 0,
			expectedDepositCount:   3,
			expectedWithdrawCount:  0,
			expectedCurrentAmount:  1500,
			expectedCurrentCount:   3,
		},
		{
			name:                   "project tier withdraw update amounts and counts correctly",
			initialDepositAmount:   1500,
			initialWithdrawAmount:  0,
			initialDepositCount:    3,
			initialWithdrawCount:   0,
			depositAmount:          500,
			operation:              "withdraw",
			expectedDepositAmount:  1500,
			expectedWithdrawAmount: 500,
			expectedDepositCount:   3,
			expectedWithdrawCount:  1,
			expectedCurrentAmount:  1000,
			expectedCurrentCount:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tier := &ProjectTier{
				totalDepositAmount:  tt.initialDepositAmount,
				totalWithdrawAmount: tt.initialWithdrawAmount,
				totalDepositCount:   tt.initialDepositCount,
				totalWithdrawCount:  tt.initialWithdrawCount,
			}
			deposit := &Deposit{
				depositAmount: tt.depositAmount,
			}

			// Execute
			if tt.operation == "deposit" {
				tier.deposit(deposit)
			} else {
				tier.withdraw(deposit)
			}

			// Verify
			uassert.Equal(t, tier.TotalDepositAmount(), tt.expectedDepositAmount)
			uassert.Equal(t, tier.TotalWithdrawAmount(), tt.expectedWithdrawAmount)
			uassert.Equal(t, tier.TotalDepositCount(), tt.expectedDepositCount)
			uassert.Equal(t, tier.TotalWithdrawCount(), tt.expectedWithdrawCount)
			uassert.Equal(t, tier.CurrentDepositAmount(), tt.expectedCurrentAmount)
			uassert.Equal(t, tier.CurrentDepositCount(), tt.expectedCurrentCount)
		})
	}
}

func TestNewProjectTier(t *testing.T) {
	tests := []struct {
		name                  string
		projectID             string
		tierDuration          int64
		totalDistributeAmount int64
		startBlockTime        int64
		endBlockTime          int64
		averageBlockTimeMs    int64
		currentBlockTime      int64
		currentBlockHeight    int64
		expectedID            string
		expectedStartHeight   int64
		expectedEndHeight     int64
		expectedHasError      bool
		expectedError         string
	}{
		{
			name:                  "project tier create valid tier with correct values",
			projectID:             "test",
			tierDuration:          180,
			totalDistributeAmount: 1000,
			startBlockTime:        1000,
			endBlockTime:          2000,
			averageBlockTimeMs:    1000,
			currentBlockTime:      500,
			currentBlockHeight:    50,
			expectedID:            "test:180",
			expectedStartHeight:   550,
			expectedEndHeight:     1550,
			expectedHasError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			tier := NewProjectTier(
				tt.projectID,
				tt.tierDuration,
				tt.totalDistributeAmount,
				tt.startBlockTime,
				tt.endBlockTime,
				tt.averageBlockTimeMs,
				tt.currentBlockTime,
				tt.currentBlockHeight,
			)

			// Verify
			uassert.Equal(t, tier.ID(), tt.expectedID)
			uassert.Equal(t, tier.StartHeight(), tt.expectedStartHeight)
			uassert.Equal(t, tier.EndHeight(), tt.expectedEndHeight)
		})
	}
}
