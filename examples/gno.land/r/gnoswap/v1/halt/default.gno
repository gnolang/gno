package halt

var (
	defaultOperations = []struct {
		opType OpType
		name   string
		desc   string
	}{
		{OpTypeWithdraw, rawWithdraw, "Withdrawal operations"},
	}

	defaultLevels = []struct {
		id      LevelID
		name    string
		desc    string
		allowed map[OpType]bool
	}{
		{
			LvNoHalt,
			rawNoHalt,
			"Normal operation",
			// enabling all operations
			map[OpType]bool{
				OpTypePool:          true,
				OpTypePosition:      true,
				OpTypeProtocolFee:   true,
				OpTypeRouter:        true,
				OpTypeStaker:        true,
				OpTypeLaunchpad:     true,
				OpTypeGovernance:    true,
				OpTypeGovStaker:     true,
				OpTypeXGns:          true,
				OpTypeCommunityPool: true,
				OpTypeEmission:      true,
				OpTypeWithdraw:      true,
			},
		},
		{
			LvContractHalt,
			rawContractHalt,
			"Specific contract operations disabled",
			map[OpType]bool{
				OpTypePool:          true,
				OpTypePosition:      true,
				OpTypeProtocolFee:   true,
				OpTypeRouter:        true,
				OpTypeStaker:        true,
				OpTypeLaunchpad:     true,
				OpTypeGovernance:    true,
				OpTypeGovStaker:     true,
				OpTypeXGns:          true,
				OpTypeCommunityPool: true,
				OpTypeEmission:      true,
				OpTypeWithdraw:      true,
			},
		},
		{
			LvEmergencyHalt,
			rawEmergencyHalt,
			"Only withdrawals allowed",
			map[OpType]bool{
				OpTypeGovernance: true,
				OpTypeWithdraw:   true,
			},
		},
		{
			LvCompleteHalt,
			rawCompleteHalt,
			"All contract disabled",
			map[OpType]bool{
				OpTypeWithdraw: true,
			},
		},
	}
)
