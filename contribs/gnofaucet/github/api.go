package github

import (
	"context"
	"time"

	"github.com/Khan/genqlient/graphql"
	"github.com/google/go-github/v74/github"
)

type GithubClient interface {
	ListRepositoryEvents(ctx context.Context, owner, repo string, page int) ([]*github.Event, *github.Response, error)
	ListPullRequests(ctx context.Context, owner string, repo string, cursor string) ([]PullRequest, string, error)
	ListIssues(ctx context.Context, owner string, repo string, page int) ([]*github.Issue, *github.Response, error)
}

var _ GithubClient = &GithubClientImpl{}

type GithubClientImpl struct {
	cl  *github.Client
	gql graphql.Client
}

// ListIssues implements GithubClient.
func (lre *GithubClientImpl) ListIssues(ctx context.Context, owner string, repo string, page int) ([]*github.Issue, *github.Response, error) {
	return lre.cl.Issues.ListByRepo(ctx, owner, repo, &github.IssueListByRepoOptions{
		Sort:      "created",
		Direction: "asc",
		ListOptions: github.ListOptions{
			PerPage: 100,
			Page:    page,
		},
	})
}

// ListPullRequests implements GithubClient.
func (lre *GithubClientImpl) ListPullRequests(ctx context.Context, owner string, repo string, cursor string) ([]PullRequest, string, error) {
	data, err := getPullRequests(ctx, lre.gql, owner, repo, cursor)
	if err != nil {
		return nil, "", err
	}

	r := data.GetRepository()

	nodes := r.GetPullRequests().Nodes
	out := make([]PullRequest, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, &PullRequestGql{n})
	}

	nextCursor := r.GetPullRequests().PageInfo.EndCursor
	if !r.GetPullRequests().PageInfo.HasNextPage {
		nextCursor = ""
	}

	return out, nextCursor, nil
}

func (lre *GithubClientImpl) ListRepositoryEvents(ctx context.Context, owner, repo string, page int) ([]*github.Event, *github.Response, error) {
	return lre.cl.Activity.ListRepositoryEvents(ctx, owner, repo, &github.ListOptions{
		PerPage: 100,
		Page:    page,
	})
}

func NewGithubClientImpl(cl *github.Client, gr graphql.Client) *GithubClientImpl {
	return &GithubClientImpl{
		cl:  cl,
		gql: gr,
	}
}

var _ PullRequest = &PullRequestGapi{}

type PullRequestGapi struct {
	pr *github.PullRequest
}

// Number implements PullRequest.
func (p *PullRequestGapi) Number() int {
	return *p.pr.Number
}

// Author implements PullRequest.
func (p *PullRequestGapi) Author() string {
	return *p.pr.User.Login
}

// CommitsCount implements PullRequest.
func (p *PullRequestGapi) CommitsCount() int {
	return *p.pr.Commits
}

// CreatedAt implements PullRequest.
func (p *PullRequestGapi) CreatedAt() time.Time {
	return p.pr.CreatedAt.Time
}

// Reviews implements PullRequest.
func (p *PullRequestGapi) Reviews() []Review {
	// we don't have this information here
	return nil
}

// State implements PullRequest.
func (p *PullRequestGapi) State() string {
	return *p.pr.State
}

// Title implements PullRequest.
func (p *PullRequestGapi) Title() string {
	return *p.pr.Title
}

var _ PullRequest = &PullRequestGql{}

type PullRequestGql struct {
	pr getPullRequestsRepositoryPullRequestsPullRequestConnectionNodesPullRequest
}

// Number implements PullRequest.
func (p *PullRequestGql) Number() int {
	return p.pr.Number
}

// Author implements PullRequest.
func (p *PullRequestGql) Author() string {
	if p.pr.GetAuthor() == nil {
		return ""
	}
	return p.pr.GetAuthor().GetLogin()
}

// CommitsCount implements PullRequest.
func (p *PullRequestGql) CommitsCount() int {
	return p.pr.GetCommits().TotalCount
}

// CreatedAt implements PullRequest.
func (p *PullRequestGql) CreatedAt() time.Time {
	return p.pr.GetCreatedAt()
}

// Reviews implements PullRequest.
func (p *PullRequestGql) Reviews() []Review {
	nodes := p.pr.Reviews.Nodes
	out := make([]Review, 0, len(nodes))
	for _, r := range nodes {
		out = append(out, &ReviewGql{r})
	}

	return out
}

// Status implements PullRequest.
func (p *PullRequestGql) State() string {
	return string(p.pr.GetState())
}

// Title implements PullRequest.
func (p *PullRequestGql) Title() string {
	return p.pr.GetTitle()
}

var _ Review = &ReviewGql{}

type ReviewGql struct {
	r getPullRequestsRepositoryPullRequestsPullRequestConnectionNodesPullRequestReviewsPullRequestReviewConnectionNodesPullRequestReview
}

// Author implements Review.
func (r *ReviewGql) Author() string {
	if r.r.GetAuthor() == nil {
		return ""
	}

	return r.r.GetAuthor().GetLogin()
}

// State implements Review.
func (r *ReviewGql) State() string {
	return string(r.r.GetState())
}

type PullRequest interface {
	CreatedAt() time.Time
	Title() string
	Number() int
	CommitsCount() int
	State() string
	Author() string
	Reviews() []Review
}

type Review interface {
	State() string
	Author() string
}
