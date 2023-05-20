package evaluation

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

type Commity struct {
	members    []std.Address // TODO - use avl tree (had an issue using a tree of addresses)
	categories avl.Tree      // A catagory is mapped to a list of evaluation criteria
	evaluation *Evaluation
}

const ApprovedStatus = "Approved"

func NewCommity() *Commity {
	c := &Commity{
		members:    []std.Address{},
		categories: avl.Tree{},
		evaluation: NewEvalutaion(),
	}
	return c
}

func (c *Commity) DesignateMembers(members []std.Address) []std.Address {
	c.members = append(c.members, members...)
	return c.members
}

func (c *Commity) DismissMembers(members []std.Address) []std.Address {
	// TODO
}

func (c *Commity) AddCategory(name string, criteria []string) bool {
	// TODO error handling
	if !c.isMember(std.GetOrigCaller()) {
		return false
	}
	category := NewCategory(name, criteria)
	c.categories.Set(name, category)
	return true
}

func (c *Commity) ApproveCategory(name string, option string) bool {
	if !c.isMember(std.GetOrigCaller()) {
		return false
	}

	if !c.categories.Has(name) {
		return false
	}

	value, _ := c.categories.Get(name)

	category := value.(*Category)

	if category.Status() == ApprovedStatus {
		return false
	}

	vote := NewVote(std.GetOrigCaller(), option)
	category.votes = append(category.votes, vote)
	category.Tally()

	// TODO Add threshold factor for a category approval
	// TODO Add quorum factor for a category approval
	// Current assumption is all members voted YES so category is approved

	result, ok := category.tallyResult.results.Get("YES")

	if result.(int) == len(c.members) {
		category.Approve()
		return true
	}

	return false
}

// TODO error handling
func (c *Commity) AddContribution(pr *PullRequest, contributor std.Address) (contributionId int, ok bool) {
	if !c.isMember(std.GetOrigCaller()) {
		return -1, false
	}
	// Check the category of the PR matches a catagory this commity evaluates
	// TODO check the category is an approved category
	if c.categories.Has(pr.category) {
		return c.evaluation.AddContribution(pr, contributor)
	}

	return -1, false
}

// TODO error handling
func (c *Commity) ApproveContribution(id int, option string) bool {
	if !c.isMember(std.GetOrigCaller()) {
		return false
	}

	if !c.evaluation.contributions.Has(ufmt.Sprintf("%d", id)) {
		return false
	}

	value, ok := c.evaluation.contributions.Get(ufmt.Sprintf("%d", id))
	contribution := value.(*Contribution)

	// Already approved
	if contribution.status == ApprovedStatus {
		return false
	}

	vote := NewVote(std.GetOrigCaller(), option)
	contribution.votes = append(contribution.votes, vote)
	contribution.Tally()

	// TODO Add threshold factor for a contribution approval
	// TODO Add quorum factor for a contribution approval
	// Current assumption is all members voted YES so contribution is approved

	result, ok := contribution.tallyResult.results.Get("YES")

	if result.(int) == len(c.members) {
		contribution.Approve()
		return true
	}

	return false
}

func (c *Commity) isMember(m std.Address) bool {
	for _, member := range c.members {
		if m == member {
			return true
		}
	}
	return false
}
