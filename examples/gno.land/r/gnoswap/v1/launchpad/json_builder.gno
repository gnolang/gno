package launchpad

import (
	"std"
	"time"

	"gno.land/p/demo/json"
	"gno.land/p/demo/ufmt"
)

// TimeInfoBuilder adds TimeInfo fields to JSON
func TimeInfoBuilder(b *json.NodeBuilder, prefix string, info TimeInfo) *json.NodeBuilder {
	return b.
		WriteString(prefix+"Height", ufmt.Sprintf("%d", info.height)).
		WriteString(prefix+"Time", ufmt.Sprintf("%d", info.time))
}

// ProjectStatsBuilder adds ProjectStats fields to JSON
func ProjectStatsBuilder(b *json.NodeBuilder, stats ProjectStats) *json.NodeBuilder {
	return b.
		WriteString("totalDeposit", ufmt.Sprintf("%d", stats.totalDeposit)).
		WriteString("actualDeposit", ufmt.Sprintf("%d", stats.actualDeposit)).
		WriteString("totalParticipant", ufmt.Sprintf("%d", stats.totalParticipant)).
		WriteString("actualParticipant", ufmt.Sprintf("%d", stats.actualParticipant)).
		WriteString("totalCollected", ufmt.Sprintf("%d", stats.totalCollected))
}

// RefundInfoBuilder adds RefundInfo fields to JSON
func RefundInfoBuilder(b *json.NodeBuilder, info RefundInfo) *json.NodeBuilder {
	return b.
		WriteString("refundedAmount", ufmt.Sprintf("%d", info.amount)).
		WriteString("refundedHeight", ufmt.Sprintf("%d", info.height)).
		WriteString("refundedTime", ufmt.Sprintf("%d", info.time))
}

// TierBuilder adds Tier fields to JSON
func TierBuilder(b *json.NodeBuilder, prefix string, tier Tier) *json.NodeBuilder {
	// Add tiers info
	b.WriteString(prefix+"Id", tier.id)
	b.WriteString(prefix+"CollectWaitDuration", ufmt.Sprintf("%d", tier.collectWaitDuration))
	b.WriteString(prefix+"TierAmount", ufmt.Sprintf("%d", tier.tierAmount))
	if tier.tierAmountPerBlockX128 != nil {
		b.WriteString(prefix+"TierAmountPerBlockX128", tier.tierAmountPerBlockX128.ToString())
	}
	TimeInfoBuilder(b, prefix+"Started", tier.started)
	TimeInfoBuilder(b, prefix+"Ended", tier.ended)
	b.WriteString(prefix+"TotalDepositAmount", ufmt.Sprintf("%d", tier.totalDepositAmount))
	b.WriteString(prefix+"ActualDepositAmount", ufmt.Sprintf("%d", tier.actualDepositAmount))
	b.WriteString(prefix+"TotalParticipant", ufmt.Sprintf("%d", tier.totalParticipant))
	b.WriteString(prefix+"ActualParticipant", ufmt.Sprintf("%d", tier.actualParticipant))
	b.WriteString(prefix+"UserCollectedAmount", ufmt.Sprintf("%d", tier.userCollectedAmount))
	b.WriteString(prefix+"CalculatedAmount", ufmt.Sprintf("%d", tier.calculatedAmount))
	return b
}

// ProjectBuilder adds Project fields to JSON
func ProjectBuilder(b *json.NodeBuilder, project Project) *json.NodeBuilder {
	b.WriteString("name", project.name)
	b.WriteString("tokenPath", project.tokenPath)
	b.WriteString("depositAmount", ufmt.Sprintf("%d", project.depositAmount))
	b.WriteString("recipient", project.recipient.String())

	tokenPathList, amountList := makeConditionsToStr(project.conditions)
	b.WriteString("conditionsToken", tokenPathList)
	b.WriteString("conditionsAmount", amountList)

	// Add tier ratios
	for _, duration := range []uint64{30, 90, 180} {
		b.WriteString(ufmt.Sprintf("tier%dRatio", duration),
			ufmt.Sprintf("%d", project.tiersRatios[duration]))
	}

	// Add time info
	TimeInfoBuilder(b, "created", project.created)
	TimeInfoBuilder(b, "started", project.started)
	TimeInfoBuilder(b, "ended", project.ended)

	// Add stats
	ProjectStatsBuilder(b, project.stats)

	// Add refund info
	RefundInfoBuilder(b, project.refund)

	return b
}

// DepositBuilder adds Deposit fields to JSON
func DepositBuilder(b *json.NodeBuilder, deposit Deposit) *json.NodeBuilder {
	return b.
		WriteString("depositId", deposit.id).
		WriteString("projectId", deposit.projectId).
		WriteString("tier", deposit.tier).
		WriteString("depositor", deposit.depositor.String()).
		WriteString("amount", ufmt.Sprintf("%d", deposit.amount)).
		WriteString("depositHeight", ufmt.Sprintf("%d", deposit.depositHeight)).
		WriteString("depositTime", ufmt.Sprintf("%d", deposit.depositTime)).
		WriteString("depositCollectHeight", ufmt.Sprintf("%d", deposit.depositCollectHeight)).
		WriteString("depositCollectTime", ufmt.Sprintf("%d", deposit.depositCollectTime)).
		WriteString("claimableHeight", ufmt.Sprintf("%d", deposit.claimableHeight)).
		WriteString("claimableTime", ufmt.Sprintf("%d", deposit.claimableTime)).
		WriteString("claimableRewardAmount", ufmt.Sprintf("%d", deposit.rewardAmount)).
		WriteString("rewardCollected", ufmt.Sprintf("%d", deposit.rewardCollected)).
		WriteString("rewardCollectHeight", ufmt.Sprintf("%d", deposit.rewardCollectHeight)).
		WriteString("rewardCollectTime", ufmt.Sprintf("%d", deposit.rewardCollectTime))
}

// MetaBuilder adds metadata fields to JSON
func MetaBuilder() *json.NodeBuilder {
	height := std.ChainHeight()
	now := time.Now().Unix()

	return json.Builder().
		WriteString("height", ufmt.Sprintf("%d", height)).
		WriteString("now", ufmt.Sprintf("%d", now))
}
