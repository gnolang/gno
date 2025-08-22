package daodemo

import "testing"

func TestInit(t *testing.T) {
	membersCount := daoPrivate.Members.MembersCount()
	if membersCount != 4 {
		t.Fatalf("Expected 4 members, got %d", membersCount)
	}

	roles := daoPrivate.Members.GetRoles()
	expectedRoles := []string{"admin", "public-relationships", "finance-officer"}
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
