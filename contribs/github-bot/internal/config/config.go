package config

import (
	"github.com/gnolang/gno/contribs/github-bot/internal/client"
	c "github.com/gnolang/gno/contribs/github-bot/internal/conditions"
	r "github.com/gnolang/gno/contribs/github-bot/internal/requirements"
	"github.com/gnolang/gno/contribs/github-bot/internal/utils"
)

type Teams []string

// Automatic check that will be performed by the bot.
type AutomaticCheck struct {
	Description string
	If          c.Condition   // If the condition is met, the rule is displayed and the requirement is executed.
	Then        r.Requirement // If the requirement is satisfied, the check passes.
}

// Manual check that will be performed by users.
type ManualCheck struct {
	Description string
	If          c.Condition // If the condition is met, a checkbox will be displayed on bot comment.
	Teams       Teams       // Members of these teams can check the checkbox to make the check pass.
}

// This is the description for a persistent rule with a non-standard behavior
// that allow maintainer to force the "success" state of the CI check
const ForceSkipDescription = "**IGNORE** the bot requirements for this PR (force green CI check)"

// This function returns the configuration of the bot consisting of automatic and manual checks
// in which the GitHub client is injected.
func Config(gh *client.GitHub) ([]AutomaticCheck, []ManualCheck) {
	auto := []AutomaticCheck{
		{
			Description: "Maintainers must be able to edit this pull request ([more info](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/working-with-forks/allowing-changes-to-a-pull-request-branch-created-from-a-fork))",
			If:          c.CreatedFromFork(),
			Then:        r.MaintainerCanModify(),
		},
		{
			Description: "Changes to 'docs' folder must be reviewed/authored by at least one devrel and one tech-staff",
			If:          c.FileChanged(gh, "^docs/"),
			Then: r.And(
				r.Or(
					r.AuthorInTeam(gh, "tech-staff"),
					r.ReviewByTeamMembers(gh, "tech-staff").WithDesiredState(utils.ReviewStateApproved),
				),
				r.Or(
					r.AuthorInTeam(gh, "devrels"),
					r.ReviewByTeamMembers(gh, "devrels").WithDesiredState(utils.ReviewStateApproved),
				),
			),
		},
		{
			Description: "Must not contain the \"don't merge\" label",
			If:          c.Label("don't merge"),
			Then:        r.Never(),
		},
		{
			Description: "Pending initial approval by a review team member, or review from tech-staff",
			If:          c.Not(c.AuthorInTeam(gh, "tech-staff")),
			Then: r.
				If(r.Or(
					r.ReviewByOrgMembers(gh).WithDesiredState(utils.ReviewStateApproved),
					r.ReviewByTeamMembers(gh, "tech-staff"),
					r.Draft(),
				)).
				// Either there was a first approval from a member, and we
				// assert that the label for triage-pending is removed...
				Then(r.Not(r.Label(gh, "review/triage-pending", r.LabelRemove))).
				// Or there was not, and we apply the triage pending label.
				// The requirement should always fail, to mark the PR is not
				// ready to be merged.
				Else(r.And(r.Label(gh, "review/triage-pending", r.LabelApply), r.Never())),
		},
	}

	manual := []ManualCheck{
		{
			// WARN: Do not edit this special rule which must remain persistent.
			Description: ForceSkipDescription,
			If:          c.Always(),
		},
		{
			Description: "The pull request description provides enough details",
			If: c.And(
				c.Not(c.AuthorInTeam(gh, "core-contributors")),
				c.Not(c.Author("dependabot[bot]")),
			),
			Teams: Teams{"core-contributors"},
		},
		{
			Description: "Determine if infra needs to be updated before merging",
			If: c.And(
				c.BaseBranch("master"),
				c.Or(
					c.FileChanged(gh, `Dockerfile`),
					c.FileChanged(gh, `^misc/deployments`),
					c.FileChanged(gh, `^misc/docker-`),
					c.FileChanged(gh, `^.github/workflows/releaser.*\.yml$`),
					c.FileChanged(gh, `^.github/workflows/portal-loop\.yml$`),
				),
			),
			Teams: Teams{"devops"},
		},
	}

	// Check for duplicates in manual rule descriptions (needs to be unique for the bot operations).
	unique := make(map[string]struct{})
	for _, rule := range manual {
		if _, exists := unique[rule.Description]; exists {
			gh.Logger.Fatalf("Manual rule descriptions must be unique (duplicate: %s)", rule.Description)
		}
		unique[rule.Description] = struct{}{}
	}

	return auto, manual
}
