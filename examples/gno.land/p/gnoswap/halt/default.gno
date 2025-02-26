package halt

var (
	defaultOperations = []struct {
		opType OpType
		name   string
		desc   string
	}{
		{OpTypeSwap, rawSwap, "Token swap operation"},
		{OpTypeLiquidity, rawLiquidity, "Liquidity provision operations"},
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
			map[OpType]bool{
				OpTypeSwap:      true,
				OpTypeLiquidity: true,
				OpTypeWithdraw:  true,
			},
		},
		{
			LvSwapHalt,
			rawSwapHalt,
			"Swaps disabled",
			map[OpType]bool{
				OpTypeSwap:      false,
				OpTypeLiquidity: true,
				OpTypeWithdraw:  true,
			},
		},
		{
			LvEmergencyHalt,
			rawEmergencyHalt,
			"Only withdrawals allowed",
			map[OpType]bool{
				OpTypeSwap:      false,
				OpTypeLiquidity: false,
				OpTypeWithdraw:  true,
			},
		},
		{
			LvCompleteHalt,
			rawCompleteHalt,
			"All operations disabled",
			map[OpType]bool{
				OpTypeSwap:      false,
				OpTypeLiquidity: false,
				OpTypeWithdraw:  false,
			},
		},
	}
)
