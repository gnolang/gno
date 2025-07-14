package launchpad

import (
	"std"
	"testing"

	"gno.land/p/demo/uassert"
)

func TestProject_Getters(t *testing.T) {
	tests := []struct {
		name             string
		id               string
		projectName      string
		tokenPath        string
		depositAmount    int64
		recipient        std.Address
		createdTime      int64
		createdHeight    int64
		getterFunc       string
		expectedValue    interface{}
		expectedHasError bool
		expectedError    string
	}{
		{
			name:          "project get id",
			id:            "gno.land/r/demo/gns:100",
			projectName:   "Test Project",
			tokenPath:     "gno.land/r/demo/gns",
			depositAmount: 1000,
			recipient:     std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			createdTime:   1000,
			createdHeight: 100,
			getterFunc:    "ID",
			expectedValue: "gno.land/r/demo/gns:100",
		},
		{
			name:          "project get name",
			id:            "gno.land/r/demo/gns:100",
			projectName:   "Test Project",
			tokenPath:     "gno.land/r/demo/gns",
			depositAmount: 1000,
			recipient:     std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			createdTime:   1000,
			createdHeight: 100,
			getterFunc:    "Name",
			expectedValue: "Test Project",
		},
		{
			name:          "project get token path",
			id:            "gno.land/r/demo/gns:100",
			projectName:   "Test Project",
			tokenPath:     "gno.land/r/demo/gns",
			depositAmount: 1000,
			recipient:     std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			createdTime:   1000,
			createdHeight: 100,
			getterFunc:    "TokenPath",
			expectedValue: "gno.land/r/demo/gns",
		},
		{
			name:          "project get deposit amount",
			id:            "gno.land/r/demo/gns:100",
			projectName:   "Test Project",
			tokenPath:     "gno.land/r/demo/gns",
			depositAmount: 1000,
			recipient:     std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			createdTime:   1000,
			createdHeight: 100,
			getterFunc:    "DepositAmount",
			expectedValue: int64(1000),
		},
		{
			name:          "project get recipient",
			id:            "gno.land/r/demo/gns:100",
			projectName:   "Test Project",
			tokenPath:     "gno.land/r/demo/gns",
			depositAmount: 1000,
			recipient:     std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			createdTime:   1000,
			createdHeight: 100,
			getterFunc:    "Recipient",
			expectedValue: std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
		},
		{
			name:          "project get created at",
			id:            "gno.land/r/demo/gns:100",
			projectName:   "Test Project",
			tokenPath:     "gno.land/r/demo/gns",
			depositAmount: 1000,
			recipient:     std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			createdTime:   1000,
			createdHeight: 100,
			getterFunc:    "CreatedAt",
			expectedValue: int64(1000),
		},
		{
			name:          "project get created height",
			id:            "gno.land/r/demo/gns:100",
			projectName:   "Test Project",
			tokenPath:     "gno.land/r/demo/gns",
			depositAmount: 1000,
			recipient:     std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			createdTime:   1000,
			createdHeight: 100,
			getterFunc:    "CreatedHeight",
			expectedValue: int64(100),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			project := &Project{
				id:            tt.id,
				name:          tt.projectName,
				tokenPath:     tt.tokenPath,
				depositAmount: tt.depositAmount,
				recipient:     tt.recipient,
				conditions:    make(map[string]*ProjectCondition),
				tiers:         make(map[int64]*ProjectTier),
				createdBlockTimeInfo: &BlockTimeInfo{
					blockTime:   tt.createdTime,
					blockHeight: tt.createdHeight,
				},
			}

			// Execute and verify
			var result interface{}
			switch tt.getterFunc {
			case "ID":
				result = project.ID()
			case "Name":
				result = project.Name()
			case "TokenPath":
				result = project.TokenPath()
			case "DepositAmount":
				result = project.DepositAmount()
			case "Recipient":
				result = project.Recipient()
			case "CreatedAt":
				result = project.CreatedAt()
			case "CreatedHeight":
				result = project.CreatedHeight()
			}
			uassert.Equal(t, result, tt.expectedValue)
		})
	}
}

func TestProject_Status(t *testing.T) {
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
			name:             "status inactive and not ended when before start",
			startHeight:      150,
			endHeight:        200,
			currentHeight:    120,
			expectedIsActive: false,
			expectedIsEnded:  false,
		},
		{
			name:             "status active and not ended during active period",
			startHeight:      150,
			endHeight:        200,
			currentHeight:    175,
			expectedIsActive: true,
			expectedIsEnded:  false,
		},
		{
			name:             "status inactive and ended after end",
			startHeight:      150,
			endHeight:        200,
			currentHeight:    250,
			expectedIsActive: false,
			expectedIsEnded:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			project := &Project{
				tiers: map[int64]*ProjectTier{
					projectTier180: {
						startBlockTimeInfo: &BlockTimeInfo{
							blockHeight: tt.startHeight,
						},
						endBlockTimeInfo: &BlockTimeInfo{
							blockHeight: tt.endHeight,
						},
					},
				},
			}

			// Execute and verify
			isActive := project.IsActive(tt.currentHeight)
			isEnded := project.IsEnded(tt.currentHeight)

			uassert.Equal(t, isActive, tt.expectedIsActive)
			uassert.Equal(t, isEnded, tt.expectedIsEnded)
		})
	}
}

func TestProject_CheckConditions(t *testing.T) {
	tests := []struct {
		name             string
		conditions       map[string]*ProjectCondition
		caller           std.Address
		balances         map[string]int64
		expectedHasError bool
		expectedError    string
	}{
		{
			name: "check conditions succeed when all conditions met",
			conditions: map[string]*ProjectCondition{
				"gno.land/r/demo/gns": NewProjectCondition("gno.land/r/demo/gns", 1000),
				"gno.land/r/demo/btc": NewProjectCondition("gno.land/r/demo/btc", 500),
			},
			caller: std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			balances: map[string]int64{
				"gno.land/r/demo/gns": 1500,
				"gno.land/r/demo/btc": 1000,
			},
			expectedHasError: false,
		},
		{
			name: "check conditions fail when btc condition not met",
			conditions: map[string]*ProjectCondition{
				"gno.land/r/demo/gns": NewProjectCondition("gno.land/r/demo/gns", 1000),
				"gno.land/r/demo/btc": NewProjectCondition("gno.land/r/demo/btc", 500),
			},
			caller: std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			balances: map[string]int64{
				"gno.land/r/demo/gns": 1500,
				"gno.land/r/demo/btc": 200,
			},
			expectedHasError: true,
			expectedError:    "input amount(200) is less than minimum amount(500)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			project := &Project{
				conditions: tt.conditions,
			}

			mockBalanceFunc := func(tokenPath string, caller std.Address) int64 {
				return tt.balances[tokenPath]
			}

			// Execute
			err := project.CheckConditions(tt.caller, mockBalanceFunc)

			// Verify
			if tt.expectedHasError {
				uassert.ErrorContains(t, err, tt.expectedError)
			} else {
				uassert.NoError(t, err)
			}
		})
	}
}

func TestProject_RemainingAmount(t *testing.T) {
	tests := []struct {
		name             string
		tiers            map[int64]*ProjectTier
		expectedAmount   int64
		expectedHasError bool
		expectedError    string
	}{
		{
			name: "remaining amount return correct total remaining amount",
			tiers: map[int64]*ProjectTier{
				projectTier30: {
					totalDistributeAmount: 1000,
					totalCollectedAmount:  300,
				},
				projectTier90: {
					totalDistributeAmount: 2000,
					totalCollectedAmount:  500,
				},
				projectTier180: {
					totalDistributeAmount: 3000,
					totalCollectedAmount:  1000,
				},
			},
			expectedAmount:   4200, // (1000-300) + (2000-500) + (3000-1000)
			expectedHasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			project := &Project{
				tiers: tt.tiers,
			}

			// Execute
			result := project.RemainingAmount()

			// Verify
			uassert.Equal(t, result, tt.expectedAmount)
		})
	}
}

func TestNewProject(t *testing.T) {
	tests := []struct {
		name             string
		projectName      string
		tokenPath        string
		depositAmount    int64
		recipient        std.Address
		createdTime      int64
		createdHeight    int64
		expectedID       string
		expectedHasError bool
		expectedError    string
	}{
		{
			name:             "new project create valid project with correct values",
			projectName:      "Test Project",
			tokenPath:        "gno.land/r/demo/gns",
			depositAmount:    1000,
			recipient:        std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
			createdTime:      1000,
			createdHeight:    100,
			expectedID:       "gno.land/r/demo/gns:100",
			expectedHasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			blockTimeInfo := &BlockTimeInfo{
				blockTime:   tt.createdTime,
				blockHeight: tt.createdHeight,
			}

			// Execute
			project := NewProject(
				tt.projectName,
				tt.tokenPath,
				tt.depositAmount,
				tt.recipient,
				blockTimeInfo,
			)

			// Verify
			uassert.Equal(t, project.ID(), tt.expectedID)
			uassert.Equal(t, project.Name(), tt.projectName)
			uassert.Equal(t, project.TokenPath(), tt.tokenPath)
			uassert.Equal(t, project.DepositAmount(), tt.depositAmount)
			uassert.Equal(t, project.Recipient(), tt.recipient)
		})
	}
}
