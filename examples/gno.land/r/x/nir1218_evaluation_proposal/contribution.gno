package evaluation

import "std"

var contributionStatus = map[string]string{}

type Contribution struct {
	id          int
	contributor std.Address
	status      string // approved, proposed, negotiation, discussion, evaluation, etc.
	votes       []Vote
	tallyResult TallyResult
}

func init() {
	contributionStatus = make(map[string]string)
	contributionStatus["Proposed"] = "Proposed"
	contributionStatus["Approved"] = "Approved"
	contributionStatus["Evaluated"] = "Evaluated"
	contributionStatus["Negotiated"] = "Negotiated"
}

func NewContribution(id int, contributor std.Address) *Contribution {
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
	for _, v := range c.votes {
		if c.tallyResult.results.Has(v.option) {
			value, ok := c.tallyResult.results.Get(v.option)
			count := value.(int)
			c.tallyResult.results.Set(v.option, count)
		} else {
			c.tallyResult.results.Set(v.option, 0)
		}
	}
}
