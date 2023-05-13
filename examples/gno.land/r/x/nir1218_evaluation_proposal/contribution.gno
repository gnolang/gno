package evaluation

import "std"

var contributionStatus = map[string]string{}

type Contribution struct {
	id          int
	contributor std.Address
	status      string // approved, proposed, negotiation, discussion, evaluation, etc.
	approvers   []std.Address
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
		approvers:   []std.Address{},
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

func (c *Contribution) AddApproval(approver std.Address) bool {
	// TODO error handling
	c.approvers = append(c.approvers, approver)
	return true
}
