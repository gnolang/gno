package evaluation

/*
	1. At what stage of the PR a contribution should be evaluated?
		Should the PR be approved first?
	2. Can a contribution be re-evaluated before approved (current assumption is once a contribution is approved its state is final)?
	3. Can an evaluation criteria change up until it is approved (current assumption is that the evaluation criteria is set when the contribution is added)?
*/

import (
	"testing"

	"gno.land/p/nt/testutils"
	"gno.land/p/nt/ufmt"
)

var (
	e = NewEvalutaion()

	id          = 792
	name        = "Evaluation DAO Kick Off"
	description = "The PR is to initiate a discussion regarding the evaluation DAO"
	status      = "Draft"
	category    = "feat"
	criteria    = map[string]int32{"simplicity": 1, "usefullnes": 1, "quality": 1}
	address_XXX = testutils.TestAddress("contributor")
)

func TestEvaluationAddContribution(t *testing.T) {
	pr := NewPullRequest(id, name, description, status, category)
	contributionId, _ := e.AddContribution(pr, address_XXX)

	t.Run("", func(t *testing.T) {
		if contributionId != id {
			t.Errorf("Got Contribution Id %d expected %d", contributionId, id)
		}
	})

	t.Run("Contribution added using the pull request id", func(t *testing.T) {
		c, _ := e.contributions.Get(ufmt.Sprintf("%d", id))
		contribtution := c.(*Contribution)
		if contribtution.Id() != id {
			t.Errorf("Got Contribution Id %d expected %d", contribtution.Id(), id)
		}
	})

	t.Run("Pull Request added using the pull request id", func(t *testing.T) {
		pr, _ := e.pullrequests.Get(ufmt.Sprintf("%d", id))
		pullrequest := pr.(*PullRequest)
		if pullrequest.Id() != id {
			t.Errorf("Got Pull Request Id %d expected %d", pullrequest.Id(), id)
		}
	})
}

func TestEvaluationUpdateContribution(t *testing.T) {
	t.Run("", func(t *testing.T) {
		status := "Negotiated"
		ok := e.UpdateContribution(id, status)
		if !ok {
			t.Error("Expected evaluation to update contribution's status successfully but failed")
		}
	})

	t.Run("Contribution doesn't exist", func(t *testing.T) {
		id := 1
		status := "Negotiated"
		ok := e.UpdateContribution(id, status)
		if ok {
			t.Error("Expected evaluation to fail but pass")
		}
	})
}
