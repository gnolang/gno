package grc20

import "testing"

func TestAdminTokenImpl(t *testing.T) {
	dummy := NewAdminToken("Dummy", "DUMMY", 4)
	if dummy == nil {
		t.Errorf("should not be nil")
	}
}
