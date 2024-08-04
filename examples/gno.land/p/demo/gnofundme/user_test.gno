package gnofundme

import (
	"testing"

	"gno.land/p/demo/testutils"
)

func TestNewUser(t *testing.T) {
	name := "John Doe"
	email := "johndoe@example.com"
	address := testutils.TestAddress("John")

	user := NewUser(name, email, address)

	if user.GetName() != name {
		t.Errorf("Expected name to be %s, got %s", name, user.GetName())
	}
	if user.GetEmail() != email {
		t.Errorf("Expected email to be %s, got %s", email, user.GetEmail())
	}
	if user.GetAddress() != address {
		t.Errorf("Expected address to be %s, got %s", address, user.GetAddress())
	}
}
