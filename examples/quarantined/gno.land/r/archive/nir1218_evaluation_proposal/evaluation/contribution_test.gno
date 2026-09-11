package evaluation

import "testing"

func TestContributionUpdateStatus(t *testing.T) {
	c := NewContribution(1, "contributor")

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
