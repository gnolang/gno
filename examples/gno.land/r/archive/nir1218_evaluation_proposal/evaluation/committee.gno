package evaluation

import (
	"chain/runtime"

	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/ufmt/v0"
)

type Committee struct {
	members    []address // TODO - use avl tree or address set?
	categories avl.Tree  // A category is mapped to a list of evaluation criteria
	evaluation *Evaluation
}

const ApprovedStatus = "Approved"

func NewCommittee() *Committee {
	c := &Committee{
		members:    []address{},
		categories: avl.Tree{},
		evaluation: NewEvalutaion(),
	}
	return c
}

func (c *Committee) DesignateMembers(members []address) []address {
	c.members = append(c.members, members...)
	return c.members
}

func (c *Committee) DismissMembers(members []address) []address {
	// TODO
	return []address{}
}

func (c *Committee) AddCategory(name string, criteria []string) bool {
	// TODO error handling
	if !c.isMember(runtime.OriginCaller()) {
		return false
	}
	category := NewCategory(name, criteria)
	c.categories.Set(name, category)
	return true
}

func (c *Committee) ApproveCategory(name string, option string) bool {
	if !c.isMember(runtime.OriginCaller()) {
		return false
	}

	value, exists := c.categories.Get(name)
	if !exists {
		return false
	}
	category := value.(*Category)
	if category.Status() == ApprovedStatus {
		return false
	}

	vote := NewVote(runtime.OriginCaller(), option)
	category.votes.Set(runtime.OriginCaller().String(), vote)
	category.Tally()

	// TODO Add threshold factor for a category approval
	// TODO Add quorum factor for a category approval
	// Current assumption is all members voted YES so category is approved

	result, exists := category.tallyResult.results.Get(VoteYes)
	if !exists {
		return false
	}

	if result.(int) == len(c.members) {
		category.Approve()
		return true
	}

	return false
}

// TODO error handling
func (c *Committee) AddContribution(pr *PullRequest, contributor address) (contributionId int, ok bool) {
	if !c.isMember(runtime.OriginCaller()) {
		return -1, false
	}
	// Check the category of the PR matches a category this committee evaluates
	// TODO check the category is an approved category
	if c.categories.Has(pr.category) {
		return c.evaluation.AddContribution(pr, contributor)
	}

	return -1, false
}

// TODO error handling
func (c *Committee) ApproveContribution(id int, option string) bool {
	if !c.isMember(runtime.OriginCaller()) {
		return false
	}

	value, exists := c.evaluation.contributions.Get(ufmt.Sprintf("%d", id))
	if !exists {
		return false
	}
	contribution := value.(*Contribution)
	// Already approved
	if contribution.status == ApprovedStatus {
		return false
	}

	vote := NewVote(runtime.OriginCaller(), option)
	contribution.votes = append(contribution.votes, vote)
	contribution.Tally()

	// TODO Add threshold factor for a contribution approval
	// TODO Add quorum factor for a contribution approval
	// Current assumption is all members voted YES so contribution is approved

	result, exists := contribution.tallyResult.results.Get(VoteYes)
	if !exists {
		return false
	}

	if result.(int) == len(c.members) {
		contribution.Approve()
		return true
	}

	return false
}

func (c *Committee) isMember(m address) bool {
	for _, member := range c.members {
		if m == member {
			return true
		}
	}
	return false
}
