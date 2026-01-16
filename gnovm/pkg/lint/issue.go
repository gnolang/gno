package lint

import (
	"fmt"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
)

type Issue struct {
	RuleID     string
	Severity   Severity
	Message    string
	Filename   string
	Line       int
	Column     int
	Pos        gnolang.Pos
	Suggestion string
}

func (i Issue) Location() string {
	return fmt.Sprintf("%s:%d:%d", i.Filename, i.Line, i.Column)
}

func (i Issue) String() string {
	return fmt.Sprintf("%s: %s: %s (%s)",
		i.Location(), i.Severity, i.Message, i.RuleID)
}

func NewIssue(ruleID string, severity Severity, msg string, filename string, pos gnolang.Pos) Issue {
	return Issue{
		RuleID:   ruleID,
		Severity: severity,
		Message:  msg,
		Filename: filename,
		Line:     pos.Line,
		Column:   pos.Column,
		Pos:      pos,
	}
}
