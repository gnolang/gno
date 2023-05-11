package evaluation

import "std"

var contributionStatus = map[string]string{}

type Contribution struct {
	id          int
	contributor std.Address
	status      string           // approved, proposed, negotiation, discussion, evaluation, etc.
	criteria    map[string]int32 // {"simplicity": 10pt, "usefullnes": 20pt, "quality": 50pt}
}

func init() {
	contributionStatus = make(map[string]string)
	contributionStatus["Proposed"] = "Proposed"
	contributionStatus["Approved"] = "Approved"
	contributionStatus["Evaluated"] = "Evaluated"
	contributionStatus["Negotiated"] = "Negotiated"
}

func NewContribution(id int, contributor std.Address, criteria map[string]int32) *Contribution {
	c := &Contribution{
		id:          id,
		contributor: contributor,
		criteria:    criteria,
		status:      contributionStatus["Proposed"],
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

func (c Contribution) Evaluate() Points {
	total := int64(0)
	for _, v := range c.criteria {
		total += int64(v)
	}
	return NewPoints(total, c.criteria)
}
