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

var (
	e = NewEvalutaion()

	id          = 792
	name        = "Evaluation DAO Kick Off"
	description = "The PR is to initiate a discussion regarding the evaluation DAO"
	status      = "Draft"
	category    = "feat"
	criteria    = map[string]int32{"simplicity": 1, "usefullnes": 1, "quality": 1}
	address     = testutils.TestAddress("contributor")
)

func TestEvaluationAddContribution(t *testing.T) {
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

// Current assumption is that the evaluation criteria is set when the contribution is added
func TestEvaluationEvaluateContribution(t *testing.T) {
	t.Run("", func(t *testing.T) {
		_, ok := e.EvaluateContribution(id)
		if !ok {
			t.Error("Expected evaluate to succeed but failed")
		} else {
			c, found := e.contributions.Get(ufmt.Sprintf("%d", id))
			contribtution := c.(*Contribution)
			if contribtution.Status() != "Evaluated" {
				t.Errorf("Expected contribution status %s got %s", "Evaluated", contribtution.Status())
			}
		}
	})

	t.Run("Contribution doesn't exist", func(t *testing.T) {
		invalidId := 1
		_, ok := e.EvaluateContribution(invalidId)
		if ok {
			t.Error("Expected evaluate to fail but succeeded")
		}
	})
}
