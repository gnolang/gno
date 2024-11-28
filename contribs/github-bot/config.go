package main

import (
	"github.com/gnolang/gno/contribs/github-bot/internal/client"
	c "github.com/gnolang/gno/contribs/github-bot/internal/conditions"
	r "github.com/gnolang/gno/contribs/github-bot/internal/requirements"
)

type Teams []string

// Automatic check that will be performed by the bot.
type automaticCheck struct {
	description string
	ifC         c.Condition   // If the condition is met, the rule is displayed and the requirement is executed.
	thenR       r.Requirement // If the requirement is satisfied, the check passes.
}

// Manual check that will be performed by users.
type manualCheck struct {
	description string
	ifC         c.Condition // If the condition is met, a checkbox will be displayed on bot comment.
	teams       Teams       // Members of these teams can check the checkbox to make the check pass.
}

// This function returns the configuration of the bot consisting of automatic and manual checks
// in which the GitHub client is injected.
func config(gh *client.GitHub) ([]automaticCheck, []manualCheck) {
	auto := []automaticCheck{
		{
			description: "Maintainers must be able to edit this pull request",
			ifC:         c.Always(),
			thenR:       r.MaintainerCanModify(),
		},
		{
			description: "The pull request head branch must be up-to-date with its base",
			ifC:         c.Always(),
			thenR:       r.UpToDateWith(gh, r.PR_BASE),
		},
		{
			description: "Changes to 'docs' folder must be reviewed/authored by at least one devrel and one tech-staff",
			ifC:         c.FileChanged(gh, "^docs/"),
			thenR: r.Or(
				r.And(
					r.AuthorInTeam(gh, "devrels"),
					r.ReviewByTeamMembers(gh, "tech-staff", 1),
				),
				r.And(
					r.AuthorInTeam(gh, "tech-staff"),
					r.ReviewByTeamMembers(gh, "devrels", 1),
				),
			),
		},
	}

	manual := []manualCheck{
		{
			description: "The pull request description provides enough details",
			ifC:         c.Not(c.AuthorInTeam(gh, "core-contributors")),
			teams:       Teams{"core-contributors"},
		},
		{
			description: "Determine if infra needs to be updated before merging",
			ifC: c.And(
				c.BaseBranch("master"),
				c.Or(
					c.FileChanged(gh, `Dockerfile`),
					c.FileChanged(gh, `^misc/deployments`),
					c.FileChanged(gh, `^misc/docker-`),
					c.FileChanged(gh, `^.github/workflows/releaser.*\.yml$`),
					c.FileChanged(gh, `^.github/workflows/portal-loop\.yml$`),
				),
			),
			teams: Teams{"devops"},
		},
	}

	// Check for duplicates in manual rule descriptions (needs to be unique for the bot operations).
	unique := make(map[string]struct{})
	for _, rule := range manual {
		if _, exists := unique[rule.description]; exists {
			gh.Logger.Fatalf("Manual rule descriptions must be unique (duplicate: %s)", rule.description)
		}
		unique[rule.description] = struct{}{}
	}

	return auto, manual
}
