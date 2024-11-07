package main

import (
	"bot/client"
	c "bot/condition"
	r "bot/requirement"
)

type automaticCheck struct {
	Description string
	If          c.Condition
	Then        r.Requirement
}

type manualCheck struct {
	Description string
	If          c.Condition
	Teams       []string
}

func config(gh *client.GitHub) ([]automaticCheck, []manualCheck) {
	auto := []automaticCheck{
		{
			Description: "Changes to 'tm2' folder should be reviewed/authored by at least one member of both EU and US teams",
			If: c.And(
				c.FileChanged(gh, "tm2"),
				c.BaseBranch("main"),
			),
			Then: r.And(
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
			Description: "A maintainer must be able to edit this pull request",
			If:          c.Always(),
			Then:        r.MaintainerCanModify(),
		},
		{
			Description: "The pull request head branch must be up-to-date with its base",
			If:          c.Always(), // Or only if c.BaseBranch("main") ?
			Then:        r.UpToDateWith(gh, r.PR_BASE),
		},
	}

	manual := []manualCheck{
		{
			Description: "Determine if infra needs to be updated",
			If: c.And(
				c.BaseBranch("main"),
				c.Or(
					c.FileChanged(gh, "misc/deployments"),
					c.FileChanged(gh, `misc/docker-\.*`),
					c.FileChanged(gh, "tm2/pkg/p2p"),
				),
			),
			Teams: []string{"tech-staff"},
		},
		{
			Description: "Ensure the code style is satisfactory",
			If: c.And(
				c.BaseBranch("main"),
				c.Or(
					c.FileChanged(gh, `.*\.go`),
					c.FileChanged(gh, `.*\.js`),
				),
			),
			Teams: []string{"tech-staff"},
		},
		{
			Description: "Ensure the documentation is accurate and relevant",
			If:          c.FileChanged(gh, `.*\.md`),
			Teams: []string{
				"tech-staff",
				"devrels",
			},
		},
	}

	// Check for duplicates in manual rule descriptions
	// (needs to be unique for the bot operations)
	unique := make(map[string]struct{})
	for _, rule := range manual {
		if _, exists := unique[rule.Description]; exists {
			gh.Logger.Fatalf("Manual rule descriptions must be unique (duplicate: %s)", rule.Description)
		}
		unique[rule.Description] = struct{}{}
	}

	return auto, manual
}
