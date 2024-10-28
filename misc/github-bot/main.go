package main

import (
	"bot/client"
	"bot/param"
	"bytes"
	"text/template"

	"github.com/google/go-github/v66/github"
)

func main() {
	// Get params by parsing CLI flags and/or Github Actions context
	params := param.Get()

	// Init Github API client
	gh := client.New(params)

	// TODO:cleanup
	onCommentUpdated(gh)

	// Get a slice of pull requests to process
	var (
		prs []*github.PullRequest
		err error
	)

	// If requested, get all opened pull requests
	if params.PrAll {
		opts := &github.PullRequestListOptions{
			State:     "open",
			Sort:      "updated",
			Direction: "desc",
		}

		prs, _, err = gh.Client.PullRequests.List(gh.Ctx, gh.Owner, gh.Repo, opts)
		if err != nil {
			gh.Logger.Fatalf("Unable to get all opened pull requests : %v", err)
		}

		// Or get only specified pull request(s) (flag or Github Action context)
	} else {
		prs = make([]*github.PullRequest, len(params.PrNums))
		for i, prNum := range params.PrNums {
			pr, _, err := gh.Client.PullRequests.Get(gh.Ctx, gh.Owner, gh.Repo, prNum)
			if err != nil {
				gh.Logger.Fatalf("Unable to get specified pull request (%d) : %v", prNum, err)
			}
			prs[i] = pr
		}
	}

	tmplFile := "comment.tmpl"
	tmpl, err := template.New(tmplFile).ParseFiles(tmplFile)
	if err != nil {
		panic(err)
	}

	auto, manual := config(gh)
	// Process all pull requests
	for _, pr := range prs {
		com := Comment{}
		for _, rule := range auto {
			if rule.If.Validate(pr) {
				gh.Logger.Infof(rule.If.GetText())

				c := Auto{Description: rule.Description, Met: false}

				if !rule.Then.Validate(pr) {
					gh.Logger.Infof(rule.Then.GetText())
					c.Met = true
				}

				com.Auto = append(com.Auto, c)
			}
		}

		for _, rule := range manual {
			com.Manual = append(com.Manual, Manual{
				Description: rule.Description,
				CheckedBy:   rule.CheckedBy,
			})
		}

		var commentBytes bytes.Buffer
		err = tmpl.Execute(&commentBytes, com)
		if err != nil {
			panic(err)
		}

		comment := gh.SetBotComment(commentBytes.String(), pr.GetNumber())

		context := "Merge Requirements"
		state := "pending"
		targetURL := comment.GetHTMLURL()
		description := "Some requirements are not met yet. See bot comment."

		if _, _, err := gh.Client.Repositories.CreateStatus(
			gh.Ctx,
			gh.Owner,
			gh.Repo,
			pr.GetHead().GetSHA(),
			&github.RepoStatus{
				Context:     &context,
				State:       &state,
				TargetURL:   &targetURL,
				Description: &description,
			}); err != nil {
			gh.Logger.Errorf("Unable to create status on PR %d : %v", pr.GetNumber(), err)
		}
	}
}
