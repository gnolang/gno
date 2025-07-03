package conditions

import (
	"testing"

	"github.com/gnolang/gno/contribs/github-bot/internal/utils"
	"github.com/stretchr/testify/assert"

	"github.com/google/go-github/v64/github"
	"github.com/xlab/treeprint"
)

func TestCreatedFromFork(t *testing.T) {
	t.Parallel()

	var (
		repo = &github.PullRequestBranch{Repo: &github.Repository{Owner: &github.User{Login: github.String("main")}, Name: github.String("repo"), FullName: github.String("main/repo")}}
		fork = &github.PullRequestBranch{Repo: &github.Repository{Owner: &github.User{Login: github.String("fork")}, Name: github.String("repo"), FullName: github.String("fork/repo")}}
	)

	prFromMain := &github.PullRequest{Base: repo, Head: repo}
	prFromFork := &github.PullRequest{Base: repo, Head: fork}

	details := treeprint.New()
	assert.False(t, CreatedFromFork().IsMet(prFromMain, details))
	assert.True(t, utils.TestLastNodeStatus(t, false, details), "condition details should have a status: false")

	details = treeprint.New()
	assert.True(t, CreatedFromFork().IsMet(prFromFork, details))
	assert.True(t, utils.TestLastNodeStatus(t, true, details), "condition details should have a status: true")
}
