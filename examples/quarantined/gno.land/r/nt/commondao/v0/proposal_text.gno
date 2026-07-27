package commondao

import (
	"strings"
	"time"

	"gno.land/p/nt/commondao/v0"
)

const (
	maxTextBody         = 15_000
	maxTextTitle        = 255
	minTextVotingPeriod = time.Hour * 24
)

// NewTextPropDefinition creates a new general text proposal definition.
func NewTextPropDefinition(title, body string, votingPeriod time.Duration) TextPropDefinition {
	title = strings.TrimSpace(title)
	if title == "" {
		panic("proposal title is empty")
	}

	if len(title) > maxTextTitle {
		panic("proposal title is too long, max length is 255 chars")
	}

	body = strings.TrimSpace(body)
	if body == "" {
		panic("proposal body is empty")
	}

	if len(body) > maxTextBody {
		panic("proposal body is too long, max length is 15000 chars")
	}

	if votingPeriod < minTextVotingPeriod {
		panic("minimum proposal voting period is one day")
	}

	return TextPropDefinition{
		title:        title,
		body:         body,
		votingPeriod: votingPeriod,
	}
}

// TextPropDefinition defines a proposal type for general text proposals.
// These type of proposals are not executable so nothing happens when they pass.
type TextPropDefinition struct {
	title, body  string
	votingPeriod time.Duration
}

func (p TextPropDefinition) Title() string               { return p.title }
func (p TextPropDefinition) Body() string                { return p.body }
func (p TextPropDefinition) VotingPeriod() time.Duration { return p.votingPeriod }

// Threshold returns the tally threshold: text proposals are decided by the
// constitution's default supermajority rule.
func (TextPropDefinition) Threshold() commondao.Threshold {
	return commondao.ThresholdSupermajority
}
