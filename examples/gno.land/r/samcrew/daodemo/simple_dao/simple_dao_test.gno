package simple_dao

import "testing"

// `init()` function is executed by default from `simple_dao.gno`, hence the `membersCount != 1`.
// Refers to `simple_dao.gno` to understand the outcome of this test.
func TestInit(t *testing.T) {
	membersCount := daoPrivate.Members.MembersCount()
	if membersCount != 1 {
		t.Fatalf("Expected 1 members, got %d", membersCount)
	}

	roles := daoPrivate.Members.GetRoles()
	expectedRoles := []string{"public-relationships", "finance-officer"}
	if len(roles) != len(expectedRoles) {
		t.Fatalf("Expected %d roles, got %d", len(expectedRoles), len(roles))
	}
	for _, role := range roles {
		err := true
		for _, expectedRole := range expectedRoles {
			if role == expectedRole {
				err = false
				break
			}
		}
		if err {
			t.Fatalf("Expected roles %v, got %v", expectedRoles, roles)
		}
	}
}
