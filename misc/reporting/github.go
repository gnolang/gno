package main

import (
	"context"

	"github.com/google/go-github/v30/github"
	"golang.org/x/oauth2"
)

func getGithubBacklog(token string) ([]*github.Issue, error) {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)

	callOpts := &github.IssueListByRepoOptions{State: "all"}
	// may need to paginate at some time
	issues, _, err := client.Issues.ListByRepo(context.Background(), "gnolang", "gno", callOpts)
	if err != nil {
		return nil, err
	}

	return issues, nil
}
