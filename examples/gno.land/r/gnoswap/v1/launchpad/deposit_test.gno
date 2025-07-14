package launchpad

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
	"gno.land/p/demo/uassert"
)

func TestDeposit_ID(t *testing.T) {
	tests := []struct {
		name     string
		deposit  *Deposit
		expected string
	}{
		{
			name: "get deposit id successfully",
			deposit: &Deposit{
				id: "test_deposit_id",
			},
			expected: "test_deposit_id",
		},
		{
			name: "get empty deposit id successfully",
			deposit: &Deposit{
				id: "",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uassert.Equal(t, tt.expected, tt.deposit.ID())
		})
	}
}

func TestDeposit_ProjectID(t *testing.T) {
	tests := []struct {
		name     string
		deposit  *Deposit
		expected string
	}{
		{
			name: "get project id successfully",
			deposit: &Deposit{
				projectID: "test_project_id",
			},
			expected: "test_project_id",
		},
		{
			name: "get empty project id successfully",
			deposit: &Deposit{
				projectID: "",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uassert.Equal(t, tt.expected, tt.deposit.ProjectID())
		})
	}
}

func TestDeposit_ProjectTierID(t *testing.T) {
	tests := []struct {
		name     string
		deposit  *Deposit
		expected string
	}{
		{
			name: "get project tier id successfully",
			deposit: &Deposit{
				projectID: "test_project",
				tier:      30,
			},
			expected: "test_project:30",
		},
		{
			name: "get project tier id with empty project id",
			deposit: &Deposit{
				projectID: "",
				tier:      30,
			},
			expected: ":30",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uassert.Equal(t, tt.expected, tt.deposit.ProjectTierID())
		})
	}
}

func TestDeposit_IsOwner(t *testing.T) {
	testAddr := testutils.TestAddress("test")
	otherAddr := testutils.TestAddress("other")

	tests := []struct {
		name     string
		deposit  *Deposit
		address  std.Address
		expected bool
	}{
		{
			name: "check owner successfully - true case",
			deposit: &Deposit{
				depositor: testAddr,
			},
			address:  testAddr,
			expected: true,
		},
		{
			name: "check owner successfully - false case",
			deposit: &Deposit{
				depositor: testAddr,
			},
			address:  otherAddr,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uassert.Equal(t, tt.expected, tt.deposit.IsOwner(tt.address))
		})
	}
}

func TestDeposit_IsEnded(t *testing.T) {
	tests := []struct {
		name          string
		deposit       *Deposit
		currentHeight int64
		expected      bool
	}{
		{
			name: "check is ended successfully - true case",
			deposit: &Deposit{
				endBlockTimeInfo: newBlockTimeInfo(1000, 100),
			},
			currentHeight: 101,
			expected:      true,
		},
		{
			name: "check is ended successfully - false case",
			deposit: &Deposit{
				endBlockTimeInfo: newBlockTimeInfo(1000, 100),
			},
			currentHeight: 99,
			expected:      false,
		},
		{
			name: "check is ended successfully - equal height case",
			deposit: &Deposit{
				endBlockTimeInfo: newBlockTimeInfo(1000, 100),
			},
			currentHeight: 100,
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uassert.Equal(t, tt.expected, tt.deposit.IsEnded(tt.currentHeight))
		})
	}
}

func TestDeposit_Withdraw(t *testing.T) {
	tests := []struct {
		name              string
		deposit           *Deposit
		currentHeight     int64
		expectedAmount    int64
		expectedWithdrawn bool
		expectedRemaining int64
	}{
		{
			name: "withdraw successfully",
			deposit: &Deposit{
				depositAmount:   1000,
				withdrawnHeight: 0,
			},
			currentHeight:     100,
			expectedAmount:    1000,
			expectedWithdrawn: true,
			expectedRemaining: 0,
		},
		{
			name: "withdraw with zero amount",
			deposit: &Deposit{
				depositAmount:   0,
				withdrawnHeight: 0,
			},
			currentHeight:     100,
			expectedAmount:    0,
			expectedWithdrawn: true,
			expectedRemaining: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amount := tt.deposit.withdraw(tt.currentHeight)
			uassert.Equal(t, tt.expectedAmount, amount)
			uassert.Equal(t, tt.currentHeight, tt.deposit.withdrawnHeight)
			uassert.Equal(t, tt.expectedRemaining, tt.deposit.depositAmount)
			uassert.Equal(t, tt.expectedWithdrawn, tt.deposit.IsWithdrawn())
		})
	}
}

func TestNewDeposit(t *testing.T) {
	testAddr := testutils.TestAddress("test")
	projectID := "gno.land/r/onbloc/obl:123"

	tests := []struct {
		name               string
		projectID          string
		tier               int64
		depositor          std.Address
		depositAmount      int64
		createdBlockTime   int64
		createdBlockHeight int64
		endBlockTime       int64
		endBlockHeight     int64
		expectedID         string
	}{
		{
			name:               "create new deposit successfully",
			projectID:          projectID,
			tier:               30,
			depositor:          testAddr,
			depositAmount:      1000,
			createdBlockTime:   900,
			createdBlockHeight: 90,
			endBlockTime:       1000,
			endBlockHeight:     100,
			expectedID:         makeDepositID(projectID, 30, testAddr, 90),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createdBlockTimeInfo := newBlockTimeInfo(tt.createdBlockTime, tt.createdBlockHeight)
			endBlockTimeInfo := newBlockTimeInfo(tt.endBlockTime, tt.endBlockHeight)

			deposit := NewDeposit(
				tt.projectID,
				tt.tier,
				tt.depositor,
				tt.depositAmount,
				createdBlockTimeInfo,
				endBlockTimeInfo,
			)

			uassert.Equal(t, tt.expectedID, deposit.ID())
			uassert.Equal(t, tt.projectID, deposit.ProjectID())
			uassert.Equal(t, tt.tier, deposit.Tier())
			uassert.Equal(t, tt.depositor, deposit.Depositor())
			uassert.Equal(t, tt.depositAmount, deposit.DepositAmount())
			uassert.Equal(t, tt.createdBlockTime, deposit.DepositTime())
			uassert.Equal(t, tt.createdBlockHeight, deposit.DepositHeight())
			uassert.Equal(t, tt.endBlockTime, deposit.EndTime())
			uassert.Equal(t, tt.endBlockHeight, deposit.EndHeight())
			uassert.Equal(t, int64(0), deposit.withdrawnHeight)
		})
	}
}
