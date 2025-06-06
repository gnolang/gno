package dao

import (
	"std"
	"testing"
)

// Test DAO creation
func Test_NewDAO(t *testing.T) {
	creator := std.Address("Chaegnal")
	dao := NewDAO(creator)

	if !dao.IsAdmin(creator) {
		t.Errorf("Creator was not correctly set as admin.")
	}
}
