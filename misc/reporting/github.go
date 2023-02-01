package main

import (
	"context"

	"github.com/google/go-github/v50/github"
	"golang.org/x/oauth2"
)

func initGithubClient() *github.Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: opts.githubToken})
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)
	return client
}

func githubFetchIssues(client *github.Client, opts *github.IssueListByRepoOptions, owner string, repository string) ([]*github.Issue, error) {
	issues, _, err := client.Issues.ListByRepo(context.Background(), owner, repository, opts)
	if err != nil {
		return nil, err
	}

	return issues, nil
}
