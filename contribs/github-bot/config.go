package main

import (
	"github.com/gnolang/gno/contribs/github-bot/internal/client"
	c "github.com/gnolang/gno/contribs/github-bot/internal/conditions"
	r "github.com/gnolang/gno/contribs/github-bot/internal/requirements"
)

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
	teams       []string    // Members of these teams can check the checkbox to make the check pass.
}

// This function returns the configuration of the bot consisting of automatic and manual checks
// in which the GitHub client is injected.
func config(gh *client.GitHub) ([]automaticCheck, []manualCheck) {
	auto := []automaticCheck{
		{
			description: "Changes to 'tm2' folder should be reviewed/authored by at least one member of both EU and US teams",
			ifC: c.And(
				c.FileChanged(gh, "tm2"),
				c.BaseBranch("master"),
			),
			thenR: r.And(
				r.Or(
					r.ReviewByTeamMembers(gh, "eu", 1),
					r.AuthorInTeam(gh, "eu"),
				),
				r.Or(
					r.ReviewByTeamMembers(gh, "us", 1),
					r.AuthorInTeam(gh, "us"),
				),
			),
		},
		{
			description: "A maintainer must be able to edit this pull request",
			ifC:         c.Always(),
			thenR:       r.MaintainerCanModify(),
		},
		{
			description: "The pull request head branch must be up-to-date with its base",
			ifC:         c.Always(), // Or only if c.BaseBranch("main") ?
			thenR:       r.UpToDateWith(gh, r.PR_BASE),
		},
	}

	manual := []manualCheck{
		{
			description: "Determine if infra needs to be updated",
			ifC: c.And(
				c.BaseBranch("master"),
				c.Or(
					c.FileChanged(gh, "misc/deployments"),
					c.FileChanged(gh, `misc/docker-\.*`),
					c.FileChanged(gh, "tm2/pkg/p2p"),
				),
			),
			teams: []string{"tech-staff"},
		},
		{
			description: "Ensure the code style is satisfactory",
			ifC: c.And(
				c.BaseBranch("master"),
				c.Or(
					c.FileChanged(gh, `.*\.go`),
					c.FileChanged(gh, `.*\.js`),
				),
			),
			teams: []string{"tech-staff"},
		},
		{
			description: "Ensure the documentation is accurate and relevant",
			ifC:         c.FileChanged(gh, `.*\.md`),
			teams: []string{
				"tech-staff",
				"devrels",
			},
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
