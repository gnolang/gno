package condition

import (
	"github.com/google/go-github/v66/github"
)

// Always Condition
type always struct{}

var _ Condition = &always{}

// Validate implements Condition
func (*always) Validate(_ *github.PullRequest) bool {
	return true
}

// GetText implements Condition
func (*always) GetText() string {
	return "On every pull request"
}

func Always() Condition {
	return &always{}
}

// Never Condition
type never struct{}

var _ Condition = &never{}

// Validate implements Condition
func (*never) Validate(_ *github.PullRequest) bool {
	return false
}

// GetText implements Condition
func (*never) GetText() string {
	return "On no pull request"
}

func Never() Condition {
	return &never{}
}
