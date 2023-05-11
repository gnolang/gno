package evaluation

import "testing"

func TestContributionEvaluate(t *testing.T) {
	var (
		total    = int64(3)
		criteria = map[string]int32{"simplicity": 1, "usefullnes": 1, "quality": 1}
		c        = NewContribution(1, "contributor", criteria)
	)
	points := c.Evaluate()

	t.Run("", func(t *testing.T) {
		if points.total != total {
			t.Errorf("Total points got %d expected %d", points.total, total)
		}
	})
}

func TestContributionUpdateStatus(t *testing.T) {
	c := NewContribution(1, "contributor", criteria)

	t.Run("Status Update Negotiated", func(t *testing.T) {
		ok := c.UpdateStatus("Negotiated")
		if !ok {
			t.Error("Expected Successful Status Update but failed")
		}
	})

	t.Run("Status Update Evaluated", func(t *testing.T) {
		ok := c.UpdateStatus("Evaluated")
		if !ok {
			t.Error("Expected Successful Status Update but failed")
		}
	})

	t.Run("Status Update Approved", func(t *testing.T) {
		ok := c.UpdateStatus("Approved")
		if !ok {
			t.Error("Expected Successful Status Update but failed")
		}
	})

	t.Run("Status Update Approved Invalid", func(t *testing.T) {
		ok := c.UpdateStatus("Approved")
		if ok {
			t.Error("Expected Failed Status Update but succeded")
		}
	})
}
