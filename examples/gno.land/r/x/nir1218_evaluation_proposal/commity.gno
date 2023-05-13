package evaluation

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
)

/*
 Contributors DAO will designate its memebers to an Evaluation Commity, in the begining, the evaluation commity members could be the core development team members or any other trusted entity.
 Commity is a list of addresses of its memebers
 Commity sets categories for pull requests/contribtions
 A member can add a contribution for evaluation
 All members need to approve a contribution for a contribution be considered approved

 TODO
 Commity sets evalution criteria per pull request type and scope
 Commity members need to approve a contribution based on the evaluation criteria
 Commity negotiates contributions

 TBD
 Commity sets the minium number of approvals required as a percentage of all commity members
 Commity sets evalution criteria scoring range (1-10), scoring a contribution is essential when there are competing contributions (Game of Realm). Otherwise, the evaluaiton is a binary decision.
*/

type Commity struct {
	members    []std.Address // TODO - use avl tree (had an issue using a tree of addresses)
	categories avl.Tree
	evaluation *Evaluation
}

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

func (c *Commity) AddCategories(categories []string) []string {
	var (
		updatedCategories = make([]string, len(categories))
		index             = 0
	)

	for _, category := range categories {
		c.categories.Set(category, category)
	}

	c.categories.Iterate("", "", func(key string, value interface{}) bool {
		updatedCategories[index] = value.(string)
		index++
		return false
	})

	return updatedCategories
}

func (c *Commity) RemoveCategories(categories []string) []string {
	// TODO
}

// TODO error handling
func (c *Commity) AddContribution(pr *PullRequest, contributor std.Address) (contributionId int, ok bool) {
	if !c.isMember(std.GetOrigCaller()) {
		return -1, false
	}
	// Check the category of the PR matches a catagory this commity evaluates
	if c.categories.Has(pr.category) {
		return c.evaluation.AddContribution(pr, contributor)
	}

	return -1, false
}

// TODO error handling
func (c *Commity) ApproveContribution(id int, approved bool) bool {
	approvedStatus := "Approved"

	if !c.isMember(std.GetOrigCaller()) {
		return false
	}

	if !c.evaluation.contributions.Has(ufmt.Sprintf("%d", id)) {
		return false
	}

	value, ok := c.evaluation.contributions.Get(ufmt.Sprintf("%d", id))
	contribution := value.(*Contribution)
	// TODO Check if already in approved status
	if contribution.status == approvedStatus {
		return false
	}

	contribution.AddApproval(std.GetOrigCaller())
	//  If all members approved the contribution, the status of the contribution should change to approved.
	if contribution.approvers == c.members {
		contribution.UpdateStatus(approvedStatus)
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
