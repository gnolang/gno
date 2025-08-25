package utils

// GitHub API const.
const (
	// GitHub Actions Event Names.
	EventIssueComment      = "issue_comment"
	EventPullRequest       = "pull_request"
	EventPullRequestReview = "pull_request_review"
	EventPullRequestTarget = "pull_request_target"
	EventWorkflowDispatch  = "workflow_dispatch"

	// Pull Request States.
	PRStateOpen   = "open"
	PRStateClosed = "closed"
)

// ReviewState is the state of a PR review. See:
// https://docs.github.com/en/graphql/reference/enums#pullrequestreviewstate
type ReviewState string

// Possible values of ReviewState.
const (
	ReviewStateApproved         ReviewState = "APPROVED"
	ReviewStateChangesRequested ReviewState = "CHANGES_REQUESTED"
	ReviewStateCommented        ReviewState = "COMMENTED"
	ReviewStateDismissed        ReviewState = "DISMISSED"
)

// Valid determines whether the ReviewState is one of the known ReviewStates.
func (r ReviewState) Valid() bool {
	switch r {
	case ReviewStateApproved, ReviewStateChangesRequested, ReviewStateCommented:
		return true
	}
	return false
}
