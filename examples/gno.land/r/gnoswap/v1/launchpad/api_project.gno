package launchpad

import (
	"std"
	"strconv"

	"gno.land/p/demo/ufmt"
)

// ApiGetProjectAndTierStatisticsByProjectId retrieves project and tier statistics by project ID.
func ApiGetProjectAndTierStatisticsByProjectId(projectId string) string {
	project, err := getProject(projectId)
	if err != nil {
		return ""
	}

	builder := MetaBuilder().WriteString("projectId", projectId)
	ProjectBuilder(builder, project)

	for _, duration := range []uint64{30, 90, 180} {
		if tier, exists := project.tiers[duration]; exists {
			TierBuilder(builder, ufmt.Sprintf("tier%d", duration), tier)
		}
	}

	return marshal(builder.Node())
}

// ApiGetProjectStatisticsByProjectId retrieves project statistics by project ID.
func ApiGetProjectStatisticsByProjectId(projectId string) string {
	project, err := getProject(projectId)
	if err != nil {
		return ""
	}

	builder := MetaBuilder().WriteString("projectId", projectId)
	ProjectBuilder(builder, project)

	return marshal(builder.Node())
}

// ApiGetTierStatisticsByProjectId retrieves tier statistics by project ID.
func ApiGetTierStatisticsByProjectId(projectId string) string {
	project, err := getProject(projectId)
	if err != nil {
		return ""
	}

	builder := MetaBuilder().WriteString("projectId", projectId)

	for _, duration := range []uint64{30, 90, 180} {
		if tier, exists := project.tiers[duration]; exists {
			TierBuilder(builder, ufmt.Sprintf("tier%d", duration), tier)
		}
	}

	return marshal(builder.Node())
}

// ApiGetProjectStatisticsByProjectTierId retrieves project statistics by project tier ID.
func ApiGetProjectStatisticsByProjectTierId(tierId string) string {
	projectId, tierStr := splitProjectIdAndTier(tierId)
	project, err := getProject(projectId)
	if err != nil {
		return ""
	}

	var tier Tier
	switch tierStr {
	case "30":
		tier = project.tiers[30]
	case "90":
		tier = project.tiers[90]
	case "180":
		tier = project.tiers[180]
	default:
		return ""
	}

	builder := MetaBuilder().
		WriteString("projectId", projectId).
		WriteString("tierId", tierId).
		WriteString("collectWaitDuration", strconv.FormatUint(tier.collectWaitDuration, 10)).
		WriteString("tierAmount", strconv.FormatUint(tier.tierAmount, 10)).
		WriteString("tierAmountPerBlockX128", tier.tierAmountPerBlockX128.ToString()).
		WriteString("started.height", strconv.FormatUint(tier.started.height, 10)).
		WriteString("started.time", strconv.FormatUint(tier.started.time, 10)).
		WriteString("ended.height", strconv.FormatUint(tier.ended.height, 10)).
		WriteString("ended.time", strconv.FormatUint(tier.ended.time, 10)).
		WriteString("tierTotalDepositAmount", strconv.FormatUint(tier.totalDepositAmount, 10)).
		WriteString("tierActualDepositAmount", strconv.FormatUint(tier.actualDepositAmount, 10)).
		WriteString("tierTotalParticipant", strconv.FormatUint(tier.totalParticipant, 10)).
		WriteString("tierActualParticipant", strconv.FormatUint(tier.actualParticipant, 10)).
		WriteString("userCollectedAmount", strconv.FormatUint(tier.userCollectedAmount, 10)).
		WriteString("calculatedAmount", strconv.FormatUint(tier.calculatedAmount, 10))

	return marshal(builder.Node())
}

// ApiGetProjectActiveOf retrieves project active status by project ID.
func ApiGetProjectActiveOf(projectId string) string {
	project, err := getProject(projectId)
	if err != nil {
		return ""
	}
	projectActiveResult := isProjectActive(project, uint64(std.ChainHeight()))
	builder := (MetaBuilder().
		WriteString("projectId", project.id).
		WriteString("isActive", strconv.FormatBool(projectActiveResult)).
		WriteString("currentHeight", strconv.FormatInt(std.ChainHeight(), 10))).
		WriteString("started.height", strconv.FormatUint(project.started.height, 10))
	return marshal(builder.Node())
}
