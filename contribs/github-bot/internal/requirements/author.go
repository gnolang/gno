package requirements

import (
	"github.com/gnolang/gno/contribs/github-bot/internal/client"
	"github.com/gnolang/gno/contribs/github-bot/internal/conditions"

	"github.com/google/go-github/v64/github"
	"github.com/xlab/treeprint"
)

// Author Requirement.
type author struct {
	c conditions.Condition // Alias Author requirement to identical condition.
}

var _ Requirement = &author{}

func (a *author) IsSatisfied(pr *github.PullRequest, details treeprint.Tree) bool {
	return a.c.IsMet(pr, details)
}

func Author(user string) Requirement {
	return &author{conditions.Author(user)}
}

// AuthorInTeam Requirement.
type authorInTeam struct {
	c conditions.Condition // Alias AuthorInTeam requirement to identical condition.
}

var _ Requirement = &authorInTeam{}

func (a *authorInTeam) IsSatisfied(pr *github.PullRequest, details treeprint.Tree) bool {
	return a.c.IsMet(pr, details)
}

func AuthorInTeam(gh *client.GitHub, team string) Requirement {
	return &authorInTeam{conditions.AuthorInTeam(gh, team)}
}
