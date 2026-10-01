package state

import "testing"

func TestNewRequiresClient(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when Client is nil")
		}
	}()
	_ = New(Deps{})
}
