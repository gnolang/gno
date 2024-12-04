package conditions

import (
	"fmt"
	"regexp"

	"github.com/gnolang/gno/contribs/github-bot/internal/utils"

	"github.com/google/go-github/v64/github"
	"github.com/xlab/treeprint"
)

// Label Condition.
type label struct {
	pattern *regexp.Regexp
}

var _ Condition = &label{}

func (l *label) IsMet(pr *github.PullRequest, details treeprint.Tree) bool {
	detail := fmt.Sprintf("A label matches this pattern: %s", l.pattern.String())

	for _, label := range pr.Labels {
		if l.pattern.MatchString(label.GetName()) {
			return utils.AddStatusNode(true, fmt.Sprintf("%s (label: %s)", detail, label.GetName()), details)
		}
	}

	return utils.AddStatusNode(false, detail, details)
}

func Label(pattern string) Condition {
	return &label{pattern: regexp.MustCompile(pattern)}
}
