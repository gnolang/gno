package launchpad

import (
	"std"
	"strings"
	"time"

	"gno.land/p/demo/json"
	"gno.land/p/demo/ufmt"
)

// timeInfoBuilder adds TimeInfo fields to JSON
func timeInfoBuilder(b *json.NodeBuilder, prefix string, info *BlockTimeInfo) *json.NodeBuilder {
	return b.
		WriteString(prefix+"Height", ufmt.Sprintf("%d", info.BlockHeight())).
		WriteString(prefix+"Time", ufmt.Sprintf("%d", info.BlockTime()))
}

// projectStatsBuilder adds ProjectStats fields to JSON
func projectStatsBuilder(b *json.NodeBuilder, project *Project) *json.NodeBuilder {
	return b.
		WriteString("totalDeposit", ufmt.Sprintf("%d", project.TotalDepositAmount())).
		WriteString("actualDeposit", ufmt.Sprintf("%d", project.CurrentDepositAmount())).
		WriteString("totalParticipant", ufmt.Sprintf("%d", project.TotalDepositCount())).
		WriteString("actualParticipant", ufmt.Sprintf("%d", project.CurrentDepositCount())).
		WriteString("totalCollected", ufmt.Sprintf("%d", project.TotalCollectedAmount()))
}

// refundInfoBuilder adds RefundInfo fields to JSON
func refundInfoBuilder(b *json.NodeBuilder, project *Project) *json.NodeBuilder {
	return b.
		WriteString("refundedAmount", ufmt.Sprintf("%d", project.RemainingAmount()))
}

// TierBuilder adds Tier fields to JSON
func tierBuilder(b *json.NodeBuilder, prefix string, tier *ProjectTier) *json.NodeBuilder {
	// Add tiers info
	b.WriteString(prefix+"Id", tier.id)
	b.WriteString(prefix+"TierAmount", ufmt.Sprintf("%d", tier.TotalDistributeAmount()))
	b.WriteString(prefix+"TierAmountPerBlockX128", tier.DistributeAmountPerBlockX128().ToString())
	timeInfoBuilder(b, prefix+"Started", tier.startBlockTimeInfo)
	timeInfoBuilder(b, prefix+"Ended", tier.endBlockTimeInfo)
	b.WriteString(prefix+"TotalDepositAmount", ufmt.Sprintf("%d", tier.TotalDepositAmount()))
	b.WriteString(prefix+"ActualDepositAmount", ufmt.Sprintf("%d", tier.CurrentDepositAmount()))
	b.WriteString(prefix+"TotalParticipant", ufmt.Sprintf("%d", tier.TotalDepositCount()))
	b.WriteString(prefix+"ActualParticipant", ufmt.Sprintf("%d", tier.CurrentDepositCount()))
	b.WriteString(prefix+"UserCollectedAmount", ufmt.Sprintf("%d", tier.TotalCollectedAmount()))
	return b
}

// ProjectBuilder adds Project fields to JSON
func projectBuilder(b *json.NodeBuilder, project *Project) *json.NodeBuilder {
	b.WriteString("name", project.name)
	b.WriteString("tokenPath", project.tokenPath)
	b.WriteString("depositAmount", ufmt.Sprintf("%d", project.depositAmount))
	b.WriteString("recipient", project.recipient.String())

	tokenPaths := []string{}
	amounts := []string{}

	for _, condition := range project.getConditions() {
		tokenPaths = append(tokenPaths, condition.TokenPath())
		amounts = append(amounts, ufmt.Sprintf("%d", condition.MinimumAmount()))
	}
	b.WriteString("conditionsToken", strings.Join(tokenPaths, ","))
	b.WriteString("conditionsAmount", strings.Join(amounts, ","))

	// Add time info
	timeInfoBuilder(b, "created", project.CreatedBlockTimeInfo())
	timeInfoBuilder(b, "started", project.standardTier().startBlockTimeInfo)
	timeInfoBuilder(b, "ended", project.standardTier().endBlockTimeInfo)

	// Add refund info
	refundInfoBuilder(b, project)

	return b
}

// DepositBuilder adds Deposit fields to JSON
func depositBuilder(b *json.NodeBuilder, deposit *Deposit) *json.NodeBuilder {
	return b.
		WriteString("depositId", deposit.id).
		WriteString("projectId", deposit.ProjectID()).
		WriteString("tier", ufmt.Sprintf("%d", deposit.Tier())).
		WriteString("depositor", deposit.Depositor().String()).
		WriteString("amount", ufmt.Sprintf("%d", deposit.DepositAmount())).
		WriteString("depositHeight", ufmt.Sprintf("%d", deposit.DepositHeight())).
		WriteString("depositTime", ufmt.Sprintf("%d", deposit.DepositTime()))
}

// MetaBuilder adds metadata fields to JSON
func metaBuilder() *json.NodeBuilder {
	height := std.ChainHeight()
	now := time.Now().Unix()

	return json.Builder().
		WriteString("height", ufmt.Sprintf("%d", height)).
		WriteString("now", ufmt.Sprintf("%d", now))
}

// Marshals a JSON node to a string, panics if marshalling fails
func marshal(data *json.Node) string {
	b, err := json.Marshal(data)
	if err != nil {
		panic(err.Error())
	}

	return string(b)
}
