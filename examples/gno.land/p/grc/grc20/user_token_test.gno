package grc20

import "testing"

func TestUserTokenImpl(t *testing.T) {
	dummyAdmin := NewAdminToken("Dummy", "DUMMY", 4)
	dummyUser := dummyAdmin.GRC20()
	_ = dummyUser
}
