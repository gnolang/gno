package grc20

import "testing"

func TestTokenInterface(t *testing.T) {
	var _ IGRC20 = &token{}
}

func TestImpl(t *testing.T) {
	dummy := NewToken("Dummy", "DUMMY", 4)
	if dummy == nil {
		t.Errorf("should not be nil")
	}
}
