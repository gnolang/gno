package client

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/gnolang/gno/contribs/github-bot/logger"
	p "github.com/gnolang/gno/contribs/github-bot/params"

	"github.com/google/go-github/v64/github"
)

// PageSize is the number of items to load for each iteration when fetching a list
const PageSize = 100

var ErrBotCommentNotFound = errors.New("bot comment not found")

type GitHub struct {
	Client *github.Client
	Ctx    context.Context
	DryRun bool
	Logger logger.Logger
	Owner  string
	Repo   string
}

func (gh *GitHub) GetBotComment(prNum int) (*github.IssueComment, error) {
	// List existing comments
	const (
		sort      = "created"
		direction = "desc"
	)

	// Get current user (bot)
	currentUser, _, err := gh.Client.Users.Get(gh.Ctx, "")
	if err != nil {
		return nil, fmt.Errorf("unable to get current user: %v", err)
	}

	// Pagination option
	opts := &github.IssueListCommentsOptions{
		Sort:      github.String(sort),
		Direction: github.String(direction),
		ListOptions: github.ListOptions{
			PerPage: PageSize,
		},
	}

	for {
		comments, response, err := gh.Client.Issues.ListComments(
			gh.Ctx,
			gh.Owner,
			gh.Repo,
			prNum,
			opts,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to list comments for PR %d: %v", prNum, err)
		}

		// Get the comment created by current user
		for _, comment := range comments {
			if comment.GetUser().GetLogin() == currentUser.GetLogin() {
				return comment, nil
			}
		}

		if response.NextPage == 0 {
			break
		}
		opts.Page = response.NextPage
	}

	return nil, errors.New("bot comment not found")
}

func (gh *GitHub) SetBotComment(body string, prNum int) (*github.IssueComment, error) {
	// Create bot comment if it does not already exist
	comment, err := gh.GetBotComment(prNum)
	if err == ErrBotCommentNotFound {
		newComment, _, err := gh.Client.Issues.CreateComment(
			gh.Ctx,
			gh.Owner,
			gh.Repo,
			prNum,
			&github.IssueComment{Body: &body},
		)
		if err != nil {
			return nil, fmt.Errorf("unable to create bot comment for PR %d: %v", prNum, err)
		}
		return newComment, nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to get bot comment: %v", err)
	}

	comment.Body = &body
	editComment, _, err := gh.Client.Issues.EditComment(
		gh.Ctx,
		gh.Owner,
		gh.Repo,
		comment.GetID(),
		comment,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to edit bot comment with ID %d: %v", comment.GetID(), err)
	}

	return editComment, nil
}

func (gh *GitHub) ListTeamMembers(team string) ([]*github.User, error) {
	var (
		allMembers []*github.User
		opts       = &github.TeamListTeamMembersOptions{
			ListOptions: github.ListOptions{
				PerPage: PageSize,
			},
		}
	)

	for {
		members, response, err := gh.Client.Teams.ListTeamMembersBySlug(
			gh.Ctx,
			gh.Owner,
			team,
			opts,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to list members for team %s: %v", team, err)
		}

		allMembers = append(allMembers, members...)

		if response.NextPage == 0 {
			break
		}
		opts.Page = response.NextPage
	}

	return allMembers, nil
}

func (gh *GitHub) IsUserInTeams(user string, teams []string) bool {
	for _, team := range teams {
		teamMembers, err := gh.ListTeamMembers(team)
		if err != nil {
			gh.Logger.Errorf("unable to check if user %s in team %s", user, team)
			continue
		}

		for _, member := range teamMembers {
			if member.GetLogin() == user {
				return true
			}
		}
	}

	return false
}

func (gh *GitHub) ListPRReviewers(prNum int) (*github.Reviewers, error) {
	var (
		allReviewers = &github.Reviewers{}
		opts         = &github.ListOptions{
			PerPage: PageSize,
		}
	)

	for {
		reviewers, response, err := gh.Client.PullRequests.ListReviewers(
			gh.Ctx,
			gh.Owner,
			gh.Repo,
			prNum,
			opts,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to list reviewers for PR %d: %v", prNum, err)
		}

		allReviewers.Teams = append(allReviewers.Teams, reviewers.Teams...)
		allReviewers.Users = append(allReviewers.Users, reviewers.Users...)

		if response.NextPage == 0 {
			break
		}
		opts.Page = response.NextPage
	}

	return allReviewers, nil
}

func (gh *GitHub) ListPRReviews(prNum int) ([]*github.PullRequestReview, error) {
	var (
		allReviews []*github.PullRequestReview
		opts       = &github.ListOptions{
			PerPage: PageSize,
		}
	)

	for {
		reviews, response, err := gh.Client.PullRequests.ListReviews(
			gh.Ctx,
			gh.Owner,
			gh.Repo,
			prNum,
			opts,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to list reviews for PR %d: %v", prNum, err)
		}

		allReviews = append(allReviews, reviews...)

		if response.NextPage == 0 {
			break
		}
		opts.Page = response.NextPage
	}

	return allReviews, nil
}

func New(ctx context.Context, params *p.Params) (*GitHub, error) {
	gh := &GitHub{
		Ctx:    ctx,
		Owner:  params.Owner,
		Repo:   params.Repo,
		DryRun: params.DryRun,
	}

	// Detect if the current process was launched by a GitHub Action and return
	// a logger suitable for terminal output or the GitHub Actions web interface
	gh.Logger = logger.NewLogger(params.Verbose)

	// Retrieve GitHub API token from env
	token, set := os.LookupEnv("GITHUB_TOKEN")
	if !set {
		return nil, errors.New("GITHUB_TOKEN is not set in env")
	}

	// Init GitHub API client using token
	gh.Client = github.NewClient(nil).WithAuthToken(token)

	return gh, nil
}
