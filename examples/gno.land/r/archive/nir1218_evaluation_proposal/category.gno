package evaluation

import (
	"gno.land/p/nt/avl/v0"
)

type Category struct {
	name        string
	criteria    []string
	status      string
	votes       avl.Tree
	tallyResult TallyResult
}

func NewCategory(name string, criteria []string) *Category {
	tallyResult := TallyResult{}
	tallyResult.results.Set(VoteYes, 0)
	tallyResult.results.Set(VoteNo, 0)

	c := &Category{
		name:        name,
		criteria:    criteria,
		status:      "Proposed",
		votes:       avl.Tree{},
		tallyResult: tallyResult,
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
	// TODO error handling
	c.votes.Iterate("", "", func(address_XXX string, vote any) bool {
		v := vote.(Vote)
		value, exists := c.tallyResult.results.Get(v.option)
		if !exists {
			return false
		}
		count := value.(int)
		c.tallyResult.results.Set(v.option, count+1)
		return true
	})
}
