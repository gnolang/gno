package launchpad

import (
	"std"

	gs "gno.land/r/gnoswap/v1/gov/staker"
)

// ApiGetProjectRecipientRewardByProjectId retrieves the claimable reward for a project recipient by project ID.
func ApiGetProjectRecipientRewardByProjectId(projectId string) string {
	project, exist := projects[projectId]
	if !exist {
		return "0"
	}

	return gs.GetClaimableRewardByAddress(project.recipient)
}

// ApiGetProjectRecipientRewardByAddress retrieves the claimable reward for a recipient by address.
func ApiGetProjectRecipientRewardByAddress(address std.Address) string {
	if !address.IsValid() {
		return "0"
	}

	return gs.GetClaimableRewardByAddress(address)
}

// ApiGetDepositRewardByDepositId calculates the reward for a deposit by its ID.
func ApiGetDepositRewardByDepositId(depositId string) uint64 {
	deposit, exist := deposits[depositId]
	if !exist {
		return 0
	}

	rewardState, err := rewardStates.Get(deposit.projectId, deposit.tier)
	if err != nil {
		return 0
	}

	return rewardState.CalculateReward(depositId)
}

// ApiGetDepositRewardByAddress calculates the total reward for all deposits associated with a given address.
func ApiGetDepositRewardByAddress(address std.Address) uint64 {
	if !address.IsValid() {
		return 0
	}

	depositIds, exist := depositsByUser[address]
	if !exist {
		return 0
	}

	totalReward := uint64(0)
	for _, depositId := range depositIds {
		deposit, exist := deposits[depositId]
		if !exist {
			continue
		}
		rewardState, err := rewardStates.Get(deposit.projectId, deposit.tier)
		if err != nil {
			continue
		}
		totalReward += rewardState.CalculateReward(depositId)
	}

	return totalReward
}
