package launchpad

import (
	"std"

	"gno.land/p/demo/ufmt"
	gov_staker "gno.land/r/gnoswap/v1/gov/staker"
)

// CollectProtocolFee collects protocol fee from gov/staker
// each project's recipient wallet will be rewarded
// ref: https://docs.gnoswap.io/contracts/launchpad/launchpad_reward.gno#collectprotocolfee
func CollectProtocolFee(cur realm) {
	callerAddress := std.PreviousRealm().Address()
	hasProject := false

	// check if the caller has a project,
	// if owner, break the loop
	projects.Iterate("", "", func(key string, value interface{}) bool {
		project := value.(*Project)
		hasProject = project.IsOwner(callerAddress)

		// if true, break the loop
		return hasProject
	})

	if !hasProject {
		panic(makeErrorWithDetails(
			errInvalidOwner,
			ufmt.Sprintf("caller %s is not the owner of any project", callerAddress.String()),
		))
	}

	gov_staker.CollectRewardFromLaunchPad(cross, callerAddress)
}
