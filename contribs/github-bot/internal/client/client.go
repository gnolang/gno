package client

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/gnolang/gno/contribs/github-bot/internal/logger"
	"github.com/gnolang/gno/contribs/github-bot/internal/utils"
	"github.com/google/go-github/v64/github"
)

// PageSize is the number of items to load for each iteration when fetching a list.
const PageSize = 100

var ErrBotCommentNotFound = errors.New("bot comment not found")

// GitHub contains everything necessary to interact with the GitHub API,
// including the client, a context (which must be passed with each request),
// a logger, etc. This object will be passed to each condition or requirement
// that requires fetching additional information or modifying things on GitHub.
// The object also provides methods for performing more complex operations than
// a simple API call.
type GitHub struct {
	Client *github.Client
	Ctx    context.Context
	DryRun bool
	Logger logger.Logger
	Owner  string
	Repo   string
}

type Config struct {
	Owner   string
	Repo    string
	Verbose bool
	DryRun  bool
}

// GetBotComment retrieves the bot's (current user) comment on provided PR number.
func (gh *GitHub) GetBotComment(prNum int) (*github.IssueComment, error) {
	// List existing comments.
	const (
		sort      = "created"
		direction = "desc"
	)

	// Get current user (bot).
	currentUser, _, err := gh.Client.Users.Get(gh.Ctx, "")
	if err != nil {
		return nil, fmt.Errorf("unable to get current user: %w", err)
	}

	// Pagination option.
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
			return nil, fmt.Errorf("unable to list comments for PR %d: %w", prNum, err)
		}

		// Get the comment created by current user.
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

	return nil, ErrBotCommentNotFound
}

// SetBotComment creates a bot's comment on the provided PR number
// or updates it if it already exists.
func (gh *GitHub) SetBotComment(body string, prNum int) (*github.IssueComment, error) {
	// Prevent updating anything in dry run mode.
	if gh.DryRun {
		return nil, errors.New("should not write bot comment in dry run mode")
	}

	// Create bot comment if it does not already exist.
	comment, err := gh.GetBotComment(prNum)
	if errors.Is(err, ErrBotCommentNotFound) {
		newComment, _, err := gh.Client.Issues.CreateComment(
			gh.Ctx,
			gh.Owner,
			gh.Repo,
			prNum,
			&github.IssueComment{Body: &body},
		)
		if err != nil {
			return nil, fmt.Errorf("unable to create bot comment for PR %d: %w", prNum, err)
		}
		return newComment, nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to get bot comment: %w", err)
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
		return nil, fmt.Errorf("unable to edit bot comment with ID %d: %w", comment.GetID(), err)
	}

	return editComment, nil
}

func (gh *GitHub) GetOpenedPullRequest(prNum int) (*github.PullRequest, error) {
	pr, _, err := gh.Client.PullRequests.Get(gh.Ctx, gh.Owner, gh.Repo, prNum)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve specified pull request (%d): %w", prNum, err)
	} else if pr.GetState() != utils.PRStateOpen {
		return nil, fmt.Errorf("pull request %d is not opened, actual state: %s", prNum, pr.GetState())
	}

	return pr, nil
}

// ListTeamMembers lists the members of the specified team.
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
			return nil, fmt.Errorf("unable to list members for team %s: %w", team, err)
		}

		allMembers = append(allMembers, members...)

		if response.NextPage == 0 {
			break
		}
		opts.Page = response.NextPage
	}

	return allMembers, nil
}

// IsUserInTeams checks if the specified user is a member of any of the
// provided teams.
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

// ListPRReviewers returns the list of reviewers for the specified PR number.
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
			return nil, fmt.Errorf("unable to list reviewers for PR %d: %w", prNum, err)
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

// ListPRReviewers returns the list of reviews for the specified PR number.
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
			return nil, fmt.Errorf("unable to list reviews for PR %d: %w", prNum, err)
		}

		allReviews = append(allReviews, reviews...)

		if response.NextPage == 0 {
			break
		}
		opts.Page = response.NextPage
	}

	return allReviews, nil
}

// ListPR returns the list of pull requests in the specified state.
func (gh *GitHub) ListPR(state string) ([]*github.PullRequest, error) {
	var prs []*github.PullRequest

	opts := &github.PullRequestListOptions{
		State:     state,
		Sort:      "updated",
		Direction: "desc",
		ListOptions: github.ListOptions{
			PerPage: PageSize,
		},
	}

	for {
		prsPage, response, err := gh.Client.PullRequests.List(gh.Ctx, gh.Owner, gh.Repo, opts)
		if err != nil {
			return nil, fmt.Errorf("unable to list pull requests with state %s: %w", state, err)
		}

		prs = append(prs, prsPage...)

		if response.NextPage == 0 {
			break
		}
		opts.Page = response.NextPage
	}

	return prs, nil
}

// New initializes the API client, the logger, and creates an instance of GitHub.
func New(ctx context.Context, cfg *Config) (*GitHub, error) {
	gh := &GitHub{
		Ctx:    ctx,
		Owner:  cfg.Owner,
		Repo:   cfg.Repo,
		DryRun: cfg.DryRun,
	}

	// Detect if the current process was launched by a GitHub Action and return
	// a logger suitable for terminal output or the GitHub Actions web interface.
	gh.Logger = logger.NewLogger(cfg.Verbose)

	// Retrieve GitHub API token from env.
	token, set := os.LookupEnv("GITHUB_TOKEN")
	if !set {
		return nil, errors.New("GITHUB_TOKEN is not set in env")
	}

	// Init GitHub API client using token.
	gh.Client = github.NewClient(nil).WithAuthToken(token)

	return gh, nil
}
