package launchpad

import (
	"std"

	gs "gno.land/r/gnoswap/v1/gov/staker"
)

// ApiGetProjectRecipientRewardByProjectId retrieves the claimable reward for a project recipient by project ID.
func ApiGetProjectRecipientRewardByProjectId(projectId string) string {
	project, exist := projects.Get(projectId)
	if !exist {
		return "0"
	}

	return gs.GetClaimableRewardByAddress(project.(*Project).recipient)
}

// ApiGetProjectRecipientRewardByAddress retrieves the claimable reward for a recipient by address.
func ApiGetProjectRecipientRewardByAddress(address std.Address) string {
	if !address.IsValid() {
		return "0"
	}

	return gs.GetClaimableRewardByAddress(address)
}
