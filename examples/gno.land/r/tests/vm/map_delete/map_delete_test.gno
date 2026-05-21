package mapdelete

import "testing"

func TestGetMap(t *testing.T) {
	if !(GetMap(3)) {
		t.Error("Expected true, got ", GetMap(3))
	}
}

func TestDeleteMap(t *testing.T) {
	if !GetMap(3) {
		t.Fatal("precondition: key 3 should exist before delete")
	}
	DeleteMap(cross, 3)
	if GetMap(3) {
		t.Error("Expected false after delete, got true")
	}
}
