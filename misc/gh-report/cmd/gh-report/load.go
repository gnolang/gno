package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// raw shapes mirror the GraphQL response.
type rawResponse struct {
	Data struct {
		Repository struct {
			Issues       rawList `json:"issues"`
			PullRequests rawList `json:"pullRequests"`
		} `json:"repository"`
	} `json:"data"`
}

type rawList struct {
	Nodes []rawNode `json:"nodes"`
}

type rawNode struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	IsDraft   bool      `json:"isDraft"`
	Mergeable string    `json:"mergeable"`

	Author            *rawUser `json:"author"`
	AuthorAssociation string   `json:"authorAssociation"`

	Assignees      rawUserList    `json:"assignees"`
	ReviewRequests rawReviewReqs  `json:"reviewRequests"`
	Labels         rawLabelList   `json:"labels"`
	Reactions      rawCount       `json:"reactions"`
	Comments       rawCommentList `json:"comments"`
	Reviews        rawReviewList  `json:"reviews"`
	Commits        rawCommitList  `json:"commits"`
	TimelineItems  rawTimeline    `json:"timelineItems"`
}

type rawUser struct {
	Typename  string    `json:"__typename"`
	Login     string    `json:"login"`
	CreatedAt time.Time `json:"createdAt"`
}

type rawUserList struct {
	Nodes []rawUser `json:"nodes"`
}

type rawReviewReqs struct {
	Nodes []struct {
		RequestedReviewer rawUser `json:"requestedReviewer"`
	} `json:"nodes"`
}

type rawLabelList struct {
	Nodes []struct {
		Name string `json:"name"`
	} `json:"nodes"`
}

type rawCount struct {
	TotalCount int `json:"totalCount"`
}

type rawCommentList struct {
	TotalCount int `json:"totalCount"`
	Nodes      []struct {
		Author    *rawUser  `json:"author"`
		CreatedAt time.Time `json:"createdAt"`
		Body      string    `json:"body"`
	} `json:"nodes"`
}

type rawReviewList struct {
	Nodes []struct {
		Author      *rawUser  `json:"author"`
		State       string    `json:"state"`
		SubmittedAt time.Time `json:"submittedAt"`
	} `json:"nodes"`
}

type rawCommitList struct {
	Nodes []struct {
		Commit struct {
			StatusCheckRollup *struct {
				State string `json:"state"`
			} `json:"statusCheckRollup"`
		} `json:"commit"`
	} `json:"nodes"`
}

type rawTimeline struct {
	Nodes []struct {
		Typename string `json:"__typename"`
	} `json:"nodes"`
}

// LoadRepoJSON parses one GraphQL response payload for `repo` (e.g. "gnolang/gno")
// and returns a flat list of Entry, issues then PRs.
func LoadRepoJSON(repo string, data []byte) ([]Entry, error) {
	var raw rawResponse
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	var out []Entry
	for _, n := range raw.Data.Repository.Issues.Nodes {
		out = append(out, toEntry(repo, KindIssue, n))
	}
	for _, n := range raw.Data.Repository.PullRequests.Nodes {
		out = append(out, toEntry(repo, KindPR, n))
	}
	return out, nil
}

func toEntry(repo string, kind EntryKind, n rawNode) Entry {
	e := Entry{
		Repo:              repo,
		Number:            n.Number,
		Kind:              kind,
		Title:             n.Title,
		URL:               n.URL,
		CreatedAt:         n.CreatedAt,
		UpdatedAt:         n.UpdatedAt,
		AuthorAssociation: n.AuthorAssociation,
		Reactions:         n.Reactions.TotalCount,
		Comments:          n.Comments.TotalCount,
		IsDraft:           n.IsDraft,
		Mergeable:         n.Mergeable,
	}
	if n.Author != nil {
		e.Author = n.Author.Login
		e.AuthorIsBot = n.Author.Typename == "Bot" || strings.HasSuffix(n.Author.Login, "[bot]")
		if !n.Author.CreatedAt.IsZero() {
			e.AuthorAccountAge = time.Since(n.Author.CreatedAt)
		}
	}
	for _, a := range n.Assignees.Nodes {
		e.Assignees = append(e.Assignees, a.Login)
	}
	for _, rr := range n.ReviewRequests.Nodes {
		if rr.RequestedReviewer.Login != "" {
			e.RequestedReviewer = append(e.RequestedReviewer, rr.RequestedReviewer.Login)
		}
	}
	for _, l := range n.Labels.Nodes {
		e.Labels = append(e.Labels, l.Name)
	}
	for _, c := range n.Comments.Nodes {
		cm := Comment{CreatedAt: c.CreatedAt, Body: c.Body}
		if c.Author != nil {
			cm.Author = c.Author.Login
			cm.IsBot = c.Author.Typename == "Bot" || strings.HasSuffix(c.Author.Login, "[bot]")
		}
		e.RecentComments = append(e.RecentComments, cm)
	}
	for _, r := range n.Reviews.Nodes {
		rv := Review{State: r.State, SubmittedAt: r.SubmittedAt}
		if r.Author != nil {
			rv.Author = r.Author.Login
		}
		e.Reviews = append(e.Reviews, rv)
	}
	if len(n.Commits.Nodes) > 0 && n.Commits.Nodes[0].Commit.StatusCheckRollup != nil {
		e.StatusCheckRoll = n.Commits.Nodes[0].Commit.StatusCheckRollup.State
	}
	for _, ti := range n.TimelineItems.Nodes {
		if ti.Typename == "ReviewRequestedEvent" {
			e.ReviewRequested = true
			break
		}
	}
	return e
}
