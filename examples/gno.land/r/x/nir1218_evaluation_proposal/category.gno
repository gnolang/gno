package evaluation

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

type Category struct {
	name        string
	criteria    []string
	status      string
	votes       []Vote // TODO - use avl tree (had an issue using a tree of addresses)
	tallyResult TallyResult
}

func NewCategory(name string, criteria []string) *Category {
	c := &Category{
		name:        name,
		criteria:    criteria,
		status:      "Proposed",
		votes:       []Vote{},
		tallyResult: TallyResult{},
	}
	return c
}

func (c *Category) Approve() {
	// TODO error handling
	c.status = "Approved"
}

func (c Category) Status() string {
	return c.status
}

func (c *Category) Tally() {
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
