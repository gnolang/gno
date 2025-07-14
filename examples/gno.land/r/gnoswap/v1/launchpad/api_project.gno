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

	builder := metaBuilder().WriteString("projectId", projectId)
	projectBuilder(builder, project)

	for _, duration := range []int64{30, 90, 180} {
		if tier, err := project.getTier(duration); err == nil {
			tierBuilder(builder, ufmt.Sprintf("tier%d", duration), tier)
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

	builder := metaBuilder().WriteString("projectId", projectId)
	projectBuilder(builder, project)

	return marshal(builder.Node())
}

// ApiGetTierStatisticsByProjectId retrieves tier statistics by project ID.
func ApiGetTierStatisticsByProjectId(projectId string) string {
	project, err := getProject(projectId)
	if err != nil {
		return ""
	}

	builder := metaBuilder().WriteString("projectId", projectId)

	for _, duration := range []int64{30, 90, 180} {
		if tier, err := project.getTier(duration); err == nil {
			tierBuilder(builder, ufmt.Sprintf("tier%d", duration), tier)
		}
	}

	return marshal(builder.Node())
}

// ApiGetProjectStatisticsByProjectTierId retrieves project statistics by project tier ID.
func ApiGetProjectStatisticsByProjectTierId(tierId string) string {
	projectId, duration := parseProjectTierID(tierId)
	project, err := getProject(projectId)
	if err != nil {
		return ""
	}

	tier, err := project.getTier(duration)
	if err != nil {
		return ""
	}

	builder := metaBuilder().WriteString("projectId", projectId)
	tierBuilder(builder, "tier", tier)

	return marshal(builder.Node())
}

// ApiGetProjectActiveOf retrieves project active status by project ID.
func ApiGetProjectActiveOf(projectId string) string {
	project, err := getProject(projectId)
	if err != nil {
		return ""
	}
	projectActiveResult := project.IsActive(std.ChainHeight())
	builder := (metaBuilder().
		WriteString("projectId", project.id).
		WriteString("isActive", strconv.FormatBool(projectActiveResult)).
		WriteString("currentHeight", strconv.FormatInt(std.ChainHeight(), 10))).
		WriteString("started.height", strconv.FormatInt(project.standardTier().startBlockTimeInfo.BlockHeight(), 10))
	return marshal(builder.Node())
}
