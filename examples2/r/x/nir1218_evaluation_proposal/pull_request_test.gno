package evaluation

import "testing"

func TestPullRequestUpdateStatus(t *testing.T) {
	var (
		id          = 792
		name        = "Evaluation DAO Kick Off"
		description = "The PR is to initiate a discussion regarding the evaluation DAO"
		status      = "Draft"
		category    = "feat"
	)

	validPR := NewPullRequest(id, name, description, status, category)

	t.Run("Valid Status Approved", func(t *testing.T) {
		status := "Approved"
		if !validPR.UpdateStatus(status) {
			t.Error("expected validation to pass")
		}
	})

	t.Run("Valid Status Draft", func(t *testing.T) {
		status := "Draft"
		if !validPR.UpdateStatus(status) {
			t.Error("expected validation to pass")
		}
	})

	t.Run("Valid Status Changes requested", func(t *testing.T) {
		status := "Changes requested"
		if !validPR.UpdateStatus(status) {
			t.Error("expected validation to pass")
		}
	})

	t.Run("Valid Status Review required", func(t *testing.T) {
		status := "Review required"
		if !validPR.UpdateStatus(status) {
			t.Error("expected validation to pass")
		}
	})

	t.Run("Invalid Status", func(t *testing.T) {
		status := "Junk"
		if validPR.UpdateStatus(status) {
			t.Error("expected validation to fail")
		}
	})
}
