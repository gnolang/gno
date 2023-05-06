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
