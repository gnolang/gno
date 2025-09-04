package evaluation

import (
	"chain"
	"time"
)

var contributionStatus = map[string]string{}

type Contribution struct {
	id               int
	contributor      chain.Address
	status           string // approved, proposed, negotiation, discussion, evaluation, etc.
	votes            []Vote
	tallyResult      TallyResult
	submitTime       time.Time
	lastEvaluateTime time.Time
	approveTime      time.Time
}

func init() {
	contributionStatus = make(map[string]string)
	contributionStatus["Proposed"] = "Proposed"
	contributionStatus["Approved"] = "Approved"
	contributionStatus["Evaluated"] = "Evaluated"
	contributionStatus["Negotiated"] = "Negotiated"
}

func NewContribution(id int, contributor chain.Address) *Contribution {
	c := &Contribution{
		id:          id,
		contributor: contributor,
		status:      contributionStatus["Proposed"],
		votes:       []Vote{},
		tallyResult: TallyResult{},
	}
	return c
}

func (c Contribution) Id() int {
	return c.id
}

func (c Contribution) Status() string {
	return c.status
}

func (c *Contribution) UpdateStatus(status string) bool {
	if c.status == contributionStatus["Approved"] {
		return false
	}
	c.status = status
	return true
}

func (c *Contribution) Approve() {
	// TODO error handling
	c.status = "Approved"
}

func (c *Contribution) Tally() {
	// TODO error handling
	for _, v := range c.votes {
		if c.tallyResult.results.Has(v.option) {
			value, _ := c.tallyResult.results.Get(v.option)
			count := value.(int)
			c.tallyResult.results.Set(v.option, count+1)
		}
	}
}
