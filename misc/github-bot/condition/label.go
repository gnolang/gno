package condition

import (
	"fmt"
	"regexp"

	"github.com/google/go-github/v66/github"
)

// Label Condition
type label struct {
	pattern *regexp.Regexp
}

var _ Condition = &label{}

// Validate implements Condition
func (l *label) Validate(pr *github.PullRequest) bool {
	for _, label := range pr.Labels {
		if l.pattern.MatchString(label.GetName()) {
			return true
		}
	}
	return false
}

// GetText implements Condition
func (l *label) GetText() string {
	return fmt.Sprintf("A label match this pattern : %s", l.pattern.String())
}

func Label(pattern string) Condition {
	return &label{pattern: regexp.MustCompile(pattern)}
}
