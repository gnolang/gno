package launchpad

import (
	"gno.land/p/demo/ufmt"
)

// ApiGetProjectAndTierStatisticsByProjectId retrieves project and tier statistics by project ID.
func ApiGetProjectAndTierStatisticsByProjectId(projectId string) string {
	project, exist := projects[projectId]
	if !exist {
		return ""
	}

	builder := MetaBuilder().
		WriteString("projectId", projectId)
	ProjectBuilder(builder, project)

	return marshal(builder.Node())
}

// ApiGetProjectStatisticsByProjectId retrieves project statistics by project ID.
func ApiGetProjectStatisticsByProjectId(projectId string) string {
	project, exist := projects[projectId]
	if !exist {
		return ""
	}

	builder := MetaBuilder().
		WriteString("projectId", projectId)
	ProjectBuilder(builder, project)

	return marshal(builder.Node())
}

// ApiGetProjectStatisticsByProjectTierId retrieves project statistics by project tier ID.
func ApiGetProjectStatisticsByProjectTierId(tierId string) string {
	projectId, tierStr := splitProjectIdAndTier(tierId)
	project, exist := projects[projectId]
	if !exist {
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
		WriteString("tierAmount", ufmt.Sprintf("%d", tier.tierAmount)).
		WriteString("tierTotalDepositAmount", ufmt.Sprintf("%d", tier.totalDepositAmount)).
		WriteString("tierActualDepositAmount", ufmt.Sprintf("%d", tier.actualDepositAmount)).
		WriteString("tierTotalParticipant", ufmt.Sprintf("%d", tier.totalParticipant)).
		WriteString("tierActualParticipant", ufmt.Sprintf("%d", tier.actualParticipant))

	return marshal(builder.Node())
}
