package github

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/go-github/v74/github"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

type GHFetcher struct {
	ghClient    GithubClient
	redisClient *redis.Client

	repos         map[string][]string
	logger        *slog.Logger
	queryInterval time.Duration // block query interval
}

func NewGHFetcher(
	ghClient GithubClient,
	rClient *redis.Client,
	repos map[string][]string,
	logger *slog.Logger,
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
	if err := f.fetchHistory(ctx); err != nil {
		return err
	}

	f.logger.Info("finished getting history, going for events")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := f.fetchEvents(ctx); err != nil {
				return err
			}
		}
	}
}

func (f *GHFetcher) fetchEvents(ctx context.Context) error {
	for org, names := range f.repos {
		for _, n := range names {
			f.logger.Info("fetching events for repo", "org", org, "repo", n)
			pipe := f.redisClient.TxPipeline()
			if !f.iterateEvents(ctx, pipe, org, n) {
				f.logger.Error("error fetching events. Aborting the pipeline", "org", org, "repo", n)
				continue
			}
			_, err := pipe.Exec(ctx)
			if err != nil {
				f.logger.Error("error executing redis pipeline", "org", org, "repo", n, "err", err)
				return err
			} else {
				f.logger.Info("events correctly iterated", "org", org, "repo", n)
			}
		}
	}

	return nil
}

func (f *GHFetcher) fetchHistory(ctx context.Context) error {
	for org, names := range f.repos {
		for _, n := range names {
			f.logger.Info("Fetching history for repo", "org", org, "repo", n)
			pipe := f.redisClient.TxPipeline()

			if !f.iterateIssues(ctx, pipe, org, n) {
				return fmt.Errorf("error iterating issues")
			}
			if !f.iteratePullRequests(ctx, pipe, org, n) {
				return fmt.Errorf("error iterating pull requests")
			}

			_, err := pipe.Exec(ctx)
			if err != nil {
				return fmt.Errorf("error executing redis pipeline: %w", err)
			} else {
				f.logger.Info("history saved", "org", org, "repo", n)
			}
		}
	}

	return nil
}

func (f *GHFetcher) processIssue(ctx context.Context, pipe redis.Pipeliner, org, repo string, issue *github.Issue) {
	if issue.User == nil {
		return
	}
	u := issue.User.Login
	if u == nil {
		return
	}

	f.logger.Info("processing issue", "user", *u, "org", org, "repo", repo, "issue", issue.Number)

	if err := pipe.Incr(ctx, issueCountKey(*u)).Err(); err != nil {
		f.logger.Error("error adding issue to the count", "user", *u, "org", org, "repo", repo, "err", err)
	}
}

func (f *GHFetcher) processPullRequest(ctx context.Context, pipe redis.Pipeliner, org, repo string, pr PullRequest) {
	u := pr.Author()
	if u == "" {
		return
	}

	f.logger.Info("processing pr", "user", u, "org", org, "repo", repo, "pr", pr.Number())

	if err := pipe.Incr(ctx, prCountKey(u)).Err(); err != nil {
		f.logger.Error("error adding prs to the count", "user", u, "org", org, "repo", repo, "err", err)
		return
	}
	// if PR is not merged, do not count the commits
	if pr.State() != string(PullRequestStateMerged) {
		return
	}
	if pr.CommitsCount() != 0 {
		if err := pipe.IncrBy(ctx, commitCountKey(u), int64(pr.CommitsCount())).Err(); err != nil {
			f.logger.Error("error adding commits to the count", "user", u, "org", org, "repo", repo, "err", err)
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
			f.logger.Error("error adding review to the count", "user", ru, "org", org, "repo", repo, "err", err)
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
	f.logger.Info("review", "state", state)
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

	f.logger.Info("processing review", "user", *u, "org", org, "repo", repo)

	if err := pipe.Incr(ctx, prReviewCountKey(*u)).Err(); err != nil {
		f.logger.Error("error adding review to the count", "user", *u, "org", org, "repo", repo, "err", err)
	}
}

func (f *GHFetcher) iterateIssues(ctx context.Context, pipe redis.Pipeliner, org, repo string) bool {
	page := 0
	for {
		issues, res, err := f.ghClient.ListIssues(ctx, org, repo, page)
		if err != nil {
			f.logger.Error("error getting issues", "org", org, "repo", repo, "err", err)
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
				f.logger.Error("error setting last event date", "org", org, "repo", repo, "err", err)
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
			f.logger.Error("error getting pull requests", "org", org, "repo", repo, "err", err)
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
				f.logger.Error("error setting last event date", "org", org, "repo", repo, "err", err)
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
			f.logger.Error("error getting repository events", "org", org, "repo", repo, "err", err)
			return false
		}

		latestEventDate := f.getLatestDate(ctx, org, repo)
		for _, ev := range events {
			if !ev.CreatedAt.GetTime().After(latestEventDate) {
				continue
			}

			var eType, login string

			if ev.Type != nil {
				eType = *ev.Type
			} else {
				eType = "UNKNOWN"
			}

			if ev.Actor != nil && ev.Actor.Login != nil {
				login = *ev.Actor.Login
			} else {
				login = "UNKNOWN"
			}

			f.logger.Info("processing new event", "event", eType, "user", login, "org", org, "repo", repo)

			et, err := ev.ParsePayload()
			if err != nil {
				f.logger.Error("error parsing event payload", "org", org, "repo", repo, "err", err)
				continue
			}

			switch pe := et.(type) {
			case *github.IssuesEvent:
				if pe.Action == nil {
					continue
				}

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

				if pr.Merged == nil {
					continue
				}

				if !*pr.Merged {
					continue
				}

				f.processPullRequest(ctx, pipe, org, repo, &PullRequestGapi{pe.PullRequest})
			case *github.PullRequestReviewEvent:
				review := pe.Review
				if review == nil {
					continue
				}

				if review.State == nil {
					continue
				}

				state := *review.State
				if state != "approved" && state != "changes_requested" {
					continue
				}

				f.processEventReview(ctx, pipe, org, repo, pe.Review)
			}

			if err := pipe.Set(ctx, lastRepoFetchKey(org, repo), ev.CreatedAt.GetTime(), 0).Err(); err != nil {
				f.logger.Error("error setting lastEventDate", "org", org, "repo", repo, "err", err)
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
		f.logger.Error("error getting lastFetch", "org", org, "repo", repo, "err", err)
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
