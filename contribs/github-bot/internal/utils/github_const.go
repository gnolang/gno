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

	// PR Review state.
	// https://docs.github.com/en/graphql/reference/enums#pullrequestreviewstate
	ReviewStateApproved         = "APPROVED"
	ReviewStateChangesRequested = "CHANGES_REQUESTED"
	ReviewStateCommented        = "COMMENTED"
)
