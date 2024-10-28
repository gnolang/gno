package condition

import (
	"fmt"
	"regexp"

	"github.com/google/go-github/v66/github"
)

// BaseBranch Condition
type baseBranch struct {
	pattern *regexp.Regexp
}

var _ Condition = &baseBranch{}

// Validate implements Condition
func (b *baseBranch) Validate(pr *github.PullRequest) bool {
	return b.pattern.MatchString(pr.GetBase().GetRef())
}

// GetText implements Condition
func (b *baseBranch) GetText() string {
	return fmt.Sprintf("The base branch match this pattern : %s", b.pattern.String())
}

func BaseBranch(pattern string) Condition {
	return &baseBranch{pattern: regexp.MustCompile(pattern)}
}

// HeadBranch Condition
type headBranch struct {
	pattern *regexp.Regexp
}

var _ Condition = &headBranch{}

// Validate implements Condition
func (h *headBranch) Validate(pr *github.PullRequest) bool {
	return h.pattern.MatchString(pr.GetHead().GetRef())
}

// GetText implements Condition
func (h *headBranch) GetText() string {
	return fmt.Sprintf("The head branch match this pattern : %s", h.pattern.String())
}

func HeadBranch(pattern string) Condition {
	return &headBranch{pattern: regexp.MustCompile(pattern)}
}
