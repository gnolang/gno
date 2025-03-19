package launchpad

import (
	"std"
	"strings"

	"gno.land/p/demo/ufmt"
)

// ApiGetClaimableDepositByAddress calculates the total claimable deposit for a given address.
func ApiGetClaimableDepositByAddress(address std.Address) uint64 {
	if !address.IsValid() {
		return 0
	}

	gnsToUser := uint64(0)
	for _, depositId := range depositsByUser[address] {
		deposit := deposits[depositId]

		currentHeight := uint64(std.ChainHeight())
		if (deposit.depositCollectHeight != 0) || (!deposit.isClaimable(currentHeight)) {
			continue
		}

		project, err := getProject(deposit.projectId)
		if err != nil {
			continue
		}

		tier, err := project.Tier(convertTierTypeStrToUint64(deposit.Tier()))
		if err != nil {
			continue
		}
		reward := tier.Reward()
		rewardPerDeposit, err := reward.finalize(currentHeight, tier.ActualDepositAmount(), tier.TierAmountPerBlockX128())
		if err != nil {
			panic(err.Error())
		}
		reward.addRewardPerDeposit(rewardPerDeposit)
		reward.setLastHeight(currentHeight)
		rewardAmount := reward.deductReward(depositId, currentHeight)
		if rewardAmount == 0 {
			continue
		}

		gnsToUser += rewardAmount
	}

	return gnsToUser
}

// ApiGetDepositByDepositId retrieves deposit information by deposit ID.
func ApiGetDepositByDepositId(depositId string) string {
	deposit, exist := deposits[depositId]
	if !exist {
		return ""
	}

	builder := MetaBuilder()
	DepositBuilder(builder, deposit)

	return marshal(builder.Node())
}

// ApiGetDepositFullByDepositId retrieves full deposit information including project and tier details.
func ApiGetDepositFullByDepositId(depositId string) string {
	deposit, exist := deposits[depositId]
	if !exist {
		return ""
	}

	project, exist := projects[deposit.projectId]
	if !exist {
		return ""
	}

	var tier Tier
	switch deposit.tier {
	case "30":
		tier = project.tiers[30]
	case "90":
		tier = project.tiers[90]
	case "180":
		tier = project.tiers[180]
	}

	builder := MetaBuilder()

	// Add project info
	ProjectBuilder(builder, project)

	// Add tier info
	TierBuilder(builder, "", tier)

	// Add deposit info
	DepositBuilder(builder, deposit)

	return marshal(builder.Node())
}

// makeConditionsToStr converts a map of conditions to two strings: token paths and amounts.
func makeConditionsToStr(conditions map[string]Condition) (string, string) {
	var tokenPathList string
	var amountList string

	for tokenPath, condition := range conditions {
		// append with *PAD*, except last one
		tokenPathList += tokenPath + PAD_SEP
		amountList += ufmt.Sprintf("%d", condition.minAmount) + PAD_SEP
	}

	// remove last *PAD*
	tokenPathList = strings.TrimSuffix(tokenPathList, PAD_SEP)
	amountList = strings.TrimSuffix(amountList, PAD_SEP)

	return tokenPathList, amountList
}
