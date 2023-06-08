package access

import (
	"std"
	"testing"

	"gno.land/p/demo/testutils"
)

func TestSet_AddAndRemoveAdmin(t *testing.T) {
	set := New()

	// Add an admin
	admin1 := testutils.TestAddress("test1")
	set.Add(admin1)
	if !set.HasAccess(admin1) {
		t.Errorf("Expected admin1 to be added, but it was not")
	}

	// Remove the admin
	set.Del(admin1)
	if set.HasAccess(admin1) {
		t.Errorf("Expected admin1 to be removed, but it was still present")
	}
}

func TestSet_ListAdmins(t *testing.T) {
	// Add multiple admins
	admin1 := testutils.TestAddress("test2")
	admin2 := testutils.TestAddress("test3")
	set := NewWithAddress(admin1)
	std.TestSetOrigCaller(admin1)
	set.Add(admin2)

	// Get the list of admins
	admins := set.List()

	// Verify the correct number of admins
	expectedAdminsCount := 2
	if len(admins) != expectedAdminsCount {
		t.Errorf("Expected %d admins, but got %d", expectedAdminsCount, len(admins))
	}

	// Verify the admins in the list
	expectedAdmins := []std.Address{admin1, admin2}
	for _, expectedAdmin := range expectedAdmins {
		found := false
		for _, admin := range admins {
			if admin == expectedAdmin {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected admin %s to be in the list, but it was not found", expectedAdmin)
		}
	}
}

func TestSet_CurrentHasAccess(t *testing.T) {
	// Set the original caller as admin
	admin := testutils.TestAddress("test4")
	std.TestSetOrigCaller(admin)
	set := New()

	// Verify CurrentHasAccess returns true for the admin
	if !set.CurrentHasAccess() {
		t.Errorf("Expected CurrentHasAccess to return true for the admin, but it returned false")
	}

	// Verify CurrentHasAccess returns false for a non-admin
	nonAdmin := testutils.TestAddress("test5")
	std.TestSetOrigCaller(nonAdmin)
	if set.CurrentHasAccess() {
		t.Errorf("Expected CurrentHasAccess to return false for a non-admin, but it returned true")
	}
}

func TestSet_AddAndRemoveAdmin_WontPanic(t *testing.T) {
	// Add an admin
	admin := testutils.TestAddress("test6")
	std.TestSetOrigCaller(admin)
	set := New()

	// Set a non-admin caller
	nonAdmin := testutils.TestAddress("test7")
	std.TestSetOrigCaller(nonAdmin)

	// Verify Add won't panic in Privileged mode
	set.Add(nonAdmin)

	// Verify Del won't panic in Privileged mode
	set.Del(nonAdmin)
}
