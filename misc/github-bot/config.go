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
	// TODO: remomve that
	CheckedBy string
}

func config(gh *client.Github) ([]automaticCheck, []manualCheck) {
	return []automaticCheck{
			{
				Description: "Changes on 'tm2' folder should be reviewed/authored at least one member of both EU and US teams",
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
			}, {
				Description: "Maintainer must be able to edit this pull request",
				If:          c.Always(),
				Then:        r.MaintainerCanModify(),
			},
		}, []manualCheck{
			{
				Description: "Manual check #1",
				CheckedBy:   "",
			},
			{
				Description: "Manual check #2",
				CheckedBy:   "aeddi",
			},
			{
				Description: "Manual check #3",
				CheckedBy:   "moul",
			},
		}
}
