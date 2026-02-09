package commondao

import (
	"strings"
	"time"

	"gno.land/p/nt/commondao"
)

const (
	maxTextBody         = 15_000
	maxTextTitle        = 255
	minTextVotingPeriod = time.Hour * 24
)

// NewTextPropDefinition creates a new general text proposal definition.
func NewTextPropDefinition(title, body string, quorum float64, votingPeriod time.Duration) TextPropDefinition {
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

	if quorum < commondao.QuorumOneThird {
		panic("minimum quorum allowed is one third (0.33)")
	}

	return TextPropDefinition{
		title:        title,
		body:         body,
		votingPeriod: votingPeriod,
		quorum:       quorum,
	}
}

// TextPropDefinition defines a proposal type for general text proposals.
// These type of proposals are not executable so nothing happens when they pass.
type TextPropDefinition struct {
	title, body  string
	votingPeriod time.Duration
	quorum       float64
}

func (p TextPropDefinition) Title() string               { return p.title }
func (p TextPropDefinition) Body() string                { return p.body }
func (p TextPropDefinition) VotingPeriod() time.Duration { return p.votingPeriod }

func (p TextPropDefinition) Tally(ctx commondao.VotingContext) (bool, error) {
	if !commondao.IsQuorumReached(p.quorum, ctx.VotingRecord, ctx.Members) {
		return false, commondao.ErrNoQuorum
	}

	c, success := commondao.SelectChoiceByPlurality(ctx.VotingRecord)
	if success {
		return c == commondao.ChoiceYes, nil
	}
	return false, nil
}
