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
			If: c.And(
				c.BaseBranch("^master$"),
				c.CreatedFromFork(),
			),
			Then: r.MaintainerCanModify(),
		},
		{
			Description: "Changes to 'docs' folder must be reviewed/authored by at least one devrel and one tech-staff",
			If: c.And(
				c.BaseBranch("^master$"),
				c.FileChanged(gh, "^docs/"),
			),
			Then: r.And(
				r.Or(
					r.AuthorInTeam(gh, "tech-staff"),
					r.ReviewByTeamMembers(gh, "tech-staff", r.RequestIgnore).WithDesiredState(utils.ReviewStateApproved),
				),
				r.Or(
					r.AuthorInTeam(gh, "devrels"),
					r.ReviewByTeamMembers(gh, "devrels", r.RequestApply).WithDesiredState(utils.ReviewStateApproved),
				),
			),
		},
		{
			Description: "Changes related to gnoweb must be reviewed by its codeowners",
			If: c.And(
				c.BaseBranch("^master$"),
				c.FileChanged(gh, "^gno.land/pkg/gnoweb/"),
			),
			Then: r.Or(
				// If alexiscolin or gfanton is the author of the PR, the other must review it.
				r.Or(
					r.And(
						r.Author("alexiscolin"),
						r.ReviewByUser(gh, "gfanton", r.RequestApply).WithDesiredState(utils.ReviewStateApproved),
					),
					r.And(
						r.Author("gfanton"),
						r.ReviewByUser(gh, "alexiscolin", r.RequestApply).WithDesiredState(utils.ReviewStateApproved),
					),
				),
				// If neither of them is the author of the PR, at least one of them must review it.
				r.And(
					r.Not(r.Author("alexiscolin")),
					r.Not(r.Author("gfanton")),
					r.Or(
						r.ReviewByUser(gh, "alexiscolin", r.RequestApply).WithDesiredState(utils.ReviewStateApproved),
						r.ReviewByUser(gh, "gfanton", r.RequestApply).WithDesiredState(utils.ReviewStateApproved),
					),
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
			If: c.And(
				c.BaseBranch("^master$"),
				c.Not(c.AuthorInTeam(gh, "tech-staff")),
			),
			Then: r.
				// Decide whether to apply the review/triage-pending label.
				// The PR should be either a) approved by any review team member
				// b) reviewed by any member of tech staff
				// c) be a draft
				If(r.Or(
					r.ReviewByAnyUser(gh,
						"jefft0", "notJoon", "omarsy", "MikaelVallenet",
					).WithDesiredState(utils.ReviewStateApproved),
					r.ReviewByTeamMembers(gh, "tech-staff", r.RequestIgnore),
					r.Draft(),
				)).
				Then(
					r.Not(r.Label(gh, "review/triage-pending", r.LabelRemove)),
				).
				Else(
					r.And(
						r.Label(gh, "review/triage-pending", r.LabelApply),
						r.Never(), // Always fail the requirement.
					),
				),
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
				c.BaseBranch("^master$"),
				c.Or(
					c.FileChanged(gh, `Dockerfile`),
					c.FileChanged(gh, `^misc/deployments`),
					c.FileChanged(gh, `^misc/docker-`),
					c.FileChanged(gh, `^.github/workflows/releaser.*\.yml$`),
					c.FileChanged(gh, `^.github/workflows/staging\.yml$`),
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
