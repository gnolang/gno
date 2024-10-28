package requirement

import (
	"bot/client"
	"bot/condition"
)

func Author(user string) Requirement {
	return condition.Author(user)
}

func AuthorInTeam(gh *client.Github, team string) Requirement {
	return condition.AuthorInTeam(gh, team)
}
