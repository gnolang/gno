package main

import "time"

// EntryKind is "issue" or "PR".
type EntryKind string

const (
	KindIssue EntryKind = "issue"
	KindPR    EntryKind = "PR"
)

// Entry is one issue or PR from a repo.
type Entry struct {
	Repo      string // "owner/name"
	Number    int
	Kind      EntryKind
	Title     string
	URL       string
	CreatedAt time.Time
	UpdatedAt time.Time

	Author            string
	AuthorIsBot       bool
	AuthorAccountAge  time.Duration // since user.createdAt; 0 if unknown
	AuthorAssociation string        // FIRST_TIMER, MEMBER, OWNER, NONE, ...

	Assignees         []string
	RequestedReviewer []string // PR only
	Labels            []string

	Reactions int
	Comments  int // totalCount

	// Last comments (up to 5), used for mentions and recent-activity detection.
	RecentComments []Comment

	// PR-only fields. Empty for issues.
	IsDraft         bool
	Reviews         []Review
	StatusCheckRollup string // SUCCESS, FAILURE, PENDING, EXPECTED, ERROR, or ""
	Mergeable       string // MERGEABLE, CONFLICTING, UNKNOWN

	// True if the PR has a timeline event "review requested" (for Stuck detection).
	ReviewRequested bool
}

// Comment is a minimal comment record.
type Comment struct {
	Author    string
	IsBot     bool
	CreatedAt time.Time
	Body      string
}

// Review is a minimal review record (PR only).
type Review struct {
	Author      string
	State       string // APPROVED, CHANGES_REQUESTED, COMMENTED, DISMISSED, PENDING
	SubmittedAt time.Time
}

// Section is one bucket of the report.
type Section struct {
	Name    string  `json:"name"`
	Entries []Entry `json:"entries"`
}

func (s Section) Count() int { return len(s.Entries) }

// Report is the full classified output.
type Report struct {
	GeneratedAt time.Time `json:"generated_at"`
	WindowDays  int       `json:"window_days"`
	Sections    []Section `json:"sections"`
}
