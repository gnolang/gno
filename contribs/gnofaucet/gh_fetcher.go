package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v74/github"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type GHFetcher struct {
	ghClient    GithubClient
	redisClient *redis.Client

	repos         map[string][]string
	logger        *zap.Logger
	queryInterval time.Duration // block query interval
}

func NewGHFetcher(
	ghClient GithubClient,
	rClient *redis.Client,
	repos map[string][]string,
	logger *zap.Logger,
	interval time.Duration) *GHFetcher {
	return &GHFetcher{
		ghClient:      ghClient,
		redisClient:   rClient,
		repos:         repos,
		queryInterval: interval,
		logger:        logger,
	}
}

func (f *GHFetcher) Fetch(ctx context.Context) error {
	ticker := time.NewTicker(f.queryInterval)
	defer ticker.Stop()

	ctx = rateLimited(ctx)
	if !f.fetchHistory(ctx) {
		f.logger.Warn("problems found fetching repositories history. Scores might not be 100% accurate")
	}

	f.logger.Info("finished getting history, going for events")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			f.fetchEvents(ctx)
		}
	}
}

func (f *GHFetcher) fetchEvents(ctx context.Context) {
	for org, names := range f.repos {
		for _, n := range names {
			f.logger.Info("fetching events for repo", zap.String("org", org), zap.String("repo", n))
			pipe := f.redisClient.TxPipeline()
			if !f.iterateEvents(ctx, pipe, org, n) {
				f.logger.Error("error fetching events. Aborting the pipeline", zap.String("org", org), zap.String("repo", n))
				continue
			}
			_, err := pipe.Exec(ctx)
			if err != nil {
				f.logger.Error("error executing redis pipeline", zap.String("org", org), zap.String("repo", n), zap.Error(err))
			} else {
				f.logger.Info("events correctly iterated", zap.String("org", org), zap.String("repo", n))
			}
		}
	}
}

func (f *GHFetcher) fetchHistory(ctx context.Context) bool {
	for org, names := range f.repos {
		for _, n := range names {
			f.logger.Info("Fetching history for repo", zap.String("org", org), zap.String("repo", n))
			pipe := f.redisClient.TxPipeline()

			if !f.iterateIssues(ctx, pipe, org, n) {
				return false
			}
			if !f.iteratePullRequests(ctx, pipe, org, n) {
				return false
			}

			_, err := pipe.Exec(ctx)
			if err != nil {
				f.logger.Error("error executing redis pipeline", zap.String("org", org), zap.String("repo", n), zap.Error(err))
			} else {
				f.logger.Info("history saved", zap.String("org", org), zap.String("repo", n))
			}
		}
	}

	return true
}

func (f *GHFetcher) processIssue(ctx context.Context, pipe redis.Pipeliner, org, repo string, issue *github.Issue) {
	if issue.User == nil {
		return
	}
	u := issue.User.Login
	if u == nil {
		return
	}

	f.logger.Info("processing issue", zap.String("user", *u), zap.String("org", org), zap.String("repo", repo), zap.Any("issue", issue.Number))

	if err := pipe.Incr(ctx, issueCountKey(*u)).Err(); err != nil {
		f.logger.Error("error adding issue to the count", zap.String("user", *u), zap.String("org", org), zap.String("repo", repo), zap.Error(err))
	}
}

func (f *GHFetcher) processPullRequest(ctx context.Context, pipe redis.Pipeliner, org, repo string, pr PullRequest) {
	u := pr.Author()
	if u == "" {
		return
	}

	f.logger.Info("processing pr", zap.String("user", u), zap.String("org", org), zap.String("repo", repo), zap.Any("pr", pr.Number()))

	if err := pipe.Incr(ctx, prCountKey(u)).Err(); err != nil {
		f.logger.Error("error adding prs to the count", zap.String("user", u), zap.String("org", org), zap.String("repo", repo), zap.Error(err))
		return
	}
	// if PR is not merged, do not count the commits
	if pr.State() != string(PullRequestStateMerged) {
		return
	}
	if pr.CommitsCount() != 0 {
		if err := pipe.IncrBy(ctx, commitCountKey(u), int64(pr.CommitsCount())).Err(); err != nil {
			f.logger.Error("error adding commits to the count", zap.String("user", u), zap.String("org", org), zap.String("repo", repo), zap.Error(err))
		}
	}

	for _, r := range pr.Reviews() {
		ru := r.Author()
		if ru == "" {
			continue
		}
		if r.State() != "APPROVED" && r.State() != "CHANGES_REQUESTED" {
			continue
		}

		if err := pipe.Incr(ctx, prReviewCountKey(ru)).Err(); err != nil {
			f.logger.Error("error adding review to the count", zap.String("user", ru), zap.String("org", org), zap.String("repo", repo), zap.Error(err))
		}
	}
}

func (f *GHFetcher) processEventReview(ctx context.Context, pipe redis.Pipeliner, org, repo string, review *github.PullRequestReview) {
	if review == nil {
		return
	}
	if review.State == nil {
		return
	}
	state := *review.State
	f.logger.Info("review", zap.Any("state", state))
	if state != "approved" && state != "changes_requested" {
		return
	}
	if review.User == nil {
		return
	}
	u := review.User.Login
	if u == nil {
		return
	}

	f.logger.Info("processing review", zap.String("user", *u), zap.String("org", org), zap.String("repo", repo))

	if err := pipe.Incr(ctx, prReviewCountKey(*u)).Err(); err != nil {
		f.logger.Error("error adding review to the count", zap.String("user", *u), zap.String("org", org), zap.String("repo", repo), zap.Error(err))
	}
}

func (f *GHFetcher) iterateIssues(ctx context.Context, pipe redis.Pipeliner, org, repo string) bool {
	page := 0
	for {
		issues, res, err := f.ghClient.ListIssues(ctx, org, repo, page)
		if err != nil {
			f.logger.Error("error getting issues", zap.String("org", org), zap.String("repo", repo), zap.Error(err))
			return false
		}

		var lastCreatedAt time.Time
		ca := f.getLatestDate(ctx, org, repo)
		for _, issue := range issues {
			if issue.CreatedAt.Before(ca) {
				return true // return early if we already processed previous issues
			}

			f.processIssue(ctx, pipe, org, repo, issue)
			lastCreatedAt = issue.CreatedAt.Time
		}

		if lastCreatedAt.After(ca) {
			if err := pipe.Set(ctx, lastRepoFetchKey(org, repo), lastCreatedAt, 0).Err(); err != nil {
				f.logger.Error("error setting last event date", zap.String("org", org), zap.String("repo", repo), zap.Error(err))
				return false
			}
		}

		if page == res.LastPage || res.NextPage == 0 {
			break
		}
		page = res.NextPage
	}

	return true
}

func (f *GHFetcher) iteratePullRequests(ctx context.Context, pipe redis.Pipeliner, org, repo string) bool {
	cursor := ""
	for {
		prs, nextCursor, err := f.ghClient.ListPullRequests(ctx, org, repo, cursor)
		if err != nil {
			f.logger.Error("error getting pull requests", zap.String("org", org), zap.String("repo", repo), zap.Error(err))
			return false
		}

		var lastCreatedAt time.Time
		for _, pr := range prs {
			ca := f.getLatestDate(ctx, org, repo)

			if pr.CreatedAt().Before(ca) {
				return true // we return early if we already processed the events
			}

			f.processPullRequest(ctx, pipe, org, repo, pr)

			lastCreatedAt = pr.CreatedAt()
		}

		t := f.getLatestDate(ctx, org, repo)

		if lastCreatedAt.After(t) {
			if err := pipe.Set(ctx, lastRepoFetchKey(org, repo), lastCreatedAt, 0).Err(); err != nil {
				f.logger.Error("error setting last event date", zap.String("org", org), zap.String("repo", repo), zap.Error(err))
				return false
			}
		}

		if nextCursor == "" {
			break
		}
		cursor = nextCursor
	}

	return true
}

func (f *GHFetcher) iterateEvents(ctx context.Context, pipe redis.Pipeliner, org, repo string) bool {
	page := 0
	for {
		events, res, err := f.ghClient.ListRepositoryEvents(ctx, org, repo, page)
		if err != nil {
			f.logger.Error("error getting repository events", zap.String("org", org), zap.String("repo", repo), zap.Error(err))
			return false
		}

		latestEventDate := f.getLatestDate(ctx, org, repo)
		for _, ev := range events {
			if !ev.CreatedAt.GetTime().After(latestEventDate) {
				continue
			}

			f.logger.Info("processing new event", zap.String("event", *ev.Type), zap.String("user", *ev.Actor.Login), zap.String("org", org), zap.String("repo", repo))

			et, err := ev.ParsePayload()
			if err != nil {
				f.logger.Error("error parsing event payload", zap.String("org", org), zap.String("repo", repo), zap.Error(err))
				continue
			}

			switch pe := et.(type) {
			case *github.IssuesEvent:
				if *pe.Action != "opened" {
					continue
				}
				if pe.Issue != nil {
					f.processIssue(ctx, pipe, org, repo, pe.Issue)
				}
			case *github.PullRequestEvent:
				if *pe.Action != "closed" {
					continue
				}

				pr := pe.PullRequest
				if pr == nil {
					continue
				}

				if *pr.Merged == false {
					continue
				}

				f.processPullRequest(ctx, pipe, org, repo, &PullRequestGapi{pe.PullRequest})
			case *github.PullRequestReviewEvent:
				review := pe.Review
				if review == nil {
					continue
				}

				state := *review.State
				if state != "approved" && state != "changes_requested" {
					continue
				}

				f.processEventReview(ctx, pipe, org, repo, pe.Review)
			}

			if err := pipe.Set(ctx, lastRepoFetchKey(org, repo), ev.CreatedAt.GetTime(), 0).Err(); err != nil {
				f.logger.Error("error setting lastEventDate", zap.String("org", org), zap.String("repo", repo), zap.Error(err))
				return false
			}
		}
		if page == res.LastPage || res.NextPage == 0 {
			break
		}
		page = res.NextPage
	}

	return true
}

func (f *GHFetcher) getLatestDate(ctx context.Context, org, repo string) time.Time {
	d, err := f.redisClient.Get(ctx, lastRepoFetchKey(org, repo)).Time()
	if errors.Is(err, redis.Nil) {
		return time.Time{}
	}
	if err != nil {
		f.logger.Error("error getting lastFetch", zap.String("org", org), zap.String("repo", repo), zap.Error(err))
		return time.Time{}
	}

	return d
}

func rateLimited(ctx context.Context) context.Context {
	return context.WithValue(ctx, github.SleepUntilPrimaryRateLimitResetWhenRateLimited, true)
}

func lastRepoFetchKey(org, repo string) string {
	return fmt.Sprintf("lastFetch:%s:%s", org, repo)
}

func issueCountKey(user string) string {
	return fmt.Sprintf("issue:%s", user)
}

func prCountKey(user string) string {
	return fmt.Sprintf("pr:%s", user)
}

func prReviewCountKey(user string) string {
	return fmt.Sprintf("prr:%s", user)
}

func commitCountKey(user string) string {
	return fmt.Sprintf("commit:%s", user)
}
