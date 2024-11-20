package client

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gnolang/gno/contribs/github-bot/logger"
	p "github.com/gnolang/gno/contribs/github-bot/params"

	"github.com/google/go-github/v64/github"
)

// PageSize is the number of items to load for each iteration when fetching a list
const PageSize = 100

type GitHub struct {
	Client *github.Client
	Ctx    context.Context
	DryRun bool
	Logger logger.Logger
	Owner  string
	Repo   string
	cancel context.CancelFunc
}

func (gh *GitHub) GetBotComment(prNum int) *github.IssueComment {
	// List existing comments
	const (
		sort      = "created"
		direction = "desc"
	)

	// Get current user (bot)
	currentUser, _, err := gh.Client.Users.Get(gh.Ctx, "")
	if err != nil {
		gh.Logger.Errorf("Unable to get current user: %v", err)
		return nil
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
			gh.Logger.Errorf("Unable to list comments for PR %d: %v", prNum, err)
			return nil
		}

		// Get the comment created by current user
		for _, comment := range comments {
			if comment.GetUser().GetLogin() == currentUser.GetLogin() {
				return comment
			}
		}

		if response.NextPage == 0 {
			break
		}
		opts.Page = response.NextPage
	}

	return nil
}

func (gh *GitHub) SetBotComment(body string, prNum int) *github.IssueComment {
	// Create bot comment if it does not already exist
	comment := gh.GetBotComment(prNum)
	if comment == nil {
		newComment, _, err := gh.Client.Issues.CreateComment(
			gh.Ctx,
			gh.Owner,
			gh.Repo,
			prNum,
			&github.IssueComment{Body: &body},
		)
		if err != nil {
			gh.Logger.Errorf("Unable to create bot comment for PR %d: %v", prNum, err)
			return nil
		}
		return newComment
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
		gh.Logger.Errorf("Unable to edit bot comment with ID %d: %v", comment.GetID(), err)
		return nil
	}
	return editComment
}

func (gh *GitHub) ListTeamMembers(team string) []*github.User {
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
			gh.Logger.Errorf("Unable to list members for team %s: %v", team, err)
			return nil
		}

		allMembers = append(allMembers, members...)

		if response.NextPage == 0 {
			break
		}
		opts.Page = response.NextPage
	}

	return allMembers
}

func (gh *GitHub) IsUserInTeams(user string, teams []string) bool {
	for _, team := range teams {
		for _, member := range gh.ListTeamMembers(team) {
			if member.GetLogin() == user {
				return true
			}
		}
	}

	return false
}

func (gh *GitHub) ListPRReviewers(prNum int) *github.Reviewers {
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
			gh.Logger.Errorf("Unable to list reviewers for PR %d: %v", prNum, err)
			return nil
		}

		allReviewers.Teams = append(allReviewers.Teams, reviewers.Teams...)
		allReviewers.Users = append(allReviewers.Users, reviewers.Users...)

		if response.NextPage == 0 {
			break
		}
		opts.Page = response.NextPage
	}

	return allReviewers
}

func (gh *GitHub) ListPRReviews(prNum int) []*github.PullRequestReview {
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
			gh.Logger.Errorf("Unable to list reviews for PR %d: %v", prNum, err)
			return nil
		}

		allReviews = append(allReviews, reviews...)

		if response.NextPage == 0 {
			break
		}
		opts.Page = response.NextPage
	}

	return allReviews
}

func (gh *GitHub) Close() {
	if gh.cancel != nil {
		gh.cancel()
	}
}

func New(params *p.Params) *GitHub {
	gh := &GitHub{
		Owner:  params.Owner,
		Repo:   params.Repo,
		DryRun: params.DryRun,
	}

	// Detect if the current process was launched by a GitHub Action and return
	// a logger suitable for terminal output or the GitHub Actions web interface
	gh.Logger = logger.NewLogger(params.Verbose)

	// Create context with timeout if specified in the parameters
	if params.Timeout > 0 {
		gh.Ctx, gh.cancel = context.WithTimeout(context.Background(), time.Duration(params.Timeout)*time.Millisecond)
	} else {
		gh.Ctx = context.Background()
	}

	// Init GitHub API Client using token from env
	token, set := os.LookupEnv("GITHUB_TOKEN")
	if !set {
		log.Fatalf("GITHUB_TOKEN is not set in env")
	}
	gh.Client = github.NewClient(nil).WithAuthToken(token)

	return gh
}
