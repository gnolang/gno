package mapdelete

import (
	"testing"
)

func TestGetMap(t *testing.T) {
	if !(GetMap(3)) {
		t.Error("Expected true, got ", GetMap(3))
	}
}

func TestDeleteMap(t *testing.T) {
	DeleteMap(3)
}

func TestGetMapAfterDelete(t *testing.T) {
	if GetMap(3) {
		t.Error("Expected false, got ", GetMap(3))
	}
}
