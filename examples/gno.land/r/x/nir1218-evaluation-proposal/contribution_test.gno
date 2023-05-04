package evaluation

import "testing"

func TestEvaluate(t *testing.T) {
	var (
		total    = int64(3)
		criteria = map[string]int32{"simplicity": 1, "usefullnes": 1, "quality": 1}
		c        = NewContribution(1, "contributor", criteria)
	)

	points := c.Evaluate()
	if points.total != total {
		t.Errorf("Total points got %d expected %d", points.total, total)
	}
}
