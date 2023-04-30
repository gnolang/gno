package evaluation

import "std"

type Contribution struct {
	id          int
	contributor std.Address
	points      Points
	status      string           // approved, proposed, negotiation, discussion, evaluation, etc.
	criteria    map[string]int32 // {"simplicity": 10pt, "usefullnes": 20pt, "quality": 50pt}
}

func NewContribution(id int, contributor std.Address, points Points) *Contribution {
	c := &Contribution{
		id:          id,
		contributor: contributor,
		points:      points,
		status:      "Proposed",
	}
	return c
}

func (c Contribution) Id() int {
	return c.id
}

func (c Contribution) Status() string {
	return c.status
}

func (c Contribution) UpdateStatus(status string) {
	c.status = status
}

func (c Contribution) Evaluate() Points {
	total := int64(0)
	for _, v := range c.criteria {
		total += int64(v)
	}
	return NewPoints(total, c.criteria)
}
