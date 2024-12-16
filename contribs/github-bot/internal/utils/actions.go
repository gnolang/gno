package utils

import (
	"fmt"

	"github.com/sethvargo/go-githubactions"
)

// Recursively search for nested values using the keys provided.
func IndexMap(m map[string]any, keys ...string) any {
	if len(keys) == 0 {
		return m
	}

	if val, ok := m[keys[0]]; ok {
		if keys = keys[1:]; len(keys) == 0 {
			return val
		}
		subMap, _ := val.(map[string]any)
		return IndexMap(subMap, keys...)
	}

	return nil
}

// Retrieve PR number from GitHub Actions context.
func GetPRNumFromActionsCtx(actionCtx *githubactions.GitHubContext) (int, error) {
	firstKey := ""

	switch actionCtx.EventName {
	case EventIssueComment:
		firstKey = "issue"
	case EventPullRequest, EventPullRequestReview, EventPullRequestTarget:
		firstKey = "pull_request"
	default:
		return 0, fmt.Errorf("unsupported event: %s", actionCtx.EventName)
	}

	num, ok := IndexMap(actionCtx.Event, firstKey, "number").(float64)
	if !ok || num <= 0 {
		return 0, fmt.Errorf("invalid value: %d", int(num))
	}

	return int(num), nil
}
