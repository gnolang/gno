package evaluation

/*
	1. At what stage of the PR a contribution should be evaluated?
		Should the PR be approved first?
	2. Can a contribution be re-evaluated before approved (current assumption is once a contribution is approved its state is final)?
	3. Can an evaluation criteria change up until it is approved (current assumption is that the evaluation criteria is set when the contribution is added)?
*/

import (
	"std"
	"testing"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/testutils"
	"gno.land/p/demo/ufmt"
)

var e = NewEvalutaion()

func TestEvaluationAddContribution(t *testing.T) {
	var (
		id          = 792
		name        = "Evaluation DAO Kick Off"
		description = "The PR is to initiate a discussion regarding the evaluation DAO"
		status      = "Draft"
		category    = "feat"
		criteria    = map[string]int32{"simplicity": 1, "usefullnes": 1, "quality": 1}
		address     = testutils.TestAddress("contributor")
	)

	pr := NewPullRequest(id, name, description, status, category)
	contributionId, ok := e.AddContribution(pr, address, criteria)

	t.Run("", func(t *testing.T) {
		if contributionId != id {
			t.Errorf("Got Contribution Id %d expected %d", contributionId, id)
		}
	})

	t.Run("Contribution added using the pull request id", func(t *testing.T) {
		c, found := e.contributions.Get(ufmt.Sprintf("%d", id))
		contribtution := c.(*Contribution)
		if contribtution.Id() != id {
			t.Errorf("Got Contribution Id %d expected %d", contribtution.Id(), id)
		}
	})

	t.Run("Pull Request added using the pull request id", func(t *testing.T) {
		pr, found := e.pullrequests.Get(ufmt.Sprintf("%d", id))
		pullrequest := pr.(*PullRequest)
		if pullrequest.Id() != id {
			t.Errorf("Got Pull Request Id %d expected %d", pullrequest.Id(), id)
		}
	})
}

// func (e *Evaluation) UpdateContribution(id int, status string) bool {
// 	c, found := e.contributions.Get(ufmt.Sprintf("%d", id))
// 	if found == true {
// 		contribtution := c.(*Contribution)
// 		contribtution.UpdateStatus(status)
// 		return true
// 	}
// 	return false
// }

// Current assumption is that the evaluation criteria is set when the contribution is added
// func (e Evaluation) EvaluateContribution(id int) (int64, bool) {
// 	c, found := e.contributions.Get(ufmt.Sprintf("%d", id))
// 	var total int64
// 	if found == true {
// 		contribtution := c.(*Contribution)
// 		total = int64(contribtution.Evaluate())
// 		return total, found
// 	}
// 	return int64(0), false
// }
