package gnofundme

import (
	"testing"
	"time"

	"gno.land/p/demo/testutils"
)

func TestNewContributor(t *testing.T) {
	user := NewUser("User", "", testutils.TestAddress("user"))
	amount := uint(500)
	display := true
	contributor := NewContributor(user, amount, display)

	if contributor.GetUser() != user {
		t.Errorf("Expected contributor user to be %s, got %s", user, contributor.GetUser())
	}

	if contributor.GetDate().IsZero() {
		t.Error("Expected contributor date to be non-zero")
	}

	if contributor.GetAmount() != amount {
		t.Errorf("Expected contributor amount to be %d, got %d", amount, contributor.GetAmount())
	}

	if contributor.GetDisplay() != display {
		t.Errorf("Expected contributor display to be %t, got %t", display, contributor.GetDisplay())
	}
}
